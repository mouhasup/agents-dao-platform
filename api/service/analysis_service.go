// service/analysis_service.go
package service

import (
	"agents-dao-platform/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

// DaoFolderKey est la clé (chemin) du sous-dossier contenant le DAO original
const DaoFolderKey = "01_Dossier_Consultation_Original"

var (
	ErrAnalysisRunning = errors.New("une analyse est déjà en cours pour ce projet")
	ErrNoDaoFolder     = errors.New("le dossier Drive du DAO est introuvable pour ce projet")
	ErrNoDaoDocuments  = errors.New("aucun document PDF trouvé dans le dossier du DAO")
)

// ConsolidatedAnalysis est le résultat JSON stocké dans Project.AnalyseResultJSON
type ConsolidatedAnalysis struct {
	Documents      []string           `json:"documents"`
	Administrative *AdminAnalysis     `json:"administrative,omitempty"`
	Technique      *TechnicalAnalysis `json:"technique,omitempty"`
	Erreurs        []string           `json:"erreurs,omitempty"`
}

type AnalysisService interface {
	// DaoFolderID retourne l'ID Drive du dossier du DAO original
	DaoFolderID(p *model.Project) string
	ListDaoDocuments(ctx context.Context, token *oauth2.Token, p *model.Project) ([]*drive.File, error)
	// StartAnalysis valide les préconditions puis lance l'analyse en arrière-plan
	StartAnalysis(ctx context.Context, token *oauth2.Token, projectID string) error
}

type analysisService struct {
	repo    ProjectRepository
	drive   DriveService
	agents  AgentService
	mu      sync.Mutex
	running map[string]bool
}

func NewAnalysisService(repo ProjectRepository, ds DriveService, as AgentService) AnalysisService {
	return &analysisService{repo: repo, drive: ds, agents: as, running: map[string]bool{}}
}

func (s *analysisService) DaoFolderID(p *model.Project) string {
	for _, f := range p.Folders {
		if f.Categorie == DaoFolderKey || f.Nom == DaoFolderKey {
			return f.GoogleFolderID
		}
	}
	return ""
}

func isAnalysable(mime string) bool {
	return mime == "application/pdf" ||
		strings.HasPrefix(mime, "application/vnd.google-apps.document")
}

func (s *analysisService) ListDaoDocuments(ctx context.Context, token *oauth2.Token, p *model.Project) ([]*drive.File, error) {
	folderID := s.DaoFolderID(p)
	if folderID == "" {
		return nil, ErrNoDaoFolder
	}
	files, err := s.drive.ListFolderFiles(ctx, token, folderID)
	if err != nil {
		return nil, err
	}
	out := make([]*drive.File, 0, len(files))
	for _, f := range files {
		if isAnalysable(f.MimeType) {
			out = append(out, f)
		}
	}
	return out, nil
}

func (s *analysisService) StartAnalysis(ctx context.Context, token *oauth2.Token, projectID string) error {
	project, err := s.repo.GetProject(projectID)
	if err != nil {
		return err
	}

	s.mu.Lock()
	if s.running[projectID] {
		s.mu.Unlock()
		return ErrAnalysisRunning
	}
	s.running[projectID] = true
	s.mu.Unlock()

	release := func() {
		s.mu.Lock()
		delete(s.running, projectID)
		s.mu.Unlock()
	}

	files, err := s.ListDaoDocuments(ctx, token, project)
	if err != nil {
		release()
		return err
	}
	if len(files) == 0 {
		release()
		return ErrNoDaoDocuments
	}

	project.Statut = model.StatusEnAnalyse
	project.AnalyseError = ""
	project.UpdatedAt = time.Now()
	if _, err := s.repo.UpdateProject(project); err != nil {
		release()
		return err
	}

	// Copie du token : la requête HTTP se termine avant la fin de l'analyse
	tok := *token
	go func() {
		defer release()
		bg, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
		defer cancel()
		s.run(bg, &tok, projectID, files)
	}()
	return nil
}

func (s *analysisService) fail(projectID string, err error) {
	log.Printf("❌ Analyse %s échouée : %v", projectID, err)
	p, gerr := s.repo.GetProject(projectID)
	if gerr != nil {
		return
	}
	p.Statut = model.StatusBrouillon
	p.AnalyseError = err.Error()
	p.UpdatedAt = time.Now()
	p.Tasks, p.Folders = nil, nil
	_, _ = s.repo.UpdateProject(p)
}

func (s *analysisService) run(ctx context.Context, token *oauth2.Token, projectID string, files []*drive.File) {
	// 1. Téléchargement des documents du DAO
	docs := make([]Document, 0, len(files))
	names := make([]string, 0, len(files))
	for _, f := range files {
		data, mime, err := s.drive.DownloadFile(ctx, token, f.Id)
		if err != nil {
			s.fail(projectID, fmt.Errorf("téléchargement de %s : %v", f.Name, err))
			return
		}
		docs = append(docs, Document{Name: f.Name, Bytes: data, MimeType: mime})
		names = append(names, f.Name)
	}

	// 2. Exécution des agents en parallèle
	result := ConsolidatedAnalysis{Documents: names}
	var wg sync.WaitGroup
	var mu sync.Mutex
	addErr := func(e string) { mu.Lock(); result.Erreurs = append(result.Erreurs, e); mu.Unlock() }

	instruction := fmt.Sprintf("Voici le dossier d'appel d'offres (%d document(s) : %s). Analyse-le selon ton rôle.",
		len(names), strings.Join(names, ", "))

	wg.Add(2)
	go func() {
		defer wg.Done()
		out, err := s.agents.ExecuteAgentTask(ctx, AdminAgent, docs, instruction)
		if err != nil {
			addErr("Conformité administrative : " + err.Error())
			return
		}
		var a AdminAnalysis
		if err := json.Unmarshal([]byte(out), &a); err != nil {
			addErr("Conformité administrative : réponse JSON invalide")
			return
		}
		mu.Lock()
		result.Administrative = &a
		mu.Unlock()
	}()
	go func() {
		defer wg.Done()
		// Léger décalage pour éviter d'envoyer deux gros fichiers PDF exactement à la même milliseconde
		time.Sleep(1500 * time.Millisecond)
		out, err := s.agents.ExecuteAgentTask(ctx, TechnicalAgent, docs, instruction)
		if err != nil {
			addErr("Ingénierie technique : " + err.Error())
			return
		}
		var t TechnicalAnalysis
		if err := json.Unmarshal([]byte(out), &t); err != nil {
			addErr("Ingénierie technique : réponse JSON invalide")
			return
		}
		mu.Lock()
		result.Technique = &t
		mu.Unlock()
	}()
	wg.Wait()

	if result.Administrative == nil && result.Technique == nil {
		s.fail(projectID, fmt.Errorf("tous les agents ont échoué : %s", strings.Join(result.Erreurs, " | ")))
		return
	}

	// 3. Persistance du résultat consolidé
	project, err := s.repo.GetProject(projectID)
	if err != nil {
		log.Printf("❌ Projet %s introuvable après analyse : %v", projectID, err)
		return
	}
	raw, _ := json.Marshal(result)
	now := time.Now()
	project.AnalyseResultJSON = string(raw)
	project.AnalysedAt = &now
	project.Statut = model.StatusEnMontage
	project.UpdatedAt = now
	if len(result.Erreurs) > 0 {
		project.AnalyseError = "Analyse partielle : " + strings.Join(result.Erreurs, " | ")
	} else {
		project.AnalyseError = ""
	}
	if a := result.Administrative; a != nil {
		if a.Acheteur != "" {
			project.Acheteur = a.Acheteur
		}
		if a.NumeroAppelOffres != "" {
			project.NumeroDAO = a.NumeroAppelOffres
		}
		if d := parseDate(a.DateLimiteDepot); d != nil {
			project.DateLimiteDepot = d
		}
	}

	// 4. Génération des tâches (on ne recrée pas celles qui existent déjà)
	existing := map[string]bool{}
	for _, t := range project.Tasks {
		existing[strings.ToLower(t.Titre)] = true
	}
	for _, t := range buildTasks(projectID, &result, project.DateLimiteDepot) {
		if existing[strings.ToLower(t.Titre)] {
			continue
		}
		existing[strings.ToLower(t.Titre)] = true
		task := t
		if err := s.repo.AddTask(&task); err != nil {
			log.Printf("⚠️ Création tâche impossible : %v", err)
		}
	}

	project.Tasks, project.Folders = nil, nil // évite la ré-écriture des associations par GORM
	if _, err := s.repo.UpdateProject(project); err != nil {
		log.Printf("❌ Sauvegarde analyse %s : %v", projectID, err)
		return
	}
	log.Printf("✅ Analyse du projet %s terminée (%d document(s))", projectID, len(names))
}

func parseDate(s string) *time.Time {
	s = strings.TrimSpace(s)
	for _, layout := range []string{"2006-01-02", "02/01/2006", "2006-01-02T15:04:05Z07:00"} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}

func buildTasks(projectID string, r *ConsolidatedAnalysis, deadline *time.Time) []model.ProjectTask {
	var tasks []model.ProjectTask
	seq := time.Now().UnixNano()
	add := func(titre, desc, cat, prio string, due *time.Time) {
		titre = strings.TrimSpace(titre)
		if titre == "" {
			return
		}
		if len(titre) > 250 {
			titre = titre[:250]
		}
		seq++
		tasks = append(tasks, model.ProjectTask{
			ID:          fmt.Sprintf("task-%d", seq),
			ProjectID:   projectID,
			Titre:       titre,
			Description: desc,
			Categorie:   cat,
			Priorite:    prio,
			Echeance:    due,
			Statut:      model.TaskAFaire,
			CreatedAt:   time.Now(),
		})
	}

	if a := r.Administrative; a != nil {
		for _, p := range a.PiecesAdministratives {
			add("Fournir : "+p, "Pièce administrative exigée par le DAO.", "ADMINISTRATIF", "HAUTE", deadline)
		}
		for _, lot := range a.Lots {
			if lot.Cautionnement != "" {
				add(fmt.Sprintf("Caution de soumission lot %d (%s)", lot.NumeroLot, lot.Cautionnement),
					lot.Intitule, "FINANCIER", "HAUTE", deadline)
			}
			for _, p := range lot.PiecesRequises {
				add(fmt.Sprintf("Lot %d : %s", lot.NumeroLot, p), lot.Intitule, "ADMINISTRATIF", "MOYENNE", deadline)
			}
		}
		for _, e := range a.Calendrier {
			l := strings.ToLower(e.Etape)
			if strings.Contains(l, "visite") || strings.Contains(l, "réunion") || strings.Contains(l, "reunion") {
				add(e.Etape, "Étape clé du calendrier du DAO.", "VISITE_SITE", "HAUTE", parseDate(e.Date))
			}
		}
	}
	if t := r.Technique; t != nil {
		if len(t.PlanMemoireTechnique) > 0 {
			add("Rédiger le mémoire technique", "Plan proposé :\n- "+strings.Join(t.PlanMemoireTechnique, "\n- "),
				"TECHNIQUE", "HAUTE", deadline)
		}
		for _, p := range t.PersonnelCle {
			add("Mobiliser / CV : "+p, "Personnel clé exigé.", "TECHNIQUE", "MOYENNE", deadline)
		}
		for _, m := range t.MaterielRequis {
			add("Justifier le matériel : "+m, "Matériel exigé par le DAO.", "TECHNIQUE", "MOYENNE", deadline)
		}
	}
	add("Établir l'offre financière (bordereau des prix / DQE)", "", "FINANCIER", "HAUTE", deadline)
	return tasks
}

// service/technical_service.go
package service

import (
	"agents-dao-platform/model"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type TechnicalOfferService interface {
	GetTechnicalOffer(ctx context.Context, projectID string) (*model.TechnicalOfferData, error)
	GenerateSection(ctx context.Context, projectID, sectionID, customInstruction string) (*model.MemoireSection, error)
	SaveTechnicalOffer(ctx context.Context, projectID string, data *model.TechnicalOfferData) error
	ExportTechnicalOfferToDrive(ctx context.Context, token *oauth2.Token, projectID string) (string, string, error)
	CopyPersonnelCVsToProject(ctx context.Context, token *oauth2.Token, projectID string) (int, string, error)
	GenerateDocx(ctx context.Context, projectID string) ([]byte, string, error)
	ProposePersonnel(ctx context.Context, projectID string) ([]model.PersonnelAffectation, error)
	ProposeMateriel(ctx context.Context, projectID string) ([]model.MaterielMobilisation, error)
}

type technicalOfferService struct {
	repo   ProjectRepository
	agents AgentService
	drive  DriveService
}

func NewTechnicalOfferService(repo ProjectRepository, as AgentService, ds DriveService) TechnicalOfferService {
	return &technicalOfferService{
		repo:   repo,
		agents: as,
		drive:  ds,
	}
}

func cleanKeyword(s string) string {
	s = strings.ToLower(s)
	replacer := strings.NewReplacer(
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"à", "a", "â", "a", "ä", "a",
		"î", "i", "ï", "i",
		"ô", "o", "ö", "o",
		"ù", "u", "û", "u", "ü", "u",
		"ç", "c",
		"'", " ", "-", " ", "_", " ", "/", " ", ",", " ", ";", " ", ":", " ", ".", " ",
	)
	return replacer.Replace(s)
}

func (s *technicalOfferService) GetTechnicalOffer(ctx context.Context, projectID string) (*model.TechnicalOfferData, error) {
	project, err := s.repo.GetProject(projectID)
	if err != nil {
		return nil, err
	}

	// Si déjà initialisé dans le projet
	if strings.TrimSpace(project.MemoireTechniqueJSON) != "" {
		var data model.TechnicalOfferData
		if err := json.Unmarshal([]byte(project.MemoireTechniqueJSON), &data); err == nil && len(data.Sections) > 0 {
			if data.Personnel == nil {
				data.Personnel = make([]model.PersonnelAffectation, 0)
			}
			if data.Materiel == nil {
				data.Materiel = make([]model.MaterielMobilisation, 0)
			}
			return &data, nil
		}
	}

	// Initialisation par défaut basée sur l'analyse technique IA
	data := &model.TechnicalOfferData{
		ProjectID: project.ID,
		Titre:     fmt.Sprintf("Mémoire Technique - %s", project.Titre),
		Sections:  make([]model.MemoireSection, 0),
		Personnel: make([]model.PersonnelAffectation, 0),
		Materiel:  make([]model.MaterielMobilisation, 0),
	}

	var ca ConsolidatedAnalysis
	hasAnalysis := false
	if project.AnalyseResultJSON != "" {
		if err := json.Unmarshal([]byte(project.AnalyseResultJSON), &ca); err == nil {
			hasAnalysis = true
		}
	}

	// 1. Sections basées sur PlanMemoireTechnique de l'IA ou structure standard BTP
	var planItems []string
	if hasAnalysis && ca.Technique != nil && len(ca.Technique.PlanMemoireTechnique) > 0 {
		planItems = ca.Technique.PlanMemoireTechnique
	} else {
		planItems = []string{
			"1. Compréhension du Projet, Contexte & Enjeux",
			"2. Méthodologie d'Exécution & Phasage des Travaux",
			"3. Calendrier d'Exécution & Chemin Critique",
			"4. Organisation du Chantier, Personnel Clé & Encadrement",
			"5. Moyens Matériels & Engins Affectés au Chantier",
			"6. Plan d'Assurance Qualité (PAQ) & Contrôles",
			"7. Plan d'Hygiène, Sécurité et Environnement (HSE)",
		}
	}

	for i, chapter := range planItems {
		data.Sections = append(data.Sections, model.MemoireSection{
			ID:      fmt.Sprintf("sec-%d", i+1),
			Titre:   chapter,
			Contenu: "",
			Statut:  "A_REDIGER",
		})
	}

	// 2. Personnel clé
	companyDocs, _ := s.repo.ListCompanyDocuments()
	if hasAnalysis && ca.Technique != nil {
		for _, p := range ca.Technique.PersonnelCle {
			aff := model.PersonnelAffectation{
				RoleRequis: p,
			}
			// Rechercher un CV correspondant dans le coffre-fort
			cleanRole := cleanKeyword(p)
			for _, doc := range companyDocs {
				if doc.TypeDoc == model.DocCV || strings.Contains(cleanKeyword(doc.Titre), "cv") {
					cleanDoc := cleanKeyword(doc.Titre)
					if strings.Contains(cleanDoc, cleanRole) || strings.Contains(cleanRole, cleanDoc) ||
						strings.Contains(cleanDoc, "ingenieur") || strings.Contains(cleanDoc, "conducteur") {
						aff.Nom = doc.Titre
						aff.DocumentID = doc.ID
						aff.DocumentURL = doc.WebViewLink
						aff.Qualifications = doc.Description
						break
					}
				}
			}
			data.Personnel = append(data.Personnel, aff)
		}

		// 3. Matériel requis
		for _, m := range ca.Technique.MaterielRequis {
			mat := model.MaterielMobilisation{
				TypeRequis: m,
				Statut:     "PROPRIETE",
			}
			cleanMat := cleanKeyword(m)
			for _, doc := range companyDocs {
				if doc.TypeDoc == model.DocMateriel || strings.Contains(cleanKeyword(doc.Titre), "materiel") {
					if strings.Contains(cleanKeyword(doc.Titre), cleanMat) {
						mat.Designation = doc.Titre
						mat.DocumentID = doc.ID
						mat.DocumentURL = doc.WebViewLink
						break
					}
				}
			}
			data.Materiel = append(data.Materiel, mat)
		}
	}

	// Sauvegarder l'état initial
	raw, _ := json.Marshal(data)
	project.MemoireTechniqueJSON = string(raw)
	_, _ = s.repo.UpdateProject(project)

	return data, nil
}

func (s *technicalOfferService) GenerateSection(ctx context.Context, projectID, sectionID, customInstruction string) (*model.MemoireSection, error) {
	data, err := s.GetTechnicalOffer(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var targetSection *model.MemoireSection
	for i := range data.Sections {
		if data.Sections[i].ID == sectionID {
			targetSection = &data.Sections[i]
			break
		}
	}
	if targetSection == nil {
		return nil, errors.New("section introuvable")
	}

	project, err := s.repo.GetProject(projectID)
	if err != nil {
		return nil, err
	}

	profile, _ := s.repo.GetCompanyProfile()
	companyName := "Notre Entreprise"
	companyNinea := ""
	if profile != nil {
		companyName = profile.Nom
		companyNinea = profile.NINEA
	}

	var ca ConsolidatedAnalysis
	if project.AnalyseResultJSON != "" {
		_ = json.Unmarshal([]byte(project.AnalyseResultJSON), &ca)
	}

	consistance := ""
	contraintes := ""
	normes := ""
	vigilance := ""
	if ca.Technique != nil {
		consistance = ca.Technique.ConsistanceTravaux
		contraintes = strings.Join(ca.Technique.ContraintesExecution, "\n- ")
		normes = strings.Join(ca.Technique.NormesExigees, ", ")
		vigilance = ca.Technique.PointsVigilance
	}

	prompt := fmt.Sprintf(`Tu es le Rédacteur Technique principal de l'entreprise %s (NINEA: %s).
Rédige le chapitre suivant du Mémoire Technique d'Exécution pour le projet :
TITRE DU PROJET : %s
ACHETEUR / MAÎTRE D'OUVRAGE : %s
N° DE DOSSIER : %s

SPÉCIFICATIONS TECHNIQUES DU MARCHÉ (CCTP) :
- Consistance des travaux : %s
- Normes techniques imposées : %s
- Contraintes d'exécution et délais :
%s
- Points de vigilance majeurs : %s

CHAPITRE À RÉDIGER :
« %s »

INSTRUCTIONS COMPLÉMENTAIRES DE L'ENTREPRISE :
%s

DIRECTIVES DE RÉDACTION :
1. Rédige un contenu détaillé, technique et opérationnel directement prêt à être intégré dans le mémoire final.
2. Structure en sous-sections claires (avec ### et ####), listes à puces et étapes méthodologiques pas-à-pas.
3. Évite le bavardage générique : sois précis sur les matériaux (pavés, sable de pose, graveleux latéritique, etc.), les tolérances géométriques, le phasage et les contrôles qualité.
4. Rédige en français soigné et affirmatif au nom de l'entreprise %s.`,
		companyName, companyNinea,
		project.Titre, project.Acheteur, project.NumeroDAO,
		consistance, normes, contraintes, vigilance,
		targetSection.Titre,
		customInstruction,
		companyName,
	)

	out, err := s.agents.ExecuteAgentTask(ctx, TechnicalWriterAgent, nil, prompt)
	if err != nil {
		return nil, fmt.Errorf("erreur génération section IA : %v", err)
	}

	now := time.Now()
	targetSection.Contenu = strings.TrimSpace(out)
	targetSection.Statut = "GENERE"
	targetSection.UpdatedAt = &now

	// Sauvegarder
	if err := s.SaveTechnicalOffer(ctx, projectID, data); err != nil {
		return nil, err
	}

	return targetSection, nil
}

func (s *technicalOfferService) SaveTechnicalOffer(ctx context.Context, projectID string, data *model.TechnicalOfferData) error {
	project, err := s.repo.GetProject(projectID)
	if err != nil {
		return err
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	project.MemoireTechniqueJSON = string(raw)
	project.UpdatedAt = time.Now()
	project.Tasks, project.Folders = nil, nil
	_, err = s.repo.UpdateProject(project)
	return err
}

func sanitizeFileName(name string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9_\-\.]+`)
	clean := reg.ReplaceAllString(name, "_")
	clean = strings.Trim(clean, "_")
	if len(clean) > 50 {
		clean = clean[:50]
	}
	return clean
}

func (s *technicalOfferService) GenerateDocx(ctx context.Context, projectID string) ([]byte, string, error) {
	data, err := s.GetTechnicalOffer(ctx, projectID)
	if err != nil {
		return nil, "", err
	}

	project, err := s.repo.GetProject(projectID)
	if err != nil {
		return nil, "", err
	}

	profile, _ := s.repo.GetCompanyProfile()

	generator := NewDocxGenerator()
	docxBytes, err := generator.GenerateMemoireDocx(project, profile, data)
	if err != nil {
		return nil, "", fmt.Errorf("erreur génération docx: %w", err)
	}

	cleanTitle := sanitizeFileName(project.Titre)
	if cleanTitle == "" {
		cleanTitle = project.ID
	}
	fileName := fmt.Sprintf("Memoire_Technique_%s.docx", cleanTitle)

	return docxBytes, fileName, nil
}

func (s *technicalOfferService) ExportTechnicalOfferToDrive(ctx context.Context, token *oauth2.Token, projectID string) (string, string, error) {
	data, err := s.GetTechnicalOffer(ctx, projectID)
	if err != nil {
		return "", "", err
	}

	project, err := s.repo.GetProject(projectID)
	if err != nil {
		return "", "", err
	}

	// Rechercher le dossier 03_Offre_Technique/Memoire_Technique ou 03_Offre_Technique
	targetFolderID := ""
	for _, f := range project.Folders {
		if strings.Contains(f.Nom, "Memoire_Technique") || strings.Contains(f.Categorie, "Memoire_Technique") {
			targetFolderID = f.GoogleFolderID
			break
		}
	}
	if targetFolderID == "" {
		for _, f := range project.Folders {
			if strings.Contains(f.Nom, "03_Offre_Technique") || f.Categorie == "03_Offre_Technique" {
				targetFolderID = f.GoogleFolderID
				break
			}
		}
	}

	// Si absent, créer sous le dossier racine du projet
	if targetFolderID == "" && project.DriveFolderID != "" {
		created, fErr := s.drive.CreateFolder(ctx, token, "03_Offre_Technique", project.DriveFolderID)
		if fErr == nil {
			targetFolderID = created.Id
		}
	}

	if targetFolderID == "" {
		return "", "", errors.New("dossier 03_Offre_Technique introuvable sur Google Drive")
	}

	// Génération du fichier Word (.docx) professionnel
	docxBytes, fileName, err := s.GenerateDocx(ctx, projectID)
	if err != nil {
		return "", "", fmt.Errorf("erreur génération du document Word: %w", err)
	}

	contentReader := bytes.NewReader(docxBytes)
	uploaded, err := s.drive.UploadFile(
		ctx,
		token,
		targetFolderID,
		fileName,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		contentReader,
	)
	if err != nil {
		return "", "", fmt.Errorf("erreur téléversement Word sur Drive : %v", err)
	}

	now := time.Now()
	data.DriveFileID = uploaded.Id
	data.DriveFileURL = uploaded.WebViewLink
	data.LastExportedAt = &now
	_ = s.SaveTechnicalOffer(ctx, projectID, data)

	// Passer la tâche "Rédiger le mémoire technique" à TERMINEE
	for _, t := range project.Tasks {
		cleanTask := strings.ToLower(t.Titre)
		if strings.Contains(cleanTask, "mémoire") || strings.Contains(cleanTask, "memoire") {
			_ = s.repo.UpdateTaskStatus(t.ID, model.TaskTerminee)
		}
	}

	// Copie automatique des CVs validés dans 03_Offre_Technique/Personnel_Cle_et_CV
	_, _, _ = s.CopyPersonnelCVsToProject(ctx, token, projectID)

	return uploaded.Id, uploaded.WebViewLink, nil
}

// CopyPersonnelCVsToProject copie automatiquement les fichiers PDF des CVs validés vers 03_Offre_Technique/Personnel_Cle_et_CV
func (s *technicalOfferService) CopyPersonnelCVsToProject(ctx context.Context, token *oauth2.Token, projectID string) (int, string, error) {
	data, err := s.GetTechnicalOffer(ctx, projectID)
	if err != nil {
		return 0, "", err
	}

	project, err := s.repo.GetProject(projectID)
	if err != nil {
		return 0, "", err
	}

	// 1. Rechercher le dossier 03_Offre_Technique/Personnel_Cle_et_CV
	cvFolderID := ""
	for _, f := range project.Folders {
		if strings.Contains(f.Nom, "Personnel_Cle_et_CV") || strings.Contains(f.Categorie, "Personnel_Cle_et_CV") {
			cvFolderID = f.GoogleFolderID
			break
		}
	}

	// Si introuvable dans project.Folders, chercher le dossier parent 03_Offre_Technique pour le créer
	if cvFolderID == "" {
		techFolderID := ""
		for _, f := range project.Folders {
			if strings.Contains(f.Nom, "03_Offre_Technique") || f.Categorie == "03_Offre_Technique" {
				techFolderID = f.GoogleFolderID
				break
			}
		}
		if techFolderID == "" && project.DriveFolderID != "" {
			techFolder, _ := s.drive.CreateFolder(ctx, token, "03_Offre_Technique", project.DriveFolderID)
			if techFolder != nil {
				techFolderID = techFolder.Id
			}
		}
		if techFolderID != "" {
			createdFolder, fErr := s.drive.CreateFolder(ctx, token, "Personnel_Cle_et_CV", techFolderID)
			if fErr == nil && createdFolder != nil {
				cvFolderID = createdFolder.Id
			}
		}
	}

	if cvFolderID == "" {
		return 0, "", errors.New("dossier 03_Offre_Technique/Personnel_Cle_et_CV introuvable sur Google Drive")
	}

	// Lister les fichiers déjà présents dans le dossier pour éviter les doublons
	existingFiles, _ := s.drive.ListFolderFiles(ctx, token, cvFolderID)
	existingNames := make(map[string]bool)
	for _, ef := range existingFiles {
		existingNames[strings.ToLower(ef.Name)] = true
	}

	copiedCount := 0
	for _, p := range data.Personnel {
		if p.DocumentID == "" {
			continue
		}

		cleanNom := strings.TrimSpace(p.Nom)
		if cleanNom == "" {
			cleanNom = "Expert"
		}
		targetFileName := fmt.Sprintf("CV_%s.pdf", strings.ReplaceAll(cleanNom, " ", "_"))

		// Si le fichier existe déjà dans le dossier, ne pas le dupliquer
		if existingNames[strings.ToLower(targetFileName)] {
			continue
		}

		_, copyErr := s.drive.CopyFile(ctx, token, p.DocumentID, cvFolderID, targetFileName)
		if copyErr == nil {
			copiedCount++
			existingNames[strings.ToLower(targetFileName)] = true
		}
	}

	// Mettre à jour la tâche "Collecter les CV" si elle existe
	for _, t := range project.Tasks {
		cleanTask := strings.ToLower(t.Titre)
		if strings.Contains(cleanTask, "cv") || strings.Contains(cleanTask, "personnel") {
			_ = s.repo.UpdateTaskStatus(t.ID, model.TaskTerminee)
		}
	}

	cvFolderURL := fmt.Sprintf("https://drive.google.com/drive/folders/%s", cvFolderID)
	return copiedCount, cvFolderURL, nil
}

type PersonnelProposalResponse struct {
	Personnel []struct {
		RoleRequis       string `json:"role_requis"`
		Nom              string `json:"nom"`
		ExperienceAnnees int    `json:"experience_annees"`
		Qualifications   string `json:"qualifications"`
	} `json:"personnel"`
}

func (s *technicalOfferService) ProposePersonnel(ctx context.Context, projectID string) ([]model.PersonnelAffectation, error) {
	data, err := s.GetTechnicalOffer(ctx, projectID)
	if err != nil {
		return nil, err
	}
	project, err := s.repo.GetProject(projectID)
	if err != nil {
		return nil, err
	}
	companyDocs, _ := s.repo.ListCompanyDocuments()

	var roles []string
	for _, p := range data.Personnel {
		if strings.TrimSpace(p.RoleRequis) != "" {
			roles = append(roles, p.RoleRequis)
		}
	}
	if len(roles) == 0 {
		return data.Personnel, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Projet : %s\n", project.Titre))
	if project.Acheteur != "" {
		sb.WriteString(fmt.Sprintf("Autorité Contractante / Acheteur : %s\n", project.Acheteur))
	}
	sb.WriteString(fmt.Sprintf("Rôles exigés par le DAO : %s\n\n", strings.Join(roles, ", ")))

	if len(companyDocs) > 0 {
		sb.WriteString("Documents et CVs disponibles dans le coffre-fort d'entreprise :\n")
		for _, doc := range companyDocs {
			sb.WriteString(fmt.Sprintf("- [%s] %s : %s\n", doc.TypeDoc, doc.Titre, doc.Description))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Pour chacun des rôles exigés ci-dessus, propose l'affectation d'un expert optimal (Nom, nombre d'années d'expérience, diplômes et spécialités).")

	rawJSON, err := s.agents.ExecuteAgentTask(ctx, PersonnelMatchingAgent, nil, sb.String())
	if err != nil {
		return nil, fmt.Errorf("erreur agent affectation personnel : %w", err)
	}

	var resp PersonnelProposalResponse
	if err := json.Unmarshal([]byte(rawJSON), &resp); err != nil {
		return nil, fmt.Errorf("erreur décodage proposition personnel : %w", err)
	}

	for i := range data.Personnel {
		roleClean := cleanKeyword(data.Personnel[i].RoleRequis)
		for _, prop := range resp.Personnel {
			propClean := cleanKeyword(prop.RoleRequis)
			if strings.Contains(propClean, roleClean) || strings.Contains(roleClean, propClean) || len(resp.Personnel) == 1 {
				data.Personnel[i].Nom = prop.Nom
				data.Personnel[i].ExperienceAnnees = prop.ExperienceAnnees
				data.Personnel[i].Qualifications = prop.Qualifications

				// Vérifier si un CV du coffre-fort correspond
				for _, doc := range companyDocs {
					if doc.TypeDoc == model.DocCV || strings.Contains(cleanKeyword(doc.Titre), "cv") {
						if strings.Contains(cleanKeyword(doc.Titre), cleanKeyword(prop.Nom)) || strings.Contains(cleanKeyword(prop.Nom), cleanKeyword(doc.Titre)) {
							data.Personnel[i].DocumentID = doc.GoogleDriveID
							data.Personnel[i].DocumentURL = doc.WebViewLink
							break
						}
					}
				}
				break
			}
		}
	}

	_ = s.SaveTechnicalOffer(ctx, projectID, data)
	return data.Personnel, nil
}

type MaterielProposalResponse struct {
	Materiel []struct {
		TypeRequis  string `json:"type_requis"`
		Designation string `json:"designation"`
		Statut      string `json:"statut"`
	} `json:"materiel"`
}

func (s *technicalOfferService) ProposeMateriel(ctx context.Context, projectID string) ([]model.MaterielMobilisation, error) {
	data, err := s.GetTechnicalOffer(ctx, projectID)
	if err != nil {
		return nil, err
	}
	project, err := s.repo.GetProject(projectID)
	if err != nil {
		return nil, err
	}
	companyDocs, _ := s.repo.ListCompanyDocuments()

	var typesReq []string
	for _, m := range data.Materiel {
		if strings.TrimSpace(m.TypeRequis) != "" {
			typesReq = append(typesReq, m.TypeRequis)
		}
	}
	if len(typesReq) == 0 {
		return data.Materiel, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Projet : %s\n", project.Titre))
	sb.WriteString(fmt.Sprintf("Équipements et matériels exigés par le DAO : %s\n\n", strings.Join(typesReq, ", ")))

	if len(companyDocs) > 0 {
		sb.WriteString("Matériels enregistrés dans le coffre-fort d'entreprise :\n")
		for _, doc := range companyDocs {
			if doc.TypeDoc == model.DocMateriel || strings.Contains(cleanKeyword(doc.Titre), "materiel") {
				sb.WriteString(fmt.Sprintf("- %s : %s\n", doc.Titre, doc.Description))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Pour chaque équipement exigé par le DAO, propose la marque et le modèle précis d'engin de chantier et son statut juridique.")

	rawJSON, err := s.agents.ExecuteAgentTask(ctx, MaterielMatchingAgent, nil, sb.String())
	if err != nil {
		return nil, fmt.Errorf("erreur agent allocation matériel : %w", err)
	}

	var resp MaterielProposalResponse
	if err := json.Unmarshal([]byte(rawJSON), &resp); err != nil {
		return nil, fmt.Errorf("erreur décodage proposition matériel : %w", err)
	}

	for i := range data.Materiel {
		typeClean := cleanKeyword(data.Materiel[i].TypeRequis)
		for _, prop := range resp.Materiel {
			propClean := cleanKeyword(prop.TypeRequis)
			if strings.Contains(propClean, typeClean) || strings.Contains(typeClean, propClean) || len(resp.Materiel) == 1 {
				data.Materiel[i].Designation = prop.Designation
				if prop.Statut != "" {
					data.Materiel[i].Statut = prop.Statut
				}
				break
			}
		}
	}

	_ = s.SaveTechnicalOffer(ctx, projectID, data)
	return data.Materiel, nil
}

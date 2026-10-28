// service/agent_service.go
package service

import (
	"agents-dao-platform/config"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"google.golang.org/genai"
)

// AgentConfig définit l'identité et le comportement d'un agent spécifique
type AgentConfig struct {
	Name         string
	SystemPrompt string
	Temperature  float32
	Schema       *genai.Schema // Schéma JSON de sortie propre à l'agent
}

// Document représente une pièce transmise à un agent (PDF, etc.)
type Document struct {
	Name     string
	Bytes    []byte
	MimeType string
}

// AdminAnalysis est la sortie structurée de l'agent Conformité Administrative
type AdminAnalysis struct {
	Acheteur              string          `json:"acheteur"`
	ObjetMarche           string          `json:"objet_marche"`
	NumeroAppelOffres     string          `json:"numero_appel_offres"`
	DateLimiteDepot       string          `json:"date_limite_depot"` // format AAAA-MM-JJ si connue
	CriteresEliminatoires []string        `json:"criteres_eliminatoires"`
	PiecesAdministratives []string        `json:"pieces_administratives"`
	Calendrier            []EtapeCalendar `json:"calendrier"`
	Lots                  []Lot           `json:"lots"`
	Recommandations       string          `json:"recommandations"`
}

type EtapeCalendar struct {
	Etape string `json:"etape"`
	Date  string `json:"date"` // AAAA-MM-JJ si connue
}

type Lot struct {
	NumeroLot      int      `json:"numero_lot"`
	Intitule       string   `json:"intitule"`
	Cautionnement  string   `json:"cautionnement"`
	PiecesRequises []string `json:"pieces_requises"`
}

// TechnicalAnalysis est la sortie structurée de l'agent Technique
type TechnicalAnalysis struct {
	ConsistanceTravaux   string   `json:"consistance_travaux"`
	ContraintesExecution []string `json:"contraintes_execution"`
	NormesExigees        []string `json:"normes_exigees"`
	MaterielRequis       []string `json:"materiel_requis"`
	PersonnelCle         []string `json:"personnel_cle"`
	PlanMemoireTechnique []string `json:"plan_memoire_technique"`
	PointsVigilance      string   `json:"points_vigilance"`
}

type AgentService interface {
	ExecuteAgentTask(ctx context.Context, agent AgentConfig, docs []Document, userInstruction string) (string, error)
}

type agentService struct {
	cfg    *config.AppConfig
	once   sync.Once
	client *genai.Client
	err    error
}

func NewAgentService(cfg *config.AppConfig) AgentService {
	return &agentService{cfg: cfg}
}

func (s *agentService) getClient(ctx context.Context) (*genai.Client, error) {
	s.once.Do(func() {
		httpClient := &http.Client{Timeout: 120 * time.Second}
		if s.cfg.GeminiAPIKey != "" {
			s.client, s.err = genai.NewClient(ctx, &genai.ClientConfig{
				APIKey:     s.cfg.GeminiAPIKey,
				Backend:    genai.BackendGeminiAPI,
				HTTPClient: httpClient,
			})
		} else {
			s.client, s.err = genai.NewClient(ctx, &genai.ClientConfig{
				Project:    s.cfg.ProjectID,
				Location:   s.cfg.Region,
				Backend:    genai.BackendVertexAI,
				HTTPClient: httpClient,
			})
		}
	})
	if s.err != nil {
		return nil, fmt.Errorf("erreur initialisation client genai : %v", s.err)
	}
	return s.client, nil
}

func strArr() *genai.Schema {
	return &genai.Schema{Type: genai.TypeArray, Items: &genai.Schema{Type: genai.TypeString}}
}

// AdminSchema : schéma de sortie de l'agent Conformité Administrative
var AdminSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"acheteur":               {Type: genai.TypeString, Description: "Entité qui émet l'appel d'offres"},
		"objet_marche":           {Type: genai.TypeString, Description: "Objet global du marché"},
		"numero_appel_offres":    {Type: genai.TypeString, Description: "Numéro d'identification du DAO"},
		"date_limite_depot":      {Type: genai.TypeString, Description: "Date limite de dépôt au format AAAA-MM-JJ, vide si inconnue"},
		"criteres_eliminatoires": strArr(),
		"pieces_administratives": strArr(),
		"calendrier": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"etape": {Type: genai.TypeString},
					"date":  {Type: genai.TypeString, Description: "AAAA-MM-JJ, vide si inconnue"},
				},
				Required: []string{"etape"},
			},
		},
		"lots": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"numero_lot":      {Type: genai.TypeInteger},
					"intitule":        {Type: genai.TypeString},
					"cautionnement":   {Type: genai.TypeString},
					"pieces_requises": strArr(),
				},
				Required: []string{"numero_lot", "intitule"},
			},
		},
		"recommandations": {Type: genai.TypeString, Description: "Points de vigilance majeurs pour la soumission"},
	},
	Required: []string{"acheteur", "objet_marche", "pieces_administratives", "criteres_eliminatoires"},
}

// TechnicalSchema : schéma de sortie de l'agent Technique
var TechnicalSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"consistance_travaux":    {Type: genai.TypeString},
		"contraintes_execution":  strArr(),
		"normes_exigees":         strArr(),
		"materiel_requis":        strArr(),
		"personnel_cle":          strArr(),
		"plan_memoire_technique": strArr(),
		"points_vigilance":       {Type: genai.TypeString},
	},
	Required: []string{"consistance_travaux", "plan_memoire_technique"},
}

// PersonnelMatchingSchema : schéma pour la proposition et affectation du personnel clé par IA
var PersonnelMatchingSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"personnel": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"role_requis":       {Type: genai.TypeString},
					"nom":               {Type: genai.TypeString},
					"experience_annees": {Type: genai.TypeInteger},
					"qualifications":    {Type: genai.TypeString},
				},
				Required: []string{"role_requis", "nom", "experience_annees", "qualifications"},
			},
		},
	},
	Required: []string{"personnel"},
}

// MaterielMatchingSchema : schéma pour la proposition d'engins et matériels de chantier par IA
var MaterielMatchingSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"materiel": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"type_requis": {Type: genai.TypeString},
					"designation": {Type: genai.TypeString},
					"statut":      {Type: genai.TypeString},
				},
				Required: []string{"type_requis", "designation", "statut"},
			},
		},
	},
	Required: []string{"materiel"},
}

// ExtractedCVProfile structure les informations extraites directement d'un fichier CV par l'IA
type ExtractedCVProfile struct {
	NomComplet           string   `json:"nom_complet"`
	SpecialitePrincipale string   `json:"specialite_principale"`
	TitreProfessionnel   string   `json:"titre_professionnel"`
	ExperienceAnnees     int      `json:"experience_annees"`
	Diplomes             []string `json:"diplomes"`
	CompetencesCles      []string `json:"competences_cles"`
	ResumeAnalytique     string   `json:"resume_analytique"`
}

// ExtractedAdminDocProfile structure les informations extraites d'un scan administratif ou fiscal
type ExtractedAdminDocProfile struct {
	TypeDocument      string `json:"type_document"`      // "QUITUS_FISCAL", "IPRES", "CSS", "INSPECTION_TRAVAIL", "NINEA", "RCCM", "AUTRE"
	IntituleOfficiel  string `json:"intitule_officiel"`  // ex: "Quitus Fiscal (Attestation de Régularité Fiscale DGID)"
	OrganismeEmetteur string `json:"organisme_emetteur"` // ex: "DGID", "IPRES", "Caisse de Sécurité Sociale (CSS)", "Inspection du Travail"
	Beneficiaire      string `json:"beneficiaire"`       // Raison sociale ou NINEA
	DateEmission      string `json:"date_emission"`      // Format AAAA-MM-JJ
	DateExpiration    string `json:"date_expiration"`    // Format AAAA-MM-JJ (obligatoire : date limite ou fin de trimestre)
	EstValide         bool   `json:"est_valide"`
	Commentaire       string `json:"commentaire"`
}

// AdminDocParserSchema : schéma pour l'analyse OCR et sémantique de scans administratifs (JPG, PNG, PDF)
var AdminDocParserSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"type_document": {
			Type:        genai.TypeString,
			Description: "Type exact du document: QUITUS_FISCAL, IPRES, CSS, INSPECTION_TRAVAIL, NINEA, RCCM, ou AUTRE",
		},
		"intitule_officiel": {
			Type:        genai.TypeString,
			Description: "Intitulé officiel complet (ex: Quitus Fiscal (Attestation de Régularité Fiscale DGID), Attestation de Régularité IPRES)",
		},
		"organisme_emetteur": {
			Type:        genai.TypeString,
			Description: "Organisme émetteur (ex: DGID, IPRES, CSS, Inspection du Travail)",
		},
		"beneficiaire": {
			Type:        genai.TypeString,
			Description: "Nom de l'entreprise ou contribuable figurant sur la pièce",
		},
		"date_emission": {
			Type:        genai.TypeString,
			Description: "Date d'émission au format AAAA-MM-JJ (ou vide si absente)",
		},
		"date_expiration": {
			Type:        genai.TypeString,
			Description: "Date de fin de validité ou d'expiration AU FORMAT STRICT AAAA-MM-JJ (cherche 'valable jusqu au', 'expire le', ou fin de période trimestrielle)",
		},
		"est_valide": {
			Type:        genai.TypeBoolean,
			Description: "true si l'attestation est toujours valide",
		},
		"commentaire": {
			Type:        genai.TypeString,
			Description: "Brève synthèse de la régularité et de la période couverte",
		},
	},
	Required: []string{"type_document", "intitule_officiel", "date_expiration", "commentaire"},
}

// CVParserSchema : schéma pour l'analyse et l'extraction sémantique d'un CV PDF
var CVParserSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"nom_complet":           {Type: genai.TypeString, Description: "Nom et prénom de l'expert"},
		"specialite_principale": {Type: genai.TypeString, Description: "Spécialité clé (ex: Environnementaliste / HSE, Ingénieur Génie Civil, Géomètre-Topographe, Chef de Chantier, Hydraulique, etc.)"},
		"titre_professionnel":   {Type: genai.TypeString, Description: "Titre ou poste actuel"},
		"experience_annees":     {Type: genai.TypeInteger, Description: "Nombre total d'années d'expérience professionnelle"},
		"diplomes":              strArr(),
		"competences_cles":      strArr(),
		"resume_analytique":     {Type: genai.TypeString, Description: "Résumé analytique percutant du profil en 1 à 2 phrases pour les dossiers de soumission"},
	},
	Required: []string{"nom_complet", "specialite_principale", "resume_analytique"},
}

func (s *agentService) ExecuteAgentTask(ctx context.Context, agent AgentConfig, docs []Document, userInstruction string) (string, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return "", err
	}

	parts := []*genai.Part{genai.NewPartFromText(userInstruction)}
	for _, d := range docs {
		if len(d.Bytes) == 0 {
			continue
		}
		mime := d.MimeType
		if mime == "" {
			mime = "application/pdf"
		}
		parts = append(parts, genai.NewPartFromBytes(d.Bytes, mime))
	}

	temp := agent.Temperature
	respMime := "application/json"
	if agent.Schema == nil {
		respMime = "text/plain"
	}
	genConfig := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(agent.SystemPrompt, genai.RoleUser),
		Temperature:       &temp,
		ResponseMIMEType:  respMime,
		ResponseSchema:    agent.Schema,
	}

	// Liste des modèles candidats avec modèle principal en premier
	candidateModels := []string{s.cfg.GeminiModel}
	if s.cfg.GeminiModel != "gemini-3.5-flash" {
		candidateModels = append(candidateModels, "gemini-3.5-flash")
	}
	if s.cfg.GeminiModel != "gemini-3-flash-preview" {
		candidateModels = append(candidateModels, "gemini-3-flash-preview")
	}

	var lastErr error
	for _, modelName := range candidateModels {
		for attempt := 1; attempt <= 3; attempt++ {
			result, err := client.Models.GenerateContent(ctx, modelName,
				[]*genai.Content{genai.NewContentFromParts(parts, genai.RoleUser)},
				genConfig,
			)
			if err == nil && result != nil {
				return result.Text(), nil
			}

			lastErr = err
			if err != nil {
				errStr := err.Error()
				// Si erreur de charge temporaire Google (503, 429, deadline), attendre et réessayer
				if strings.Contains(errStr, "503") ||
					strings.Contains(errStr, "high demand") ||
					strings.Contains(errStr, "UNAVAILABLE") ||
					strings.Contains(errStr, "DEADLINE_EXCEEDED") ||
					strings.Contains(errStr, "429") {
					time.Sleep(time.Duration(attempt*2) * time.Second)
					continue
				}
			}

			// Autre type d'erreur, essayer le modèle suivant
			break
		}
	}

	return "", fmt.Errorf("erreur exécution agent %s : %v", agent.Name, lastErr)
}

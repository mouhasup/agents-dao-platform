package model

import (
	"time"
)

// ProjectStatus définit l'état d'avancement d'un appel d'offres
type ProjectStatus string

const (
	StatusBrouillon ProjectStatus = "BROUILLON"
	StatusEnAnalyse ProjectStatus = "EN_ANALYSE"
	StatusEnMontage ProjectStatus = "EN_MONTAGE"
	StatusPretDepot ProjectStatus = "PRET_DEPOT"
	StatusDepose    ProjectStatus = "DEPOSE"
	StatusGagne     ProjectStatus = "GAGNE"
	StatusPerdu     ProjectStatus = "PERDU"
)

// Project représente un dossier de soumission à un appel d'offres
type Project struct {
	ID                string        `json:"id" gorm:"primaryKey"`
	Titre             string        `json:"titre" gorm:"not null"`
	NumeroDAO         string        `json:"numero_dao"`
	Acheteur          string        `json:"acheteur"`
	Statut            ProjectStatus `json:"statut" gorm:"default:'BROUILLON'"`
	DriveFolderID     string        `json:"drive_folder_id"` // ID du dossier racine créé sur Google Drive
	DriveFolderURL    string        `json:"drive_folder_url"`
	DateLimiteDepot   *time.Time    `json:"date_limite_depot"`
	AnalyseResultJSON   string        `json:"analyse_result_json" gorm:"type:text"` // Résultat structuré JSON consolidé
	AnalyseError        string        `json:"analyse_error" gorm:"type:text"`
	AnalysedAt          *time.Time    `json:"analysed_at"`
	MemoireTechniqueJSON string       `json:"memoire_technique_json" gorm:"type:text"` // Mémoire technique et offre technique rédigée
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`

	Tasks   []ProjectTask `json:"tasks,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	Folders []DriveFolder `json:"folders,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}

// TaskStatus définit le statut d'une tâche de préparation
type TaskStatus string

const (
	TaskAFaire   TaskStatus = "A_FAIRE"
	TaskEnCours  TaskStatus = "EN_COURS"
	TaskTerminee TaskStatus = "TERMINEE"
)

// ProjectTask représente une tâche générée par l'IA ou créée par l'utilisateur
type ProjectTask struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	ProjectID   string     `json:"project_id" gorm:"index;not null"`
	Titre       string     `json:"titre" gorm:"not null"`
	Description string     `json:"description" gorm:"type:text"`
	Categorie   string     `json:"categorie"` // "ADMINISTRATIF", "TECHNIQUE", "FINANCIER", "VISITE_SITE"
	Echeance    *time.Time `json:"echeance"`
	Priorite    string     `json:"priorite"` // "HAUTE", "MOYENNE", "BASSE"
	Statut      TaskStatus `json:"statut" gorm:"default:'A_FAIRE'"`
	CreatedAt   time.Time  `json:"created_at"`
}

// DriveFolder représente un sous-dossier de l'arborescence normalisée du projet sur Google Drive
type DriveFolder struct {
	ID             string `json:"id" gorm:"primaryKey"`
	ProjectID      string `json:"project_id" gorm:"index;not null"`
	GoogleFolderID string `json:"google_folder_id" gorm:"not null"`
	ParentFolderID string `json:"parent_folder_id"`
	Nom            string `json:"nom" gorm:"not null"`
	Categorie      string `json:"categorie"` // "01_DAO_ORIGINAL", "02_ADMIN", "03_TECHNIQUE", "04_FINANCIER"
}

// TypeDocumentEntreprise définit la nature d'un document du coffre-fort RAG
type TypeDocumentEntreprise string

const (
	DocCV          TypeDocumentEntreprise = "CV_EXPERT"
	DocAttestation TypeDocumentEntreprise = "ATTESTATION_FISCALE_SOCIALE"
	DocReference   TypeDocumentEntreprise = "REFERENCE_SERVICE_FAIT"
	DocMateriel    TypeDocumentEntreprise = "JUSTIFICATIF_MATERIEL"
	DocBilan       TypeDocumentEntreprise = "BILAN_FINANCIER"
	DocAutre       TypeDocumentEntreprise = "AUTRE"
)

// CompanyDocument représente une pièce du référentiel d'entreprise (RAG)
type CompanyDocument struct {
	ID              string                 `json:"id" gorm:"primaryKey"`
	ProfileID       string                 `json:"profile_id" gorm:"index"`
	TypeDoc         TypeDocumentEntreprise `json:"type_doc" gorm:"not null"`
	Titre           string                 `json:"titre" gorm:"not null"`
	Description     string                 `json:"description" gorm:"type:text"`
	GoogleDriveID   string                 `json:"google_drive_id,omitempty"`
	WebViewLink     string                 `json:"web_view_link,omitempty"`
	FileName        string                 `json:"file_name,omitempty"`
	DateExpiration  *time.Time             `json:"date_expiration,omitempty"`
	MetadonneesJSON string                 `json:"metadonnees_json,omitempty" gorm:"type:text"` // ex: {"annees_experience": 10}
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// CompanyProfile représente la fiche signalétique et le coffre-fort documentaire de l'entreprise
type CompanyProfile struct {
	ID                 string            `json:"id" gorm:"primaryKey"`
	Nom                string            `json:"nom" gorm:"not null"`
	NINEA              string            `json:"ninea"`
	RCCM               string            `json:"rccm"`
	Email              string            `json:"email"`
	Telephone          string            `json:"telephone"`
	Adresse            string            `json:"adresse"`
	WorkspaceFolderID  string            `json:"workspace_folder_id"`  // ID du dossier AGENTS_DAO_WORKSPACE
	WorkspaceFolderURL string            `json:"workspace_folder_url"`
	RagFolderID        string            `json:"rag_folder_id"`        // ID du dossier 01_COFFRE_FORT_ENTREPRISE
	ProjectsFolderID   string            `json:"projects_folder_id"`   // ID du dossier 02_PROJETS_DAO
	Documents          []CompanyDocument `json:"documents,omitempty" gorm:"foreignKey:ProfileID"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

// MemoireSection représente un chapitre du mémoire technique
type MemoireSection struct {
	ID        string     `json:"id"`
	Titre     string     `json:"titre"`
	Contenu   string     `json:"contenu"` // Markdown rédigé par l'IA ou ajusté par l'utilisateur
	Statut    string     `json:"statut"`  // "A_REDIGER", "GENERE", "VALIDE"
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// PersonnelAffectation représente l'affectation d'un membre du personnel clé
type PersonnelAffectation struct {
	RoleRequis       string `json:"role_requis"`
	Nom              string `json:"nom"`
	ExperienceAnnees int    `json:"experience_annees"`
	Qualifications   string `json:"qualifications"`
	DocumentID       string `json:"document_id,omitempty"` // ID CompanyDocument du CV
	DocumentURL      string `json:"document_url,omitempty"`
}

// MaterielMobilisation représente un engin ou équipement mobilisé
type MaterielMobilisation struct {
	TypeRequis      string `json:"type_requis"`
	Designation     string `json:"designation"`
	Immatriculation string `json:"immatriculation,omitempty"`
	Statut          string `json:"statut"` // "PROPRIETE", "LOCATION", "LEASING"
	DocumentID      string `json:"document_id,omitempty"`
	DocumentURL     string `json:"document_url,omitempty"`
}

// TechnicalOfferData représente l'état consolidé de l'offre technique
type TechnicalOfferData struct {
	ProjectID      string                 `json:"project_id"`
	Titre          string                 `json:"titre"`
	Sections       []MemoireSection       `json:"sections"`
	Personnel      []PersonnelAffectation `json:"personnel"`
	Materiel       []MaterielMobilisation `json:"materiel"`
	DriveFileID    string                 `json:"drive_file_id,omitempty"`
	DriveFileURL   string                 `json:"drive_folder_url,omitempty"`
	LastExportedAt *time.Time             `json:"last_exported_at,omitempty"`
}

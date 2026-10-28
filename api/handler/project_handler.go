package handler

import (
	"agents-dao-platform/model"
	"agents-dao-platform/service"
	"agents-dao-platform/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

const maxDaoUploadBytes = 30 << 20 // 30 Mo

type ProjectHandler struct {
	projectRepo  service.ProjectRepository
	driveService service.DriveService
	analysis     service.AnalysisService
	sessions     service.SessionRepository
	technical    service.TechnicalOfferService
	agentService service.AgentService
}

func NewProjectHandler(repo service.ProjectRepository, ds service.DriveService, an service.AnalysisService, sessions service.SessionRepository, tech service.TechnicalOfferService, agent service.AgentService) *ProjectHandler {
	return &ProjectHandler{
		projectRepo:  repo,
		driveService: ds,
		analysis:     an,
		sessions:     sessions,
		technical:    tech,
		agentService: agent,
	}
}

// requireToken renvoie le token OAuth ou répond 401
func (h *ProjectHandler) requireToken(c *gin.Context) *oauth2.Token {
	token, _ := c.Request.Context().Value("oauth_token").(*oauth2.Token)
	if token == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Connectez-vous à Google Drive pour continuer.", "needs_reauth": true})
	}
	return token
}

// handleDriveError traduit les erreurs Drive (token expiré…) en réponses HTTP
func (h *ProjectHandler) handleDriveError(c *gin.Context, err error) {
	errStr := err.Error()
	if strings.Contains(errStr, "invalid_grant") || strings.Contains(errStr, "token has been expired") {
		if sessionID, cookieErr := c.Cookie(utils.SessionCookieName); cookieErr == nil {
			_ = h.sessions.Delete(sessionID)
		}
		c.SetCookie(utils.SessionCookieName, "", -1, "/", "", false, true)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Votre session Google a expiré. Veuillez vous reconnecter.", "needs_reauth": true})
		return
	}
	c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("Erreur Google Drive : %v", err)})
}

// ListDaoDocuments gère GET /api/projects/:id/documents
func (h *ProjectHandler) ListDaoDocuments(c *gin.Context) {
	project, err := h.projectRepo.GetProject(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Projet introuvable"})
		return
	}
	token := h.requireToken(c)
	if token == nil {
		return
	}
	files, err := h.analysis.ListDaoDocuments(c.Request.Context(), token, project)
	if errors.Is(err, service.ErrNoDaoFolder) {
		c.JSON(http.StatusOK, gin.H{"documents": []any{}, "has_folder": false})
		return
	}
	if err != nil {
		h.handleDriveError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"documents": files, "has_folder": true})
}

// UploadDaoDocument gère POST /api/projects/:id/documents (multipart, champ "file")
func (h *ProjectHandler) UploadDaoDocument(c *gin.Context) {
	project, err := h.projectRepo.GetProject(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Projet introuvable"})
		return
	}
	token := h.requireToken(c)
	if token == nil {
		return
	}
	folderID := h.analysis.DaoFolderID(project)
	if folderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ce projet n'a pas d'arborescence Drive (dossier du DAO manquant)."})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxDaoUploadBytes+1<<20)
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Fichier manquant ou trop volumineux (max 30 Mo)."})
		return
	}
	if fh.Size > maxDaoUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Fichier trop volumineux (max 30 Mo)."})
		return
	}
	if strings.ToLower(filepath.Ext(fh.Filename)) != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Seuls les fichiers PDF sont acceptés."})
		return
	}
	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lecture du fichier impossible."})
		return
	}
	defer f.Close()

	uploaded, err := h.driveService.UploadFile(c.Request.Context(), token, folderID, filepath.Base(fh.Filename), "application/pdf", f)
	if err != nil {
		h.handleDriveError(c, err)
		return
	}
	c.JSON(http.StatusCreated, uploaded)
}

// AnalyzeProject gère POST /api/projects/:id/analyze (asynchrone, 202)
func (h *ProjectHandler) AnalyzeProject(c *gin.Context) {
	token := h.requireToken(c)
	if token == nil {
		return
	}
	err := h.analysis.StartAnalysis(c.Request.Context(), token, c.Param("id"))
	switch {
	case err == nil:
		c.JSON(http.StatusAccepted, gin.H{"statut": model.StatusEnAnalyse})
	case errors.Is(err, service.ErrAnalysisRunning):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrNoDaoDocuments), errors.Is(err, service.ErrNoDaoFolder):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		h.handleDriveError(c, err)
	}
}

type updateTaskRequest struct {
	Statut model.TaskStatus `json:"statut" binding:"required"`
}

// UpdateTaskStatus gère PATCH /api/tasks/:taskId
func (h *ProjectHandler) UpdateTaskStatus(c *gin.Context) {
	var req updateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Statut requis"})
		return
	}
	switch req.Statut {
	case model.TaskAFaire, model.TaskEnCours, model.TaskTerminee:
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Statut invalide"})
		return
	}
	if err := h.projectRepo.UpdateTaskStatus(c.Param("taskId"), req.Statut); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": c.Param("taskId"), "statut": req.Statut})
}

// ListDriveFolders gère GET /api/drive/folders
func (h *ProjectHandler) ListDriveFolders(c *gin.Context) {
	parentID := c.DefaultQuery("parent", "root")
	search := strings.TrimSpace(c.Query("q"))
	folderID := strings.TrimSpace(c.Query("folder_id"))

	token, _ := c.Request.Context().Value("oauth_token").(*oauth2.Token)
	if token == nil {
		c.JSON(http.StatusOK, gin.H{"folders": []any{}, "authenticated": false})
		return
	}

	// Si l'utilisateur demande des informations sur un dossier précis (ex: validation d'un lien/ID collé)
	if folderID != "" {
		folder, err := h.driveService.GetFolder(c.Request.Context(), token, folderID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "authenticated": true})
			return
		}
		c.JSON(http.StatusOK, gin.H{"folder": folder, "authenticated": true})
		return
	}

	folders, err := h.driveService.ListFolders(c.Request.Context(), token, parentID, search)
	if err != nil {
		errStr := err.Error()
		// Token révoqué ou expiré → effacer la session et demander reconnexion
		if strings.Contains(errStr, "invalid_grant") || strings.Contains(errStr, "token has been expired") {
			if sessionID, cookieErr := c.Cookie(utils.SessionCookieName); cookieErr == nil {
				_ = h.sessions.Delete(sessionID)
			}
			c.SetCookie(utils.SessionCookieName, "", -1, "/", "", false, true)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":         "Votre session Google a expiré. Veuillez vous reconnecter.",
				"authenticated": false,
				"needs_reauth":  true,
			})
			return
		}
		// Périmètre insuffisant → demander reconnexion avec Drive coché
		if strings.Contains(errStr, "insufficient") || strings.Contains(errStr, "403") {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "L'accès à Google Drive n'a pas été autorisé. Reconnectez-vous en cochant la case Google Drive.",
				"authenticated": true,
				"needs_reauth":  true,
			})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("Erreur récupération dossiers: %v", err), "authenticated": true})
		return
	}

	c.JSON(http.StatusOK, gin.H{"folders": folders, "authenticated": true})
}

// ListProjects gère GET /api/projects
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	projects, err := h.projectRepo.ListProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erreur récupération projets: %v", err)})
		return
	}
	c.JSON(http.StatusOK, projects)
}

// GetProject gère GET /api/projects/:id
func (h *ProjectHandler) GetProject(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID projet manquant"})
		return
	}

	project, err := h.projectRepo.GetProject(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Projet introuvable"})
		return
	}
	c.JSON(http.StatusOK, project)
}

type CreateProjectRequest struct {
	Titre          string `json:"titre" binding:"required"`
	DriveParentID  string `json:"drive_parent_id"`
	OriginalFileID string `json:"original_file_id"`
}

// InitWorkspace gère POST /api/workspace/init
func (h *ProjectHandler) InitWorkspace(c *gin.Context) {
	token := h.requireToken(c)
	if token == nil {
		return
	}

	var body struct {
		ParentID string `json:"parent_id"`
	}
	_ = c.ShouldBindJSON(&body)

	profile, err := h.projectRepo.GetCompanyProfile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de charger le profil entreprise"})
		return
	}

	wsID, ragID, projID, wsURL, err := h.driveService.EnsureWorkspaceTree(c.Request.Context(), token, body.ParentID)
	if err != nil {
		h.handleDriveError(c, err)
		return
	}

	profile.WorkspaceFolderID = wsID
	profile.WorkspaceFolderURL = wsURL
	profile.RagFolderID = ragID
	profile.ProjectsFolderID = projID
	profile.UpdatedAt = time.Now()

	if err := h.projectRepo.UpdateCompanyProfile(profile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erreur sauvegarde profil: %v", err)})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// CreateProject gère POST /api/projects
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides : le titre est requis"})
		return
	}

	projectID := fmt.Sprintf("proj-%d", time.Now().Unix())
	project := &model.Project{
		ID:        projectID,
		Titre:     req.Titre,
		Statut:    model.StatusBrouillon,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Création de l'arborescence Google Drive si un token est disponible
	token, _ := c.Request.Context().Value("oauth_token").(*oauth2.Token)
	if token != nil {
		parentID := req.DriveParentID
		// Si parentID est vide ou "root", utiliser automatiquement le dossier 02_PROJETS_DAO
		if parentID == "" || parentID == "root" {
			profile, err := h.projectRepo.GetCompanyProfile()
			if err == nil {
				if profile.ProjectsFolderID != "" {
					parentID = profile.ProjectsFolderID
				} else {
					wsID, ragID, projID, wsURL, wsErr := h.driveService.EnsureWorkspaceTree(c.Request.Context(), token, "")
					if wsErr == nil {
						profile.WorkspaceFolderID = wsID
						profile.WorkspaceFolderURL = wsURL
						profile.RagFolderID = ragID
						profile.ProjectsFolderID = projID
						_ = h.projectRepo.UpdateCompanyProfile(profile)
						parentID = projID
					}
				}
			}
		}

		rootFolderName := fmt.Sprintf("[DAO] %s", req.Titre)
		rootID, folderMap, err := h.driveService.CreateProjectTree(c.Request.Context(), token, rootFolderName, parentID)
		if err == nil {
			project.DriveFolderID = rootID
			project.DriveFolderURL = fmt.Sprintf("https://drive.google.com/drive/folders/%s", rootID)

			for path, folderID := range folderMap {
				project.Folders = append(project.Folders, model.DriveFolder{
					ID:             fmt.Sprintf("f-%d-%s", time.Now().UnixNano(), folderID),
					ProjectID:      projectID,
					GoogleFolderID: folderID,
					Nom:            path,
					Categorie:      path,
				})
			}
		}
	}

	created, err := h.projectRepo.CreateProject(project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erreur création projet: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// GetCompanyProfile gère GET /api/company
func (h *ProjectHandler) GetCompanyProfile(c *gin.Context) {
	profile, err := h.projectRepo.GetCompanyProfile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur profil entreprise"})
		return
	}

	docs, _ := h.projectRepo.ListCompanyDocuments()
	profile.Documents = docs
	c.JSON(http.StatusOK, profile)
}

// AddCompanyDocument gère POST /api/company/documents
func (h *ProjectHandler) AddCompanyDocument(c *gin.Context) {
	var doc model.CompanyDocument
	if err := c.ShouldBindJSON(&doc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données du document invalides"})
		return
	}

	profile, err := h.projectRepo.GetCompanyProfile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de charger le profil entreprise"})
		return
	}

	doc.ProfileID = profile.ID
	doc.ID = fmt.Sprintf("doc-%d", time.Now().UnixNano())
	if err := h.projectRepo.AddCompanyDocument(&doc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erreur ajout document: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

// UpdateCompanyDocument gère PUT /api/company/documents/:id
func (h *ProjectHandler) UpdateCompanyDocument(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Titre       string                       `json:"titre"`
		Description string                       `json:"description"`
		TypeDoc     model.TypeDocumentEntreprise `json:"type_doc"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides"})
		return
	}

	docs, err := h.projectRepo.ListCompanyDocuments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lecture documents"})
		return
	}

	var targetDoc *model.CompanyDocument
	for _, d := range docs {
		if d.ID == id {
			docCopy := d
			targetDoc = &docCopy
			break
		}
	}

	if targetDoc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document introuvable"})
		return
	}

	if req.Titre != "" {
		targetDoc.Titre = req.Titre
	}
	if req.Description != "" {
		targetDoc.Description = req.Description
	}
	if req.TypeDoc != "" {
		targetDoc.TypeDoc = req.TypeDoc
	}
	targetDoc.UpdatedAt = time.Now()

	if err := h.projectRepo.UpdateCompanyDocument(targetDoc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erreur mise à jour: %v", err)})
		return
	}

	c.JSON(http.StatusOK, targetDoc)
}

// getVaultSubfolderForDoc détermine le nom du sous-dossier thématique pour un type de document donné
func getVaultSubfolderForDoc(typeDoc model.TypeDocumentEntreprise, fileName, titre string) string {
	lower := strings.ToLower(fileName + " " + titre)
	switch typeDoc {
	case model.DocCV:
		return "02_Ressources_Humaines_et_CV"
	case model.DocReference:
		return "03_References_et_Attestations"
	case model.DocMateriel:
		return "04_Materiels_et_Equipements"
	case model.DocBilan:
		return "05_Finances_et_Banque"
	case model.DocAttestation:
		return "01_Juridique_et_Fiscal"
	default:
		if strings.Contains(lower, "cv") || strings.Contains(lower, "expert") || strings.Contains(lower, "curriculum") || strings.Contains(lower, "ingenieur") {
			return "02_Ressources_Humaines_et_CV"
		}
		if strings.Contains(lower, "reference") || strings.Contains(lower, "bonne_fin") || strings.Contains(lower, "chantier") || strings.Contains(lower, "pv") {
			return "03_References_et_Attestations"
		}
		if strings.Contains(lower, "materiel") || strings.Contains(lower, "engin") || strings.Contains(lower, "carte_grise") || strings.Contains(lower, "camion") {
			return "04_Materiels_et_Equipements"
		}
		if strings.Contains(lower, "bilan") || strings.Contains(lower, "financier") || strings.Contains(lower, "banque") || strings.Contains(lower, "credit") {
			return "05_Finances_et_Banque"
		}
		return "01_Juridique_et_Fiscal"
	}
}

// UploadCompanyDocument gère POST /api/company/documents/upload (multipart, téléversement classé dans les sous-dossiers de 01_COFFRE_FORT_ENTREPRISE)
func (h *ProjectHandler) UploadCompanyDocument(c *gin.Context) {
	token := h.requireToken(c)
	if token == nil {
		return
	}

	profile, err := h.projectRepo.GetCompanyProfile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de charger le profil entreprise"})
		return
	}

	// S'assurer de la présence du dossier 01_COFFRE_FORT_ENTREPRISE
	if profile.RagFolderID == "" {
		wsID, ragID, projID, wsURL, wsErr := h.driveService.EnsureWorkspaceTree(c.Request.Context(), token, "")
		if wsErr != nil {
			h.handleDriveError(c, wsErr)
			return
		}
		profile.WorkspaceFolderID = wsID
		profile.WorkspaceFolderURL = wsURL
		profile.RagFolderID = ragID
		profile.ProjectsFolderID = projID
		_ = h.projectRepo.UpdateCompanyProfile(profile)
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxDaoUploadBytes+1<<20)
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Fichier manquant ou trop volumineux (max 30 Mo)"})
		return
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if ext != ".pdf" && ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".docx" && ext != ".doc" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Seuls les fichiers PDF, Word (.docx, .doc) et images (PNG, JPG) sont acceptés."})
		return
	}

	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Impossible de lire le fichier"})
		return
	}
	defer f.Close()

	mimeType := "application/pdf"
	if ext == ".png" {
		mimeType = "image/png"
	} else if ext == ".jpg" || ext == ".jpeg" {
		mimeType = "image/jpeg"
	} else if ext == ".docx" {
		mimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	} else if ext == ".doc" {
		mimeType = "application/msword"
	}

	titre := strings.TrimSpace(c.PostForm("titre"))
	if titre == "" {
		titre = strings.TrimSuffix(fh.Filename, filepath.Ext(fh.Filename))
	}
	typeDoc := strings.TrimSpace(c.PostForm("type_document"))
	if typeDoc == "" {
		typeDoc = "AUTRE"
	}
	desc := strings.TrimSpace(c.PostForm("description"))

	// Cibler automatiquement le sous-dossier thématique approprié
	targetFolderID := profile.RagFolderID
	subfolders, sErr := h.driveService.EnsureVaultSubfolders(c.Request.Context(), token, profile.RagFolderID)
	if sErr == nil && subfolders != nil {
		targetSubName := getVaultSubfolderForDoc(model.TypeDocumentEntreprise(typeDoc), fh.Filename, titre)
		if fID, exists := subfolders[targetSubName]; exists {
			targetFolderID = fID
		}
	}

	uploaded, err := h.driveService.UploadFile(c.Request.Context(), token, targetFolderID, fh.Filename, mimeType, f)
	if err != nil {
		h.handleDriveError(c, err)
		return
	}

	var dateExp *time.Time
	if expStr := strings.TrimSpace(c.PostForm("date_expiration")); expStr != "" {
		for _, layout := range []string{"2006-01-02", "02/01/2006", time.RFC3339} {
			if parsed, pErr := time.Parse(layout, expStr); pErr == nil {
				dateExp = &parsed
				break
			}
		}
	}

	doc := &model.CompanyDocument{
		ID:             fmt.Sprintf("doc-%d", time.Now().UnixNano()),
		ProfileID:      profile.ID,
		TypeDoc:        model.TypeDocumentEntreprise(typeDoc),
		Titre:          titre,
		Description:    desc,
		GoogleDriveID:  uploaded.Id,
		WebViewLink:    uploaded.WebViewLink,
		FileName:       fh.Filename,
		DateExpiration: dateExp,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Analyse sémantique immédiate si c'est un CV
	if doc.TypeDoc == model.DocCV && uploaded.Id != "" {
		if parsed, parseErr := h.parseCVContent(c.Request.Context(), token, uploaded.Id, fh.Filename); parseErr == nil && parsed != nil {
			if parsed.NomComplet != "" {
				doc.Titre = fmt.Sprintf("CV %s", parsed.NomComplet)
			}
			doc.Description = fmt.Sprintf("[%s] %s (%d ans d'exp.) - %s", parsed.SpecialitePrincipale, parsed.NomComplet, parsed.ExperienceAnnees, parsed.ResumeAnalytique)
			if metaBytes, mErr := json.Marshal(parsed); mErr == nil {
				doc.MetadonneesJSON = string(metaBytes)
			}
		}
	}

	if err := h.projectRepo.AddCompanyDocument(doc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erreur enregistrement métadonnées: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

// SyncCompanyDriveDocuments gère POST /api/company/sync-drive (scan récursif des sous-dossiers thématiques)
func (h *ProjectHandler) SyncCompanyDriveDocuments(c *gin.Context) {
	token := h.requireToken(c)
	if token == nil {
		return
	}

	profile, err := h.projectRepo.GetCompanyProfile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de charger le profil entreprise"})
		return
	}

	if profile.RagFolderID == "" {
		wsID, ragID, projID, wsURL, wsErr := h.driveService.EnsureWorkspaceTree(c.Request.Context(), token, "")
		if wsErr != nil {
			h.handleDriveError(c, wsErr)
			return
		}
		profile.WorkspaceFolderID = wsID
		profile.WorkspaceFolderURL = wsURL
		profile.RagFolderID = ragID
		profile.ProjectsFolderID = projID
		_ = h.projectRepo.UpdateCompanyProfile(profile)
	}

	// S'assurer que les 5 sous-dossiers thématiques existent
	subfolders, err := h.driveService.EnsureVaultSubfolders(c.Request.Context(), token, profile.RagFolderID)
	if err != nil {
		h.handleDriveError(c, err)
		return
	}

	existingDocs, _ := h.projectRepo.ListCompanyDocuments()
	existingDriveIDs := make(map[string]bool)
	for _, ed := range existingDocs {
		if ed.GoogleDriveID != "" {
			existingDriveIDs[ed.GoogleDriveID] = true
		}
	}

	importedCount := 0

	// 1. Scanner les fichiers à la racine de 01_COFFRE_FORT_ENTREPRISE pour les classer dans les sous-dossiers
	rootFiles, _ := h.driveService.ListFolderFiles(c.Request.Context(), token, profile.RagFolderID)
	for _, f := range rootFiles {
		if f.MimeType == "application/vnd.google-apps.folder" {
			continue
		}

		typeDoc := deduceDocTypeFromName(f.Name)
		targetSubName := getVaultSubfolderForDoc(typeDoc, f.Name, f.Name)
		if targetFolderID, ok := subfolders[targetSubName]; ok {
			_ = h.driveService.MoveFile(c.Request.Context(), token, f.Id, targetFolderID)
		}

		if existingDriveIDs[f.Id] {
			continue
		}

		doc := &model.CompanyDocument{
			ID:            fmt.Sprintf("doc-%d", time.Now().UnixNano()),
			ProfileID:     profile.ID,
			TypeDoc:       typeDoc,
			Titre:         strings.TrimSuffix(f.Name, filepath.Ext(f.Name)),
			Description:   fmt.Sprintf("Document synchronisé depuis Google Drive (%s)", f.Name),
			GoogleDriveID: f.Id,
			WebViewLink:   f.WebViewLink,
			FileName:      f.Name,
		}

		if typeDoc == model.DocCV && f.Id != "" {
			if parsed, parseErr := h.parseCVContent(c.Request.Context(), token, f.Id, f.Name); parseErr == nil && parsed != nil {
				if parsed.NomComplet != "" {
					doc.Titre = fmt.Sprintf("CV %s", parsed.NomComplet)
				}
				doc.Description = fmt.Sprintf("[%s] %s (%d ans d'exp.) - %s", parsed.SpecialitePrincipale, parsed.NomComplet, parsed.ExperienceAnnees, parsed.ResumeAnalytique)
				if metaBytes, mErr := json.Marshal(parsed); mErr == nil {
					doc.MetadonneesJSON = string(metaBytes)
				}
			}
		} else if typeDoc == model.DocAttestation && f.Id != "" {
			if parsed, parseErr := h.parseAdminDocContent(c.Request.Context(), token, f.Id, f.Name); parseErr == nil && parsed != nil {
				if parsed.IntituleOfficiel != "" {
					doc.Titre = parsed.IntituleOfficiel
				}
				doc.Description = fmt.Sprintf("[%s] %s - %s", parsed.OrganismeEmetteur, parsed.IntituleOfficiel, parsed.Commentaire)
				if parsed.DateExpiration != "" {
					for _, layout := range []string{"2006-01-02", "02/01/2006", time.RFC3339} {
						if exp, pErr := time.Parse(layout, parsed.DateExpiration); pErr == nil {
							doc.DateExpiration = &exp
							break
						}
					}
				}
				if metaBytes, mErr := json.Marshal(parsed); mErr == nil {
					doc.MetadonneesJSON = string(metaBytes)
				}
			}
		}

		if err := h.projectRepo.AddCompanyDocument(doc); err == nil {
			importedCount++
			existingDriveIDs[f.Id] = true
		}
	}

	// 2. Scanner chaque sous-dossier thématique
	subfolderTypeMap := map[string]model.TypeDocumentEntreprise{
		"01_Juridique_et_Fiscal":       model.DocAttestation,
		"02_Ressources_Humaines_et_CV": model.DocCV,
		"03_References_et_Attestations": model.DocReference,
		"04_Materiels_et_Equipements":  model.DocMateriel,
		"05_Finances_et_Banque":        model.DocBilan,
	}

	for subName, subFolderID := range subfolders {
		defaultType := subfolderTypeMap[subName]
		files, fErr := h.driveService.ListFolderFiles(c.Request.Context(), token, subFolderID)
		if fErr != nil {
			continue
		}

		for _, f := range files {
			if f.MimeType == "application/vnd.google-apps.folder" {
				continue
			}
			if existingDriveIDs[f.Id] {
				continue
			}

			typeDoc := defaultType
			lowerName := strings.ToLower(f.Name)
			if subName == "01_Juridique_et_Fiscal" {
				if strings.Contains(lowerName, "ninea") || strings.Contains(lowerName, "rccm") || strings.Contains(lowerName, "statut") {
					typeDoc = model.DocAutre
				}
			}

			doc := &model.CompanyDocument{
				ID:            fmt.Sprintf("doc-%d", time.Now().UnixNano()),
				ProfileID:     profile.ID,
				TypeDoc:       typeDoc,
				Titre:         strings.TrimSuffix(f.Name, filepath.Ext(f.Name)),
				Description:   fmt.Sprintf("Document synchronisé depuis %s (%s)", subName, f.Name),
				GoogleDriveID: f.Id,
				WebViewLink:   f.WebViewLink,
				FileName:      f.Name,
			}

			if typeDoc == model.DocCV && f.Id != "" {
				if parsed, parseErr := h.parseCVContent(c.Request.Context(), token, f.Id, f.Name); parseErr == nil && parsed != nil {
					if parsed.NomComplet != "" {
						doc.Titre = fmt.Sprintf("CV %s", parsed.NomComplet)
					}
					doc.Description = fmt.Sprintf("[%s] %s (%d ans d'exp.) - %s", parsed.SpecialitePrincipale, parsed.NomComplet, parsed.ExperienceAnnees, parsed.ResumeAnalytique)
					if metaBytes, mErr := json.Marshal(parsed); mErr == nil {
						doc.MetadonneesJSON = string(metaBytes)
					}
				}
			} else if (subName == "01_Juridique_et_Fiscal" || typeDoc == model.DocAttestation) && f.Id != "" {
				if parsed, parseErr := h.parseAdminDocContent(c.Request.Context(), token, f.Id, f.Name); parseErr == nil && parsed != nil {
					if parsed.IntituleOfficiel != "" {
						doc.Titre = parsed.IntituleOfficiel
					}
					doc.Description = fmt.Sprintf("[%s] %s - %s", parsed.OrganismeEmetteur, parsed.IntituleOfficiel, parsed.Commentaire)
					if parsed.DateExpiration != "" {
						for _, layout := range []string{"2006-01-02", "02/01/2006", time.RFC3339} {
							if exp, pErr := time.Parse(layout, parsed.DateExpiration); pErr == nil {
								doc.DateExpiration = &exp
								break
							}
						}
					}
					if metaBytes, mErr := json.Marshal(parsed); mErr == nil {
						doc.MetadonneesJSON = string(metaBytes)
					}
				}
			}

			if err := h.projectRepo.AddCompanyDocument(doc); err == nil {
				importedCount++
				existingDriveIDs[f.Id] = true
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        fmt.Sprintf("%d nouveau(x) document(s) synchronisé(s) et classé(s) dans le Coffre-fort Drive.", importedCount),
		"imported_count": importedCount,
	})
}

// deduceDocTypeFromName devine le type de document d'après le nom de fichier
func deduceDocTypeFromName(name string) model.TypeDocumentEntreprise {
	lowerName := strings.ToLower(name)
	if strings.Contains(lowerName, "cv") || strings.Contains(lowerName, "expert") || strings.Contains(lowerName, "curriculum") || strings.Contains(lowerName, "ingenieur") {
		return model.DocCV
	} else if strings.Contains(lowerName, "quitus") || strings.Contains(lowerName, "fiscal") || strings.Contains(lowerName, "css") || strings.Contains(lowerName, "ipres") || strings.Contains(lowerName, "social") || strings.Contains(lowerName, "dgid") || strings.Contains(lowerName, "arcop") || strings.Contains(lowerName, "armp") || strings.Contains(lowerName, "travail") || strings.Contains(lowerName, "inspection") || strings.Contains(lowerName, "it ") || strings.HasPrefix(lowerName, "it") {
		return model.DocAttestation
	} else if strings.Contains(lowerName, "reference") || strings.Contains(lowerName, "bonne_fin") || strings.Contains(lowerName, "attestation") || strings.Contains(lowerName, "chantier") {
		return model.DocReference
	} else if strings.Contains(lowerName, "materiel") || strings.Contains(lowerName, "carte_grise") || strings.Contains(lowerName, "engin") || strings.Contains(lowerName, "camion") {
		return model.DocMateriel
	} else if strings.Contains(lowerName, "bilan") || strings.Contains(lowerName, "financier") || strings.Contains(lowerName, "banque") || strings.Contains(lowerName, "credit") {
		return model.DocBilan
	}
	return model.DocAutre
}

// OrganizeCompanyVault gère POST /api/company/organize-drive (range tous les fichiers existants dans les sous-dossiers thématiques)
func (h *ProjectHandler) OrganizeCompanyVault(c *gin.Context) {
	token := h.requireToken(c)
	if token == nil {
		return
	}

	profile, err := h.projectRepo.GetCompanyProfile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de charger le profil entreprise"})
		return
	}

	if profile.RagFolderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Le coffre-fort d'entreprise n'est pas encore initialisé sur Google Drive."})
		return
	}

	subfolders, err := h.driveService.EnsureVaultSubfolders(c.Request.Context(), token, profile.RagFolderID)
	if err != nil {
		h.handleDriveError(c, err)
		return
	}

	docs, err := h.projectRepo.ListCompanyDocuments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lecture des documents"})
		return
	}

	movedCount := 0
	subfolderCounts := make(map[string]int)

	for _, doc := range docs {
		if doc.GoogleDriveID == "" {
			continue
		}

		targetSubName := getVaultSubfolderForDoc(doc.TypeDoc, doc.FileName, doc.Titre)
		targetFolderID, exists := subfolders[targetSubName]
		if !exists {
			continue
		}

		if err := h.driveService.MoveFile(c.Request.Context(), token, doc.GoogleDriveID, targetFolderID); err == nil {
			movedCount++
			subfolderCounts[targetSubName]++
		}
	}

	// Déplacer également les fichiers orphelins restant à la racine de 01_COFFRE_FORT_ENTREPRISE
	rootFiles, _ := h.driveService.ListFolderFiles(c.Request.Context(), token, profile.RagFolderID)
	for _, f := range rootFiles {
		if f.MimeType == "application/vnd.google-apps.folder" {
			continue
		}
		typeDoc := deduceDocTypeFromName(f.Name)
		targetSubName := getVaultSubfolderForDoc(typeDoc, f.Name, f.Name)
		if targetFolderID, ok := subfolders[targetSubName]; ok {
			_ = h.driveService.MoveFile(c.Request.Context(), token, f.Id, targetFolderID)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         fmt.Sprintf("Coffre-fort réorganisé avec succès sur Google Drive (%d fichiers classés dans leurs sous-dossiers respectifs).", movedCount),
		"organized_count": movedCount,
		"subfolders":      subfolders,
		"counts_by_sub":   subfolderCounts,
	})
}

// parseCVContent télécharge le fichier depuis Google Drive et extrait le profil sémantique via Gemini
func (h *ProjectHandler) parseCVContent(ctx context.Context, token *oauth2.Token, driveID string, fileName string) (*service.ExtractedCVProfile, error) {
	bytes, mime, err := h.driveService.DownloadFile(ctx, token, driveID)
	if err != nil {
		return nil, fmt.Errorf("téléchargement Drive impossible: %w", err)
	}
	if len(bytes) == 0 {
		return nil, fmt.Errorf("fichier vide")
	}

	doc := service.Document{
		Name:     fileName,
		Bytes:    bytes,
		MimeType: mime,
	}

	instruction := "Analyse rigoureusement ce CV. Extrais le nom complet, la spécialité principale (ex: Environnementaliste/HSE, Ingénieur Génie Civil, Géomètre-Topographe, etc.), le titre professionnel, les années d'expérience, les diplômes, les compétences clés et un résumé analytique percutant."

	rawJSON, err := h.agentService.ExecuteAgentTask(ctx, service.CVParserAgent, []service.Document{doc}, instruction)
	if err != nil {
		return nil, fmt.Errorf("erreur analyse Gemini: %w", err)
	}

	var parsed service.ExtractedCVProfile
	if err := json.Unmarshal([]byte(rawJSON), &parsed); err != nil {
		return nil, fmt.Errorf("décodage extraction CV impossible: %w", err)
	}

	return &parsed, nil
}

// parseAdminDocContent télécharge le fichier (scan JPG, PNG ou PDF) depuis Google Drive et extrait les informations officielles et la date de validité via Gemini Vision
func (h *ProjectHandler) parseAdminDocContent(ctx context.Context, token *oauth2.Token, driveID string, fileName string) (*service.ExtractedAdminDocProfile, error) {
	bytes, mime, err := h.driveService.DownloadFile(ctx, token, driveID)
	if err != nil {
		return nil, fmt.Errorf("téléchargement Drive impossible: %w", err)
	}
	if len(bytes) == 0 {
		return nil, fmt.Errorf("fichier vide")
	}

	doc := service.Document{
		Name:     fileName,
		Bytes:    bytes,
		MimeType: mime,
	}

	instruction := "Analyse rigoureusement ce scan officiel de pièce administrative / fiscale. Identifie la nature du document (QUITUS_FISCAL, IPRES, CSS, INSPECTION_TRAVAIL), son intitulé officiel précis, son organisme émetteur, et SURTOUT la date d'expiration ou de fin de validité au format strict AAAA-MM-JJ."

	rawJSON, err := h.agentService.ExecuteAgentTask(ctx, service.AdminDocParserAgent, []service.Document{doc}, instruction)
	if err != nil {
		return nil, fmt.Errorf("erreur analyse Gemini: %w", err)
	}

	var parsed service.ExtractedAdminDocProfile
	if err := json.Unmarshal([]byte(rawJSON), &parsed); err != nil {
		return nil, fmt.Errorf("décodage extraction administrative impossible: %w", err)
	}

	return &parsed, nil
}

// AnalyzeCompanyCV gère POST /api/company/documents/:id/analyze (analyse intelligente par IA d'un document: CV ou Scan administratif)
func (h *ProjectHandler) AnalyzeCompanyCV(c *gin.Context) {
	token := h.requireToken(c)
	if token == nil {
		return
	}
	id := c.Param("id")

	docs, err := h.projectRepo.ListCompanyDocuments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lecture documents"})
		return
	}

	var targetDoc *model.CompanyDocument
	for _, d := range docs {
		if d.ID == id {
			docCopy := d
			targetDoc = &docCopy
			break
		}
	}

	if targetDoc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document introuvable"})
		return
	}

	if targetDoc.GoogleDriveID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ce document n'a pas d'identifiant Google Drive"})
		return
	}

	lower := strings.ToLower(targetDoc.FileName + " " + targetDoc.Titre)
	isCV := targetDoc.TypeDoc == model.DocCV || strings.Contains(lower, "cv") || strings.Contains(lower, "expert") || strings.Contains(lower, "curriculum")

	if isCV {
		parsed, err := h.parseCVContent(c.Request.Context(), token, targetDoc.GoogleDriveID, targetDoc.FileName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erreur analyse CV: %v", err)})
			return
		}

		if parsed.NomComplet != "" {
			targetDoc.Titre = fmt.Sprintf("CV %s", parsed.NomComplet)
		}
		targetDoc.TypeDoc = model.DocCV
		targetDoc.Description = fmt.Sprintf("[%s] %s (%d ans d'exp.) - %s", parsed.SpecialitePrincipale, parsed.NomComplet, parsed.ExperienceAnnees, parsed.ResumeAnalytique)
		if metaBytes, err := json.Marshal(parsed); err == nil {
			targetDoc.MetadonneesJSON = string(metaBytes)
		}
		targetDoc.UpdatedAt = time.Now()

		if err := h.projectRepo.UpdateCompanyDocument(targetDoc); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erreur sauvegarde document: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  fmt.Sprintf("CV de %s analysé et indexé avec succès (%s)", parsed.NomComplet, parsed.SpecialitePrincipale),
			"document": targetDoc,
			"profile":  parsed,
		})
		return
	}

	// Analyse scan administratif (Quitus fiscal, IPRES, CSS, IT, etc.)
	parsed, err := h.parseAdminDocContent(c.Request.Context(), token, targetDoc.GoogleDriveID, targetDoc.FileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erreur analyse scan administratif: %v", err)})
		return
	}

	if parsed.IntituleOfficiel != "" {
		targetDoc.Titre = parsed.IntituleOfficiel
	}
	targetDoc.TypeDoc = model.DocAttestation
	if parsed.TypeDocument == "NINEA" || parsed.TypeDocument == "RCCM" {
		targetDoc.TypeDoc = model.DocAutre
	}
	targetDoc.Description = fmt.Sprintf("[%s] %s - %s", parsed.OrganismeEmetteur, parsed.IntituleOfficiel, parsed.Commentaire)
	if parsed.DateExpiration != "" {
		for _, layout := range []string{"2006-01-02", "02/01/2006", time.RFC3339} {
			if exp, pErr := time.Parse(layout, parsed.DateExpiration); pErr == nil {
				targetDoc.DateExpiration = &exp
				break
			}
		}
	}
	if metaBytes, err := json.Marshal(parsed); err == nil {
		targetDoc.MetadonneesJSON = string(metaBytes)
	}
	targetDoc.UpdatedAt = time.Now()

	if err := h.projectRepo.UpdateCompanyDocument(targetDoc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erreur sauvegarde document: %v", err)})
		return
	}

	expText := "Non précisée"
	if targetDoc.DateExpiration != nil {
		expText = targetDoc.DateExpiration.Format("02/01/2006")
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("%s analysé avec succès (Expire le : %s)", targetDoc.Titre, expText),
		"document": targetDoc,
		"profile":  parsed,
	})
}

// AnalyzeAllCompanyAdminDocs gère POST /api/company/analyze-all-admin (analyse tous les scans administratifs)
func (h *ProjectHandler) AnalyzeAllCompanyAdminDocs(c *gin.Context) {
	token := h.requireToken(c)
	if token == nil {
		return
	}

	docs, err := h.projectRepo.ListCompanyDocuments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lecture documents"})
		return
	}

	analyzedCount := 0
	var results []gin.H

	for _, d := range docs {
		lower := strings.ToLower(d.FileName + " " + d.Titre)
		isCV := d.TypeDoc == model.DocCV || strings.Contains(lower, "cv")
		if isCV || d.GoogleDriveID == "" {
			continue
		}

		parsed, err := h.parseAdminDocContent(c.Request.Context(), token, d.GoogleDriveID, d.FileName)
		if err != nil {
			fmt.Printf("Erreur analyse scan admin %s: %v\n", d.FileName, err)
			continue
		}

		docCopy := d
		if parsed.IntituleOfficiel != "" {
			docCopy.Titre = parsed.IntituleOfficiel
		}
		docCopy.TypeDoc = model.DocAttestation
		if parsed.TypeDocument == "NINEA" || parsed.TypeDocument == "RCCM" {
			docCopy.TypeDoc = model.DocAutre
		}
		docCopy.Description = fmt.Sprintf("[%s] %s - %s", parsed.OrganismeEmetteur, parsed.IntituleOfficiel, parsed.Commentaire)
		if parsed.DateExpiration != "" {
			for _, layout := range []string{"2006-01-02", "02/01/2006", time.RFC3339} {
				if exp, pErr := time.Parse(layout, parsed.DateExpiration); pErr == nil {
					docCopy.DateExpiration = &exp
					break
				}
			}
		}
		if metaBytes, err := json.Marshal(parsed); err == nil {
			docCopy.MetadonneesJSON = string(metaBytes)
		}
		docCopy.UpdatedAt = time.Now()

		if err := h.projectRepo.UpdateCompanyDocument(&docCopy); err == nil {
			analyzedCount++
			expFormatted := "Inconnue"
			if docCopy.DateExpiration != nil {
				expFormatted = docCopy.DateExpiration.Format("02/01/2006")
			}
			results = append(results, gin.H{
				"id":         docCopy.ID,
				"titre":      docCopy.Titre,
				"organisme":  parsed.OrganismeEmetteur,
				"expiration": expFormatted,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        fmt.Sprintf("%d document(s) administratif(s) analysé(s) avec extraction de validité", analyzedCount),
		"analyzed_count": analyzedCount,
		"results":        results,
	})
}

// DeleteCompanyDocument gère DELETE /api/company/documents/:id
func (h *ProjectHandler) DeleteCompanyDocument(c *gin.Context) {
	id := c.Param("id")
	if err := h.projectRepo.DeleteCompanyDocument(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erreur suppression: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Document supprimé avec succès."})
}

// AnalyzeAllCompanyVault gère POST /api/company/analyze-all (analyse l'ensemble des pièces du coffre-fort : CVs et Scans administratifs)
func (h *ProjectHandler) AnalyzeAllCompanyVault(c *gin.Context) {
	token := h.requireToken(c)
	if token == nil {
		return
	}

	docs, err := h.projectRepo.ListCompanyDocuments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lecture documents"})
		return
	}

	analyzedCVCount := 0
	analyzedAdminCount := 0
	var results []gin.H

	for _, d := range docs {
		if d.GoogleDriveID == "" {
			continue
		}

		lowerTitre := strings.ToLower(d.Titre)
		lowerFile := strings.ToLower(d.FileName)
		isCV := d.TypeDoc == model.DocCV || strings.Contains(lowerTitre, "cv") || strings.Contains(lowerFile, "cv") || strings.Contains(lowerFile, "resume")

		if isCV {
			parsed, err := h.parseCVContent(c.Request.Context(), token, d.GoogleDriveID, d.FileName)
			if err != nil {
				fmt.Printf("Erreur analyse CV %s: %v\n", d.FileName, err)
				continue
			}

			docCopy := d
			if parsed.NomComplet != "" {
				docCopy.Titre = fmt.Sprintf("CV %s", parsed.NomComplet)
			}
			docCopy.TypeDoc = model.DocCV
			docCopy.Description = fmt.Sprintf("[%s] %s (%d ans d'exp.) - %s", parsed.SpecialitePrincipale, parsed.NomComplet, parsed.ExperienceAnnees, parsed.ResumeAnalytique)
			if metaBytes, err := json.Marshal(parsed); err == nil {
				docCopy.MetadonneesJSON = string(metaBytes)
			}
			docCopy.UpdatedAt = time.Now()

			if err := h.projectRepo.UpdateCompanyDocument(&docCopy); err == nil {
				analyzedCVCount++
				results = append(results, gin.H{
					"id":     docCopy.ID,
					"type":   "CV",
					"titre":  docCopy.Titre,
					"detail": parsed.SpecialitePrincipale,
				})
			}
		} else {
			parsed, err := h.parseAdminDocContent(c.Request.Context(), token, d.GoogleDriveID, d.FileName)
			if err != nil {
				fmt.Printf("Erreur analyse scan admin %s: %v\n", d.FileName, err)
				continue
			}

			docCopy := d
			if parsed.IntituleOfficiel != "" {
				docCopy.Titre = parsed.IntituleOfficiel
			}
			docCopy.TypeDoc = model.DocAttestation
			if parsed.TypeDocument == "NINEA" || parsed.TypeDocument == "RCCM" {
				docCopy.TypeDoc = model.DocAutre
			}
			docCopy.Description = fmt.Sprintf("[%s] %s - %s", parsed.OrganismeEmetteur, parsed.IntituleOfficiel, parsed.Commentaire)
			if parsed.DateExpiration != "" {
				for _, layout := range []string{"2006-01-02", "02/01/2006", time.RFC3339} {
					if exp, pErr := time.Parse(layout, parsed.DateExpiration); pErr == nil {
						docCopy.DateExpiration = &exp
						break
					}
				}
			}
			if metaBytes, err := json.Marshal(parsed); err == nil {
				docCopy.MetadonneesJSON = string(metaBytes)
			}
			docCopy.UpdatedAt = time.Now()

			if err := h.projectRepo.UpdateCompanyDocument(&docCopy); err == nil {
				analyzedAdminCount++
				expFormatted := "Non spécifiée"
				if docCopy.DateExpiration != nil {
					expFormatted = docCopy.DateExpiration.Format("02/01/2006")
				}
				results = append(results, gin.H{
					"id":     docCopy.ID,
					"type":   "ADMIN",
					"titre":  docCopy.Titre,
					"detail": fmt.Sprintf("%s (Exp: %s)", parsed.OrganismeEmetteur, expFormatted),
				})
			}
		}
	}

	totalAnalyzed := analyzedCVCount + analyzedAdminCount
	c.JSON(http.StatusOK, gin.H{
		"message":        fmt.Sprintf("%d pièce(s) analysée(s) avec succès par l'IA (%d CV(s), %d scan(s) administratif(s)).", totalAnalyzed, analyzedCVCount, analyzedAdminCount),
		"analyzed_count": totalAnalyzed,
		"cv_count":       analyzedCVCount,
		"admin_count":    analyzedAdminCount,
		"results":        results,
	})
}

// AnalyzeAllCompanyCVs gère POST /api/company/analyze-all-cvs (analyse en lot de tous les CVs)
func (h *ProjectHandler) AnalyzeAllCompanyCVs(c *gin.Context) {
	token := h.requireToken(c)
	if token == nil {
		return
	}

	docs, err := h.projectRepo.ListCompanyDocuments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lecture documents"})
		return
	}

	analyzedCount := 0
	var results []gin.H

	for _, d := range docs {
		lowerTitre := strings.ToLower(d.Titre)
		lowerFile := strings.ToLower(d.FileName)
		isCV := d.TypeDoc == model.DocCV || strings.Contains(lowerTitre, "cv") || strings.Contains(lowerFile, "cv") || strings.Contains(lowerFile, "resume")

		if !isCV || d.GoogleDriveID == "" {
			continue
		}

		parsed, err := h.parseCVContent(c.Request.Context(), token, d.GoogleDriveID, d.FileName)
		if err != nil {
			fmt.Printf("Erreur analyse CV %s: %v\n", d.FileName, err)
			continue
		}

		docCopy := d
		if parsed.NomComplet != "" {
			docCopy.Titre = fmt.Sprintf("CV %s", parsed.NomComplet)
		}
		docCopy.TypeDoc = model.DocCV
		docCopy.Description = fmt.Sprintf("[%s] %s (%d ans d'exp.) - %s", parsed.SpecialitePrincipale, parsed.NomComplet, parsed.ExperienceAnnees, parsed.ResumeAnalytique)
		if metaBytes, err := json.Marshal(parsed); err == nil {
			docCopy.MetadonneesJSON = string(metaBytes)
		}
		docCopy.UpdatedAt = time.Now()

		if err := h.projectRepo.UpdateCompanyDocument(&docCopy); err == nil {
			analyzedCount++
			results = append(results, gin.H{
				"id":         docCopy.ID,
				"nom":        parsed.NomComplet,
				"specialite": parsed.SpecialitePrincipale,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        fmt.Sprintf("%d CV(s) analysé(s) et indexé(s) avec succès par l'IA", analyzedCount),
		"analyzed_count": analyzedCount,
		"results":        results,
	})
}

// ==========================================
// TYPES & LOGIQUE DU CONTRÔLE DE CONFORMITÉ
// ==========================================

type RequirementStatus string

const (
	ReqConforme   RequirementStatus = "CONFORME"
	ReqARenouveler RequirementStatus = "A_RENOUVELER"
	ReqManquant   RequirementStatus = "MANQUANT"
)

type RequirementMatch struct {
	Requirement     string                 `json:"requirement"`
	Category        string                 `json:"category"`
	IsEliminatory   bool                   `json:"is_eliminatory"`
	Status          RequirementStatus      `json:"status"`
	MatchedDocument *model.CompanyDocument `json:"matched_document,omitempty"`
	Comment         string                 `json:"comment"`
	MatchingTaskID  string                 `json:"matching_task_id,omitempty"`
}

type ComplianceCheckResult struct {
	ProjectID             string             `json:"project_id"`
	GlobalStatus          string             `json:"global_status"` // CONFORME, ATTENTION, NON_CONFORME
	TotalRequirements     int                `json:"total_requirements"`
	MatchedCount          int                `json:"matched_count"`
	MissingCount          int                `json:"missing_count"`
	ExpiredCount          int                `json:"expired_count"`
	Requirements          []RequirementMatch `json:"requirements"`
	CompanyDocumentsCount int                `json:"company_documents_count"`
	OffreAdminFolderID    string             `json:"offre_admin_folder_id,omitempty"`
	OffreAdminFolderURL   string             `json:"offre_admin_folder_url,omitempty"`
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

func findMatchingCompanyDocument(reqText string, docs []model.CompanyDocument) *model.CompanyDocument {
	cleanedReq := cleanKeyword(reqText)

	rules := []struct {
		keys     []string
		docTypes []string
		docWords []string
	}{
		{
			keys:     []string{"ninea", "identification nationale"},
			docTypes: []string{"NINEA"},
			docWords: []string{"ninea"},
		},
		{
			keys:     []string{"rccm", "registre du commerce", "credit mobilier"},
			docTypes: []string{"RCCM"},
			docWords: []string{"rccm", "commerce"},
		},
		{
			keys:     []string{"fiscal", "fiscale", "dgid", "impot", "impots", "regularite fiscale", "quitus fiscal", "recouvrement fiscal"},
			docTypes: []string{"QUITUS_FISCAL", "QUITUS", "DGID"},
			docWords: []string{"fiscal", "dgid", "impot"},
		},
		{
			keys:     []string{"css", "ipres", "caisse de securite", "securite sociale", "regularite sociale", "prevoyance retraite", "cotisation"},
			docTypes: []string{"ATTESTATION_CSS_IPRES", "CSS", "IPRES"},
			docWords: []string{"css", "ipres", "sociale", "retraite", "cotisation"},
		},
		{
			keys:     []string{"inspection du travail", "inspecteur du travail"},
			docTypes: []string{"INSPECTION_TRAVAIL"},
			docWords: []string{"inspection", "travail"},
		},
		{
			keys:     []string{"arcop", "redevance de regulation", "redevance arcop"},
			docTypes: []string{"ARCOP"},
			docWords: []string{"arcop", "redevance"},
		},
		{
			keys:     []string{"agrement", "qualification", "classification btp", "certificat de qualification", "classement btp"},
			docTypes: []string{"AGREMENT_TECHNIQUE", "AGREMENT", "QUALIFICATION"},
			docWords: []string{"agrement", "qualification", "classification", "classement"},
		},
		{
			keys:     []string{"faillite", "non faillite", "liquidation", "declaration sur l honneur", "charte", "transparence"},
			docTypes: []string{"DECLARATION_HONNEUR", "NON_FAILLITE", "CHARTE"},
			docWords: []string{"faillite", "honneur", "charte", "transparence"},
		},
		{
			keys:     []string{"bilan", "etats financiers", "chiffre d affaires", "chiffres d affaires"},
			docTypes: []string{"ETATS_FINANCIERS", "BILAN", "CHIFFRE_AFFAIRES"},
			docWords: []string{"bilan", "financier", "chiffre"},
		},
		{
			keys:     []string{"ligne de credit", "capacite financiere", "attestation bancaire", "banque", "caution", "garantie bancaire"},
			docTypes: []string{"ATTESTATION_BANCAIRE", "LIGNE_CREDIT", "BANQUE"},
			docWords: []string{"banque", "credit", "bancaire", "caution"},
		},
		{
			keys:     []string{"bonne execution", "bonne fin", "reference", "service fait", "attestation de travaux"},
			docTypes: []string{"ATTESTATION_BONNE_EXECUTION", "REFERENCE"},
			docWords: []string{"execution", "reference", "bonne fin"},
		},
		{
			keys:     []string{"cv", "curriculum", "diplome", "personnel cle", "conducteur", "ingenieur", "technicien"},
			docTypes: []string{"CV", "DIPLOME", "PERSONNEL"},
			docWords: []string{"cv", "curriculum", "diplome", "personnel"},
		},
	}

	for _, rule := range rules {
		matchedKey := false
		for _, k := range rule.keys {
			if strings.Contains(cleanedReq, k) {
				matchedKey = true
				break
			}
		}
		if matchedKey {
			for i := range docs {
				doc := &docs[i]
				cleanDocTitle := cleanKeyword(doc.Titre)
				cleanDocFile := cleanKeyword(doc.FileName)
				cleanDocType := cleanKeyword(string(doc.TypeDoc))

				for _, dt := range rule.docTypes {
					if strings.EqualFold(string(doc.TypeDoc), dt) {
						return doc
					}
				}
				for _, dw := range rule.docWords {
					if strings.Contains(cleanDocTitle, dw) || strings.Contains(cleanDocFile, dw) || strings.Contains(cleanDocType, dw) {
						return doc
					}
				}
			}
		}
	}

	// Deuxième passe : intersection des mots significatifs (>= 4 lettres, hors mots génériques)
	stopWords := map[string]bool{
		"attestation": true, "attestations": true, "certificat": true, "certificats": true,
		"justificatif": true, "justificatifs": true, "document": true, "documents": true,
		"lettre": true, "piece": true, "pieces": true, "dossier": true, "absence": true,
		"fournir": true, "pour": true, "dans": true, "avec": true, "sans": true,
		"candidat": true, "soumission": true, "offre": true, "marche": true,
	}

	reqWords := strings.Fields(cleanedReq)
	for i := range docs {
		doc := &docs[i]
		cleanDocTitle := cleanKeyword(doc.Titre)
		cleanDocFile := cleanKeyword(doc.FileName)
		for _, w := range reqWords {
			if len(w) >= 4 && !stopWords[w] && (strings.Contains(cleanDocTitle, w) || strings.Contains(cleanDocFile, w)) {
				return doc
			}
		}
	}

	return nil
}

func (h *ProjectHandler) buildComplianceCheck(project *model.Project, companyDocs []model.CompanyDocument) *ComplianceCheckResult {
	result := &ComplianceCheckResult{
		ProjectID:             project.ID,
		GlobalStatus:          "CONFORME",
		CompanyDocumentsCount: len(companyDocs),
		Requirements:          make([]RequirementMatch, 0),
	}

	// Rechercher le dossier de l'Offre Administrative sur Drive
	for _, f := range project.Folders {
		if strings.Contains(f.Nom, "02_Offre_Administrative") || f.Categorie == "02_Offre_Administrative" {
			result.OffreAdminFolderID = f.GoogleFolderID
			result.OffreAdminFolderURL = fmt.Sprintf("https://drive.google.com/drive/folders/%s", f.GoogleFolderID)
			break
		}
	}

	type reqItem struct {
		Text          string
		IsEliminatory bool
		Category      string
	}
	var rawItems []reqItem

	// 1. Parser AnalyseResultJSON si présent
	if project.AnalyseResultJSON != "" {
		var ca service.ConsolidatedAnalysis
		if err := json.Unmarshal([]byte(project.AnalyseResultJSON), &ca); err == nil && ca.Administrative != nil {
			for _, crit := range ca.Administrative.CriteresEliminatoires {
				crit = strings.TrimSpace(crit)
				if crit != "" {
					rawItems = append(rawItems, reqItem{Text: crit, IsEliminatory: true, Category: "CRITERE_ELIMINATOIRE"})
				}
			}
			for _, piece := range ca.Administrative.PiecesAdministratives {
				piece = strings.TrimSpace(piece)
				if piece != "" {
					rawItems = append(rawItems, reqItem{Text: piece, IsEliminatory: false, Category: "ADMINISTRATIF"})
				}
			}
		}
	}

	// 2. Si aucune pièce extraite, utiliser les tâches de la catégorie ADMINISTRATIF
	if len(rawItems) == 0 {
		for _, t := range project.Tasks {
			if t.Categorie == "ADMINISTRATIF" {
				titre := strings.TrimPrefix(t.Titre, "Fournir : ")
				rawItems = append(rawItems, reqItem{Text: titre, IsEliminatory: false, Category: "ADMINISTRATIF"})
			}
		}
	}

	// Dédoublonnage des exigences
	seen := make(map[string]bool)
	var uniqueItems []reqItem
	for _, it := range rawItems {
		clean := cleanKeyword(it.Text)
		if !seen[clean] {
			seen[clean] = true
			uniqueItems = append(uniqueItems, it)
		}
	}

	now := time.Now()
	for _, it := range uniqueItems {
		match := RequirementMatch{
			Requirement:   it.Text,
			Category:      it.Category,
			IsEliminatory: it.IsEliminatory,
		}

		// Trouver la tâche correspondante si elle existe
		cleanReq := cleanKeyword(it.Text)
		for _, t := range project.Tasks {
			cleanTask := cleanKeyword(t.Titre)
			if strings.Contains(cleanTask, cleanReq) || strings.Contains(cleanReq, cleanTask) {
				match.MatchingTaskID = t.ID
				break
			}
		}

		doc := findMatchingCompanyDocument(it.Text, companyDocs)
		if doc == nil {
			match.Status = ReqManquant
			match.Comment = "Pièce absente du coffre-fort de l'entreprise"
			result.MissingCount++
			if it.IsEliminatory {
				result.GlobalStatus = "NON_CONFORME"
			} else if result.GlobalStatus != "NON_CONFORME" {
				result.GlobalStatus = "ATTENTION"
			}
		} else {
			match.MatchedDocument = doc
			if doc.DateExpiration != nil {
				if doc.DateExpiration.Before(now) {
					match.Status = ReqARenouveler
					match.Comment = fmt.Sprintf("Expiré le %s", doc.DateExpiration.Format("02/01/2006"))
					result.ExpiredCount++
					if it.IsEliminatory {
						result.GlobalStatus = "NON_CONFORME"
					} else if result.GlobalStatus != "NON_CONFORME" {
						result.GlobalStatus = "ATTENTION"
					}
				} else if project.DateLimiteDepot != nil && doc.DateExpiration.Before(*project.DateLimiteDepot) {
					match.Status = ReqARenouveler
					match.Comment = fmt.Sprintf("Expire le %s (avant la date limite DAO : %s)",
						doc.DateExpiration.Format("02/01/2006"), project.DateLimiteDepot.Format("02/01/2006"))
					result.ExpiredCount++
					if it.IsEliminatory {
						result.GlobalStatus = "NON_CONFORME"
					} else if result.GlobalStatus != "NON_CONFORME" {
						result.GlobalStatus = "ATTENTION"
					}
				} else {
					match.Status = ReqConforme
					match.Comment = fmt.Sprintf("Valide jusqu'au %s", doc.DateExpiration.Format("02/01/2006"))
					result.MatchedCount++
				}
			} else {
				match.Status = ReqConforme
				match.Comment = "Document permanent valide"
				result.MatchedCount++
			}
		}

		result.Requirements = append(result.Requirements, match)
	}

	result.TotalRequirements = len(result.Requirements)
	return result
}

// CheckProjectCompliance gère GET /api/projects/:id/compliance-check
func (h *ProjectHandler) CheckProjectCompliance(c *gin.Context) {
	project, err := h.projectRepo.GetProject(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Projet introuvable"})
		return
	}

	docs, err := h.projectRepo.ListCompanyDocuments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de charger les documents du coffre-fort"})
		return
	}

	compliance := h.buildComplianceCheck(project, docs)
	c.JSON(http.StatusOK, compliance)
}

// CopyRagDocumentsToProject gère POST /api/projects/:id/copy-rag-documents
func (h *ProjectHandler) CopyRagDocumentsToProject(c *gin.Context) {
	token := h.requireToken(c)
	if token == nil {
		return
	}

	project, err := h.projectRepo.GetProject(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Projet introuvable"})
		return
	}

	docs, err := h.projectRepo.ListCompanyDocuments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de charger les documents de l'entreprise"})
		return
	}

	compliance := h.buildComplianceCheck(project, docs)
	targetFolderID := compliance.OffreAdminFolderID

	// Si le dossier n'a pas été trouvé dans les métadonnées, le rechercher ou le créer sous le dossier racine du projet
	if targetFolderID == "" && project.DriveFolderID != "" {
		targetFolder, fErr := h.driveService.CreateFolder(c.Request.Context(), token, "02_Offre_Administrative", project.DriveFolderID)
		if fErr == nil {
			targetFolderID = targetFolder.Id
			compliance.OffreAdminFolderID = targetFolderID
			compliance.OffreAdminFolderURL = fmt.Sprintf("https://drive.google.com/drive/folders/%s", targetFolderID)
		}
	}

	if targetFolderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dossier 02_Offre_Administrative introuvable sur Google Drive pour ce projet."})
		return
	}

	type CopiedDocInfo struct {
		Requirement   string `json:"requirement"`
		Titre         string `json:"titre"`
		DriveID       string `json:"drive_id"`
		FileName      string `json:"file_name"`
		WebViewLink   string `json:"web_view_link"`
		UpdatedTaskID string `json:"updated_task_id,omitempty"`
	}

	var copied []CopiedDocInfo
	copiedDriveIDs := make(map[string]bool)

	for _, reqMatch := range compliance.Requirements {
		if reqMatch.MatchedDocument == nil || reqMatch.MatchedDocument.GoogleDriveID == "" {
			continue
		}
		// Ne pas recopier le même document Drive deux fois
		if copiedDriveIDs[reqMatch.MatchedDocument.GoogleDriveID] {
			continue
		}

		copiedFile, copyErr := h.driveService.CopyFile(
			c.Request.Context(),
			token,
			reqMatch.MatchedDocument.GoogleDriveID,
			targetFolderID,
			reqMatch.MatchedDocument.FileName,
		)
		if copyErr != nil {
			continue
		}

		copiedDriveIDs[reqMatch.MatchedDocument.GoogleDriveID] = true
		info := CopiedDocInfo{
			Requirement: reqMatch.Requirement,
			Titre:       reqMatch.MatchedDocument.Titre,
			DriveID:     copiedFile.Id,
			FileName:    copiedFile.Name,
			WebViewLink: copiedFile.WebViewLink,
		}

		// Si une tâche DAO correspond à cette exigence, la passer à TERMINEE
		if reqMatch.MatchingTaskID != "" {
			_ = h.projectRepo.UpdateTaskStatus(reqMatch.MatchingTaskID, model.TaskTerminee)
			info.UpdatedTaskID = reqMatch.MatchingTaskID
		}

		copied = append(copied, info)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           fmt.Sprintf("%d document(s) transféré(s) avec succès vers l'Offre Administrative sur Drive.", len(copied)),
		"copied_count":      len(copied),
		"copied_documents":  copied,
		"target_folder_id":  targetFolderID,
		"target_folder_url": compliance.OffreAdminFolderURL,
	})
}

// MoveProjectToWorkspace gère POST /api/projects/:id/move-to-workspace
func (h *ProjectHandler) MoveProjectToWorkspace(c *gin.Context) {
	token := h.requireToken(c)
	if token == nil {
		return
	}

	project, err := h.projectRepo.GetProject(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Projet introuvable"})
		return
	}

	if project.DriveFolderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ce projet n'a pas de dossier Google Drive associé."})
		return
	}

	profile, err := h.projectRepo.GetCompanyProfile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de charger le profil entreprise"})
		return
	}

	// S'assurer que l'arborescence AGENTS_DAO_WORKSPACE / 02_PROJETS_DAO existe
	if profile.ProjectsFolderID == "" {
		wsID, ragID, projID, wsURL, wsErr := h.driveService.EnsureWorkspaceTree(c.Request.Context(), token, "")
		if wsErr != nil {
			h.handleDriveError(c, wsErr)
			return
		}
		profile.WorkspaceFolderID = wsID
		profile.WorkspaceFolderURL = wsURL
		profile.RagFolderID = ragID
		profile.ProjectsFolderID = projID
		_ = h.projectRepo.UpdateCompanyProfile(profile)
	}

	// Déplacer le dossier racine du projet dans 02_PROJETS_DAO
	if err := h.driveService.MoveFolder(c.Request.Context(), token, project.DriveFolderID, profile.ProjectsFolderID); err != nil {
		h.handleDriveError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Projet classé avec succès dans AGENTS_DAO_WORKSPACE / 02_PROJETS_DAO",
		"projects_folder_id": profile.ProjectsFolderID,
		"project_id":         project.ID,
	})
}

// GetTechnicalOffer gère GET /api/projects/:id/technical-offer
func (h *ProjectHandler) GetTechnicalOffer(c *gin.Context) {
	data, err := h.technical.GetTechnicalOffer(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erreur récupération offre technique: %v", err)})
		return
	}
	c.JSON(http.StatusOK, data)
}

type GenerateSectionRequest struct {
	SectionID         string `json:"section_id" binding:"required"`
	CustomInstruction string `json:"custom_instruction"`
}

// GenerateTechnicalSection gère POST /api/projects/:id/technical-offer/generate-section
func (h *ProjectHandler) GenerateTechnicalSection(c *gin.Context) {
	var req GenerateSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de section requis"})
		return
	}

	sec, err := h.technical.GenerateSection(c.Request.Context(), c.Param("id"), req.SectionID, req.CustomInstruction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erreur génération section: %v", err)})
		return
	}

	c.JSON(http.StatusOK, sec)
}

// SaveTechnicalOffer gère POST /api/projects/:id/technical-offer/save
func (h *ProjectHandler) SaveTechnicalOffer(c *gin.Context) {
	var data model.TechnicalOfferData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données de l'offre technique invalides"})
		return
	}

	if err := h.technical.SaveTechnicalOffer(c.Request.Context(), c.Param("id"), &data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erreur sauvegarde: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Offre technique sauvegardée avec succès"})
}

// ExportTechnicalOfferToDrive gère POST /api/projects/:id/technical-offer/export-drive
func (h *ProjectHandler) ExportTechnicalOfferToDrive(c *gin.Context) {
	token := h.requireToken(c)
	if token == nil {
		return
	}

	fileID, fileURL, err := h.technical.ExportTechnicalOfferToDrive(c.Request.Context(), token, c.Param("id"))
	if err != nil {
		h.handleDriveError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Mémoire technique exporté avec succès sur Google Drive au format Word (.docx) dans 03_Offre_Technique.",
		"file_id":   fileID,
		"file_url":  fileURL,
	})
}

// CopyPersonnelCVsToProject gère POST /api/projects/:id/technical-offer/copy-cvs
func (h *ProjectHandler) CopyPersonnelCVsToProject(c *gin.Context) {
	token := h.requireToken(c)
	if token == nil {
		return
	}

	copiedCount, folderURL, err := h.technical.CopyPersonnelCVsToProject(c.Request.Context(), token, c.Param("id"))
	if err != nil {
		h.handleDriveError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      fmt.Sprintf("%d CV(s) d'expert(s) copié(s) avec succès dans 03_Offre_Technique/Personnel_Cle_et_CV sur Google Drive.", copiedCount),
		"copied_count": copiedCount,
		"folder_url":   folderURL,
	})
}

// DownloadTechnicalOfferDocx gère GET /api/projects/:id/technical-offer/download
func (h *ProjectHandler) DownloadTechnicalOfferDocx(c *gin.Context) {
	docxBytes, fileName, err := h.technical.GenerateDocx(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erreur génération document Word : %v", err)})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	c.Header("Content-Length", fmt.Sprintf("%d", len(docxBytes)))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docxBytes)
}

// ProposePersonnel gère POST /api/projects/:id/technical-offer/propose-personnel
func (h *ProjectHandler) ProposePersonnel(c *gin.Context) {
	personnel, err := h.technical.ProposePersonnel(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erreur proposition personnel : %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "Profils d'experts proposés avec succès par l'agent IA",
		"personnel": personnel,
	})
}

// ProposeMateriel gère POST /api/projects/:id/technical-offer/propose-materiel
func (h *ProjectHandler) ProposeMateriel(c *gin.Context) {
	materiel, err := h.technical.ProposeMateriel(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erreur proposition matériel : %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "Équipements et matériels proposés avec succès par l'agent IA",
		"materiel": materiel,
	})
}

// service/drive_service.go
package service

import (
	"agents-dao-platform/config"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type DriveService interface {
	ListFolderFiles(ctx context.Context, token *oauth2.Token, folderID string) ([]*drive.File, error)
	ListFolders(ctx context.Context, token *oauth2.Token, parentID string, search string) ([]*drive.File, error)
	GetFolder(ctx context.Context, token *oauth2.Token, folderID string) (*drive.File, error)
	DownloadFile(ctx context.Context, token *oauth2.Token, fileID string) ([]byte, string, error)
	CreateFolder(ctx context.Context, token *oauth2.Token, name string, parentFolderID string) (*drive.File, error)
	CreateProjectTree(ctx context.Context, token *oauth2.Token, rootFolderName string, parentID string) (string, map[string]string, error)
	UploadFile(ctx context.Context, token *oauth2.Token, parentID, name, mimeType string, content io.Reader) (*drive.File, error)
	EnsureWorkspaceTree(ctx context.Context, token *oauth2.Token, parentID string) (workspaceID, ragID, projectsID, workspaceURL string, err error)
	CopyFile(ctx context.Context, token *oauth2.Token, fileID, destinationFolderID, newName string) (*drive.File, error)
	MoveFolder(ctx context.Context, token *oauth2.Token, folderID, newParentID string) error
}

type driveService struct {
	cfg *config.AppConfig
}

func NewDriveService(cfg *config.AppConfig) DriveService {
	return &driveService{cfg: cfg}
}

// getClient crée un client Drive éphémère et sécurisé en utilisant le token de l'utilisateur connecté
func (s *driveService) getClient(ctx context.Context, token *oauth2.Token) (*drive.Service, error) {
	client := s.cfg.OAuth.Client(ctx, token)
	driveService, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("impossible de créer le client Google Drive : %v", err)
	}
	return driveService, nil
}

// ListFolderFiles liste les fichiers présents dans un dossier spécifique du Drive de l'utilisateur
func (s *driveService) ListFolderFiles(ctx context.Context, token *oauth2.Token, folderID string) ([]*drive.File, error) {
	srv, err := s.getClient(ctx, token)
	if err != nil {
		return nil, err
	}

	// Construction de la requête pour cibler un dossier et éviter les fichiers mis à la corbeille
	query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)

	r, err := srv.Files.List().
		Context(ctx).
		Q(query).
		Fields("files(id, name, mimeType, size, webViewLink, createdTime)").
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Do()
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des fichiers : %v", err)
	}

	return r.Files, nil
}

// ListFolders liste les dossiers disponibles dans Google Drive pour sélection par l'utilisateur
func (s *driveService) ListFolders(ctx context.Context, token *oauth2.Token, parentID string, search string) ([]*drive.File, error) {
	srv, err := s.getClient(ctx, token)
	if err != nil {
		return nil, err
	}

	query := "mimeType = 'application/vnd.google-apps.folder' and trashed = false"
	if search != "" {
		cleanSearch := strings.ReplaceAll(search, "'", "\\'")
		query += fmt.Sprintf(" and name contains '%s'", cleanSearch)
	} else if parentID != "" && parentID != "root" {
		query += fmt.Sprintf(" and '%s' in parents", parentID)
	} else {
		// Par défaut, lister les dossiers à la racine de Mon Drive
		query += " and 'root' in parents"
	}

	r, err := srv.Files.List().
		Context(ctx).
		Q(query).
		OrderBy("name").
		Fields("files(id, name, mimeType, webViewLink, parents)").
		PageSize(100).
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Do()
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des dossiers Drive : %v", err)
	}

	return r.Files, nil
}

// GetFolder récupère les informations d'un dossier précis par son identifiant
func (s *driveService) GetFolder(ctx context.Context, token *oauth2.Token, folderID string) (*drive.File, error) {
	srv, err := s.getClient(ctx, token)
	if err != nil {
		return nil, err
	}

	file, err := srv.Files.Get(folderID).
		SupportsAllDrives(true).
		Fields("id, name, mimeType, webViewLink, parents").
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("impossible de trouver le dossier Drive (%s) : %v", folderID, err)
	}

	if file.MimeType != "application/vnd.google-apps.folder" {
		return nil, fmt.Errorf("l'élément sélectionné n'est pas un dossier")
	}

	return file, nil
}

// DownloadFile récupère le contenu binaire d'un fichier et retourne ses octets ainsi que son type MIME réels
func (s *driveService) DownloadFile(ctx context.Context, token *oauth2.Token, fileID string) ([]byte, string, error) {
	srv, err := s.getClient(ctx, token)
	if err != nil {
		return nil, "", err
	}

	// 1. Récupérer d'abord les métadonnées pour connaître le type de fichier
	fileMeta, err := srv.Files.Get(fileID).SupportsAllDrives(true).Fields("mimeType", "name").Do()
	if err != nil {
		return nil, "", fmt.Errorf("impossible de récupérer les métadonnées du fichier : %v", err)
	}

	var resp *http.Response

	// 2. Gestion différenciée : Fichier natif Google vs Fichier binaire standard (PDF, Docx...)
	if strings.HasPrefix(fileMeta.MimeType, "application/vnd.google-apps.") {
		// C'est un document Google Docs/Sheets/Slides : on l'exporte au format PDF
		exportMimeType := "application/pdf"
		res, err := srv.Files.Export(fileID, exportMimeType).Context(ctx).Download()
		if err != nil {
			return nil, "", fmt.Errorf("erreur lors de l'export du document Google : %v", err)
		}
		resp = res
		fileMeta.MimeType = exportMimeType // On met à jour le type pour Gemini
	} else {
		// Fichier binaire standard (ex: vrai PDF) : téléchargement direct du flux
		res, err := srv.Files.Get(fileID).SupportsAllDrives(true).Context(ctx).Download()
		if err != nil {
			return nil, "", fmt.Errorf("erreur lors du téléchargement du fichier : %v", err)
		}
		resp = res
	}
	defer resp.Body.Close()

	// 3. Lecture du flux pour le convertir en tableau d'octets
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("erreur lors de la lecture des données du fichier : %v", err)
	}

	return data, fileMeta.MimeType, nil
}

// UploadFile téléverse un fichier dans un dossier Drive
func (s *driveService) UploadFile(ctx context.Context, token *oauth2.Token, parentID, name, mimeType string, content io.Reader) (*drive.File, error) {
	srv, err := s.getClient(ctx, token)
	if err != nil {
		return nil, err
	}
	meta := &drive.File{Name: name, MimeType: mimeType}
	if parentID != "" {
		meta.Parents = []string{parentID}
	}
	f, err := srv.Files.Create(meta).
		SupportsAllDrives(true).
		Media(content).
		Fields("id, name, mimeType, size, webViewLink, createdTime").
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("impossible de téléverser %s : %v", name, err)
	}
	return f, nil
}

// CreateFolder crée un dossier sur Google Drive dans un dossier parent spécifié
func (s *driveService) CreateFolder(ctx context.Context, token *oauth2.Token, name string, parentFolderID string) (*drive.File, error) {
	srv, err := s.getClient(ctx, token)
	if err != nil {
		return nil, err
	}

	folderMetadata := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
	}

	if parentFolderID != "" {
		folderMetadata.Parents = []string{parentFolderID}
	}

	folder, err := srv.Files.Create(folderMetadata).
		SupportsAllDrives(true).
		Fields("id, name, webViewLink").
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("impossible de créer le dossier %s : %v", name, err)
	}

	return folder, nil
}

// CreateProjectTree génère automatiquement l'arborescence standard d'un DAO sur Google Drive
func (s *driveService) CreateProjectTree(ctx context.Context, token *oauth2.Token, rootFolderName string, parentID string) (string, map[string]string, error) {
	// 1. Création du dossier racine du projet
	root, err := s.CreateFolder(ctx, token, rootFolderName, parentID)
	if err != nil {
		return "", nil, err
	}

	subFolders := []string{
		"01_Dossier_Consultation_Original",
		"02_Offre_Administrative",
		"03_Offre_Technique",
		"03_Offre_Technique/Personnel_Cle_et_CV",
		"03_Offre_Technique/References_et_Attestations",
		"03_Offre_Technique/Memoire_Technique",
		"04_Offre_Financiere",
	}

	createdMap := make(map[string]string)
	createdMap["ROOT"] = root.Id

	for _, path := range subFolders {
		parts := strings.Split(path, "/")
		currentParent := root.Id

		if len(parts) > 1 {
			// Le parent est le premier dossier créé
			if pID, exists := createdMap[parts[0]]; exists {
				currentParent = pID
			}
		}

		folderName := parts[len(parts)-1]
		f, err := s.CreateFolder(ctx, token, folderName, currentParent)
		if err != nil {
			return root.Id, createdMap, fmt.Errorf("erreur création sous-dossier %s : %v", folderName, err)
		}
		createdMap[path] = f.Id
	}

	return root.Id, createdMap, nil
}

// EnsureWorkspaceTree crée ou retrouve le dossier racine AGENTS_DAO_WORKSPACE et ses sous-dossiers 01_COFFRE_FORT_ENTREPRISE et 02_PROJETS_DAO
func (s *driveService) EnsureWorkspaceTree(ctx context.Context, token *oauth2.Token, parentID string) (workspaceID, ragID, projectsID, workspaceURL string, err error) {
	srv, err := s.getClient(ctx, token)
	if err != nil {
		return "", "", "", "", err
	}

	workspaceName := "AGENTS_DAO_WORKSPACE"
	ragName := "01_COFFRE_FORT_ENTREPRISE"
	projectsName := "02_PROJETS_DAO"

	findFolder := func(name, parent string) (*drive.File, error) {
		q := fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and trashed = false and name = '%s'", name)
		if parent != "" && parent != "root" {
			q += fmt.Sprintf(" and '%s' in parents", parent)
		} else {
			q += " and 'root' in parents"
		}
		res, err := srv.Files.List().Context(ctx).Q(q).SupportsAllDrives(true).IncludeItemsFromAllDrives(true).Fields("files(id, name, webViewLink)").PageSize(1).Do()
		if err != nil {
			return nil, err
		}
		if len(res.Files) > 0 {
			return res.Files[0], nil
		}
		return nil, nil
	}

	// 1. Chercher ou créer AGENTS_DAO_WORKSPACE
	wsFolder, err := findFolder(workspaceName, parentID)
	if err != nil {
		return "", "", "", "", fmt.Errorf("recherche workspace: %v", err)
	}
	if wsFolder == nil {
		wsFolder, err = s.CreateFolder(ctx, token, workspaceName, parentID)
		if err != nil {
			return "", "", "", "", fmt.Errorf("création workspace: %v", err)
		}
	}
	workspaceID = wsFolder.Id
	workspaceURL = wsFolder.WebViewLink
	if workspaceURL == "" {
		workspaceURL = fmt.Sprintf("https://drive.google.com/drive/folders/%s", workspaceID)
	}

	// 2. Chercher ou créer 01_COFFRE_FORT_ENTREPRISE
	ragFolder, err := findFolder(ragName, workspaceID)
	if err != nil {
		return "", "", "", "", fmt.Errorf("recherche coffre-fort: %v", err)
	}
	if ragFolder == nil {
		ragFolder, err = s.CreateFolder(ctx, token, ragName, workspaceID)
		if err != nil {
			return "", "", "", "", fmt.Errorf("création coffre-fort: %v", err)
		}
	}
	ragID = ragFolder.Id

	// 3. Chercher ou créer 02_PROJETS_DAO
	projFolder, err := findFolder(projectsName, workspaceID)
	if err != nil {
		return "", "", "", "", fmt.Errorf("recherche projets: %v", err)
	}
	if projFolder == nil {
		projFolder, err = s.CreateFolder(ctx, token, projectsName, workspaceID)
		if err != nil {
			return "", "", "", "", fmt.Errorf("création dossier projets: %v", err)
		}
	}
	projectsID = projFolder.Id

	return workspaceID, ragID, projectsID, workspaceURL, nil
}

// CopyFile copie un fichier Google Drive vers un dossier de destination spécifié
func (s *driveService) CopyFile(ctx context.Context, token *oauth2.Token, fileID, destinationFolderID, newName string) (*drive.File, error) {
	srv, err := s.getClient(ctx, token)
	if err != nil {
		return nil, err
	}

	copyMeta := &drive.File{
		Parents: []string{destinationFolderID},
	}
	if newName != "" {
		copyMeta.Name = newName
	}

	copied, err := srv.Files.Copy(fileID, copyMeta).
		SupportsAllDrives(true).
		Fields("id, name, mimeType, size, webViewLink, createdTime").
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("impossible de copier le fichier : %v", err)
	}
	return copied, nil
}

// MoveFolder déplace un dossier Drive vers un nouveau dossier parent
func (s *driveService) MoveFolder(ctx context.Context, token *oauth2.Token, folderID, newParentID string) error {
	srv, err := s.getClient(ctx, token)
	if err != nil {
		return err
	}

	file, err := srv.Files.Get(folderID).
		SupportsAllDrives(true).
		Fields("id, parents").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("impossible d'accéder au dossier : %v", err)
	}

	// Si le dossier est déjà dans newParentID, rien à faire
	for _, p := range file.Parents {
		if p == newParentID {
			return nil
		}
	}

	var previousParents string
	if len(file.Parents) > 0 {
		previousParents = strings.Join(file.Parents, ",")
	}

	call := srv.Files.Update(folderID, nil).
		AddParents(newParentID).
		SupportsAllDrives(true).
		Fields("id, parents")

	if previousParents != "" {
		call = call.RemoveParents(previousParents)
	}

	_, err = call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("impossible de déplacer le dossier : %v", err)
	}

	return nil
}
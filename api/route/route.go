// package route
package route

import (
	"agents-dao-platform/config"
	"agents-dao-platform/handler"
	"agents-dao-platform/service"
	"agents-dao-platform/utils"

	"github.com/gin-gonic/gin"
)

// SetupRouter initialise le moteur Gin et centralise l'ensemble des routes
func SetupRouter(hm *handler.HandlerManager, sm *service.ServiceManager, cfg *config.AppConfig) *gin.Engine {
	r := gin.Default()

	// Middlewares globaux
	r.Use(utils.GinCORS(cfg.FrontendURL))
	r.Use(utils.SessionMiddleware(sm.Session))

	// Authentification Google OAuth
	r.GET("/", hm.Auth.Home)
	r.GET("/login", hm.Auth.Login)
	r.GET("/callback", hm.Auth.Callback)

	// Groupe d'API REST
	api := r.Group("/api")
	{
		// Session
		api.GET("/auth/me", hm.Auth.Me)
		api.POST("/auth/logout", hm.Auth.Logout)

		// Espace de travail global & Drive
		api.POST("/workspace/init", hm.Project.InitWorkspace)

		// Gestion des Projets / DAO
		api.GET("/projects", hm.Project.ListProjects)
		api.POST("/projects", hm.Project.CreateProject)
		api.GET("/projects/:id", hm.Project.GetProject)
		api.GET("/projects/:id/documents", hm.Project.ListDaoDocuments)
		api.POST("/projects/:id/documents", hm.Project.UploadDaoDocument)
		api.POST("/projects/:id/analyze", hm.Project.AnalyzeProject)
		api.GET("/projects/:id/compliance-check", hm.Project.CheckProjectCompliance)
		api.POST("/projects/:id/copy-rag-documents", hm.Project.CopyRagDocumentsToProject)
		api.POST("/projects/:id/move-to-workspace", hm.Project.MoveProjectToWorkspace)
		api.PATCH("/tasks/:taskId", hm.Project.UpdateTaskStatus)

		// Offre Technique & Mémoire Technique IA
		api.GET("/projects/:id/technical-offer", hm.Project.GetTechnicalOffer)
		api.GET("/projects/:id/technical-offer/download", hm.Project.DownloadTechnicalOfferDocx)
		api.POST("/projects/:id/technical-offer/generate-section", hm.Project.GenerateTechnicalSection)
		api.POST("/projects/:id/technical-offer/propose-personnel", hm.Project.ProposePersonnel)
		api.POST("/projects/:id/technical-offer/propose-materiel", hm.Project.ProposeMateriel)
		api.POST("/projects/:id/technical-offer/save", hm.Project.SaveTechnicalOffer)
		api.POST("/projects/:id/technical-offer/export-drive", hm.Project.ExportTechnicalOfferToDrive)
		api.POST("/projects/:id/technical-offer/copy-cvs", hm.Project.CopyPersonnelCVsToProject)

		// Gestion du Référentiel Entreprise (RAG)
		api.GET("/company", hm.Project.GetCompanyProfile)
		api.POST("/company/documents", hm.Project.AddCompanyDocument)
		api.PUT("/company/documents/:id", hm.Project.UpdateCompanyDocument)
		api.POST("/company/documents/:id/analyze", hm.Project.AnalyzeCompanyCV)
		api.POST("/company/analyze-all-cvs", hm.Project.AnalyzeAllCompanyCVs)
		api.POST("/company/documents/upload", hm.Project.UploadCompanyDocument)
		api.POST("/company/sync-drive", hm.Project.SyncCompanyDriveDocuments)

		// Explorateur Google Drive
		api.GET("/drive/folders", hm.Project.ListDriveFolders)
	}

	return r
}

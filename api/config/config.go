// config/config.go
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

type AppConfig struct {
	OAuth        *oauth2.Config
	ProjectID    string
	Region       string
	Port         string
	GeminiModel  string
	GeminiAPIKey string
	DatabaseDSN  string
	FrontendURL  string
}

func LoadConfig() *AppConfig {

	// 1. Charger le fichier .env en local (ne bloque pas si absent en prod, ex: conteneur Docker)
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️ Aucun fichier .env trouvé, lecture des variables d'environnement système.")
	}

	// 2. Récupérer les variables d'environnement
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Port par défaut
	}
	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	region := os.Getenv("GOOGLE_REGION")
	redirectURL := os.Getenv("OAUTH_REDIRECT_URL")
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	geminiModel := os.Getenv("GEMINI_MODEL")
	if geminiModel == "" || geminiModel == "gemini-2.5-flash" || geminiModel == "gemini-1.5-flash" {
		geminiModel = "gemini-3.5-flash" // Modèle par défaut supporté
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgrespassword"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "agents_dao_db"
	}
	dbSSL := os.Getenv("DB_SSLMODE")
	if dbSSL == "" {
		dbSSL = "disable"
	}

	dbDSN := os.Getenv("DATABASE_URL")
	if dbDSN == "" {
		dbDSN = "host=" + dbHost + " user=" + dbUser + " password=" + dbPassword + " dbname=" + dbName + " port=" + dbPort + " sslmode=" + dbSSL
	}

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Impossible de lire credentials.json : %v", err)
	}

	// Définition des privilèges requis
	config, err := google.ConfigFromJSON(b,
		drive.DriveScope,
		"https://www.googleapis.com/auth/userinfo.profile",
	)
	if err != nil {
		log.Fatalf("Erreur lors du parsing du JSON OAuth2 : %v", err)
	}

	config.RedirectURL = redirectURL

	return &AppConfig{
		OAuth:        config,
		ProjectID:    projectID,
		Region:       region,
		Port:         port,
		GeminiModel:  geminiModel,
		GeminiAPIKey: geminiAPIKey,
		DatabaseDSN:  dbDSN,
		FrontendURL:  frontendURL,
	}
}
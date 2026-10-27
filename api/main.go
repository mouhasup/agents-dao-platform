// main.go
package main

import (
	"fmt"
	"log"

	"agents-dao-platform/config"
	"agents-dao-platform/database"
	"agents-dao-platform/handler"
	"agents-dao-platform/route"
	"agents-dao-platform/service"
)

func main() {
	// 1. Initialisation de la configuration (.env)
	cfg := config.LoadConfig()

	// 2. Connexion PostgreSQL via GORM (avec tolérance si la base est en cours de démarrage)
	db, err := database.InitDB(cfg.DatabaseDSN)
	if err != nil {
		log.Printf("⚠️ Impossible de se connecter à PostgreSQL (%v). Démarrage en mode mémoire temporaire.", err)
		log.Printf("💡 Conseil : Lancez 'make db-up' pour démarrer PostgreSQL dans Docker.")
	}

	// 3. Initialisation des Services Métier
	sm := service.NewServiceManager(cfg, db)

	// 4. Initialisation des Handlers
	hm := handler.NewHandlerManager(sm, cfg)

	// 5. Initialisation du moteur Gin avec ses routes et middlewares
	r := route.SetupRouter(hm, sm, cfg)

	fmt.Printf("🚀 Serveur Gin & GORM démarré sur http://localhost:%s\n", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Erreur serveur : %v", err)
	}
}

package database

import (
	"agents-dao-platform/model"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB connecte l'application à PostgreSQL et lance les migrations automatiques
func InitDB(dsn string) (*gorm.DB, error) {
	log.Println("🔌 Connexion à PostgreSQL...")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	log.Println("🔄 Exécution des migrations automatiques (GORM AutoMigrate)...")
	err = db.AutoMigrate(
		&model.CompanyProfile{},
		&model.CompanyDocument{},
		&model.Project{},
		&model.ProjectTask{},
		&model.DriveFolder{},
		&model.UserSession{},
	)
	if err != nil {
		return nil, err
	}

	// Création du profil entreprise par défaut s'il n'existe pas encore
	var count int64
	db.Model(&model.CompanyProfile{}).Count(&count)
	if count == 0 {
		defaultProfile := model.CompanyProfile{
			ID:        "cp-default",
			Nom:       "Mon Entreprise BTP & Ingénierie",
			NINEA:     "001234567 2V3",
			RCCM:      "SN.DKR.2024.B.12345",
			Email:     "contact@mon-entreprise.sn",
			Telephone: "+221 33 800 00 00",
			Adresse:   "Dakar, Sénégal",
		}
		db.Create(&defaultProfile)
		log.Println("🏢 Profil d'entreprise par défaut initialisé.")
	}

	log.Println("✅ Base de données PostgreSQL connectée et migrée avec succès !")
	return db, nil
}

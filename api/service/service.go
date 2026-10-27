// service/service.go
package service

import (
	"agents-dao-platform/config"

	"gorm.io/gorm"
)

// ServiceManager centralise tous les services métiers
type ServiceManager struct {
	Auth      AuthService
	Drive     DriveService
	Gemini    GeminiService
	Agent     AgentService
	Project   ProjectRepository
	Session   SessionRepository
	Analysis  AnalysisService
	Technical TechnicalOfferService
}

func NewServiceManager(cfg *config.AppConfig, db *gorm.DB) *ServiceManager {
	var repo ProjectRepository
	var sessions SessionRepository
	if db != nil {
		repo = NewGormProjectRepository(db)
		sessions = NewGormSessionRepository(db)
	} else {
		repo = NewMemoryProjectRepository()
		sessions = NewMemorySessionRepository()
	}

	drive := NewDriveService(cfg)
	agent := NewAgentService(cfg)
	technical := NewTechnicalOfferService(repo, agent, drive)

	return &ServiceManager{
		Auth:      NewAuthService(cfg),
		Drive:     drive,
		Gemini:    NewGeminiService(cfg),
		Agent:     agent,
		Project:   repo,
		Session:   sessions,
		Analysis:  NewAnalysisService(repo, drive, agent),
		Technical: technical,
	}
}

// handler/handler.go
package handler

import (
	"agents-dao-platform/config"
	"agents-dao-platform/service"
)

type HandlerManager struct {
	Auth    *AuthHandler
	Project *ProjectHandler
}

func NewHandlerManager(sm *service.ServiceManager, cfg *config.AppConfig) *HandlerManager {
	return &HandlerManager{
		Auth:    NewAuthHandler(sm.Auth, sm.Session, cfg.FrontendURL),
		Project: NewProjectHandler(sm.Project, sm.Drive, sm.Analysis, sm.Session, sm.Technical, sm.Agent),
	}
}

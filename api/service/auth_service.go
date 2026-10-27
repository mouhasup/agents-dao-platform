// service/auth_service.go
package service

import (
	"context"
	"agents-dao-platform/config"
	"golang.org/x/oauth2"
)

type AuthService interface {
	GetLoginURL() string
	ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error)
}

type authService struct {
	cfg *config.AppConfig
}

func NewAuthService(cfg *config.AppConfig) AuthService {
	return &authService{cfg: cfg}
}

func (s *authService) GetLoginURL() string {
	return s.cfg.OAuth.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

func (s *authService) ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.cfg.OAuth.Exchange(ctx, code)
}
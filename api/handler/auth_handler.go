// handler/auth_handler.go
package handler

import (
	"fmt"
	"net/http"
	"strings"

	"agents-dao-platform/service"
	"agents-dao-platform/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	authService  service.AuthService
	sessions     service.SessionRepository
	frontendURL  string
}

func NewAuthHandler(as service.AuthService, sessions service.SessionRepository, frontendURL string) *AuthHandler {
	return &AuthHandler{
		authService: as,
		sessions:    sessions,
		frontendURL: frontendURL,
	}
}

func (h *AuthHandler) Home(c *gin.Context) {
	c.Redirect(http.StatusTemporaryRedirect, h.frontendURL)
}

func (h *AuthHandler) Login(c *gin.Context) {
	c.Redirect(http.StatusTemporaryRedirect, h.authService.GetLoginURL())
}

// Callback échange le code, persiste le token en session et redirige vers le frontend
func (h *AuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code manquant"})
		return
	}

	token, err := h.authService.ExchangeCodeForToken(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erreur échange token : %v", err)})
		return
	}

	// Google permet de décocher des permissions : on vérifie que l'accès Drive a bien été accordé
	if granted, _ := token.Extra("scope").(string); !strings.Contains(granted, "https://www.googleapis.com/auth/drive") {
		c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/?drive_error=scope")
		return
	}

	sessionID, err := h.sessions.Create(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de créer la session"})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	// 30 jours, HttpOnly. Passer secure=true en production HTTPS.
	c.SetCookie(utils.SessionCookieName, sessionID, 30*24*3600, "/", "", false, true)

	c.Redirect(http.StatusTemporaryRedirect, h.frontendURL)
}

// Me indique si l'utilisateur est connecté à Google
func (h *AuthHandler) Me(c *gin.Context) {
	token, _ := c.Request.Context().Value("oauth_token").(*oauth2.Token)
	c.JSON(http.StatusOK, gin.H{"authenticated": token != nil})
}

// Logout supprime la session
func (h *AuthHandler) Logout(c *gin.Context) {
	if sessionID, err := c.Cookie(utils.SessionCookieName); err == nil {
		_ = h.sessions.Delete(sessionID)
	}
	c.SetCookie(utils.SessionCookieName, "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"authenticated": false})
}

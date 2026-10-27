package utils

import (
	"context"
	"net/http"

	"agents-dao-platform/service"

	"github.com/gin-gonic/gin"
)

// SessionCookieName est le nom du cookie contenant l'identifiant de session
const SessionCookieName = "dao_session"

// GinCORS autorise le frontend avec envoi de cookies (credentials)
func GinCORS(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = allowedOrigin
		}
		// Avec credentials, l'origine doit être explicite (pas de "*")
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Vary", "Origin")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// SessionMiddleware lit le cookie de session et injecte le token OAuth dans le contexte
func SessionMiddleware(sessions service.SessionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(SessionCookieName)
		if err == nil && sessionID != "" {
			if token, err := sessions.GetToken(sessionID); err == nil {
				ctx := context.WithValue(c.Request.Context(), "oauth_token", token)
				c.Request = c.Request.WithContext(ctx)
			}
		}
		c.Next()
	}
}

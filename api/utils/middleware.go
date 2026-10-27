package utils

import (
	"log"
	"net/http"
	"time"
)

// Middleware représente un décorateur http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain permet de chaîner plusieurs middlewares.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// ChainHandlerFunc applique des middlewares à un http.HandlerFunc standard.
func ChainHandlerFunc(h http.HandlerFunc, middlewares ...Middleware) http.Handler {
	return Chain(h, middlewares...)
}

// Logging est un middleware qui journalise les requêtes HTTP (méthode, URI et temps d'exécution).
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("⏳ Début %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("✅ Fin %s %s en %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// Recovery est un middleware qui intercepte les paniques pour éviter que le serveur ne s'arrête.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("🚨 Panique récupérée : %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORS est un middleware qui ajoute les en-têtes CORS nécessaires pour autoriser les requêtes cross-origin.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

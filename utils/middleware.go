// utils/middleware.go
package utils

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware enregistre les détails de chaque requête
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Temps de début de la requête
		start := time.Now()

		// Traiter la requête
		next.ServeHTTP(w, r)

		// Calculer la durée
		duration := time.Since(start)

		// Journaliser les détails
		if LogFile != nil {
			clientIP := GetVisitorIP(r)
			logLine := fmt.Sprintf("[%s] %s - %s %s %s - %v\n",
				time.Now().Format("2006-01-02 15:04:05"),
				clientIP,
				r.Method,
				r.URL.Path,
				r.Proto,
				duration)
			LogFile.WriteString(logLine)
		}

		// Afficher également dans la console si en mode debug
		if Config.DebugLog {
			log.Printf("%s %s %s - %v", r.Method, r.URL.Path, GetVisitorIP(r), duration)
		}
	})
}

// RecoveryMiddleware récupère les panics dans les handlers
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Journaliser l'erreur
				if LogFile != nil {
					logLine := fmt.Sprintf("[%s] PANIC RECOVERED: %v - %s %s\n",
						time.Now().Format("2006-01-02 15:04:05"),
						err,
						r.Method,
						r.URL.Path)
					LogFile.WriteString(logLine)
				}

				// Logger dans la console
				log.Printf("PANIC RECOVERED: %v", err)

				// Renvoyer une erreur 500
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware ajoute les en-têtes CORS nécessaires
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Définir les en-têtes CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Gérer les requêtes preflight OPTIONS
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware ajoute des en-têtes de sécurité
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// En-têtes de sécurité de base
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' https://cdn.tailwindcss.com https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com;")

		next.ServeHTTP(w, r)
	})
}

// RateLimiterMiddleware limite le nombre de requêtes par IP
// (Implémentation simple - pour production, utiliser un middleware dédié)
func RateLimiterMiddleware(next http.Handler) http.Handler {
	// Map pour suivre les requêtes par IP avec leurs timestamps
	var (
		requestCounts = make(map[string][]time.Time)
		maxRequests   = 100              // Nombre maximum de requêtes
		timeWindow    = 60 * time.Second // Fenêtre de temps
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Récupérer l'IP du client
		clientIP := GetVisitorIP(r)

		// Horodatage actuel
		now := time.Now()

		// Nettoyer les anciennes requêtes
		if timestamps, exists := requestCounts[clientIP]; exists {
			var validTimestamps []time.Time
			for _, ts := range timestamps {
				if now.Sub(ts) < timeWindow {
					validTimestamps = append(validTimestamps, ts)
				}
			}
			requestCounts[clientIP] = validTimestamps
		}

		// Vérifier le nombre de requêtes
		if timestamps, exists := requestCounts[clientIP]; exists && len(timestamps) >= maxRequests {
			http.Error(w, "Trop de requêtes, veuillez réessayer plus tard", http.StatusTooManyRequests)

			// Journaliser
			if LogFile != nil {
				logLine := fmt.Sprintf("[%s] RATE LIMIT EXCEEDED: %s - %s %s\n",
					time.Now().Format("2006-01-02 15:04:05"),
					clientIP,
					r.Method,
					r.URL.Path)
				LogFile.WriteString(logLine)
			}

			return
		}

		// Ajouter cette requête
		requestCounts[clientIP] = append(requestCounts[clientIP], now)

		next.ServeHTTP(w, r)
	})
}

// MiddlewareChain combine plusieurs middlewares
func MiddlewareChain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

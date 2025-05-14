package utils

import (
	"BkC/blockchain"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// sessionMutex protège l'accès concurrent à la gestion des sessions.
var sessionMutex sync.Mutex

// GetVisitorIP extrait l'adresse IP du visiteur, en prenant en compte les proxies.
func GetVisitorIP(r *http.Request) string {
	// Vérifier d'abord les en-têtes communs pour les proxies
	ipHeaders := []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"CF-Connecting-IP", // Cloudflare
		"True-Client-IP",   // Akamai
	}

	for _, header := range ipHeaders {
		if ip := r.Header.Get(header); ip != "" {
			// X-Forwarded-For peut contenir plusieurs IPs (client, proxy1, proxy2...)
			// Nous voulons la première (celle du client)
			if strings.Contains(ip, ",") {
				return strings.TrimSpace(strings.Split(ip, ",")[0])
			}
			return strings.TrimSpace(ip)
		}
	}

	// Si aucun en-tête n'est trouvé, utiliser l'adresse distante
	remoteAddr := r.RemoteAddr
	if strings.ContainsRune(remoteAddr, ':') {
		// IPv4 ou IPv6 avec port, enlever le port
		return remoteAddr[:strings.LastIndex(remoteAddr, ":")]
	}

	return remoteAddr
}

// TrackVisitor suit les connexions et déconnexions d'un visiteur
func TrackVisitor(clientIP string, isConnected bool, sessions map[string]*UserSession, bc *blockchain.Blockchain) {
	now := time.Now()
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	if isConnected {
		// Nouvel utilisateur connecté
		if _, exists := sessions[clientIP]; !exists {
			// Enregistrer la connexion dans la blockchain
			connectionData := fmt.Sprintf("Connexion de %s à %v", clientIP, now)
			bc.AddBlockAsync(connectionData, 2) // Difficulté plus faible pour connexions
		}
	} else {
		// Utilisateur déconnecté
		if session, exists := sessions[clientIP]; exists {
			// Enregistrer la déconnexion
			sessionDuration := now.Sub(session.LastSeen)
			if sessionDuration.Minutes() > 1 { // Éviter les déconnexions trop rapides
				disconnectData := fmt.Sprintf("Déconnexion de %s après %v minutes",
					clientIP, int(sessionDuration.Minutes()))
				bc.AddBlockAsync(disconnectData, 2)
			}
		}
	}
}

// ManageSession gère la logique de session en fonction de l'IP du client.
func ManageSession(clientIP string, sessions map[string]*UserSession, bc *blockchain.Blockchain) {
	now := time.Now()
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	session, exists := sessions[clientIP]

	if !exists {
		sessions[clientIP] = &UserSession{
			IP:        clientIP,
			StartTime: now,
			LastSeen:  now,
		}
		// Enregistrer la visite (mais pas comme bloc pour éviter de surcharger)
		log.Printf("📡 Nouvelle visite de %s", clientIP)
	} else {
		if now.Sub(session.LastSeen) >= 5*time.Minute {
			sessionData := fmt.Sprintf("Session de %s démarrée à %v", clientIP, session.StartTime)
			bc.AddBlock(sessionData, 4)
			session.StartTime = now
		}
		session.LastSeen = now
	}
}

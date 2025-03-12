package utils

import (
	"BkC/blockchain"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// sessionMutex protège l'accès concurrent à la gestion des sessions.
var sessionMutex sync.Mutex

// GetVisitorIP extrait l'adresse IP du visiteur.
func GetVisitorIP(r *http.Request) string {
	remoteAddr := r.RemoteAddr
	return remoteAddr[:strings.LastIndex(remoteAddr, ":")]
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
	} else {
		if now.Sub(session.LastSeen) >= 5*time.Minute {
			sessionData := fmt.Sprintf("Session de %s démarrée à %v", clientIP, session.StartTime)
			// Utiliser MineBlock au lieu de AddBlock
			bc.MineBlock("system", sessionData)
			session.StartTime = now
		}
		session.LastSeen = now
	}
}

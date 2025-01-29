package utils

import (
	"fmt"
	"mon-projet/blockchain"
	"net/http"
	"strings"
	"time"
)

type UserSession struct {
	IP        string
	StartTime time.Time
	LastSeen  time.Time
}

// GetVisitorIP extrait l'adresse IP du visiteur
func GetVisitorIP(r *http.Request) string {
	remoteAddr := r.RemoteAddr
	return remoteAddr[:strings.LastIndex(remoteAddr, ":")]
}

// ManageSession gère la logique de session
func ManageSession(clientIP string, sessions map[string]*UserSession, bc *blockchain.Blockchain) {
	now := time.Now()
	session, exists := sessions[clientIP]

	if !exists {
		sessions[clientIP] = &UserSession{IP: clientIP, StartTime: now, LastSeen: now}
	} else {
		if now.Sub(session.StartTime) >= 5*time.Minute {
			sessionData := fmt.Sprintf("Session from %s started at %v", clientIP, session.StartTime)
			bc.AddBlock(sessionData, 4)
			session.StartTime = now
		}
		session.LastSeen = now
	}
}

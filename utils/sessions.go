package utils

import "time"

// UserSession définit la structure d'une session utilisateur.
type UserSession struct {
	Username  string
	IP        string
	StartTime time.Time
	LastSeen  time.Time
}

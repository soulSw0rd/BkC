package utils

import (
	"net"
	"time"
)

// UserStatus représente le statut d'un utilisateur
type UserStatus int

const (
	StatusOffline UserStatus = iota
	StatusOnline
	StatusAway
	StatusBusy
)

// UserSession définit la structure d'une session utilisateur.
type UserSession struct {
	Username       string
	IP             string
	NetworkInfo    *NetworkInfo // Information réseau détaillée
	StartTime      time.Time
	LastSeen       time.Time
	Status         UserStatus
	IsRegistered   bool           // S'il s'agit d'un utilisateur enregistré ou juste un visiteur
	UserAgent      string         // Agent utilisateur du navigateur
	Visits         int            // Nombre de visites
	MiningActivity map[string]int // Suivi de l'activité de minage ("blocksMinés", "dernierMinage", etc.)
}

// NetworkInfo contient des informations détaillées sur la connexion réseau
type NetworkInfo struct {
	RawIP       string // Adresse IP complète
	IsIPv6      bool   // Si l'adresse est IPv6
	ISP         string // Fournisseur d'accès Internet (déterminé si possible)
	CountryCode string // Code pays (déterminé si possible)
	City        string // Ville (déterminé si possible)
}

// NewNetworkInfo crée une structure NetworkInfo à partir d'une adresse IP
func NewNetworkInfo(ipStr string) *NetworkInfo {
	info := &NetworkInfo{
		RawIP: ipStr,
	}

	// Analyser l'adresse IP
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return info
	}

	// Détecter si c'est IPv6
	info.IsIPv6 = ip.To4() == nil

	// Dans une implémentation réelle, on pourrait utiliser une API de géolocalisation
	// pour déterminer l'ISP, le pays et la ville

	// Par défaut, nous allons simplement vérifier si c'est une adresse locale
	if ip.IsLoopback() || ip.IsPrivate() {
		info.ISP = "Réseau local"
		info.CountryCode = "FR" // Supposons que c'est en France
		info.City = "Local"
	} else {
		info.ISP = "Indéterminé"
		info.CountryCode = "??"
		info.City = "Inconnue"
	}

	return info
}

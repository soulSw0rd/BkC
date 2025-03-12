package handlers

import (
	"BkC/blockchain"
	"BkC/utils"
	"html/template"
	"net/http"
	"sync"
	"time"
)

// Gestion des utilisateurs et des sessions.
var (
	users    = map[string]string{"admin": "admin"} // Utilisateurs autorisés (admin/admin)
	sessions = make(map[string]*utils.UserSession) // Stocke les sessions actives
	mu       sync.Mutex                            // Protection contre les accès concurrents
)

// LoginHandler affiche la page de connexion.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement de la page de connexion", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]string{
		"Error": r.URL.Query().Get("error"),
	})
}

// LoginSubmitHandler traite le formulaire de connexion.
func LoginSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if pass, ok := users[username]; ok && pass == password {
		mu.Lock()
		sessions[username] = &utils.UserSession{
			Username:  username,
			IP:        r.RemoteAddr,
			StartTime: time.Now(),
			LastSeen:  time.Now(),
		}
		mu.Unlock()

		http.SetCookie(w, &http.Cookie{
			Name:  "session",
			Value: username,
			Path:  "/",
		})

		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
}

// LogoutHandler supprime la session et redirige vers la page de connexion.
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		mu.Lock()
		delete(sessions, cookie.Value)
		mu.Unlock()

		http.SetCookie(w, &http.Cookie{
			Name:   "session",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// HomeHandler affiche la page d'accueil uniquement si l'utilisateur est connecté.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	mu.Lock()
	session, exists := sessions[cookie.Value]
	mu.Unlock()
	if !exists || session == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement de la page d'accueil", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Stats structure pour les statistiques.
type Stats struct {
	VisitorCount       int               `json:"visitorCount"`
	ActiveSessions     int               `json:"activeSessions"`
	DailyTransactions  int               `json:"dailyTransactions"`
	LastBlock          *blockchain.Block `json:"lastBlock"`
	TransactionHistory struct {
		Dates  []string `json:"dates"`
		Counts []int    `json:"counts"`
	} `json:"transactionHistory"`
}

// StatsHandler affiche la page des statistiques si l'utilisateur est connecté.
func StatsHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		mu.Lock()
		session, exists := sessions[cookie.Value]
		mu.Unlock()
		if !exists || session == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		var lastBlock *blockchain.Block
		emptyTime := time.Time{} // Une valeur zero/vide pour time.Time

		if len(bc.Blocks) > 0 {
			lastBlock = bc.Blocks[len(bc.Blocks)-1]
		} else {
			lastBlock = &blockchain.Block{
				Index:     0,
				Timestamp: emptyTime, // Utiliser time.Time{} au lieu de "N/A"
				Data:      "Aucun bloc",
				Hash:      "N/A",
				PrevHash:  "N/A",
			}
		}

		stats := Stats{
			VisitorCount:      120,
			ActiveSessions:    len(sessions),
			DailyTransactions: 342,
			LastBlock:         lastBlock,
			TransactionHistory: struct {
				Dates  []string `json:"dates"`
				Counts []int    `json:"counts"`
			}{
				Dates:  []string{"Lundi", "Mardi", "Mercredi", "Jeudi", "Vendredi"},
				Counts: []int{50, 70, 90, 120, 150},
			},
		}

		tmpl, err := template.ParseFiles("templates/stats.html")
		if err != nil {
			http.Error(w, "Erreur lors du chargement de la page des statistiques", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, stats)
	}
}

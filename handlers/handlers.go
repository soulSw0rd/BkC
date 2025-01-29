package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"sync"
	"time"

	"BkC/blockchain"
	"BkC/utils"
)

// Clé de session (nom du cookie)
var sessionName = "user_session"

// Variables globales pour les statistiques
var (
	visitorMutex   sync.Mutex
	uniqueVisitors = make(map[string]bool)
	visitorCount   int
	activeSessions int
)

// Stats représente les statistiques à afficher
type Stats struct {
	VisitorCount       int
	ActiveSessionCount int
	LastBlock          *blockchain.Block
}

// HomeHandler protège la page d'accueil
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	utils.LogRequest(r)

	// Vérifier si l'utilisateur est connecté
	if !isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement de la page d'accueil", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Erreur lors du rendu de la page", http.StatusInternalServerError)
	}
}

// StatsHandler protège la page des statistiques
func StatsHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.LogRequest(r)

		// Vérifier si l'utilisateur est connecté
		if !isAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		visitorIP := utils.GetVisitorIP(r)
		visitorMutex.Lock()
		if !uniqueVisitors[visitorIP] {
			uniqueVisitors[visitorIP] = true
			visitorCount++
		}
		visitorMutex.Unlock()

		stats := generateStats(bc)

		tmpl, err := template.ParseFiles("templates/stats.html")
		if err != nil {
			http.Error(w, "Erreur lors du chargement des statistiques", http.StatusInternalServerError)
			return
		}

		lastBlockHTML := formatBlock(stats.LastBlock)

		data := struct {
			VisitorCount       int
			ActiveSessionCount int
			LastBlock          string
		}{
			VisitorCount:       stats.VisitorCount,
			ActiveSessionCount: stats.ActiveSessionCount,
			LastBlock:          lastBlockHTML,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Erreur lors du rendu des statistiques", http.StatusInternalServerError)
		}
	}
}

// Vérifie si l'utilisateur est authentifié via son cookie
func isAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie(sessionName)
	return err == nil && cookie.Value == "admin"
}

// LoginHandler affiche la page de login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	utils.LogRequest(r)

	// Si l'utilisateur est déjà connecté, on le redirige
	if isAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	http.ServeFile(w, r, "templates/login.html")
}

// LoginSubmitHandler gère l'authentification
func LoginSubmitHandler(w http.ResponseWriter, r *http.Request) {
	utils.LogRequest(r)

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Vérifier les identifiants
	if username == "admin" && password == "admin" {
		// Créer un cookie de session
		expiration := time.Now().Add(30 * time.Minute)
		cookie := http.Cookie{
			Name:     sessionName,
			Value:    "admin",
			Expires:  expiration,
			HttpOnly: true,
		}
		http.SetCookie(w, &cookie)

		// Rediriger vers la page d'accueil
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Identifiants incorrects
	http.Error(w, "Identifiants incorrects", http.StatusUnauthorized)
}

// LogoutHandler gère la déconnexion
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	utils.LogRequest(r)

	// Supprimer le cookie de session
	cookie := http.Cookie{
		Name:     sessionName,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	// Rediriger vers la page de login
	http.Redirect(w, r, "/login", http.StatusFound)
}

// formatBlock formate les données d'un bloc pour affichage HTML
func formatBlock(block *blockchain.Block) string {
	if block == nil {
		return "<p>Aucun bloc disponible.</p>"
	}

	return fmt.Sprintf(`
<p>Index : %d</p>
<p>Timestamp : %s</p>
<p>Données : %s</p>
<p>Hash : %s</p>
`, block.Index, block.Timestamp, block.Data, block.Hash)
}

// generateStats génère les statistiques de la blockchain
func generateStats(bc *blockchain.Blockchain) Stats {
	bc.Lock()
	defer bc.Unlock()

	var lastBlock *blockchain.Block
	if len(bc.Blocks) > 0 {
		lastBlock = bc.Blocks[len(bc.Blocks)-1]
	}

	visitorMutex.Lock()
	defer visitorMutex.Unlock()

	return Stats{
		VisitorCount:       visitorCount,
		ActiveSessionCount: activeSessions,
		LastBlock:          lastBlock,
	}
}

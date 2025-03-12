package handlers

import (
	"BkC/utils"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

// SecurityDashboardHandler affiche le tableau de bord de sécurité
func SecurityDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier l'authentification
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

	// Récupérer l'utilisateur connecté
	username := session.Username

	// Récupérer le score de sécurité pour l'affichage
	securityScore := utils.SecurityScore{
		Score:           100,
		LastUpdated:     time.Now(),
		Factors:         make(map[utils.RiskFactor]float64),
		RecentEvents:    []string{},
		AnomalyDetected: false,
		ThreatLevel:     utils.RiskLow,
	}

	// Préparer les données pour le template
	data := struct {
		Username      string
		SecurityScore utils.SecurityScore
	}{
		Username:      username,
		SecurityScore: securityScore,
	}

	// Enregistrer cet accès au tableau de bord dans le journal
	clientIP := utils.GetVisitorIP(r)
	fmt.Printf("Accès au tableau de bord de sécurité par %s depuis %s\n", username, clientIP)

	// Charger et exécuter le template
	tmpl, err := template.New("security").Parse("<h1>Tableau de bord de sécurité</h1><p>Utilisateur: {{.Username}}</p>")
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}

// APISecurityStatsHandler renvoie les statistiques de sécurité au format JSON
func APISecurityStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier l'authentification
	cookie, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Non authentifié"})
		return
	}

	mu.Lock()
	session, exists := sessions[cookie.Value]
	mu.Unlock()
	if !exists || session == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Non authentifié"})
		return
	}

	username := session.Username

	// Journaliser l'accès à l'API
	clientIP := utils.GetVisitorIP(r)
	fmt.Printf("Accès à l'API de sécurité par %s depuis %s\n", username, clientIP)

	// Construire la réponse
	response := struct {
		Username      string    `json:"username"`
		LastLogin     time.Time `json:"last_login"`
		SecurityScore float64   `json:"security_score"`
		IPAddress     string    `json:"ip_address"`
	}{
		Username:      username,
		LastLogin:     time.Now(),
		SecurityScore: 100,
		IPAddress:     clientIP,
	}

	// Définir les en-têtes et renvoyer la réponse
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// EnhancedLoginSubmitHandler est une version améliorée du gestionnaire de connexion
func EnhancedLoginSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	clientIP := utils.GetVisitorIP(r)

	// Journaliser la tentative de connexion
	fmt.Printf("Tentative de connexion pour l'utilisateur %s depuis %s\n", username, clientIP)

	// Vérifier les identifiants
	if pass, ok := users[username]; ok && pass == password {
		// Connexion réussie
		mu.Lock()
		sessions[username] = &utils.UserSession{
			Username:  username,
			IP:        r.RemoteAddr,
			StartTime: time.Now(),
			LastSeen:  time.Now(),
		}
		mu.Unlock()

		// Définir le cookie de session
		http.SetCookie(w, &http.Cookie{
			Name:  "session",
			Value: username,
			Path:  "/",
		})

		fmt.Printf("Connexion réussie pour %s depuis %s\n", username, clientIP)

		// Rediriger vers la page d'accueil
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	// Connexion échouée
	fmt.Printf("Échec de connexion pour %s depuis %s\n", username, clientIP)
	http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
}

// SecurityAuditHandler affiche la piste d'audit de sécurité
func SecurityAuditHandler(w http.ResponseWriter, r *http.Request) {
	// Vérification d'authentification (similaire aux autres handlers)
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

	username := session.Username
	clientIP := utils.GetVisitorIP(r)
	fmt.Printf("Accès à l'audit de sécurité par %s depuis %s\n", username, clientIP)

	// Données factices pour ce gestionnaire simplifié
	data := struct {
		Username string
		Events   []string
	}{
		Username: username,
		Events:   []string{"Connexion", "Déconnexion", "Création de bloc"},
	}

	// Charger et exécuter le template
	tmpl, err := template.New("audit").Parse("<h1>Audit de sécurité</h1><p>Utilisateur: {{.Username}}</p>")
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}

// SecurityAlertsHandler gère les alertes de sécurité
func SecurityAlertsHandler(w http.ResponseWriter, r *http.Request) {
	// Vérification d'authentification (similaire aux autres handlers)
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

	username := session.Username
	clientIP := utils.GetVisitorIP(r)
	fmt.Printf("Accès aux alertes de sécurité par %s depuis %s\n", username, clientIP)

	// Données factices pour ce gestionnaire simplifié
	data := struct {
		Username string
		Alerts   []string
	}{
		Username: username,
		Alerts:   []string{"Tentative de connexion suspecte", "Activité inhabituelle"},
	}

	// Charger et exécuter le template
	tmpl, err := template.New("alerts").Parse("<h1>Alertes de sécurité</h1><p>Utilisateur: {{.Username}}</p>")
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}

// Fonction helper pour convertir string en int
func parseInt(s string, defaultValue int) int {
	if value, err := strconv.Atoi(s); err == nil {
		return value
	}
	return defaultValue
}

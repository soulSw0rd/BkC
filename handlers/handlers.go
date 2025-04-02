package handlers

import (
	"BkC/blockchain"
	"BkC/utils"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Gestion des utilisateurs et des sessions.
var (
	users    = map[string]User{}                   // Utilisateurs stockés
	sessions = make(map[string]*utils.UserSession) // Stocke les sessions actives
	mu       sync.Mutex                            // Protection contre les accès concurrents
)

// User définit la structure d'un utilisateur
type User struct {
	Username    string    `json:"username"`
	Password    string    `json:"-"` // Le - empêche la sérialisation du mot de passe
	Email       string    `json:"email"`
	IsAdmin     bool      `json:"isAdmin"`
	CreatedAt   time.Time `json:"createdAt"`
	LastLoginAt time.Time `json:"lastLoginAt"`
}

// InitSampleUsers initialise quelques utilisateurs par défaut
func InitSampleUsers() {
	// Générer le hash du mot de passe "admin"
	adminHash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)

	// Générer le hash du mot de passe "user"
	userHash, _ := bcrypt.GenerateFromPassword([]byte("user"), bcrypt.DefaultCost)

	// Ajouter l'admin
	users["admin"] = User{
		Username:  "admin",
		Password:  string(adminHash),
		Email:     "admin@cryptochain.go",
		IsAdmin:   true,
		CreatedAt: time.Now(),
	}

	// Ajouter un utilisateur régulier
	users["user"] = User{
		Username:  "user",
		Password:  string(userHash),
		Email:     "user@cryptochain.go",
		IsAdmin:   false,
		CreatedAt: time.Now(),
	}
}

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

	mu.Lock()
	user, exists := users[username]
	mu.Unlock()

	if !exists {
		http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
		return
	}

	// Vérifier le mot de passe avec bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
		return
	}

	// Générer un token de session unique
	sessionToken := generateSessionToken()

	mu.Lock()
	// Mettre à jour la dernière connexion
	user.LastLoginAt = time.Now()
	users[username] = user

	// Créer une nouvelle session
	sessions[sessionToken] = &utils.UserSession{
		Username:  username,
		IP:        r.RemoteAddr,
		StartTime: time.Now(),
		LastSeen:  time.Now(),
		IsAdmin:   user.IsAdmin,
	}
	mu.Unlock()

	// Définir le cookie de session
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600, // 1 heure
	})

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

// generateSessionToken génère un token de session unique
func generateSessionToken() string {
	tokenBytes := make([]byte, 16)
	rand.Read(tokenBytes)
	return hex.EncodeToString(tokenBytes)
}

// RegisterHandler affiche la page d'inscription
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/register.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement de la page d'inscription", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]string{
		"Error": r.URL.Query().Get("error"),
	})
}

// RegisterSubmitHandler traite le formulaire d'inscription
func RegisterSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")

	// Vérifier si l'utilisateur existe déjà
	mu.Lock()
	_, exists := users[username]
	mu.Unlock()

	if exists {
		http.Redirect(w, r, "/register?error=exists", http.StatusSeeOther)
		return
	}

	// Hacher le mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Redirect(w, r, "/register?error=internal", http.StatusSeeOther)
		return
	}

	// Créer le nouvel utilisateur
	newUser := User{
		Username:  username,
		Password:  string(hashedPassword),
		Email:     email,
		IsAdmin:   false, // Les nouveaux utilisateurs ne sont pas admin par défaut
		CreatedAt: time.Now(),
	}

	mu.Lock()
	users[username] = newUser
	mu.Unlock()

	// Rediriger vers la page de connexion
	http.Redirect(w, r, "/login?registered=1", http.StatusSeeOther)
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

	// Récupérer l'utilisateur pour afficher des infos personnalisées
	mu.Lock()
	user, _ := users[session.Username]
	mu.Unlock()

	data := map[string]interface{}{
		"Username": user.Username,
		"IsAdmin":  user.IsAdmin,
	}

	tmpl.Execute(w, data)
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
		if len(bc.Blocks) > 0 {
			lastBlock = bc.Blocks[len(bc.Blocks)-1]
		} else {
			lastBlock = &blockchain.Block{
				Index:      0,
				Timestamp:  time.Now(),
				MerkleRoot: "N/A",
				Hash:       "N/A",
				PrevHash:   "N/A",
			}
		}

		// Total des transactions dans toute la blockchain
		totalTx := 0
		for _, block := range bc.Blocks {
			totalTx += len(block.Transactions)
		}

		// Calculer les transactions des 5 derniers jours
		dates := []string{"Lundi", "Mardi", "Mercredi", "Jeudi", "Vendredi"}
		counts := []int{50, 70, 90, 120, 150} // Données fictives pour l'exemple

		stats := Stats{
			VisitorCount:      len(users),
			ActiveSessions:    len(sessions),
			DailyTransactions: totalTx,
			LastBlock:         lastBlock,
			TransactionHistory: struct {
				Dates  []string `json:"dates"`
				Counts []int    `json:"counts"`
			}{
				Dates:  dates,
				Counts: counts,
			},
		}

		// Vérifier si le client demande du JSON
		if r.Header.Get("Accept") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(stats)
			return
		}

		// Sinon, rendre le template HTML
		tmpl, err := template.ParseFiles("templates/stats.html")
		if err != nil {
			http.Error(w, "Erreur lors du chargement de la page des statistiques", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, stats)
	}
}

// ProfileHandler affiche la page de profil de l'utilisateur
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
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

	mu.Lock()
	user, _ := users[session.Username]
	mu.Unlock()

	tmpl, err := template.ParseFiles("templates/profile.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement de la page de profil", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, map[string]interface{}{
		"User":     user,
		"Username": user.Username,
		"Email":    user.Email,
		"IsAdmin":  user.IsAdmin,
		"JoinDate": user.CreatedAt.Format("02/01/2006"),
	})
}

// AdminHandler affiche le panneau d'administration (réservé aux admins)
func AdminHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		mu.Lock()
		session, exists := sessions[cookie.Value]
		mu.Unlock()
		if !exists || session == nil || !session.IsAdmin {
			http.Error(w, "Accès non autorisé", http.StatusForbidden)
			return
		}

		tmpl, err := template.ParseFiles("templates/admin.html")
		if err != nil {
			http.Error(w, "Erreur lors du chargement de la page d'administration", http.StatusInternalServerError)
			return
		}

		mu.Lock()
		usersList := make([]User, 0, len(users))
		for _, u := range users {
			// Créer une copie sans mot de passe
			userCopy := User{
				Username:    u.Username,
				Email:       u.Email,
				IsAdmin:     u.IsAdmin,
				CreatedAt:   u.CreatedAt,
				LastLoginAt: u.LastLoginAt,
			}
			usersList = append(usersList, userCopy)
		}
		mu.Unlock()

		data := map[string]interface{}{
			"Users":      usersList,
			"BlockCount": len(bc.Blocks),
			"Difficulty": bc.MiningDifficulty,
			"Sessions":   len(sessions),
		}

		tmpl.Execute(w, data)
	}
}

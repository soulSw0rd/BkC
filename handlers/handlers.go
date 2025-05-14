package handlers

import (
	"BkC/blockchain"
	"BkC/utils"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// Gestion des utilisateurs et des sessions.
var (
	users       = map[string]string{"admin": "admin"} // Utilisateurs autorisés sans hash (admin/admin)
	sessions    = make(map[string]*utils.UserSession) // Stocke les sessions actives
	mu          sync.Mutex                            // Protection contre les accès concurrents
	usersOnline = 0                                   // Nombre d'utilisateurs en ligne
	bc          *blockchain.Blockchain                // Référence globale à la blockchain
)

// InitGlobalBC initialise la référence globale à la blockchain
func InitGlobalBC(blockchain *blockchain.Blockchain) {
	bc = blockchain

	// Charger les utilisateurs au démarrage
	if err := LoadUsers(); err != nil {
		log.Printf("Erreur lors du chargement des utilisateurs: %v", err)
	}

	// Charger les sessions au démarrage
	if err := LoadSessions(); err != nil {
		log.Printf("Erreur lors du chargement des sessions: %v", err)
	}
}

// Sauvegarde les utilisateurs dans un fichier.
func SaveUsers() error {
	mu.Lock()
	defer mu.Unlock()

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile("users.json", data, 0644)
}

// Charge les utilisateurs depuis un fichier.
func LoadUsers() error {
	if _, err := os.Stat("users.json"); os.IsNotExist(err) {
		// Le fichier n'existe pas, on utilise les utilisateurs par défaut
		return nil
	}

	data, err := ioutil.ReadFile("users.json")
	if err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	return json.Unmarshal(data, &users)
}

// SaveSessions sauvegarde les sessions dans un fichier
func SaveSessions() error {
	mu.Lock()
	defer mu.Unlock()

	// Créer une copie des sessions pour la sérialisation
	sessionsCopy := make(map[string]*utils.UserSession)
	for key, session := range sessions {
		// Copier uniquement les sessions enregistrées
		if session.IsRegistered {
			sessionsCopy[key] = session
		}
	}

	data, err := json.MarshalIndent(sessionsCopy, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur lors de la sérialisation des sessions: %v", err)
	}

	return ioutil.WriteFile("sessions.json", data, 0644)
}

// LoadSessions charge les sessions depuis un fichier
func LoadSessions() error {
	if _, err := os.Stat("sessions.json"); os.IsNotExist(err) {
		// Le fichier n'existe pas, on commence avec des sessions vides
		return nil
	}

	data, err := ioutil.ReadFile("sessions.json")
	if err != nil {
		return fmt.Errorf("erreur lors de la lecture du fichier de sessions: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	// Désérialiser les sessions
	loadedSessions := make(map[string]*utils.UserSession)
	if err := json.Unmarshal(data, &loadedSessions); err != nil {
		return fmt.Errorf("erreur lors de la désérialisation des sessions: %v", err)
	}

	// Fusionner les sessions chargées avec les sessions actuelles
	for username, session := range loadedSessions {
		// Vérifier si l'utilisateur existe encore
		if _, exists := users[username]; exists {
			// Réinitialiser l'état de connexion pour les sessions chargées
			session.Status = utils.StatusOffline
			sessions[username] = session
		}
	}

	// Mettre à jour le compteur d'utilisateurs en ligne
	recountOnlineUsers()

	return nil
}

// recountOnlineUsers recalcule le nombre d'utilisateurs en ligne
func recountOnlineUsers() {
	count := 0
	for _, session := range sessions {
		if session.Status == utils.StatusOnline {
			count++
		}
	}
	usersOnline = count
}

// LoginHandler affiche la page de connexion.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement de la page de connexion", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]string{
		"Error":   r.URL.Query().Get("error"),
		"Message": r.URL.Query().Get("message"),
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

	// Récupérer l'IP réelle du client
	clientIP := utils.GetVisitorIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Vérification du mot de passe (sans hash)
	if pass, ok := users[username]; ok && pass == password {
		mu.Lock()
		// Créer la session avec informations réseau détaillées
		networkInfo := utils.NewNetworkInfo(clientIP)
		sessions[username] = &utils.UserSession{
			Username:       username,
			IP:             clientIP,
			NetworkInfo:    networkInfo,
			StartTime:      time.Now(),
			LastSeen:       time.Now(),
			Status:         utils.StatusOnline,
			IsRegistered:   true,
			UserAgent:      userAgent,
			Visits:         1,                    // Première visite
			MiningActivity: make(map[string]int), // Initialiser l'activité de minage
		}
		usersOnline++
		mu.Unlock()

		// Sauvegarder les sessions
		if err := SaveSessions(); err != nil {
			log.Printf("Erreur lors de la sauvegarde des sessions: %v", err)
		}

		// Traquer la connexion dans la blockchain
		utils.TrackVisitor(clientIP, true, sessions, bc)

		// Log de connexion
		log.Printf("👤 Connexion utilisateur: %s depuis %s [%s]", username, clientIP, networkInfo.CountryCode)

		// Définir un cookie qui dure longtemps (1 an)
		http.SetCookie(w, &http.Cookie{
			Name:    "session",
			Value:   username,
			Path:    "/",
			MaxAge:  31536000, // 365 jours en secondes
			Expires: time.Now().AddDate(1, 0, 0),
		})

		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	// Log tentative échouée
	log.Printf("⚠️ Échec d'authentification pour l'utilisateur: %s depuis %s", username, clientIP)
	http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
}

// SigninHandler affiche la page d'inscription.
func SigninHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/signin.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement de la page d'inscription", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]string{
		"Error": r.URL.Query().Get("error"),
	})
}

// SigninSubmitHandler traite le formulaire d'inscription.
func SigninSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	// Récupérer l'IP et l'User-Agent
	clientIP := utils.GetVisitorIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Vérifier que le mot de passe et sa confirmation correspondent
	if password != confirmPassword {
		http.Redirect(w, r, "/signin?error=password_mismatch", http.StatusSeeOther)
		return
	}

	// Vérifier que le nom d'utilisateur n'existe pas déjà
	mu.Lock()
	_, exists := users[username]
	if exists {
		mu.Unlock()
		http.Redirect(w, r, "/signin?error=username_exists", http.StatusSeeOther)
		return
	}

	// Ajouter l'utilisateur (sans hash)
	users[username] = password
	mu.Unlock()

	// Sauvegarder les utilisateurs dans un fichier
	if err := SaveUsers(); err != nil {
		log.Printf("Erreur lors de la sauvegarde des utilisateurs: %v", err)
	}

	// Créer automatiquement la session et connecter l'utilisateur
	mu.Lock()
	// Créer la session avec informations réseau détaillées
	networkInfo := utils.NewNetworkInfo(clientIP)
	sessions[username] = &utils.UserSession{
		Username:       username,
		IP:             clientIP,
		NetworkInfo:    networkInfo,
		StartTime:      time.Now(),
		LastSeen:       time.Now(),
		Status:         utils.StatusOnline,
		IsRegistered:   true,
		UserAgent:      userAgent,
		Visits:         1,                    // Première visite
		MiningActivity: make(map[string]int), // Initialiser l'activité de minage
	}
	usersOnline++
	mu.Unlock()

	// Sauvegarder les sessions
	if err := SaveSessions(); err != nil {
		log.Printf("Erreur lors de la sauvegarde des sessions: %v", err)
	}

	// Enregistrer un nouveau bloc pour l'inscription
	signupData := fmt.Sprintf("Inscription de %s depuis %s à %v", username, clientIP, time.Now())
	bc.AddBlockAsync(signupData, 3) // Difficulté 3 pour l'inscription

	// Log de la nouvelle inscription
	log.Printf("✅ Nouvel utilisateur: %s depuis %s [%s]", username, clientIP, networkInfo.CountryCode)

	// Définir un cookie qui dure longtemps (1 an)
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   username,
		Path:    "/",
		MaxAge:  31536000, // 365 jours en secondes
		Expires: time.Now().AddDate(1, 0, 0),
	})

	// Rediriger vers la page d'accueil
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

// LogoutHandler supprime la session et redirige vers la page de connexion.
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	clientIP := utils.GetVisitorIP(r)

	if err == nil {
		mu.Lock()
		session, exists := sessions[cookie.Value]
		if exists {
			// Récupérer le nom d'utilisateur avant de supprimer la session
			username := session.Username
			log.Printf("🚪 Déconnexion utilisateur: %s depuis %s", username, clientIP)

			// Traquer la déconnexion
			utils.TrackVisitor(clientIP, false, sessions, bc)

			// Marquer la session comme déconnecté plutôt que de la supprimer
			session.Status = utils.StatusOffline
			session.LastSeen = time.Now()

			if usersOnline > 0 {
				usersOnline--
			}
		}
		mu.Unlock()

		// Sauvegarder les sessions après modification
		if err := SaveSessions(); err != nil {
			log.Printf("Erreur lors de la sauvegarde des sessions: %v", err)
		}

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

	// Mettre à jour l'heure de la dernière visite
	mu.Lock()
	session.LastSeen = time.Now()
	mu.Unlock()

	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement de la page d'accueil", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]string{
		"Username": session.Username,
	})
}

// Stats structure pour les statistiques.
type Stats struct {
	VisitorCount       int                `json:"visitorCount"`
	ActiveSessions     int                `json:"activeSessions"`
	RegisteredUsers    int                `json:"registeredUsers"`
	DailyTransactions  int                `json:"dailyTransactions"`
	LastBlock          *blockchain.Block  `json:"lastBlock"`
	OnlineUsers        []string           `json:"onlineUsers"`
	RecentConnections  []RecentConnection `json:"recentConnections"` // Connexions récentes
	TransactionHistory struct {
		Dates  []string `json:"dates"`
		Counts []int    `json:"counts"`
	} `json:"transactionHistory"`
}

// RecentConnection contient des informations sur une connexion récente
type RecentConnection struct {
	Username   string    `json:"username"`   // Nom d'utilisateur (ou "Visiteur" si anonyme)
	Timestamp  time.Time `json:"timestamp"`  // Heure de connexion
	Country    string    `json:"country"`    // Pays (si disponible)
	UserAgent  string    `json:"userAgent"`  // Agent utilisateur
	LastAction string    `json:"lastAction"` // Dernière action effectuée
}

// StatsHandler affiche la page des statistiques si l'utilisateur est connecté.
func StatsHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Vérifier si c'est une requête AJAX (XHR) ou une demande de page
		isXHR := r.Header.Get("X-Requested-With") == "XMLHttpRequest" || r.URL.Query().Get("format") == "json"

		// Pour les requêtes API (AJAX/fetch), nous ne vérifions pas l'authentification
		if !isXHR {
			// Vérification d'authentification seulement pour l'affichage de la page
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

			// Mettre à jour l'heure de la dernière visite
			mu.Lock()
			session.LastSeen = time.Now()
			mu.Unlock()
		}

		var lastBlock *blockchain.Block
		if len(bc.Blocks) > 0 {
			lastBlock = bc.Blocks[len(bc.Blocks)-1]
		} else {
			lastBlock = &blockchain.Block{
				Index:     0,
				Timestamp: "N/A",
				Data:      "Aucun bloc",
				Hash:      "N/A",
				PrevHash:  "N/A",
			}
		}

		// Récupérer la liste des utilisateurs en ligne et préparer les données anonymisées
		mu.Lock()
		onlineUsersList := make([]string, 0, len(sessions))
		recentConnections := make([]RecentConnection, 0)

		// Counter pour les visiteurs anonymes
		anonymousCounter := 1

		// Collecter les données
		for username, session := range sessions {
			onlineUsersList = append(onlineUsersList, username)

			// Préparation des données pour l'affichage
			displayName := username
			if displayName == "" {
				displayName = fmt.Sprintf("Visiteur-%d", anonymousCounter)
				anonymousCounter++
			}

			// Préparer les données de connexion
			country := "FR" // Par défaut
			userAgent := "Inconnu"

			if session.NetworkInfo != nil && session.NetworkInfo.CountryCode != "" {
				country = session.NetworkInfo.CountryCode
			}

			if session.UserAgent != "" {
				// Simplifier l'user agent pour l'affichage
				userAgent = simplifyUserAgent(session.UserAgent)
			}

			lastAction := "Navigation"
			if time.Since(session.LastSeen) < 5*time.Minute {
				lastAction = "Actif"
			} else if time.Since(session.LastSeen) < 30*time.Minute {
				lastAction = "Inactif"
			} else {
				lastAction = "Déconnecté"
			}

			recentConnections = append(recentConnections, RecentConnection{
				Username:   displayName,
				Timestamp:  session.LastSeen,
				Country:    country,
				UserAgent:  userAgent,
				LastAction: lastAction,
			})
		}

		registeredUsers := len(users)
		visitorCount := len(sessions) // Nombre total de sessions
		mu.Unlock()

		stats := Stats{
			VisitorCount:      visitorCount,
			ActiveSessions:    usersOnline,
			RegisteredUsers:   registeredUsers,
			DailyTransactions: len(bc.Blocks) - 1, // Moins le bloc genesis
			LastBlock:         lastBlock,
			OnlineUsers:       onlineUsersList,
			RecentConnections: recentConnections,
			TransactionHistory: struct {
				Dates  []string `json:"dates"`
				Counts []int    `json:"counts"`
			}{
				Dates:  []string{"Lundi", "Mardi", "Mercredi", "Jeudi", "Vendredi"},
				Counts: []int{len(bc.Blocks) / 5, len(bc.Blocks) / 4, len(bc.Blocks) / 3, len(bc.Blocks) / 2, len(bc.Blocks)},
			},
		}

		// Pour les requêtes API, renvoyer les données en JSON
		if isXHR {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(stats)
			return
		}

		// Pour les requêtes normales, renvoyer la page HTML
		tmpl, err := template.ParseFiles("templates/stats.html")
		if err != nil {
			http.Error(w, "Erreur lors du chargement de la page des statistiques", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, stats)
	}
}

// simplifyUserAgent simplifie la chaîne User-Agent pour l'affichage
func simplifyUserAgent(ua string) string {
	if strings.Contains(ua, "Chrome") {
		return "Chrome"
	} else if strings.Contains(ua, "Firefox") {
		return "Firefox"
	} else if strings.Contains(ua, "Safari") {
		return "Safari"
	} else if strings.Contains(ua, "Edge") {
		return "Edge"
	} else if strings.Contains(ua, "MSIE") || strings.Contains(ua, "Trident") {
		return "Internet Explorer"
	} else {
		return "Navigateur inconnu"
	}
}

// MineBlockHandler permet aux utilisateurs de miner un nouveau bloc
func MineBlockHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Vérifier si l'utilisateur est connecté
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "Vous devez être connecté pour miner", http.StatusUnauthorized)
			return
		}

		mu.Lock()
		session, exists := sessions[cookie.Value]
		mu.Unlock()
		if !exists || session == nil {
			http.Error(w, "Session non valide", http.StatusUnauthorized)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		// Décoder la requête JSON
		var mineRequest struct {
			Data string `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&mineRequest); err != nil {
			http.Error(w, "Données JSON invalides", http.StatusBadRequest)
			return
		}

		// Récupérer le nom du mineur
		username := session.Username

		// Créer un message formaté incluant les informations utilisateur
		messageData := fmt.Sprintf("%s (par %s à %s)",
			mineRequest.Data,
			username,
			time.Now().Format("15:04:05 02/01/2006"))

		// Difficulté du minage
		difficulty := 4

		// Traçabilité: ajouter le bloc avec informations sur le mineur
		newBlock := bc.AddBlockWithMiner(messageData, difficulty, username)

		// Mettre à jour la dernière action de l'utilisateur
		mu.Lock()
		// Mettre à jour la session avec les informations de minage
		session.LastSeen = time.Now()
		if session.MiningActivity == nil {
			session.MiningActivity = make(map[string]int)
		}
		session.MiningActivity["blocksMinés"] = session.MiningActivity["blocksMinés"] + 1
		session.MiningActivity["dernierMinage"] = int(time.Now().Unix())
		mu.Unlock()

		// Sauvegarder les sessions après le minage
		if err := SaveSessions(); err != nil {
			log.Printf("Erreur lors de la sauvegarde des sessions: %v", err)
		}

		// Log dans la console
		log.Printf("🔗 Nouveau hash généré par %s: %.8s... (bloc #%d)", username, newBlock.Hash, newBlock.Index)

		// Répondre avec un succès et les informations du bloc
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "success",
			"message":  "Votre hash a été ajouté à la blockchain avec succès!",
			"miner":    username,
			"blockId":  newBlock.Index,
			"hash":     newBlock.Hash,
			"nonce":    newBlock.Nonce,
			"prevHash": newBlock.PrevHash,
		})
	}
}

// MinerStats représente les statistiques d'un mineur pour le classement
type MinerStats struct {
	Username       string `json:"username"`
	BlocksMined    int    `json:"blocksMined"`
	LastMiningTime int64  `json:"lastMiningTime"`
}

// MinersStatsHandler renvoie les statistiques de minage de tous les utilisateurs
func MinersStatsHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Calculer les statistiques de minage à partir des sessions enregistrées
		mu.Lock()
		minerStats := make([]MinerStats, 0, len(sessions))

		for _, session := range sessions {
			if session.IsRegistered {
				blocksMined := 0
				lastMiningTime := int64(0)

				if session.MiningActivity != nil {
					blocksMined = session.MiningActivity["blocksMinés"]
					lastMiningTime = int64(session.MiningActivity["dernierMinage"])
				}

				// Compter également les blocs minés dans la blockchain
				userBlocks := bc.GetBlocksByMiner(session.Username)
				blocksMined = len(userBlocks) // Utiliser le compte de la blockchain

				// Ajouter seulement les utilisateurs qui ont miné des blocs
				if blocksMined > 0 {
					minerStats = append(minerStats, MinerStats{
						Username:       session.Username,
						BlocksMined:    blocksMined,
						LastMiningTime: lastMiningTime,
					})
				}
			}
		}
		mu.Unlock()

		// Trier les mineurs par nombre de blocs minés (décroissant)
		sort.Slice(minerStats, func(i, j int) bool {
			return minerStats[i].BlocksMined > minerStats[j].BlocksMined
		})

		// Renvoyer les données en JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(minerStats)
	}
}

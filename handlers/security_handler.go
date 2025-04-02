package handlers

import (
	"BkC/utils"
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// LoginRequest représente les données de demande de connexion
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse représente la réponse à une demande de connexion
type LoginResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Token     string `json:"token,omitempty"`
	Username  string `json:"username,omitempty"`
	IsAdmin   bool   `json:"isAdmin,omitempty"`
	ExpiresAt int64  `json:"expiresAt,omitempty"`
}

// RegisterRequest représente les données de demande d'inscription
type RegisterRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	Email           string `json:"email"`
}

// RegisterResponse représente la réponse à une demande d'inscription
type RegisterResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Username string `json:"username,omitempty"`
}

// ApiLoginHandler traite les demandes de connexion via l'API
func ApiLoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Vérifier que la méthode est POST
	if r.Method != "POST" {
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Message: "Méthode non autorisée",
		})
		return
	}

	// Décoder le corps de la requête
	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Message: "Format de requête invalide",
		})
		return
	}

	// Vérifier les identifiants
	mu.Lock()
	user, exists := users[loginReq.Username]
	mu.Unlock()

	if !exists {
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Message: "Identifiants invalides",
		})
		return
	}

	// Utiliser bcrypt pour vérifier le mot de passe
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Message: "Identifiants invalides",
		})
		return
	}

	// Générer un token de session unique
	sessionToken := generateSessionToken()

	// Définir quand la session expire
	expiresAt := time.Now().Add(1 * time.Hour)

	mu.Lock()
	// Mettre à jour la dernière connexion
	user.LastLoginAt = time.Now()
	users[loginReq.Username] = user

	// Créer une nouvelle session
	sessions[sessionToken] = &utils.UserSession{
		Username:  loginReq.Username,
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
		Expires:  expiresAt,
	})

	// Retourner la réponse
	json.NewEncoder(w).Encode(LoginResponse{
		Success:   true,
		Message:   "Connexion réussie",
		Token:     sessionToken,
		Username:  loginReq.Username,
		IsAdmin:   user.IsAdmin,
		ExpiresAt: expiresAt.Unix(),
	})
}

// ApiRegisterHandler traite les demandes d'inscription via l'API
func ApiRegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Vérifier que la méthode est POST
	if r.Method != "POST" {
		json.NewEncoder(w).Encode(RegisterResponse{
			Success: false,
			Message: "Méthode non autorisée",
		})
		return
	}

	// Décoder le corps de la requête
	var registerReq RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&registerReq); err != nil {
		json.NewEncoder(w).Encode(RegisterResponse{
			Success: false,
			Message: "Format de requête invalide",
		})
		return
	}

	// Vérifier que les mots de passe correspondent
	if registerReq.Password != registerReq.ConfirmPassword {
		json.NewEncoder(w).Encode(RegisterResponse{
			Success: false,
			Message: "Les mots de passe ne correspondent pas",
		})
		return
	}

	// Vérifier si l'utilisateur existe déjà
	mu.Lock()
	_, exists := users[registerReq.Username]
	mu.Unlock()

	if exists {
		json.NewEncoder(w).Encode(RegisterResponse{
			Success: false,
			Message: "Ce nom d'utilisateur existe déjà",
		})
		return
	}

	// Hacher le mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerReq.Password), bcrypt.DefaultCost)
	if err != nil {
		json.NewEncoder(w).Encode(RegisterResponse{
			Success: false,
			Message: "Erreur lors de la création du compte",
		})
		return
	}

	// Créer le nouvel utilisateur
	newUser := User{
		Username:  registerReq.Username,
		Password:  string(hashedPassword),
		Email:     registerReq.Email,
		IsAdmin:   false, // Les nouveaux utilisateurs ne sont pas admin par défaut
		CreatedAt: time.Now(),
	}

	// Enregistrer l'utilisateur
	mu.Lock()
	users[registerReq.Username] = newUser
	mu.Unlock()

	// Retourner la réponse
	json.NewEncoder(w).Encode(RegisterResponse{
		Success:  true,
		Message:  "Inscription réussie",
		Username: registerReq.Username,
	})
}

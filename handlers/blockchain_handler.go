package handlers

import (
	"BkC/blockchain"
	"BkC/utils"
	"encoding/json"
	"html/template"
	"net/http"
	"sync"
	"time"
)

var sessionMutex sync.Mutex // Protéger l'accès aux sessions dans ce handler

// BlockchainPageData structure pour les données de la page blockchain
type BlockchainPageData struct {
	Username  string
	Blocks    []*blockchain.Block
	LastBlock *blockchain.Block
}

// BlockchainHandler gère les requêtes sur la blockchain.
func BlockchainHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.LogRequest(r)
		clientIP := utils.GetVisitorIP(r)

		sessionMutex.Lock()
		utils.ManageSession(clientIP, sessions, bc)
		sessionMutex.Unlock()

		// Récupérer l'utilisateur connecté (s'il y en a un)
		var username string
		cookie, err := r.Cookie("session")
		if err == nil {
			mu.Lock()
			if session, exists := sessions[cookie.Value]; exists && session != nil {
				username = session.Username
				// Mettre à jour la dernière activité de l'utilisateur
				session.LastSeen = time.Now()
			}
			mu.Unlock()
		}

		if r.Method == "POST" {
			// Rediriger vers MineBlockHandler pour avoir une meilleure traçabilité
			// des hashs générés par les utilisateurs
			if err != nil || username == "" {
				http.Error(w, "Vous devez être connecté pour générer un hash", http.StatusUnauthorized)
				return
			}

			var input struct {
				Data string `json:"data"`
			}
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				http.Error(w, "Données JSON invalides", http.StatusBadRequest)
				return
			}

			// Récupérer le handler de minage et l'exécuter
			mineHandler := MineBlockHandler(bc)
			mineHandler(w, r)
		} else if r.Method == "GET" {
			// Vérifier si le client accepte du HTML (navigateur)
			acceptHeader := r.Header.Get("Accept")
			if acceptHeader == "application/json" {
				// Si le client demande explicitement du JSON, lui envoyer
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(bc.Blocks)
			} else {
				// Sinon, afficher la page HTML
				var lastBlock *blockchain.Block
				if len(bc.Blocks) > 0 {
					lastBlock = bc.Blocks[len(bc.Blocks)-1]
				}

				pageData := BlockchainPageData{
					Username:  username,
					Blocks:    bc.Blocks,
					LastBlock: lastBlock,
				}

				tmpl, err := template.ParseFiles("templates/blockchain.html")
				if err != nil {
					http.Error(w, "Erreur lors du chargement de la page blockchain", http.StatusInternalServerError)
					return
				}
				tmpl.Execute(w, pageData)
			}
		} else {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	}
}

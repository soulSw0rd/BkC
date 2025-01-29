package handlers

import (
	"BkC/blockchain"
	"BkC/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

var (
	sessionMutex sync.Mutex
	sessions     = make(map[string]*utils.UserSession)
)

// BlockchainHandler gère les requêtes sur la blockchain
func BlockchainHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.LogRequest(r)
		clientIP := utils.GetVisitorIP(r)

		// Gestion des sessions
		sessionMutex.Lock()
		utils.ManageSession(clientIP, sessions, bc)
		sessionMutex.Unlock()

		if r.Method == "POST" {
			var input struct {
				Data string `json:"data"`
			}
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			bc.AddBlock(input.Data, 4)
			fmt.Fprintf(w, "Nouveau bloc ajouté !")
		} else if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(bc.Blocks)
		} else {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	}
}

package handlers

import (
	"BkC/blockchain"
	"BkC/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

var sessionMutex sync.Mutex // Protéger l'accès aux sessions dans ce handler

// BlockRequest représente une demande de création de bloc
type BlockRequest struct {
	MinerAddress string `json:"minerAddress"`
}

// BlockResponse représente la réponse après création d'un bloc
type BlockResponse struct {
	Success      bool    `json:"success"`
	Message      string  `json:"message"`
	BlockHash    string  `json:"blockHash,omitempty"`
	BlockIndex   int     `json:"blockIndex,omitempty"`
	Transactions int     `json:"transactions,omitempty"`
	MiningReward float64 `json:"miningReward,omitempty"`
}

// BlockchainHandler gère les requêtes sur la blockchain.
func BlockchainHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.LogRequest(r)
		w.Header().Set("Content-Type", "application/json")

		clientIP := utils.GetVisitorIP(r)

		sessionMutex.Lock()
		utils.ManageSession(clientIP, sessions, bc)
		sessionMutex.Unlock()

		if r.Method == "POST" {
			// Vérification de l'authentification pour miner un bloc
			cookie, err := r.Cookie("session")
			if err != nil {
				json.NewEncoder(w).Encode(BlockResponse{
					Success: false,
					Message: "Authentification requise pour miner un bloc",
				})
				return
			}

			mu.Lock()
			session, exists := sessions[cookie.Value]
			mu.Unlock()
			if !exists || session == nil {
				json.NewEncoder(w).Encode(BlockResponse{
					Success: false,
					Message: "Session invalide",
				})
				return
			}

			var blockReq BlockRequest
			if err := json.NewDecoder(r.Body).Decode(&blockReq); err != nil {
				json.NewEncoder(w).Encode(BlockResponse{
					Success: false,
					Message: "Requête invalide: " + err.Error(),
				})
				return
			}

			minerAddress := blockReq.MinerAddress
			if minerAddress == "" {
				minerAddress = "default_miner_address"
			}

			// Miner un nouveau bloc
			newBlock := bc.CreateBlock(minerAddress)

			json.NewEncoder(w).Encode(BlockResponse{
				Success:      true,
				Message:      "Nouveau bloc miné avec succès",
				BlockHash:    newBlock.Hash,
				BlockIndex:   newBlock.Index,
				Transactions: len(newBlock.Transactions),
				MiningReward: bc.MiningReward,
			})
		} else if r.Method == "GET" {
			blockHash := r.URL.Query().Get("hash")
			blockIndex := r.URL.Query().Get("index")

			if blockHash != "" {
				// Recherche par hash
				block := bc.GetBlockByHash(blockHash)
				if block == nil {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"success": false,
						"message": "Bloc non trouvé",
					})
					return
				}
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"block":   block,
				})
				return
			} else if blockIndex != "" {
				// Recherche par index
				index := 0
				fmt.Sscanf(blockIndex, "%d", &index)
				block := bc.GetBlockByIndex(index)
				if block == nil {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"success": false,
						"message": "Bloc non trouvé",
					})
					return
				}
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"block":   block,
				})
				return
			}

			// Pas de paramètre spécifique, retourne toute la blockchain
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"blocks":  bc.Blocks,
				"length":  len(bc.Blocks),
			})
		} else {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	}
}

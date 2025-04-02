package handlers

import (
	"BkC/blockchain"
	"BkC/utils"
	"encoding/json"
	"net/http"
)

// TransactionRequest représente une demande de création de transaction
type TransactionRequest struct {
	SenderPrivateKey string  `json:"senderPrivateKey"`
	SenderAddress    string  `json:"senderAddress"`
	RecipientAddress string  `json:"recipientAddress"`
	Amount           float64 `json:"amount"`
	Fee              float64 `json:"fee"`
}

// TransactionResponse représente la réponse après création d'une transaction
type TransactionResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	TransactionID string `json:"transactionId,omitempty"`
}

// TransactionHandler gère les requêtes relatives aux transactions
func TransactionHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.LogRequest(r)
		w.Header().Set("Content-Type", "application/json")

		// Vérification de l'authentification
		cookie, err := r.Cookie("session")
		if err != nil {
			json.NewEncoder(w).Encode(TransactionResponse{
				Success: false,
				Message: "Authentification requise",
			})
			return
		}

		mu.Lock()
		session, exists := sessions[cookie.Value]
		mu.Unlock()
		if !exists || session == nil {
			json.NewEncoder(w).Encode(TransactionResponse{
				Success: false,
				Message: "Session invalide",
			})
			return
		}

		if r.Method == "POST" {
			// Création d'une nouvelle transaction
			var txReq TransactionRequest
			if err := json.NewDecoder(r.Body).Decode(&txReq); err != nil {
				json.NewEncoder(w).Encode(TransactionResponse{
					Success: false,
					Message: "Requête invalide: " + err.Error(),
				})
				return
			}

			// Validation des données
			if txReq.RecipientAddress == "" || txReq.Amount <= 0 {
				json.NewEncoder(w).Encode(TransactionResponse{
					Success: false,
					Message: "Adresse du destinataire et montant doivent être valides",
				})
				return
			}

			// Création d'un portefeuille à partir de la clé privée (simplifiée pour l'exemple)
			wallet, err := blockchain.LoadWalletFromString(txReq.SenderPrivateKey)
			if err != nil {
				json.NewEncoder(w).Encode(TransactionResponse{
					Success: false,
					Message: "Clé privée invalide: " + err.Error(),
				})
				return
			}

			// Vérifier que l'adresse correspond
			if wallet.Address != txReq.SenderAddress {
				json.NewEncoder(w).Encode(TransactionResponse{
					Success: false,
					Message: "L'adresse ne correspond pas à la clé privée",
				})
				return
			}

			// Créer la transaction
			tx, err := wallet.CreateTransaction(txReq.RecipientAddress, txReq.Amount, txReq.Fee)
			if err != nil {
				json.NewEncoder(w).Encode(TransactionResponse{
					Success: false,
					Message: "Création de transaction échouée: " + err.Error(),
				})
				return
			}

			// Ajouter à la blockchain
			if err := bc.AddTransaction(tx); err != nil {
				json.NewEncoder(w).Encode(TransactionResponse{
					Success: false,
					Message: "Transaction rejetée: " + err.Error(),
				})
				return
			}

			json.NewEncoder(w).Encode(TransactionResponse{
				Success:       true,
				Message:       "Transaction créée et en attente de confirmation",
				TransactionID: tx.ID,
			})
			return
		} else if r.Method == "GET" {
			// Récupérer les transactions en attente
			address := r.URL.Query().Get("address")
			if address != "" {
				// Récupérer l'historique des transactions pour une adresse
				history := bc.GetTransactionHistory(address)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success":      true,
					"transactions": history,
					"balance":      bc.GetBalance(address),
				})
				return
			}

			// Récupérer toutes les transactions en attente dans le mempool
			pendingTxs := bc.MemPool.GetTransactions(100)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":      true,
				"transactions": pendingTxs,
			})
		} else {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	}
}

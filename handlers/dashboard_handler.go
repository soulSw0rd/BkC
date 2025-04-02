package handlers

import (
	"BkC/blockchain"
	"encoding/json"
	"net/http"
	"time"
)

// DashboardData représente les données pour le tableau de bord
type DashboardData struct {
	BlockCount         int                      `json:"blockCount"`
	TransactionCount   int                      `json:"transactionCount"`
	PendingTxCount     int                      `json:"pendingTxCount"`
	Difficulty         int                      `json:"difficulty"`
	Balance            float64                  `json:"balance"`
	LastBlockTime      time.Time                `json:"lastBlockTime"`
	LastBlockHash      string                   `json:"lastBlockHash"`
	RecentTransactions []blockchain.Transaction `json:"recentTransactions"`
}

// DashboardHandler renvoie les données pour le tableau de bord
func DashboardHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Vérification de l'authentification
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "Authentification requise", http.StatusUnauthorized)
			return
		}

		mu.Lock()
		session, exists := sessions[cookie.Value]
		mu.Unlock()
		if !exists || session == nil {
			http.Error(w, "Session invalide", http.StatusUnauthorized)
			return
		}

		// Récupérer les données de la blockchain
		blockCount := len(bc.Blocks)
		lastBlock := bc.Blocks[blockCount-1]

		// Compter les transactions
		transactionCount := 0
		for _, block := range bc.Blocks {
			transactionCount += len(block.Transactions)
		}

		// Récupérer le solde de l'utilisateur
		balance := bc.GetBalance(session.Username)

		// Récupérer les transactions récentes de l'utilisateur
		transactions := bc.GetTransactionHistory(session.Username)

		// Limiter à 10 transactions maximum
		recentTransactions := transactions
		if len(recentTransactions) > 10 {
			recentTransactions = recentTransactions[:10]
		}

		// Construire la réponse
		dashboardData := DashboardData{
			BlockCount:         blockCount,
			TransactionCount:   transactionCount,
			PendingTxCount:     len(bc.MemPool.Transactions),
			Difficulty:         bc.MiningDifficulty,
			Balance:            balance,
			LastBlockTime:      lastBlock.Timestamp,
			LastBlockHash:      lastBlock.Hash,
			RecentTransactions: recentTransactions,
		}

		// Renvoyer les données au format JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    dashboardData,
		})
	}
}

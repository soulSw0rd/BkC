package handlers

import (
	"BkC/blockchain"
	"BkC/utils"
	"encoding/json"
	"net/http"
	"path/filepath"
)

// WalletResponse représente la réponse pour une opération sur portefeuille
type WalletResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	Address     string `json:"address,omitempty"`
	PublicKey   []byte `json:"publicKey,omitempty"`
	PrivateKey  string `json:"privateKey,omitempty"` // Attention: sensible!
	WalletCount int    `json:"walletCount,omitempty"`
}

// WalletHandler gère les requêtes relatives aux portefeuilles
func WalletHandler(walletDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.LogRequest(r)
		w.Header().Set("Content-Type", "application/json")

		// Vérification de l'authentification
		cookie, err := r.Cookie("session")
		if err != nil {
			json.NewEncoder(w).Encode(WalletResponse{
				Success: false,
				Message: "Authentification requise",
			})
			return
		}

		mu.Lock()
		session, exists := sessions[cookie.Value]
		mu.Unlock()
		if !exists || session == nil {
			json.NewEncoder(w).Encode(WalletResponse{
				Success: false,
				Message: "Session invalide",
			})
			return
		}

		username := session.Username

		if r.Method == "POST" {
			// Création d'un nouveau portefeuille
			wallet, err := blockchain.NewWallet()
			if err != nil {
				json.NewEncoder(w).Encode(WalletResponse{
					Success: false,
					Message: "Erreur lors de la création du portefeuille: " + err.Error(),
				})
				return
			}

			// Sauvegarder le portefeuille
			userWalletPath := filepath.Join(walletDir, username)
			walletFileName := filepath.Join(userWalletPath, wallet.Address+".json")
			if err := wallet.SaveToFile(walletFileName); err != nil {
				json.NewEncoder(w).Encode(WalletResponse{
					Success: false,
					Message: "Erreur lors de la sauvegarde du portefeuille: " + err.Error(),
				})
				return
			}

			// ATTENTION: Dans un système de production réel, ne JAMAIS
			// renvoyer la clé privée dans la réponse! C'est uniquement
			// pour des raisons de démonstration ici.
			json.NewEncoder(w).Encode(WalletResponse{
				Success:    true,
				Message:    "Portefeuille créé avec succès",
				Address:    wallet.Address,
				PublicKey:  wallet.PublicKey,
				PrivateKey: "DEMO_ONLY_" + wallet.Address, // En prod, ne pas exposer!
			})
			return
		} else if r.Method == "GET" {
			// Liste des portefeuilles de l'utilisateur
			userWalletPath := filepath.Join(walletDir, username)
			wallets, err := filepath.Glob(filepath.Join(userWalletPath, "*.json"))
			if err != nil {
				json.NewEncoder(w).Encode(WalletResponse{
					Success: false,
					Message: "Erreur lors de la récupération des portefeuilles: " + err.Error(),
				})
				return
			}

			var walletList []map[string]string
			for _, walletFile := range wallets {
				wallet, err := blockchain.LoadWalletFromFile(walletFile)
				if err != nil {
					continue
				}
				walletList = append(walletList, map[string]string{
					"address": wallet.Address,
				})
			}

			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":  true,
				"wallets":  walletList,
				"count":    len(walletList),
				"username": username,
			})
		} else {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	}
}

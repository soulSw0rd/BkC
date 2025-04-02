package handlers

import (
	"BkC/blockchain"
	"BkC/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Variables pour le démarrage du serveur
var startTime = time.Now()

// APIResponse définit la structure standard de réponse API
type APIResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
	StatusCode int         `json:"-"` // Non sérialisé, utilisé en interne
}

// sendAPIResponse envoie une réponse API standardisée
func sendAPIResponse(w http.ResponseWriter, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")

	// Définir le code de statut HTTP
	if response.StatusCode != 0 {
		w.WriteHeader(response.StatusCode)
	} else if !response.Success {
		w.WriteHeader(http.StatusBadRequest)
	}

	// Encoder et envoyer la réponse
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Erreur lors de l'encodage de la réponse", http.StatusInternalServerError)
	}
}

// APIHandler gère toutes les requêtes API RESTful de l'application
func APIHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Variables pour le traitement des requêtes
		var response APIResponse

		// Obtenir l'utilisateur depuis la session si authentifié
		var username string
		cookie, err := r.Cookie("session")
		if err == nil {
			mu.Lock()
			session, exists := sessions[cookie.Value]
			mu.Unlock()
			if exists && session != nil {
				username = session.Username
			}
		}

		// Déterminer le point d'entrée API
		path := strings.TrimPrefix(r.URL.Path, "/api/")
		pathParts := strings.Split(path, "/")
		endpoint := pathParts[0]

		// Journaliser la requête API
		utils.LogAuditEvent(
			utils.EventTypeUIAccess,
			username,
			r.RemoteAddr,
			fmt.Sprintf("Accès API: %s %s", r.Method, r.URL.Path),
			utils.RiskLow,
			nil,
		)

		// Traiter la requête en fonction de l'endpoint et de la méthode HTTP
		switch endpoint {
		case "status":
			// Endpoint de status du serveur (/api/status)
			handleStatusRequest(w, r, bc)

		case "blockchain":
			// Endpoints blockchain (/api/blockchain/*)
			handleBlockchainRequest(w, r, bc, pathParts, username)

		case "transactions":
			// Endpoints transactions (/api/transactions/*)
			handleTransactionRequest(w, r, bc, pathParts, username)

		case "blocks":
			// Endpoints blocks (/api/blocks/*)
			handleBlockRequest(w, r, bc, pathParts, username)

		case "contracts":
			// Endpoints contracts (/api/contracts/*) - Redirige vers le gestionnaire de contrats
			ContractsHandler(bc)(w, r)

		case "wallet":
			// Endpoints wallet (/api/wallet/*)
			handleWalletRequest(w, r, username)

		case "mining":
			// Endpoints mining (/api/mining/*)
			handleMiningRequest(w, r, bc, username)

		case "stats":
			// Endpoints statistiques (/api/stats/*)
			handleStatsRequest(w, r, bc)

		case "docs":
			// Documentation de l'API
			handleAPIDocsRequest(w, r)

		default:
			// Endpoint non reconnu
			response = APIResponse{
				Success:    false,
				Message:    "Endpoint API non reconnu",
				StatusCode: http.StatusNotFound,
			}
			sendAPIResponse(w, response)
		}
	}
}

// handleStatusRequest traite les requêtes de statut du serveur
func handleStatusRequest(w http.ResponseWriter, r *http.Request, bc *blockchain.Blockchain) {
	// Récupérer des informations de base sur l'état du serveur
	uptime := time.Since(startTime).String()
	blockCount := len(bc.Blocks)
	lastBlockTime := bc.Blocks[blockCount-1].Timestamp

	// Créer la réponse
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"status":         "running",
			"version":        "1.2.0",
			"uptime":         uptime,
			"blockCount":     blockCount,
			"lastBlockTime":  lastBlockTime,
			"difficulty":     bc.MiningDifficulty,
			"miningReward":   bc.MiningReward,
			"pendingTxCount": len(bc.MemPool.Transactions),
		},
	}

	sendAPIResponse(w, response)
}

// handleBlockchainRequest traite les requêtes relatives à la blockchain
func handleBlockchainRequest(w http.ResponseWriter, r *http.Request, bc *blockchain.Blockchain, pathParts []string, username string) {
	// Déterminer le sous-endpoint
	subEndpoint := ""
	if len(pathParts) > 1 {
		subEndpoint = pathParts[1]
	}

	switch r.Method {
	case "GET":
		var response APIResponse

		if subEndpoint == "" {
			// Récupérer toute la blockchain
			response = APIResponse{
				Success: true,
				Data: map[string]interface{}{
					"blocks": bc.Blocks,
					"length": len(bc.Blocks),
				},
			}
		} else if subEndpoint == "height" {
			// Récupérer la hauteur actuelle de la blockchain
			response = APIResponse{
				Success: true,
				Data: map[string]interface{}{
					"height": len(bc.Blocks) - 1,
				},
			}
		} else if subEndpoint == "validate" {
			// Valider l'intégrité de la blockchain
			valid, err := bc.ValidateChain()

			if err != nil {
				response = APIResponse{
					Success: false,
					Error:   fmt.Sprintf("Erreur lors de la validation: %v", err),
				}
			} else {
				response = APIResponse{
					Success: true,
					Data: map[string]interface{}{
						"valid": valid,
					},
				}
			}
		} else {
			response = APIResponse{
				Success:    false,
				Message:    "Sous-endpoint non reconnu",
				StatusCode: http.StatusNotFound,
			}
		}

		sendAPIResponse(w, response)

	default:
		sendAPIResponse(w, APIResponse{
			Success:    false,
			Message:    "Méthode non autorisée",
			StatusCode: http.StatusMethodNotAllowed,
		})
	}
}

// handleBlockRequest traite les requêtes relatives aux blocs
func handleBlockRequest(w http.ResponseWriter, r *http.Request, bc *blockchain.Blockchain, pathParts []string, username string) {
	// Déterminer l'identifiant du bloc
	blockID := ""
	if len(pathParts) > 1 {
		blockID = pathParts[1]
	}

	switch r.Method {
	case "GET":
		var response APIResponse

		if blockID == "" {
			// Récupérer tous les blocs (avec pagination)
			page := 0
			limit := 10

			// Extraire les paramètres de requête
			pageStr := r.URL.Query().Get("page")
			limitStr := r.URL.Query().Get("limit")

			if pageStr != "" {
				if p, err := strconv.Atoi(pageStr); err == nil && p >= 0 {
					page = p
				}
			}

			if limitStr != "" {
				if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
					limit = l
				}
			}

			// Calculer les indices de début et de fin
			start := len(bc.Blocks) - 1 - (page * limit)
			end := start - limit + 1

			if start < 0 {
				start = 0
			}

			if end < 0 {
				end = 0
			}

			if start < end {
				// Inverser pour l'ordre des blocs
				start, end = end, start
			}

			// Extraire les blocs
			blocks := make([]*blockchain.Block, 0, limit)
			for i := start; i >= end && i >= 0; i-- {
				blocks = append(blocks, bc.Blocks[i])
			}

			response = APIResponse{
				Success: true,
				Data: map[string]interface{}{
					"blocks":      blocks,
					"page":        page,
					"limit":       limit,
					"totalBlocks": len(bc.Blocks),
				},
			}
		} else {
			// Récupérer un bloc spécifique
			var block *blockchain.Block

			// Vérifier si l'ID est un index ou un hash
			if index, err := strconv.Atoi(blockID); err == nil {
				block = bc.GetBlockByIndex(index)
			} else {
				block = bc.GetBlockByHash(blockID)
			}

			if block == nil {
				response = APIResponse{
					Success:    false,
					Message:    "Bloc non trouvé",
					StatusCode: http.StatusNotFound,
				}
			} else {
				response = APIResponse{
					Success: true,
					Data:    block,
				}
			}
		}

		sendAPIResponse(w, response)

	default:
		sendAPIResponse(w, APIResponse{
			Success:    false,
			Message:    "Méthode non autorisée",
			StatusCode: http.StatusMethodNotAllowed,
		})
	}
}

// handleTransactionRequest traite les requêtes relatives aux transactions
func handleTransactionRequest(w http.ResponseWriter, r *http.Request, bc *blockchain.Blockchain, pathParts []string, username string) {
	// Déterminer l'identifiant de la transaction
	txID := ""
	if len(pathParts) > 1 {
		txID = pathParts[1]
	}

	switch r.Method {
	case "GET":
		var response APIResponse

		if txID == "" {
			// Liste des transactions (avec pagination et filtres)
			address := r.URL.Query().Get("address")
			pending := r.URL.Query().Get("pending")
			limit := 20
			limitStr := r.URL.Query().Get("limit")

			if limitStr != "" {
				if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
					limit = l
				}
			}

			// Récupérer les transactions
			var transactions []blockchain.Transaction

			if address != "" {
				// Filtre par adresse
				transactions = bc.GetTransactionHistory(address)
			} else if pending == "true" {
				// Transactions en attente (mempool)
				pendingTxs := bc.MemPool.GetTransactions(limit)
				for _, tx := range pendingTxs {
					transactions = append(transactions, *tx)
				}
			} else {
				// Récupérer toutes les transactions des derniers blocs
				txCount := 0
				for i := len(bc.Blocks) - 1; i >= 0 && txCount < limit; i-- {
					for _, tx := range bc.Blocks[i].Transactions {
						transactions = append(transactions, tx)
						txCount++
						if txCount >= limit {
							break
						}
					}
				}
			}

			response = APIResponse{
				Success: true,
				Data: map[string]interface{}{
					"transactions": transactions,
					"count":        len(transactions),
				},
			}
		} else {
			// Rechercher la transaction par ID dans les blocs
			var transaction *blockchain.Transaction
			var found bool
			var blockIndex int

			// Rechercher dans le mempool
			pendingTxs := bc.MemPool.GetTransactions(0)
			for _, tx := range pendingTxs {
				if tx.ID == txID {
					transaction = tx
					found = false // Pas encore confirmée
					blockIndex = -1
					break
				}
			}

			// Si pas trouvée dans le mempool, rechercher dans les blocs
			if transaction == nil {
				for i, block := range bc.Blocks {
					for j, tx := range block.Transactions {
						if tx.ID == txID {
							transaction = &bc.Blocks[i].Transactions[j]
							found = true
							blockIndex = i
							break
						}
					}
					if transaction != nil {
						break
					}
				}
			}

			if transaction == nil {
				response = APIResponse{
					Success:    false,
					Message:    "Transaction non trouvée",
					StatusCode: http.StatusNotFound,
				}
			} else {
				response = APIResponse{
					Success: true,
					Data: map[string]interface{}{
						"transaction": transaction,
						"confirmed":   found,
						"blockIndex":  blockIndex,
					},
				}
			}
		}

		sendAPIResponse(w, response)

	case "POST":
		// Créer une nouvelle transaction
		var txReq TransactionRequest

		// Décoder le corps de la requête
		if err := json.NewDecoder(r.Body).Decode(&txReq); err != nil {
			sendAPIResponse(w, APIResponse{
				Success:    false,
				Message:    "Format de requête invalide",
				Error:      err.Error(),
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		// Validation des données
		if txReq.RecipientAddress == "" || txReq.Amount <= 0 {
			sendAPIResponse(w, APIResponse{
				Success:    false,
				Message:    "Adresse du destinataire et montant doivent être valides",
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		// Créer la transaction
		tx := &blockchain.Transaction{
			Sender:    txReq.SenderAddress,
			Recipient: txReq.RecipientAddress,
			Amount:    txReq.Amount,
			Fee:       txReq.Fee,
			Timestamp: time.Now(),
			Data:      txReq.Message,
		}

		// Dans une implémentation réelle, ici on signerait la transaction avec la clé privée

		// Ajouter à la blockchain
		if err := bc.AddTransaction(tx); err != nil {
			sendAPIResponse(w, APIResponse{
				Success:    false,
				Message:    "Transaction rejetée",
				Error:      err.Error(),
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		// Transaction acceptée
		sendAPIResponse(w, APIResponse{
			Success: true,
			Message: "Transaction créée et en attente de confirmation",
			Data: map[string]interface{}{
				"transactionId": tx.ID,
				"timestamp":     tx.Timestamp,
			},
		})

	default:
		sendAPIResponse(w, APIResponse{
			Success:    false,
			Message:    "Méthode non autorisée",
			StatusCode: http.StatusMethodNotAllowed,
		})
	}
}

// handleMiningRequest traite les requêtes de minage
func handleMiningRequest(w http.ResponseWriter, r *http.Request, bc *blockchain.Blockchain, username string) {
	switch r.Method {
	case "POST":
		// Vérification de l'authentification
		if username == "" {
			sendAPIResponse(w, APIResponse{
				Success:    false,
				Message:    "Authentification requise pour miner un bloc",
				StatusCode: http.StatusUnauthorized,
			})
			return
		}

		// Récupérer le nom du mineur depuis le corps de la requête
		var req struct {
			MinerAddress string `json:"minerAddress"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// Si le corps est vide, utiliser le nom d'utilisateur comme adresse du mineur
			req.MinerAddress = username
		}

		// Si aucune adresse n'est fournie, utiliser le nom d'utilisateur
		if req.MinerAddress == "" {
			req.MinerAddress = username
		}

		// Miner un nouveau bloc
		newBlock := bc.CreateBlock(req.MinerAddress)

		// Journaliser l'événement
		utils.LogAuditEvent(
			utils.EventTypeBlockMined,
			username,
			r.RemoteAddr,
			fmt.Sprintf("Bloc #%d miné avec succès", newBlock.Index),
			utils.RiskLow,
			map[string]interface{}{
				"block_index": newBlock.Index,
				"block_hash":  newBlock.Hash,
				"miner":       req.MinerAddress,
			},
		)

		// Renvoyer les informations sur le nouveau bloc
		sendAPIResponse(w, APIResponse{
			Success: true,
			Message: "Nouveau bloc miné avec succès",
			Data: map[string]interface{}{
				"blockHash":    newBlock.Hash,
				"blockIndex":   newBlock.Index,
				"transactions": len(newBlock.Transactions),
				"miningReward": bc.MiningReward,
				"timestamp":    newBlock.Timestamp,
				"difficulty":   newBlock.Difficulty,
			},
		})

	default:
		sendAPIResponse(w, APIResponse{
			Success:    false,
			Message:    "Méthode non autorisée",
			StatusCode: http.StatusMethodNotAllowed,
		})
	}
}

// handleWalletRequest traite les requêtes relatives aux portefeuilles
func handleWalletRequest(w http.ResponseWriter, r *http.Request, username string) {
	// Vérification de l'authentification
	if username == "" {
		sendAPIResponse(w, APIResponse{
			Success:    false,
			Message:    "Authentification requise",
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	// Router vers le gestionnaire de portefeuilles
	WalletHandler("wallets")(w, r)
}

// handleStatsRequest traite les requêtes de statistiques
func handleStatsRequest(w http.ResponseWriter, r *http.Request, bc *blockchain.Blockchain) {
	switch r.Method {
	case "GET":
		// Récupérer des statistiques générales sur la blockchain

		// Obtenir le nombre total de transactions
		txCount := 0
		for _, block := range bc.Blocks {
			txCount += len(block.Transactions)
		}

		// Calculer le hashrate à partir de la difficulté et du temps moyen entre les blocs
		hashrate := 0.0
		avgBlockTime := 0.0

		if len(bc.Blocks) > 1 {
			totalTime := bc.Blocks[len(bc.Blocks)-1].Timestamp.Sub(bc.Blocks[0].Timestamp)
			avgBlockTime = totalTime.Seconds() / float64(len(bc.Blocks)-1)

			// Formule approximative pour le hashrate
			if avgBlockTime > 0 {
				hashrate = float64(1<<uint(bc.MiningDifficulty)) / avgBlockTime
			}
		}

		// Renvoyer les statistiques
		sendAPIResponse(w, APIResponse{
			Success: true,
			Data: map[string]interface{}{
				"blockCount":        len(bc.Blocks),
				"transactionCount":  txCount,
				"pendingTxCount":    len(bc.MemPool.Transactions),
				"currentDifficulty": bc.MiningDifficulty,
				"miningReward":      bc.MiningReward,
				"averageBlockTime":  avgBlockTime,
				"estimatedHashrate": hashrate,
				"lastBlockHash":     bc.Blocks[len(bc.Blocks)-1].Hash,
			},
		})

	default:
		sendAPIResponse(w, APIResponse{
			Success:    false,
			Message:    "Méthode non autorisée",
			StatusCode: http.StatusMethodNotAllowed,
		})
	}
}

// handleAPIDocsRequest sert la documentation OpenAPI
func handleAPIDocsRequest(w http.ResponseWriter, r *http.Request) {
	// Endpoint dédié à fournir une interface Swagger ou similaire
	// Dans un vrai système, cela servirait une documentation OpenAPI complète

	apiDocs := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "CryptoChain Go API",
			"description": "API pour interagir avec la blockchain CryptoChain Go",
			"version":     "1.2.0",
		},
		"servers": []map[string]string{
			{"url": "/api", "description": "Serveur principal"},
		},
		"paths": map[string]interface{}{
			"/status": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Obtenir le statut du serveur",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Succès",
						},
					},
				},
			},
			"/blockchain": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Obtenir la blockchain complète",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Succès",
						},
					},
				},
			},
			// Autres endpoints seraient documentés ici
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiDocs)
}

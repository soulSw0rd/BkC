package handlers

import (
	"BkC/blockchain"
	"BkC/network"
	"BkC/utils"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

var networkManager *network.NetworkManager

// InitNetworkManager initialise le gestionnaire de réseau P2P
func InitNetworkManager(bc *blockchain.Blockchain, nodeURL string, isValidator bool) {
	networkManager = network.NewNetworkManager(nodeURL, bc, isValidator)

	// Démarrer la découverte des nœuds avec quelques seeds
	seedNodes := []string{
		"http://localhost:8080", // Ce nœud lui-même (ignoré par le manager)
		"http://localhost:8081", // Autres nœuds potentiels
		"http://localhost:8082",
	}

	networkManager.StartNodeDiscovery(seedNodes, 60) // Découverte toutes les 60 secondes
	networkManager.StartPeriodicSync(300)            // Synchronisation toutes les 5 minutes
}

// P2PHandler gère les communications P2P
func P2PHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.LogRequest(r)

		// Vérifier si le gestionnaire de réseau est initialisé
		if networkManager == nil {
			http.Error(w, "Gestionnaire de réseau non initialisé", http.StatusInternalServerError)
			return
		}

		if r.Method == "POST" && r.URL.Path == "/p2p/message" {
			// Traiter les messages P2P entrants
			handleP2PMessage(w, r)
			return
		} else if r.Method == "GET" && r.URL.Path == "/p2p/nodes" {
			// Renvoyer la liste des nœuds connus
			handleNodesRequest(w, r)
			return
		} else if r.Method == "POST" && r.URL.Path == "/p2p/node" {
			// Ajouter un nouveau nœud
			handleAddNodeRequest(w, r)
			return
		} else if r.Method == "DELETE" && r.URL.Path == "/p2p/node" {
			// Supprimer un nœud
			handleRemoveNodeRequest(w, r)
			return
		} else if r.Method == "GET" && r.URL.Path == "/p2p/sync" {
			// Déclencher une synchronisation manuelle
			handleSyncRequest(w, r)
			return
		}

		http.Error(w, "Méthode non autorisée ou endpoint P2P inconnu", http.StatusMethodNotAllowed)
	}
}

// handleP2PMessage traite les messages P2P entrants
func handleP2PMessage(w http.ResponseWriter, r *http.Request) {
	// Lire le corps de la requête
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Erreur lors de la lecture du corps de la requête", http.StatusBadRequest)
		return
	}

	// Décoder le message
	var message network.Message
	if err := json.Unmarshal(body, &message); err != nil {
		http.Error(w, "Erreur lors du décodage du message", http.StatusBadRequest)
		return
	}

	// Traiter le message
	response := networkManager.HandleMessage(message)

	// Renvoyer une réponse si nécessaire
	if response != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		// Réponse vide mais OK
		w.WriteHeader(http.StatusOK)
	}
}

// handleNodesRequest renvoie la liste des nœuds connus
func handleNodesRequest(w http.ResponseWriter, r *http.Request) {
	nodes := networkManager.GetKnownNodes()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"nodes":   nodes,
		"count":   len(nodes),
	})
}

// handleAddNodeRequest ajoute un nouveau nœud
func handleAddNodeRequest(w http.ResponseWriter, r *http.Request) {
	// Vérification de l'authentification
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Authentification requise", http.StatusUnauthorized)
		return
	}

	mu.Lock()
	session, exists := sessions[cookie.Value]
	mu.Unlock()

	if !exists || session == nil || !session.IsAdmin {
		http.Error(w, "Non autorisé", http.StatusForbidden)
		return
	}

	// Lire le corps de la requête
	var nodeData struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&nodeData); err != nil {
		http.Error(w, "Erreur lors du décodage de la requête", http.StatusBadRequest)
		return
	}

	if nodeData.URL == "" {
		http.Error(w, "URL du nœud manquante", http.StatusBadRequest)
		return
	}

	// Ajouter le nœud
	networkManager.AddNode(nodeData.URL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Nœud ajouté avec succès",
		"url":     nodeData.URL,
	})
}

// handleRemoveNodeRequest supprime un nœud
func handleRemoveNodeRequest(w http.ResponseWriter, r *http.Request) {
	// Vérification de l'authentification
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Authentification requise", http.StatusUnauthorized)
		return
	}

	mu.Lock()
	session, exists := sessions[cookie.Value]
	mu.Unlock()

	if !exists || session == nil || !session.IsAdmin {
		http.Error(w, "Non autorisé", http.StatusForbidden)
		return
	}

	// Lire le corps de la requête
	var nodeData struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&nodeData); err != nil {
		http.Error(w, "Erreur lors du décodage de la requête", http.StatusBadRequest)
		return
	}

	if nodeData.URL == "" {
		http.Error(w, "URL du nœud manquante", http.StatusBadRequest)
		return
	}

	// Supprimer le nœud
	networkManager.RemoveNode(nodeData.URL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Nœud supprimé avec succès",
		"url":     nodeData.URL,
	})
}

// handleSyncRequest déclenche une synchronisation manuelle
func handleSyncRequest(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Non autorisé", http.StatusForbidden)
		return
	}

	// Lancer la synchronisation en arrière-plan
	go networkManager.SyncWithNetwork()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Synchronisation démarrée",
	})
}

// AddBlockToNetworkHandler ajoute un nouveau bloc au réseau P2P
func AddBlockToNetworkHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.LogRequest(r)

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
			http.Error(w, "Non autorisé", http.StatusForbidden)
			return
		}

		// Vérifier que la méthode est POST
		if r.Method != "POST" {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		// Lire le corps de la requête
		var blockReq struct {
			MinerAddress string `json:"minerAddress"`
		}

		if err := json.NewDecoder(r.Body).Decode(&blockReq); err != nil {
			http.Error(w, "Erreur lors du décodage de la requête", http.StatusBadRequest)
			return
		}

		// Utiliser l'adresse du mineur ou celle de la session
		minerAddress := blockReq.MinerAddress
		if minerAddress == "" {
			minerAddress = session.Username
		}

		// Créer un nouveau bloc
		newBlock := bc.CreateBlock(minerAddress)

		// Diffuser le bloc sur le réseau P2P
		if networkManager != nil {
			go networkManager.BroadcastBlock(newBlock)
		} else {
			log.Println("Gestionnaire de réseau non initialisé, bloc non diffusé")
		}

		// Réponse
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":      true,
			"message":      "Nouveau bloc miné et diffusé avec succès",
			"blockHash":    newBlock.Hash,
			"blockIndex":   newBlock.Index,
			"transactions": len(newBlock.Transactions),
		})
	}
}

// AddTransactionToNetworkHandler ajoute une nouvelle transaction au réseau P2P
func AddTransactionToNetworkHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.LogRequest(r)

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
			http.Error(w, "Non autorisé", http.StatusForbidden)
			return
		}

		// Vérifier que la méthode est POST
		if r.Method != "POST" {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		// Lire le corps de la requête
		var txReq TransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&txReq); err != nil {
			http.Error(w, "Erreur lors du décodage de la requête", http.StatusBadRequest)
			return
		}

		// Validation des données
		if txReq.RecipientAddress == "" || txReq.Amount <= 0 {
			http.Error(w, "Adresse du destinataire et montant doivent être valides", http.StatusBadRequest)
			return
		}

		// Création d'un portefeuille à partir de la clé privée
		wallet, err := blockchain.LoadWalletFromString(txReq.SenderPrivateKey)
		if err != nil {
			http.Error(w, "Clé privée invalide: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Vérifier que l'adresse correspond
		if wallet.Address != txReq.SenderAddress {
			http.Error(w, "L'adresse ne correspond pas à la clé privée", http.StatusBadRequest)
			return
		}

		// Créer la transaction
		tx, err := wallet.CreateTransaction(txReq.RecipientAddress, txReq.Amount, txReq.Fee)
		if err != nil {
			http.Error(w, "Création de transaction échouée: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Ajouter à la blockchain
		if err := bc.AddTransaction(tx); err != nil {
			http.Error(w, "Transaction rejetée: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Diffuser la transaction sur le réseau P2P
		if networkManager != nil {
			go networkManager.BroadcastTransaction(tx)
		} else {
			log.Println("Gestionnaire de réseau non initialisé, transaction non diffusée")
		}

		// Réponse
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":       true,
			"message":       "Transaction créée et diffusée avec succès",
			"transactionId": tx.ID,
		})
	}
}

package network

import (
	"BkC/blockchain"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

// NodeInfo contient les informations sur un nœud du réseau
type NodeInfo struct {
	URL         string    `json:"url"`
	LastSeen    time.Time `json:"lastSeen"`
	BlockHeight int       `json:"blockHeight"`
	Version     string    `json:"version"`
	Status      string    `json:"status"`
}

// Message représente un message échangé entre les nœuds
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Sender  string          `json:"sender"`
	Time    time.Time       `json:"time"`
}

// NetworkManager gère la communication P2P entre les nœuds
type NetworkManager struct {
	NodeURL       string
	KnownNodes    map[string]*NodeInfo
	Blockchain    *blockchain.Blockchain
	IsValidator   bool
	Version       string
	mu            sync.RWMutex
	stopSync      chan struct{}
	isSyncing     bool
	pendingBlocks []*blockchain.Block
	pendingTxs    []*blockchain.Transaction
}

// NodeStatus représente les différents états d'un nœud
const (
	NodeStatusActive    = "ACTIVE"
	NodeStatusInactive  = "INACTIVE"
	NodeStatusSyncing   = "SYNCING"
	NodeStatusValidator = "VALIDATOR"
)

// MessageType représente les différents types de messages
const (
	MessageTypeBlock        = "BLOCK"
	MessageTypeTransaction  = "TRANSACTION"
	MessageTypeNodeInfo     = "NODE_INFO"
	MessageTypeBlockRequest = "BLOCK_REQUEST"
	MessageTypePeersList    = "PEERS_LIST"
	MessageTypePing         = "PING"
)

// NewNetworkManager crée une nouvelle instance de gestionnaire réseau
func NewNetworkManager(nodeURL string, bc *blockchain.Blockchain, isValidator bool) *NetworkManager {
	return &NetworkManager{
		NodeURL:       nodeURL,
		KnownNodes:    make(map[string]*NodeInfo),
		Blockchain:    bc,
		IsValidator:   isValidator,
		Version:       "1.0.0",
		stopSync:      make(chan struct{}),
		isSyncing:     false,
		pendingBlocks: []*blockchain.Block{},
		pendingTxs:    []*blockchain.Transaction{},
	}
}

// AddNode ajoute un nouveau nœud à la liste des nœuds connus
func (nm *NetworkManager) AddNode(url string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if url == nm.NodeURL {
		// Ne pas s'ajouter soi-même
		return
	}

	// Vérifier si le nœud existe déjà
	if _, exists := nm.KnownNodes[url]; !exists {
		// Ajouter le nouveau nœud
		nm.KnownNodes[url] = &NodeInfo{
			URL:      url,
			LastSeen: time.Time{}, // Jamais vu
			Status:   NodeStatusInactive,
		}

		// Envoyer un ping au nouveau nœud
		go nm.pingNode(url)
	}
}

// RemoveNode supprime un nœud de la liste des nœuds connus
func (nm *NetworkManager) RemoveNode(url string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	delete(nm.KnownNodes, url)
}

// GetKnownNodes retourne la liste des nœuds connus
func (nm *NetworkManager) GetKnownNodes() map[string]*NodeInfo {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	// Créer une copie pour éviter les problèmes de concurrence
	nodes := make(map[string]*NodeInfo)
	for url, info := range nm.KnownNodes {
		nodes[url] = &NodeInfo{
			URL:         info.URL,
			LastSeen:    info.LastSeen,
			BlockHeight: info.BlockHeight,
			Version:     info.Version,
			Status:      info.Status,
		}
	}

	return nodes
}

// BroadcastBlock diffuse un nouveau bloc à tous les nœuds connus
func (nm *NetworkManager) BroadcastBlock(block *blockchain.Block) {
	log.Printf("[P2P] Diffusion du bloc #%d aux pairs", block.Index)

	// Préparer le message
	blockData, _ := json.Marshal(block)
	message := Message{
		Type:    MessageTypeBlock,
		Payload: blockData,
		Sender:  nm.NodeURL,
		Time:    time.Now(),
	}

	// Diffuser le message à tous les nœuds connus
	nm.broadcastMessage(message)
}

// BroadcastTransaction diffuse une nouvelle transaction à tous les nœuds connus
func (nm *NetworkManager) BroadcastTransaction(tx *blockchain.Transaction) {
	log.Printf("[P2P] Diffusion de la transaction %s aux pairs", tx.ID)

	// Préparer le message
	txData, _ := json.Marshal(tx)
	message := Message{
		Type:    MessageTypeTransaction,
		Payload: txData,
		Sender:  nm.NodeURL,
		Time:    time.Now(),
	}

	// Diffuser le message à tous les nœuds connus
	nm.broadcastMessage(message)
}

// SyncWithNetwork synchronise la blockchain avec le réseau
func (nm *NetworkManager) SyncWithNetwork() {
	if nm.isSyncing {
		log.Println("[P2P] Synchronisation déjà en cours")
		return
	}

	nm.isSyncing = true
	log.Println("[P2P] Début de la synchronisation avec le réseau")

	// Récupérer la hauteur de bloc locale
	localHeight := len(nm.Blockchain.Blocks) - 1

	// Trouver le nœud avec la blockchain la plus longue
	var bestNode *NodeInfo
	var bestHeight int

	nm.mu.RLock()
	for _, node := range nm.KnownNodes {
		if node.Status == NodeStatusActive && node.BlockHeight > bestHeight {
			bestNode = node
			bestHeight = node.BlockHeight
		}
	}
	nm.mu.RUnlock()

	if bestNode == nil || bestHeight <= localHeight {
		log.Println("[P2P] Aucun nœud avec une blockchain plus longue trouvé")
		nm.isSyncing = false
		return
	}

	log.Printf("[P2P] Synchronisation avec %s (hauteur: %d)", bestNode.URL, bestHeight)

	// Demander les blocs manquants
	for i := localHeight + 1; i <= bestHeight; i++ {
		// Préparer la demande de bloc
		blockReq := struct {
			Height int `json:"height"`
		}{Height: i}

		reqData, _ := json.Marshal(blockReq)
		message := Message{
			Type:    MessageTypeBlockRequest,
			Payload: reqData,
			Sender:  nm.NodeURL,
			Time:    time.Now(),
		}

		// Envoyer la demande au nœud choisi
		messageData, _ := json.Marshal(message)
		resp, err := http.Post(bestNode.URL+"/p2p/message", "application/json", bytes.NewBuffer(messageData))

		if err != nil {
			log.Printf("[P2P] Erreur lors de la demande du bloc %d: %v", i, err)
			continue
		}

		// Traiter la réponse
		if resp.StatusCode == http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)
			var responseMsg Message

			if err := json.Unmarshal(body, &responseMsg); err != nil {
				log.Printf("[P2P] Erreur lors du décodage de la réponse: %v", err)
				resp.Body.Close()
				continue
			}

			if responseMsg.Type == MessageTypeBlock {
				var block blockchain.Block
				if err := json.Unmarshal(responseMsg.Payload, &block); err != nil {
					log.Printf("[P2P] Erreur lors du décodage du bloc: %v", err)
				} else {
					// Valider et ajouter le bloc
					nm.handleBlockMessage(&block, responseMsg.Sender)
				}
			}
		}

		resp.Body.Close()
	}

	log.Println("[P2P] Synchronisation terminée")
	nm.isSyncing = false
}

// broadcastMessage diffuse un message à tous les nœuds connus
func (nm *NetworkManager) broadcastMessage(message Message) {
	nm.mu.RLock()
	nodes := make([]string, 0, len(nm.KnownNodes))
	for url, node := range nm.KnownNodes {
		if node.Status == NodeStatusActive {
			nodes = append(nodes, url)
		}
	}
	nm.mu.RUnlock()

	// Préparer les données du message
	messageData, _ := json.Marshal(message)

	// Envoyer le message à chaque nœud
	for _, url := range nodes {
		go func(nodeURL string) {
			resp, err := http.Post(nodeURL+"/p2p/message", "application/json", bytes.NewBuffer(messageData))

			if err != nil {
				log.Printf("[P2P] Erreur lors de l'envoi du message à %s: %v", nodeURL, err)
				return
			}

			resp.Body.Close()
		}(url)
	}
}

// pingNode envoie un ping à un nœud pour vérifier s'il est actif
func (nm *NetworkManager) pingNode(url string) {
	// Préparer le message
	pingData, _ := json.Marshal(struct {
		NodeURL     string `json:"nodeUrl"`
		BlockHeight int    `json:"blockHeight"`
		Version     string `json:"version"`
		IsValidator bool   `json:"isValidator"`
	}{
		NodeURL:     nm.NodeURL,
		BlockHeight: len(nm.Blockchain.Blocks) - 1,
		Version:     nm.Version,
		IsValidator: nm.IsValidator,
	})

	message := Message{
		Type:    MessageTypePing,
		Payload: pingData,
		Sender:  nm.NodeURL,
		Time:    time.Now(),
	}

	// Encoder le message
	messageData, _ := json.Marshal(message)

	// Envoyer le ping
	resp, err := http.Post(url+"/p2p/message", "application/json", bytes.NewBuffer(messageData))

	if err != nil {
		log.Printf("[P2P] Nœud %s injoignable: %v", url, err)
		return
	}

	defer resp.Body.Close()

	// Mettre à jour l'état du nœud
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if _, exists := nm.KnownNodes[url]; exists {
		nm.KnownNodes[url].LastSeen = time.Now()
		nm.KnownNodes[url].Status = NodeStatusActive
	}
}

// RequestPeersList demande la liste des pairs à un nœud
func (nm *NetworkManager) RequestPeersList(nodeURL string) {
	// Préparer le message
	message := Message{
		Type:    MessageTypePeersList,
		Payload: []byte("{}"),
		Sender:  nm.NodeURL,
		Time:    time.Now(),
	}

	// Encoder le message
	messageData, _ := json.Marshal(message)

	// Envoyer la demande
	resp, err := http.Post(nodeURL+"/p2p/message", "application/json", bytes.NewBuffer(messageData))

	if err != nil {
		log.Printf("[P2P] Erreur lors de la demande de la liste des pairs à %s: %v", nodeURL, err)
		return
	}

	defer resp.Body.Close()

	// Traiter la réponse
	body, _ := ioutil.ReadAll(resp.Body)
	var responseMsg Message

	if err := json.Unmarshal(body, &responseMsg); err != nil {
		log.Printf("[P2P] Erreur lors du décodage de la réponse: %v", err)
		return
	}

	if responseMsg.Type == MessageTypePeersList {
		var peers []string
		if err := json.Unmarshal(responseMsg.Payload, &peers); err != nil {
			log.Printf("[P2P] Erreur lors du décodage de la liste des pairs: %v", err)
			return
		}

		// Ajouter les nouveaux pairs
		for _, peer := range peers {
			nm.AddNode(peer)
		}
	}
}

// HandleMessage traite les messages reçus des autres nœuds
func (nm *NetworkManager) HandleMessage(message Message) interface{} {
	switch message.Type {
	case MessageTypeBlock:
		var block blockchain.Block
		if err := json.Unmarshal(message.Payload, &block); err != nil {
			log.Printf("[P2P] Erreur lors du décodage du bloc: %v", err)
			return nil
		}
		return nm.handleBlockMessage(&block, message.Sender)

	case MessageTypeTransaction:
		var tx blockchain.Transaction
		if err := json.Unmarshal(message.Payload, &tx); err != nil {
			log.Printf("[P2P] Erreur lors du décodage de la transaction: %v", err)
			return nil
		}
		return nm.handleTransactionMessage(&tx, message.Sender)

	case MessageTypeNodeInfo:
		var nodeInfo NodeInfo
		if err := json.Unmarshal(message.Payload, &nodeInfo); err != nil {
			log.Printf("[P2P] Erreur lors du décodage des infos du nœud: %v", err)
			return nil
		}
		return nm.handleNodeInfoMessage(&nodeInfo, message.Sender)

	case MessageTypeBlockRequest:
		var blockReq struct {
			Height int `json:"height"`
		}
		if err := json.Unmarshal(message.Payload, &blockReq); err != nil {
			log.Printf("[P2P] Erreur lors du décodage de la demande de bloc: %v", err)
			return nil
		}
		return nm.handleBlockRequestMessage(blockReq.Height, message.Sender)

	case MessageTypePeersList:
		return nm.handlePeersListMessage(message.Sender)

	case MessageTypePing:
		var pingData struct {
			NodeURL     string `json:"nodeUrl"`
			BlockHeight int    `json:"blockHeight"`
			Version     string `json:"version"`
			IsValidator bool   `json:"isValidator"`
		}
		if err := json.Unmarshal(message.Payload, &pingData); err != nil {
			log.Printf("[P2P] Erreur lors du décodage du ping: %v", err)
			return nil
		}
		return nm.handlePingMessage(pingData, message.Sender)

	default:
		log.Printf("[P2P] Type de message inconnu: %s", message.Type)
		return nil
	}
}

// handleBlockMessage traite un message de bloc
func (nm *NetworkManager) handleBlockMessage(block *blockchain.Block, sender string) interface{} {
	// Vérifier si le bloc est déjà dans la blockchain
	for _, b := range nm.Blockchain.Blocks {
		if b.Hash == block.Hash {
			log.Printf("[P2P] Bloc #%d déjà présent dans la blockchain", block.Index)
			return nil
		}
	}

	// Vérifier si l'index du bloc est cohérent
	currentHeight := len(nm.Blockchain.Blocks) - 1
	if block.Index > currentHeight+1 {
		// Bloc trop avancé, le mettre en attente
		log.Printf("[P2P] Bloc #%d reçu trop tôt, mise en attente", block.Index)
		nm.pendingBlocks = append(nm.pendingBlocks, block)

		// Lancer une synchronisation pour récupérer les blocs manquants
		go nm.SyncWithNetwork()
		return nil
	} else if block.Index <= currentHeight {
		// Bloc déjà traité ou fork potentiel
		log.Printf("[P2P] Bloc #%d ignoré (hauteur actuelle: %d)", block.Index, currentHeight)
		return nil
	}

	// Vérifier que le bloc est valide
	if block.PrevHash != nm.Blockchain.Blocks[currentHeight].Hash {
		log.Printf("[P2P] Bloc #%d invalide: hash précédent incorrect", block.Index)
		return nil
	}

	// Valider le bloc
	if !block.VerifyProofOfWork() {
		log.Printf("[P2P] Bloc #%d invalide: preuve de travail incorrecte", block.Index)
		return nil
	}

	// Ajouter le bloc à la blockchain
	nm.Blockchain.Blocks = append(nm.Blockchain.Blocks, block)
	log.Printf("[P2P] Bloc #%d ajouté à la blockchain", block.Index)

	// Traiter les blocs en attente
	nm.processPendingBlocks()

	// Mettre à jour le nœud émetteur
	nm.mu.Lock()
	if node, exists := nm.KnownNodes[sender]; exists {
		node.BlockHeight = block.Index
		node.LastSeen = time.Now()
	}
	nm.mu.Unlock()

	return nil
}

// handleTransactionMessage traite un message de transaction
func (nm *NetworkManager) handleTransactionMessage(tx *blockchain.Transaction, sender string) interface{} {
	// Vérifier si la transaction est déjà dans le mempool
	if _, exists := nm.Blockchain.MemPool.Transactions[tx.ID]; exists {
		log.Printf("[P2P] Transaction %s déjà présente dans le mempool", tx.ID)
		return nil
	}

	// Valider la transaction
	if !tx.Verify() {
		log.Printf("[P2P] Transaction %s invalide", tx.ID)
		return nil
	}

	// Ajouter la transaction au mempool
	if err := nm.Blockchain.AddTransaction(tx); err != nil {
		log.Printf("[P2P] Erreur lors de l'ajout de la transaction %s: %v", tx.ID, err)
		return nil
	}

	log.Printf("[P2P] Transaction %s ajoutée au mempool", tx.ID)

	// Mettre à jour le nœud émetteur
	nm.mu.Lock()
	if node, exists := nm.KnownNodes[sender]; exists {
		node.LastSeen = time.Now()
	}
	nm.mu.Unlock()

	return nil
}

// handleNodeInfoMessage traite un message d'information sur un nœud
func (nm *NetworkManager) handleNodeInfoMessage(nodeInfo *NodeInfo, sender string) interface{} {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Mettre à jour les informations du nœud
	if node, exists := nm.KnownNodes[sender]; exists {
		node.LastSeen = time.Now()
		node.BlockHeight = nodeInfo.BlockHeight
		node.Version = nodeInfo.Version
		node.Status = nodeInfo.Status
	} else {
		// Ajouter le nœud s'il n'existe pas encore
		nm.KnownNodes[sender] = &NodeInfo{
			URL:         sender,
			LastSeen:    time.Now(),
			BlockHeight: nodeInfo.BlockHeight,
			Version:     nodeInfo.Version,
			Status:      nodeInfo.Status,
		}
	}

	return nil
}

// handleBlockRequestMessage traite une demande de bloc
func (nm *NetworkManager) handleBlockRequestMessage(height int, sender string) interface{} {
	// Vérifier si le bloc demandé existe
	if height < 0 || height >= len(nm.Blockchain.Blocks) {
		log.Printf("[P2P] Bloc #%d demandé par %s introuvable", height, sender)
		return nil
	}

	// Récupérer le bloc demandé
	block := nm.Blockchain.Blocks[height]

	// Préparer la réponse
	blockData, _ := json.Marshal(block)
	message := Message{
		Type:    MessageTypeBlock,
		Payload: blockData,
		Sender:  nm.NodeURL,
		Time:    time.Now(),
	}

	// Mettre à jour le nœud émetteur
	nm.mu.Lock()
	if node, exists := nm.KnownNodes[sender]; exists {
		node.LastSeen = time.Now()
	}
	nm.mu.Unlock()

	return message
}

// handlePeersListMessage traite une demande de liste des pairs
func (nm *NetworkManager) handlePeersListMessage(sender string) interface{} {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	// Construire la liste des pairs actifs
	peers := make([]string, 0)
	for url, node := range nm.KnownNodes {
		if node.Status == NodeStatusActive && url != sender {
			peers = append(peers, url)
		}
	}

	// Ajouter également ce nœud
	peers = append(peers, nm.NodeURL)

	// Préparer la réponse
	peersData, _ := json.Marshal(peers)
	message := Message{
		Type:    MessageTypePeersList,
		Payload: peersData,
		Sender:  nm.NodeURL,
		Time:    time.Now(),
	}

	// Mettre à jour le nœud émetteur
	if node, exists := nm.KnownNodes[sender]; exists {
		node.LastSeen = time.Now()
	}

	return message
}

// handlePingMessage traite un message de ping
func (nm *NetworkManager) handlePingMessage(pingData struct {
	NodeURL     string `json:"nodeUrl"`
	BlockHeight int    `json:"blockHeight"`
	Version     string `json:"version"`
	IsValidator bool   `json:"isValidator"`
}, sender string) interface{} {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Mettre à jour les informations du nœud
	nodeStatus := NodeStatusActive
	if pingData.IsValidator {
		nodeStatus = NodeStatusValidator
	}

	if node, exists := nm.KnownNodes[pingData.NodeURL]; exists {
		node.LastSeen = time.Now()
		node.BlockHeight = pingData.BlockHeight
		node.Version = pingData.Version
		node.Status = nodeStatus
	} else {
		// Ajouter le nœud s'il n'existe pas encore
		nm.KnownNodes[pingData.NodeURL] = &NodeInfo{
			URL:         pingData.NodeURL,
			LastSeen:    time.Now(),
			BlockHeight: pingData.BlockHeight,
			Version:     pingData.Version,
			Status:      nodeStatus,
		}
	}

	// Préparer une réponse (pong)
	pongData, _ := json.Marshal(struct {
		NodeURL     string `json:"nodeUrl"`
		BlockHeight int    `json:"blockHeight"`
		Version     string `json:"version"`
		IsValidator bool   `json:"isValidator"`
	}{
		NodeURL:     nm.NodeURL,
		BlockHeight: len(nm.Blockchain.Blocks) - 1,
		Version:     nm.Version,
		IsValidator: nm.IsValidator,
	})

	message := Message{
		Type:    MessageTypeNodeInfo,
		Payload: pongData,
		Sender:  nm.NodeURL,
		Time:    time.Now(),
	}

	return message
}

// processPendingBlocks traite les blocs en attente
func (nm *NetworkManager) processPendingBlocks() {
	// Trier les blocs en attente par index
	for len(nm.pendingBlocks) > 0 {
		var nextBlock *blockchain.Block
		nextIndex := len(nm.Blockchain.Blocks)

		// Chercher le prochain bloc
		for i, block := range nm.pendingBlocks {
			if block.Index == nextIndex {
				nextBlock = block
				// Retirer le bloc de la liste des blocs en attente
				nm.pendingBlocks = append(nm.pendingBlocks[:i], nm.pendingBlocks[i+1:]...)
				break
			}
		}

		if nextBlock == nil {
			// Aucun bloc correspondant au prochain index
			break
		}

		// Vérifier que le bloc est valide
		if nextBlock.PrevHash != nm.Blockchain.Blocks[nextIndex-1].Hash {
			log.Printf("[P2P] Bloc #%d en attente invalide: hash précédent incorrect", nextBlock.Index)
			continue
		}

		// Valider le bloc
		if !nextBlock.VerifyProofOfWork() {
			log.Printf("[P2P] Bloc #%d en attente invalide: preuve de travail incorrecte", nextBlock.Index)
			continue
		}

		// Ajouter le bloc à la blockchain
		nm.Blockchain.Blocks = append(nm.Blockchain.Blocks, nextBlock)
		log.Printf("[P2P] Bloc #%d en attente ajouté à la blockchain", nextBlock.Index)
	}
}

// StartNodeDiscovery démarre la découverte des nœuds
func (nm *NetworkManager) StartNodeDiscovery(seedNodes []string, interval time.Duration) {
	// Ajouter les nœuds seeds
	for _, url := range seedNodes {
		nm.AddNode(url)
	}

	// Démarrer la routine de découverte
	go func() {
		for {
			// Ping tous les nœuds connus
			nm.pingAllNodes()

			// Demander la liste des pairs à quelques nœuds actifs
			nm.requestPeersFromActiveNodes()

			// Attendre avant la prochaine découverte
			time.Sleep(interval)
		}
	}()
}

// pingAllNodes ping tous les nœuds connus
func (nm *NetworkManager) pingAllNodes() {
	nm.mu.RLock()
	nodes := make([]string, 0, len(nm.KnownNodes))
	for url := range nm.KnownNodes {
		nodes = append(nodes, url)
	}
	nm.mu.RUnlock()

	for _, url := range nodes {
		go nm.pingNode(url)
	}
}

// requestPeersFromActiveNodes demande la liste des pairs à quelques nœuds actifs
func (nm *NetworkManager) requestPeersFromActiveNodes() {
	nm.mu.RLock()
	activeNodes := make([]string, 0)
	for url, node := range nm.KnownNodes {
		if node.Status == NodeStatusActive || node.Status == NodeStatusValidator {
			activeNodes = append(activeNodes, url)
		}
	}
	nm.mu.RUnlock()

	// Choisir un ou plusieurs nœuds actifs au hasard
	if len(activeNodes) > 0 {
		// Pour simplifier, nous demandons au premier nœud actif
		go nm.RequestPeersList(activeNodes[0])
	}
}

// StartPeriodicSync démarre la synchronisation périodique avec le réseau
func (nm *NetworkManager) StartPeriodicSync(interval time.Duration) {
	go func() {
		for {
			select {
			case <-nm.stopSync:
				return
			default:
				nm.SyncWithNetwork()
				time.Sleep(interval)
			}
		}
	}()
}

// StopPeriodicSync arrête la synchronisation périodique
func (nm *NetworkManager) StopPeriodicSync() {
	nm.stopSync <- struct{}{}
}

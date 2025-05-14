package blockchain

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

// MiningData structure pour enregistrer les informations de minage
type MiningData struct {
	Miner      string    `json:"miner"`       // Nom d'utilisateur du mineur
	Content    string    `json:"content"`     // Contenu du bloc
	Timestamp  time.Time `json:"timestamp"`   // Heure de minage
	Difficulty int       `json:"difficulty"`  // Difficulté de minage
	Duration   int64     `json:"duration_ms"` // Durée du minage en millisecondes
	Nonce      int       `json:"nonce"`       // Nonce final
}

// BlockUpdate représente une mise à jour de bloc pour les notifications
type BlockUpdate struct {
	Block *Block
	Type  string // "new" pour nouveau bloc, "validate" pour validation
	Miner string // Nom d'utilisateur du mineur (si applicable)
}

// Blockchain représente la chaîne de blocs
type Blockchain struct {
	Blocks        []*Block
	mu            sync.RWMutex // Utilisez RWMutex pour permettre des lectures concurrentes
	updateChannel chan BlockUpdate
	subscribers   []chan BlockUpdate
	subMutex      sync.RWMutex
}

type block struct {
	Index     int
	Timestamp string
	Data      string
	PrevHash  string
	Hash      string
	Nonce     int
}

// CreateGenesisBlock crée le premier bloc (genesis block)
func CreateGenesisBlock() *Block {
	block := &Block{
		Index:     0,
		Timestamp: time.Now().String(),
		Data:      "Genesis Block",
		PrevHash:  "",
		Nonce:     0,
	}
	block.Hash = block.ComputeHash()
	block.ProofOfWork(4) // Par défaut, difficulté = 4
	return block
}

// NewBlockchain initialise une nouvelle blockchain
func NewBlockchain() *Blockchain {
	bc := &Blockchain{
		Blocks:        []*Block{CreateGenesisBlock()},
		updateChannel: make(chan BlockUpdate, 100), // Buffer de 100 pour éviter le blocage
		subscribers:   make([]chan BlockUpdate, 0),
	}

	// Essayer de charger la blockchain depuis un fichier
	bc.LoadFromFile()

	// Démarrer la goroutine pour traiter les mises à jour
	go bc.processUpdates()

	return bc
}

// processUpdates traite les mises à jour et les envoie aux abonnés
func (bc *Blockchain) processUpdates() {
	for update := range bc.updateChannel {
		// Envoyer l'update à tous les abonnés
		bc.subMutex.RLock()
		for _, subscriber := range bc.subscribers {
			select {
			case subscriber <- update:
				// Message envoyé avec succès
			default:
				// Canal plein, on ignore
			}
		}
		bc.subMutex.RUnlock()
	}
}

// Subscribe permet de s'abonner aux mises à jour de la blockchain
func (bc *Blockchain) Subscribe() chan BlockUpdate {
	subscriber := make(chan BlockUpdate, 10)

	bc.subMutex.Lock()
	bc.subscribers = append(bc.subscribers, subscriber)
	bc.subMutex.Unlock()

	return subscriber
}

// Unsubscribe permet de se désabonner des mises à jour
func (bc *Blockchain) Unsubscribe(subscriber chan BlockUpdate) {
	bc.subMutex.Lock()
	defer bc.subMutex.Unlock()

	for i, sub := range bc.subscribers {
		if sub == subscriber {
			// Remplacer l'élément à supprimer par le dernier élément
			bc.subscribers[i] = bc.subscribers[len(bc.subscribers)-1]
			// Réduire la taille du slice
			bc.subscribers = bc.subscribers[:len(bc.subscribers)-1]
			close(sub)
			break
		}
	}
}

// AddBlockWithMiner ajoute un nouveau bloc à la blockchain avec informations sur le mineur
func (bc *Blockchain) AddBlockWithMiner(data string, difficulty int, miner string) *Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Créer un objet MiningData
	miningData := MiningData{
		Miner:      miner,
		Content:    data,
		Timestamp:  time.Now(),
		Difficulty: difficulty,
	}

	startTime := time.Now()

	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := &Block{
		Index:     len(bc.Blocks),
		Timestamp: miningData.Timestamp.String(),
		Data:      fmt.Sprintf("%s", data),
		PrevHash:  prevBlock.Hash,
		Nonce:     0,
		Miner:     miner,
	}

	// Exécuter la preuve de travail
	newBlock.ProofOfWork(difficulty)

	// Mettre à jour les données de minage
	miningData.Duration = time.Since(startTime).Milliseconds()
	miningData.Nonce = newBlock.Nonce

	// Sérialiser les données de minage
	miningJson, _ := json.Marshal(miningData)
	newBlock.MiningInfo = string(miningJson)

	bc.Blocks = append(bc.Blocks, newBlock)

	// Envoyer une notification de mise à jour
	bc.updateChannel <- BlockUpdate{
		Block: newBlock,
		Type:  "new",
		Miner: miner,
	}

	// Sauvegarde automatique après ajout d'un bloc
	bc.SaveToFile()

	return newBlock
}

// AddBlockWithMinerAsync ajoute un bloc de manière asynchrone avec information sur le mineur
func (bc *Blockchain) AddBlockWithMinerAsync(data string, difficulty int, miner string) {
	go func() {
		bc.AddBlockWithMiner(data, difficulty, miner)
	}()
}

// GetBlocksByMiner récupère tous les blocs minés par un utilisateur spécifique
func (bc *Blockchain) GetBlocksByMiner(miner string) []*Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	var minerBlocks []*Block

	for _, block := range bc.Blocks {
		if block.Miner == miner {
			minerBlocks = append(minerBlocks, block)
		}
	}

	return minerBlocks
}

// IsMiningBlock vérifie si un bloc a été miné (commence par "Miné par")
func (bc *Blockchain) IsMiningBlock(block *Block) bool {
	return strings.HasPrefix(block.Data, "Miné par")
}

// AddBlock ajoute un nouveau bloc à la blockchain
func (bc *Blockchain) AddBlock(data string, difficulty int) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := &Block{
		Index:     len(bc.Blocks),
		Timestamp: time.Now().String(),
		Data:      data,
		PrevHash:  prevBlock.Hash,
		Nonce:     0,
	}
	newBlock.ProofOfWork(difficulty)
	bc.Blocks = append(bc.Blocks, newBlock)

	// Envoyer une notification de mise à jour
	bc.updateChannel <- BlockUpdate{
		Block: newBlock,
		Type:  "new",
	}

	// Sauvegarde automatique après ajout d'un bloc
	bc.SaveToFile()
}

// AddBlockAsync ajoute un bloc de manière asynchrone
func (bc *Blockchain) AddBlockAsync(data string, difficulty int) {
	go func() {
		bc.AddBlock(data, difficulty)
	}()
}

// AddMessageBlock ajoute un message entre utilisateurs comme un nouveau bloc
func (bc *Blockchain) AddMessageBlock(message Message, difficulty int) {
	// Convertir le message en JSON
	messageData, err := json.Marshal(message)
	if err != nil {
		return // Gestion d'erreur simple pour l'exemple
	}

	// Ajouter le bloc avec les données du message
	bc.AddBlock(string(messageData), difficulty)
}

// AddMessageBlockAsync ajoute un message de manière asynchrone
func (bc *Blockchain) AddMessageBlockAsync(message Message, difficulty int) {
	go func() {
		bc.AddMessageBlock(message, difficulty)
	}()
}

// GetMessageBlocks retourne tous les blocs qui contiennent des messages
func (bc *Blockchain) GetMessageBlocks() []Message {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	var messages []Message

	// Parcourir tous les blocs sauf le genesis
	for i := 1; i < len(bc.Blocks); i++ {
		var message Message
		// Essayer de décoder le bloc comme un message
		err := json.Unmarshal([]byte(bc.Blocks[i].Data), &message)
		if err == nil && message.ID != "" {
			messages = append(messages, message)
		}
	}

	return messages
}

// GetUserMessages retourne les messages envoyés ou reçus par un utilisateur
func (bc *Blockchain) GetUserMessages(username string) []Message {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	var userMessages []Message

	// Parcourir tous les blocs sauf le genesis
	for i := 1; i < len(bc.Blocks); i++ {
		var message Message
		// Essayer de décoder le bloc comme un message
		err := json.Unmarshal([]byte(bc.Blocks[i].Data), &message)
		if err == nil && message.ID != "" {
			// Ajouter uniquement les messages envoyés ou reçus par l'utilisateur
			if message.Sender == username || message.Recipient == username {
				userMessages = append(userMessages, message)
			}
		}
	}

	return userMessages
}

// SaveToFile sauvegarde la blockchain dans un fichier
func (bc *Blockchain) SaveToFile() error {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Convertir la blockchain en JSON
	data, err := json.MarshalIndent(bc.Blocks, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur lors de la sérialisation de la blockchain: %v", err)
	}

	// Écrire dans un fichier
	err = ioutil.WriteFile("blockchain_data.json", data, 0644)
	if err != nil {
		return fmt.Errorf("erreur lors de l'écriture de la blockchain: %v", err)
	}

	return nil
}

// LoadFromFile charge la blockchain depuis un fichier
func (bc *Blockchain) LoadFromFile() error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Vérifier si le fichier existe
	if _, err := os.Stat("blockchain_data.json"); os.IsNotExist(err) {
		return nil // Le fichier n'existe pas, utiliser la blockchain par défaut
	}

	// Lire le fichier
	data, err := ioutil.ReadFile("blockchain_data.json")
	if err != nil {
		return fmt.Errorf("erreur lors de la lecture du fichier blockchain: %v", err)
	}

	// Désérialiser les données
	var blocks []*Block
	err = json.Unmarshal(data, &blocks)
	if err != nil {
		return fmt.Errorf("erreur lors de la désérialisation de la blockchain: %v", err)
	}

	// Vérifier la validité de la chaîne
	if len(blocks) > 0 {
		bc.Blocks = blocks
	}

	return nil
}

// GetStats retourne les statistiques en JSON
func (bc *Blockchain) GetStats() string {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	stats := map[string]interface{}{
		"numBlocks": len(bc.Blocks),
	}
	jsonStats, _ := json.Marshal(stats)
	return string(jsonStats)
}

// Expose Lock et Unlock pour permettre des verrous explicites
func (bc *Blockchain) Lock() {
	bc.mu.Lock()
}

func (bc *Blockchain) Unlock() {
	bc.mu.Unlock()
}

// getLastHash retourne le hash du dernier bloc ou une valeur vide.
func (bc *Blockchain) getLastHash() string {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if len(bc.Blocks) == 0 {
		return ""
	}
	return bc.Blocks[len(bc.Blocks)-1].Hash
}

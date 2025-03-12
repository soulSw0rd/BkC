package blockchain

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Blockchain représente la chaîne de blocs
type Blockchain struct {
	Blocks          []*Block           `json:"blocks"`
	PendingTxs      []*Transaction     `json:"pending_transactions"`
	Difficulty      int                `json:"difficulty"`
	MiningReward    float64            `json:"mining_reward"`
	Balances        map[string]float64 `json:"balances"`
	LastBlockTime   time.Time          `json:"last_block_time"`
	MaxBlockSize    int                `json:"max_block_size"`
	BlockTimeTarget time.Duration      `json:"block_time_target"`
	mu              sync.RWMutex
}

// BlockchainStats contient les statistiques de la blockchain
type BlockchainStats struct {
	BlockCount        int       `json:"block_count"`
	TransactionCount  int       `json:"transaction_count"`
	PendingTxCount    int       `json:"pending_tx_count"`
	LastBlockTime     time.Time `json:"last_block_time"`
	CurrentDifficulty int       `json:"current_difficulty"`
	AverageBlockTime  float64   `json:"average_block_time"`
	HashRate          float64   `json:"hash_rate"`
}

// CreateGenesisBlock crée le premier bloc (genesis block)
func CreateGenesisBlock() *Block {
	genesisTime := time.Now()
	block := &Block{
		Index:        0,
		Timestamp:    genesisTime,
		Transactions: []*Transaction{},
		Data:         "Genesis Block",
		PrevHash:     "",
		Nonce:        0,
		Miner:        "system",
		Difficulty:   2,
	}

	block.ProofOfWork(block.Difficulty)
	return block
}

// NewBlockchain initialise une nouvelle blockchain
func NewBlockchain() *Blockchain {
	return &Blockchain{
		Blocks:          []*Block{CreateGenesisBlock()},
		PendingTxs:      []*Transaction{},
		Difficulty:      4,
		MiningReward:    10.0,
		Balances:        map[string]float64{"system": 1000.0},
		LastBlockTime:   time.Now(),
		MaxBlockSize:    10,
		BlockTimeTarget: 30 * time.Second,
	}
}

// AddTransaction ajoute une transaction à la liste d'attente
func (bc *Blockchain) AddTransaction(tx *Transaction) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Valider la transaction
	if !tx.IsValid() {
		return fmt.Errorf("transaction invalide")
	}

	// Vérifier le solde (sauf pour les transactions système)
	if tx.From != "system" {
		balance, exists := bc.Balances[tx.From]
		if !exists || balance < tx.Amount {
			return fmt.Errorf("solde insuffisant pour %s", tx.From)
		}
	}

	// Vérifier si l'ID existe déjà
	for _, pendingTx := range bc.PendingTxs {
		if pendingTx.ID == tx.ID {
			return fmt.Errorf("transaction avec ID %s existe déjà", tx.ID)
		}
	}

	// Vérifier dans les blocs existants
	for _, block := range bc.Blocks {
		for _, blockTx := range block.Transactions {
			if blockTx.ID == tx.ID {
				return fmt.Errorf("transaction avec ID %s existe déjà dans un bloc", tx.ID)
			}
		}
	}

	bc.PendingTxs = append(bc.PendingTxs, tx)
	return nil
}

// MineBlock crée un nouveau bloc avec les transactions en attente
func (bc *Blockchain) MineBlock(minerAddress string, customData string) (*Block, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if len(bc.Blocks) == 0 {
		return nil, fmt.Errorf("blockchain vide, impossible de miner un bloc")
	}

	// Déterminer combien de transactions inclure
	txCount := len(bc.PendingTxs)
	if txCount > bc.MaxBlockSize {
		txCount = bc.MaxBlockSize
	}

	// Préparer les transactions pour le bloc
	transactions := make([]*Transaction, 0, txCount+1) // +1 pour la récompense
	if txCount > 0 {
		transactions = append(transactions, bc.PendingTxs[:txCount]...)
	}

	// Ajouter la transaction de récompense
	rewardTx := NewTransaction("system", minerAddress, bc.MiningReward)
	transactions = append(transactions, rewardTx)

	// Créer le bloc
	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := &Block{
		Index:        len(bc.Blocks),
		Timestamp:    time.Now(),
		Transactions: transactions,
		Data:         customData,
		PrevHash:     prevBlock.Hash,
		Nonce:        0,
		Miner:        minerAddress,
	}

	// Miner le bloc (proof of work)
	miningTime := newBlock.ProofOfWork(bc.Difficulty)

	// Ajouter le bloc à la chaîne
	bc.Blocks = append(bc.Blocks, newBlock)

	// Ajuster la difficulté si nécessaire
	bc.adjustDifficulty(miningTime)

	// Mettre à jour les soldes
	for _, tx := range transactions {
		if tx.From != "system" {
			bc.Balances[tx.From] -= tx.Amount
		}

		// S'assurer que le destinataire a une entrée dans la map des soldes
		if _, exists := bc.Balances[tx.To]; !exists {
			bc.Balances[tx.To] = 0
		}

		bc.Balances[tx.To] += tx.Amount
	}

	// Supprimer les transactions traitées de la liste d'attente
	if txCount > 0 {
		bc.PendingTxs = bc.PendingTxs[txCount:]
	}

	// Mettre à jour le temps du dernier bloc
	bc.LastBlockTime = newBlock.Timestamp

	return newBlock, nil
}

// adjustDifficulty ajuste la difficulté en fonction du temps de minage
func (bc *Blockchain) adjustDifficulty(lastMiningTime time.Duration) {
	// Si le minage était trop rapide, augmenter la difficulté
	if lastMiningTime < bc.BlockTimeTarget/2 {
		bc.Difficulty++
		return
	}

	// Si le minage était trop lent, diminuer la difficulté (mais pas en dessous de 1)
	if lastMiningTime > bc.BlockTimeTarget*2 && bc.Difficulty > 1 {
		bc.Difficulty--
	}
}

// GetStats retourne les statistiques de la blockchain
func (bc *Blockchain) GetStats() BlockchainStats {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	stats := BlockchainStats{
		BlockCount:        len(bc.Blocks),
		TransactionCount:  0,
		PendingTxCount:    len(bc.PendingTxs),
		LastBlockTime:     bc.LastBlockTime,
		CurrentDifficulty: bc.Difficulty,
	}

	// Calculer le nombre total de transactions
	var totalTime time.Duration
	var totalTx int

	for i, block := range bc.Blocks {
		txCount := len(block.Transactions)
		stats.TransactionCount += txCount
		totalTx += txCount

		if i > 0 {
			timeDiff := block.Timestamp.Sub(bc.Blocks[i-1].Timestamp)
			totalTime += timeDiff
		}
	}

	// Calculer le temps moyen entre les blocs
	if len(bc.Blocks) > 1 {
		stats.AverageBlockTime = totalTime.Seconds() / float64(len(bc.Blocks)-1)
	}

	// Estimation du taux de hachage basée sur la difficulté et le temps moyen
	if stats.AverageBlockTime > 0 {
		// Formule approximative: 2^difficulté / temps_moyen_bloc
		stats.HashRate = math.Pow(2, float64(bc.Difficulty)) / stats.AverageBlockTime
	}

	return stats
}

// GetBalance retourne le solde d'une adresse
func (bc *Blockchain) GetBalance(address string) float64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	balance, exists := bc.Balances[address]
	if !exists {
		return 0
	}
	return balance
}

// ValidateChain vérifie l'intégrité de toute la blockchain
func (bc *Blockchain) ValidateChain() (bool, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Vérifier que la blockchain contient au moins un bloc
	if len(bc.Blocks) == 0 {
		return false, fmt.Errorf("blockchain vide")
	}

	// Parcourir tous les blocs à partir du deuxième
	for i := 1; i < len(bc.Blocks); i++ {
		currentBlock := bc.Blocks[i]
		previousBlock := bc.Blocks[i-1]

		// Vérifier que l'index est correct
		if currentBlock.Index != previousBlock.Index+1 {
			return false, fmt.Errorf("index du bloc %d invalide", i)
		}

		// Vérifier que le hash précédent correspond
		if currentBlock.PrevHash != previousBlock.Hash {
			return false, fmt.Errorf("hash précédent invalide pour le bloc %d", i)
		}

		// Recalculer le hash pour vérifier
		calculatedHash := currentBlock.ComputeHash()
		if calculatedHash != currentBlock.Hash {
			return false, fmt.Errorf("hash invalide pour le bloc %d", i)
		}

		// Vérifier la preuve de travail
		if !currentBlock.ValidateProofOfWork() {
			return false, fmt.Errorf("preuve de travail invalide pour le bloc %d", i)
		}
	}

	return true, nil
}

// SaveToFile sauvegarde la blockchain dans un fichier JSON
func (bc *Blockchain) SaveToFile(filename string) error {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Créer le répertoire parent si nécessaire
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("impossible de créer le répertoire %s: %w", dir, err)
	}

	// Structure temporaire pour la sérialisation
	data := struct {
		Blocks          []*Block           `json:"blocks"`
		PendingTxs      []*Transaction     `json:"pending_transactions"`
		Difficulty      int                `json:"difficulty"`
		MiningReward    float64            `json:"mining_reward"`
		Balances        map[string]float64 `json:"balances"`
		LastBlockTime   time.Time          `json:"last_block_time"`
		MaxBlockSize    int                `json:"max_block_size"`
		BlockTimeTarget time.Duration      `json:"block_time_target"`
	}{
		Blocks:          bc.Blocks,
		PendingTxs:      bc.PendingTxs,
		Difficulty:      bc.Difficulty,
		MiningReward:    bc.MiningReward,
		Balances:        bc.Balances,
		LastBlockTime:   bc.LastBlockTime,
		MaxBlockSize:    bc.MaxBlockSize,
		BlockTimeTarget: bc.BlockTimeTarget,
	}

	// Sérialiser avec indentation pour lisibilité
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur lors de la sérialisation: %w", err)
	}

	// Écrire dans le fichier
	if err := ioutil.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("erreur d'écriture dans le fichier %s: %w", filename, err)
	}

	return nil
}

// LoadBlockchainFromFile charge la blockchain depuis un fichier JSON
func LoadBlockchainFromFile(filename string) (*Blockchain, error) {
	// Vérifier si le fichier existe
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("le fichier %s n'existe pas", filename)
	}

	// Lire le fichier
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la lecture du fichier %s: %w", filename, err)
	}

	// Structure temporaire pour la désérialisation
	var temp struct {
		Blocks          []*Block           `json:"blocks"`
		PendingTxs      []*Transaction     `json:"pending_transactions"`
		Difficulty      int                `json:"difficulty"`
		MiningReward    float64            `json:"mining_reward"`
		Balances        map[string]float64 `json:"balances"`
		LastBlockTime   time.Time          `json:"last_block_time"`
		MaxBlockSize    int                `json:"max_block_size"`
		BlockTimeTarget time.Duration      `json:"block_time_target"`
	}

	// Désérialiser
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, fmt.Errorf("erreur lors de la désérialisation: %w", err)
	}

	// Créer une nouvelle blockchain avec les données chargées
	bc := &Blockchain{
		Blocks:          temp.Blocks,
		PendingTxs:      temp.PendingTxs,
		Difficulty:      temp.Difficulty,
		MiningReward:    temp.MiningReward,
		Balances:        temp.Balances,
		LastBlockTime:   temp.LastBlockTime,
		MaxBlockSize:    temp.MaxBlockSize,
		BlockTimeTarget: temp.BlockTimeTarget,
	}

	return bc, nil
}

// GetTransactionByID recherche une transaction par son ID
func (bc *Blockchain) GetTransactionByID(txID string) (*Transaction, bool, int) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Chercher d'abord dans les transactions en attente
	for _, tx := range bc.PendingTxs {
		if tx.ID == txID {
			return tx, false, -1 // Trouvé dans les transactions en attente
		}
	}

	// Chercher dans les blocs
	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			if tx.ID == txID {
				return tx, true, block.Index // Trouvé dans un bloc confirmé
			}
		}
	}

	return nil, false, -1 // Non trouvé
}

// GetTransactionsForAddress retourne toutes les transactions pour une adresse
func (bc *Blockchain) GetTransactionsForAddress(address string) []*Transaction {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	result := []*Transaction{}

	// Rechercher dans tous les blocs
	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			if tx.From == address || tx.To == address {
				result = append(result, tx)
			}
		}
	}

	// Ajouter les transactions en attente
	for _, tx := range bc.PendingTxs {
		if tx.From == address || tx.To == address {
			result = append(result, tx)
		}
	}

	return result
}

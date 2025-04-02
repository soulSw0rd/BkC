package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// MemPool représente la réserve de transactions en attente
type MemPool struct {
	Transactions map[string]*Transaction
	mu           sync.RWMutex
}

// NewMemPool crée un nouveau pool de transactions
func NewMemPool() *MemPool {
	return &MemPool{
		Transactions: make(map[string]*Transaction),
	}
}

// AddTransaction ajoute une transaction au mempool
func (mp *MemPool) AddTransaction(tx *Transaction) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.Transactions[tx.ID] = tx
}

// GetTransactions récupère les transactions du mempool
func (mp *MemPool) GetTransactions(limit int) []*Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	var txs []*Transaction
	count := 0

	// Trier par frais de transaction (simplification - en réalité, utilisez un tas)
	for _, tx := range mp.Transactions {
		txs = append(txs, tx)
		count++
		if count >= limit && limit > 0 {
			break
		}
	}

	return txs
}

// RemoveTransaction supprime une transaction du mempool
func (mp *MemPool) RemoveTransaction(txID string) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	delete(mp.Transactions, txID)
}

// Blockchain représente la chaîne de blocs
type Blockchain struct {
	Blocks           []*Block
	MemPool          *MemPool
	mu               sync.RWMutex
	MiningDifficulty int
	MiningReward     float64
	TargetBlockTime  time.Duration
}

// CreateGenesisBlock crée le premier bloc (genesis block)
func CreateGenesisBlock() *Block {
	genesisTime := time.Now()
	block := &Block{
		Index:        0,
		Timestamp:    genesisTime,
		Transactions: []Transaction{},
		PrevHash:     "",
		Nonce:        0,
		Difficulty:   4,
		Miner:        "system",
	}

	// Créer une transaction de récompense pour le bloc genesis
	coinbaseTx := Transaction{
		ID:        "genesis_coinbase",
		Timestamp: genesisTime,
		Sender:    "system",
		Recipient: "genesis_address",
		Amount:    50.0,
		Fee:       0,
	}

	block.AddTransaction(coinbaseTx)
	block.MerkleRoot = block.CalculateMerkleRoot()
	block.Hash = block.ComputeHash()
	block.ProofOfWork(4)

	return block
}

// NewBlockchain initialise une nouvelle blockchain
func NewBlockchain() *Blockchain {
	return &Blockchain{
		Blocks:           []*Block{CreateGenesisBlock()},
		MemPool:          NewMemPool(),
		MiningDifficulty: 4,
		MiningReward:     50.0,
		TargetBlockTime:  60 * time.Second, // 1 minute par bloc
	}
}

// CreateBlock crée et mine un nouveau bloc
func (bc *Blockchain) CreateBlock(minerAddress string) *Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	prevBlock := bc.Blocks[len(bc.Blocks)-1]

	// Créer le nouveau bloc
	newBlock := &Block{
		Index:        len(bc.Blocks),
		Timestamp:    time.Now(),
		Transactions: []Transaction{},
		PrevHash:     prevBlock.Hash,
		Nonce:        0,
		Difficulty:   bc.MiningDifficulty,
		Miner:        minerAddress,
	}

	// Créer la transaction de récompense pour le mineur
	coinbaseTx := Transaction{
		ID:        fmt.Sprintf("coinbase_%d", newBlock.Index),
		Timestamp: newBlock.Timestamp,
		Sender:    "system",
		Recipient: minerAddress,
		Amount:    bc.MiningReward,
		Fee:       0,
	}

	// Ajouter la transaction de récompense
	newBlock.AddTransaction(coinbaseTx)

	// Ajouter les transactions du mempool (jusqu'à une limite, par exemple 100)
	pendingTxs := bc.MemPool.GetTransactions(100)
	for _, tx := range pendingTxs {
		// Vérifier la validité de la transaction
		if tx.Verify() {
			newBlock.AddTransaction(*tx)
			bc.MemPool.RemoveTransaction(tx.ID)
		}
	}

	// Calculer la racine de Merkle
	newBlock.MerkleRoot = newBlock.CalculateMerkleRoot()

	// Miner le bloc
	newBlock.ProofOfWork(bc.MiningDifficulty)

	// Ajouter à la blockchain
	bc.Blocks = append(bc.Blocks, newBlock)

	// Ajuster la difficulté si nécessaire
	bc.MiningDifficulty = AdjustDifficulty(bc.Blocks, bc.TargetBlockTime)

	return newBlock
}

// ValidateChain vérifie l'intégrité de toute la blockchain
func (bc *Blockchain) ValidateChain() (bool, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for i := 1; i < len(bc.Blocks); i++ {
		currentBlock := bc.Blocks[i]
		prevBlock := bc.Blocks[i-1]

		// Vérifier la référence au bloc précédent
		if currentBlock.PrevHash != prevBlock.Hash {
			return false, errors.New("chaîne brisée: hash précédent incorrect")
		}

		// Vérifier le hash du bloc actuel
		if currentBlock.Hash != currentBlock.ComputeHash() {
			return false, errors.New("hash de bloc invalide")
		}

		// Vérifier la preuve de travail
		if !currentBlock.VerifyProofOfWork() {
			return false, errors.New("preuve de travail invalide")
		}

		// Vérifier la racine de Merkle
		if currentBlock.MerkleRoot != currentBlock.CalculateMerkleRoot() {
			return false, errors.New("racine de Merkle invalide")
		}

		// Vérifier les transactions
		for _, tx := range currentBlock.Transactions {
			if !tx.Verify() && tx.Sender != "system" {
				return false, errors.New("transaction invalide dans le bloc")
			}
		}
	}

	return true, nil
}

// GetBalance calcule le solde d'une adresse
func (bc *Blockchain) GetBalance(address string) float64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	balance := 0.0

	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			if tx.Recipient == address {
				balance += tx.Amount
			}
			if tx.Sender == address {
				balance -= (tx.Amount + tx.Fee)
			}
		}
	}

	return balance
}

// GetTransactionHistory récupère l'historique des transactions d'une adresse
func (bc *Blockchain) GetTransactionHistory(address string) []Transaction {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	var history []Transaction

	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			if tx.Sender == address || tx.Recipient == address {
				history = append(history, tx)
			}
		}
	}

	return history
}

// AddTransaction ajoute une transaction au mempool
func (bc *Blockchain) AddTransaction(tx *Transaction) error {
	// Vérifier la validité de la transaction
	if !tx.Verify() {
		return errors.New("transaction invalide")
	}

	// Vérifier que l'expéditeur a suffisamment de fonds
	if tx.Sender != "system" {
		balance := bc.GetBalance(tx.Sender)
		if balance < tx.Amount+tx.Fee {
			return errors.New("fonds insuffisants")
		}
	}

	// Ajouter au mempool
	bc.MemPool.AddTransaction(tx)

	return nil
}

// GetStats retourne les statistiques en JSON
func (bc *Blockchain) GetStats() string {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	stats := map[string]interface{}{
		"numBlocks":        len(bc.Blocks),
		"difficulty":       bc.MiningDifficulty,
		"miningReward":     bc.MiningReward,
		"pendingTxCount":   len(bc.MemPool.Transactions),
		"lastBlockHash":    bc.Blocks[len(bc.Blocks)-1].Hash,
		"blockchainHeight": len(bc.Blocks) - 1,
	}

	jsonStats, _ := json.Marshal(stats)
	return string(jsonStats)
}

// GetBlockByHash recherche un bloc par son hash
func (bc *Blockchain) GetBlockByHash(hash string) *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for _, block := range bc.Blocks {
		if block.Hash == hash {
			return block
		}
	}

	return nil
}

// GetBlockByIndex recherche un bloc par son index
func (bc *Blockchain) GetBlockByIndex(index int) *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if index < 0 || index >= len(bc.Blocks) {
		return nil
	}

	return bc.Blocks[index]
}

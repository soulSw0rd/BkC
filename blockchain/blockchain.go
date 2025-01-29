package blockchain

import (
	"encoding/json"
	"sync"
	"time"
)

// Blockchain représente la chaîne de blocs
type Blockchain struct {
	Blocks []*Block
	mu     sync.RWMutex // Utilisez RWMutex pour permettre des lectures concurrentes
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
	return &Blockchain{
		Blocks: []*Block{CreateGenesisBlock()},
	}
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

package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Block represents a block in the blockchain
type Block struct {
	Index        int           `json:"index"`
	Timestamp    time.Time     `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	MerkleRoot   string        `json:"merkleRoot"`
	PrevHash     string        `json:"prevHash"`
	Hash         string        `json:"hash"`
	Nonce        int           `json:"nonce"`
	Difficulty   int           `json:"difficulty"`
	Miner        string        `json:"miner"`
}

// ComputeHash calculates the hash of a block
func (b *Block) ComputeHash() string {
	record := fmt.Sprintf("%d%s%s%s%d%d",
		b.Index,
		b.Timestamp.String(),
		b.MerkleRoot,
		b.PrevHash,
		b.Nonce,
		b.Difficulty,
	)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// CalculateMerkleRoot calculates the Merkle tree root of transactions
func (b *Block) CalculateMerkleRoot() string {
	var txHashes [][]byte

	// If no transactions, use a default value
	if len(b.Transactions) == 0 {
		hash := sha256.Sum256([]byte("empty_block"))
		return hex.EncodeToString(hash[:])
	}

	// Collect transaction hashes
	for _, tx := range b.Transactions {
		hashBytes, _ := hex.DecodeString(tx.ID)
		txHashes = append(txHashes, hashBytes)
	}

	// While there is more than one hash
	for len(txHashes) > 1 {
		if len(txHashes)%2 != 0 {
			// Duplicate the last element if odd number
			txHashes = append(txHashes, txHashes[len(txHashes)-1])
		}

		var newLevel [][]byte

		// Combine hashes in pairs
		for i := 0; i < len(txHashes); i += 2 {
			concat := append(txHashes[i], txHashes[i+1]...)
			hash := sha256.Sum256(concat)
			newLevel = append(newLevel, hash[:])
		}

		txHashes = newLevel
	}

	// The last hash is the Merkle root
	return hex.EncodeToString(txHashes[0])
}

// AddTransaction adds a transaction to the block and recalculates the Merkle root
func (b *Block) AddTransaction(tx Transaction) {
	b.Transactions = append(b.Transactions, tx)
	b.MerkleRoot = b.CalculateMerkleRoot()
}

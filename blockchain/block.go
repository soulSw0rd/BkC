package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Block représente un bloc dans la blockchain
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

// ComputeHash calcule le hash d'un bloc
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

// CalculateMerkleRoot calcule la racine de l'arbre de Merkle des transactions
func (b *Block) CalculateMerkleRoot() string {
	var txHashes [][]byte

	// Si aucune transaction, utilisez une valeur par défaut
	if len(b.Transactions) == 0 {
		hash := sha256.Sum256([]byte("empty_block"))
		return hex.EncodeToString(hash[:])
	}

	// Collecte des hashes de transactions
	for _, tx := range b.Transactions {
		hashBytes, _ := hex.DecodeString(tx.ID)
		txHashes = append(txHashes, hashBytes)
	}

	// Tant qu'il reste plus d'un hash
	for len(txHashes) > 1 {
		if len(txHashes)%2 != 0 {
			// Dupliquer le dernier élément si nombre impair
			txHashes = append(txHashes, txHashes[len(txHashes)-1])
		}

		var newLevel [][]byte

		// Combiner les hashes par paires
		for i := 0; i < len(txHashes); i += 2 {
			concat := append(txHashes[i], txHashes[i+1]...)
			hash := sha256.Sum256(concat)
			newLevel = append(newLevel, hash[:])
		}

		txHashes = newLevel
	}

	// Le dernier hash est la racine de Merkle
	return hex.EncodeToString(txHashes[0])
}

// AddTransaction ajoute une transaction au bloc et recalcule la racine de Merkle
func (b *Block) AddTransaction(tx Transaction) {
	b.Transactions = append(b.Transactions, tx)
	b.MerkleRoot = b.CalculateMerkleRoot()
}

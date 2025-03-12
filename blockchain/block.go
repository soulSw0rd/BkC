// blockchain/block.go
package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Block représente un bloc dans la blockchain
type Block struct {
	Index        int            `json:"index"`
	Timestamp    time.Time      `json:"timestamp"`
	Transactions []*Transaction `json:"transactions"`
	Data         string         `json:"data"` // Données supplémentaires
	PrevHash     string         `json:"prev_hash"`
	Hash         string         `json:"hash"`
	Nonce        int            `json:"nonce"`
	Difficulty   int            `json:"difficulty"`
	Miner        string         `json:"miner"`       // Adresse du mineur
	MiningTime   time.Duration  `json:"mining_time"` // Temps de minage
}

// ComputeHash calcule le hash d'un bloc
func (b *Block) ComputeHash() string {
	// Sérialiser les transactions
	txData, err := json.Marshal(b.Transactions)
	if err != nil {
		// En cas d'erreur, utiliser une chaîne vide pour les transactions
		txData = []byte("")
	}

	// Construire l'enregistrement à hacher
	record := fmt.Sprintf("%d%s%s%s%s%d",
		b.Index,
		b.Timestamp.Format(time.RFC3339),
		string(txData),
		b.Data,
		b.PrevHash,
		b.Nonce)

	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// AddTransaction ajoute une transaction au bloc
func (b *Block) AddTransaction(tx *Transaction) {
	b.Transactions = append(b.Transactions, tx)
}

// CalculateMerkleRoot calcule la racine de Merkle pour les transactions du bloc
// (Simplifiée pour cette implémentation)
func (b *Block) CalculateMerkleRoot() string {
	var txHashes []string

	for _, tx := range b.Transactions {
		txHash := sha256.Sum256([]byte(tx.ID))
		txHashes = append(txHashes, hex.EncodeToString(txHash[:]))
	}

	// Si aucune transaction, retourner un hash par défaut
	if len(txHashes) == 0 {
		return "0000000000000000000000000000000000000000000000000000000000000000"
	}

	// Version simplifiée - dans une implémentation réelle, on construirait un arbre de Merkle
	combinedHash := sha256.Sum256([]byte(fmt.Sprintf("%v", txHashes)))
	return hex.EncodeToString(combinedHash[:])
}

// GetTransactionTotal calcule le montant total des transactions
func (b *Block) GetTransactionTotal() float64 {
	var total float64
	for _, tx := range b.Transactions {
		if tx.From != "system" { // Ne pas compter les récompenses de minage
			total += tx.Amount
		}
	}
	return total
}

// GetTransactionCount retourne le nombre de transactions dans le bloc
func (b *Block) GetTransactionCount() int {
	return len(b.Transactions)
}

// ToString retourne une représentation textuelle du bloc
func (b *Block) ToString() string {
	return fmt.Sprintf(
		"Bloc #%d\nTimestamp: %s\nHash: %s\nHash précédent: %s\nNonce: %d\nDifficulté: %d\nMineur: %s\nTemps de minage: %s\nTransactions: %d\nDonnées: %s",
		b.Index,
		b.Timestamp.Format(time.RFC3339),
		b.Hash,
		b.PrevHash,
		b.Nonce,
		b.Difficulty,
		b.Miner,
		b.MiningTime.String(),
		len(b.Transactions),
		b.Data,
	)
}

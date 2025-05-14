package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Block représente un bloc dans la blockchain
type Block struct {
	Index      int    `json:"index"`
	Timestamp  string `json:"timestamp"`
	Data       string `json:"data"`
	PrevHash   string `json:"prev_hash"`
	Hash       string `json:"hash"`
	Nonce      int    `json:"nonce"`
	Miner      string `json:"miner,omitempty"`       // Nom d'utilisateur du mineur
	MiningInfo string `json:"mining_info,omitempty"` // Informations de minage en JSON
}

// ComputeHash calcule le hash d'un bloc
func (b *Block) ComputeHash() string {
	record := fmt.Sprintf("%d%s%s%s%d", b.Index, b.Timestamp, b.Data, b.PrevHash, b.Nonce)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

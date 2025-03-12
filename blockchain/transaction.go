// blockchain/transaction.go
package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

// Transaction représente une transaction dans la blockchain
type Transaction struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Amount    float64   `json:"amount"`
	Timestamp time.Time `json:"timestamp"`
	Signature string    `json:"signature"` // Pour la vérification cryptographique
}

// NewTransaction crée une nouvelle transaction
func NewTransaction(from, to string, amount float64) *Transaction {
	transaction := &Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Timestamp: time.Now(),
	}
	transaction.ID = transaction.CalculateHash()
	return transaction
}

// CalculateHash calcule le hash de la transaction
func (t *Transaction) CalculateHash() string {
	data, _ := json.Marshal(struct {
		From      string    `json:"from"`
		To        string    `json:"to"`
		Amount    float64   `json:"amount"`
		Timestamp time.Time `json:"timestamp"`
	}{
		From:      t.From,
		To:        t.To,
		Amount:    t.Amount,
		Timestamp: t.Timestamp,
	})

	hash := sha256.Sum256(data)
	return string(hash[:])
}

// IsValid vérifie si une transaction est valide
func (tx *Transaction) IsValid() bool {
	// Vérifications de base
	if tx.From == "" || tx.To == "" {
		return false
	}

	if tx.Amount <= 0 {
		return false
	}

	// Les transactions du système sont toujours valides (pour les récompenses de minage)
	if tx.From == "system" {
		return true
	}

	// Pour une implémentation future: vérification de la signature
	// if !VerifySignature(tx.From, tx.SignatureData(), tx.Signature) {
	//     return false
	// }

	return true
}

// Sign ajoute une signature à la transaction (placeholder pour implémentation future)
func (tx *Transaction) Sign(privateKey string) {
	// Dans une implémentation réelle:
	// 1. Créer une représentation hachée des données de la transaction
	// 2. Signer ce hash avec la clé privée
	// 3. Stocker la signature

	// Pour l'instant, on simule simplement une signature
	signatureData := tx.SignatureData()
	hash := sha256.Sum256([]byte(signatureData + privateKey))
	tx.Signature = string(hash[:])
}

// SignatureData retourne les données à signer
func (tx *Transaction) SignatureData() string {
	return fmt.Sprintf("%s%s%f%s%s",
		tx.From,
		tx.To,
		tx.Amount,
		tx.Timestamp.Format(time.RFC3339))
}

// ToString retourne une représentation textuelle de la transaction
func (tx *Transaction) ToString() string {
	return fmt.Sprintf("ID: %s\nDe: %s\nVers: %s\nMontant: %.2f\nDate: %s",
		tx.ID, tx.From, tx.To, tx.Amount, tx.Timestamp.Format(time.RFC3339))
}

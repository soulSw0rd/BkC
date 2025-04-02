package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

// Transaction représente une transaction dans la blockchain
type Transaction struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Sender    string    `json:"sender"`
	Recipient string    `json:"recipient"`
	Amount    float64   `json:"amount"`
	Fee       float64   `json:"fee"`
	Signature []byte    `json:"signature"`
	PublicKey []byte    `json:"publicKey"`
}

// NewTransaction crée une nouvelle transaction non signée
func NewTransaction(sender, recipient string, amount, fee float64) *Transaction {
	tx := &Transaction{
		Timestamp: time.Now(),
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
		Fee:       fee,
	}
	tx.ID = tx.ComputeHash()
	return tx
}

// ComputeHash calcule le hash d'une transaction
func (tx *Transaction) ComputeHash() string {
	record := fmt.Sprintf("%s%s%.8f%.8f%s",
		tx.Timestamp.String(),
		tx.Sender,
		tx.Recipient,
		tx.Amount,
		tx.Fee,
	)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// Sign signe une transaction avec la clé privée
func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey) error {
	txHash := sha256.Sum256([]byte(tx.ID))

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, txHash[:])
	if err != nil {
		return err
	}

	// Concaténer r et s pour former la signature
	signature := append(r.Bytes(), s.Bytes()...)
	tx.Signature = signature

	// Stocker la clé publique pour la vérification
	tx.PublicKey = append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)

	return nil
}

// Verify vérifie la validité de la signature de la transaction
func (tx *Transaction) Verify() bool {
	if tx.Sender == "system" {
		// Transaction de récompense du système, pas besoin de signature
		return true
	}

	if len(tx.Signature) == 0 || len(tx.PublicKey) == 0 {
		return false
	}

	txHash := sha256.Sum256([]byte(tx.ID))

	// Extraire les composantes r et s de la signature
	sigLen := len(tx.Signature) / 2
	r := new(big.Int).SetBytes(tx.Signature[:sigLen])
	s := new(big.Int).SetBytes(tx.Signature[sigLen:])

	// Reconstruire la clé publique
	x := new(big.Int).SetBytes(tx.PublicKey[:len(tx.PublicKey)/2])
	y := new(big.Int).SetBytes(tx.PublicKey[len(tx.PublicKey)/2:])

	publicKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	return ecdsa.Verify(publicKey, txHash[:], r, s)
}

// String retourne une représentation JSON de la transaction
func (tx *Transaction) String() string {
	txJSON, _ := json.MarshalIndent(tx, "", "  ")
	return string(txJSON)
}

package blockchain

import (
	"testing"
	"time"
)

func TestBlockComputeHash(t *testing.T) {
	block := Block{
		Index:        1,
		Timestamp:    time.Now(),
		Transactions: []Transaction{},
		PrevHash:     "prevhash",
		Nonce:        0,
		Difficulty:   4,
	}

	// Premier hash
	hash1 := block.ComputeHash()

	// Vérifier que le hash n'est pas vide
	if hash1 == "" {
		t.Error("Le hash calculé est vide")
	}

	// Modifier le bloc et vérifier que le hash change
	block.Nonce = 1
	hash2 := block.ComputeHash()

	// Vérifier que le hash a changé
	if hash1 == hash2 {
		t.Error("Le hash n'a pas changé après modification du bloc")
	}
}

func TestBlockCalculateMerkleRoot(t *testing.T) {
	// Créer un bloc sans transactions
	emptyBlock := Block{
		Index:        1,
		Timestamp:    time.Now(),
		Transactions: []Transaction{},
		PrevHash:     "prevhash",
		Nonce:        0,
		Difficulty:   4,
	}

	// Calculer la racine de Merkle pour un bloc vide
	emptyRoot := emptyBlock.CalculateMerkleRoot()

	// Vérifier que la racine n'est pas vide
	if emptyRoot == "" {
		t.Error("La racine de Merkle pour un bloc vide est vide")
	}

	// Créer un bloc avec une transaction
	tx := Transaction{
		ID:        "transaction1",
		Timestamp: time.Now(),
		Sender:    "sender",
		Recipient: "recipient",
		Amount:    10.0,
		Fee:       0.1,
	}

	blockWithTx := Block{
		Index:        1,
		Timestamp:    time.Now(),
		Transactions: []Transaction{tx},
		PrevHash:     "prevhash",
		Nonce:        0,
		Difficulty:   4,
	}

	// Calculer la racine de Merkle pour un bloc avec une transaction
	rootWithTx := blockWithTx.CalculateMerkleRoot()

	// Vérifier que la racine n'est pas vide
	if rootWithTx == "" {
		t.Error("La racine de Merkle pour un bloc avec transaction est vide")
	}

	// Vérifier que les racines sont différentes
	if emptyRoot == rootWithTx {
		t.Error("Les racines de Merkle sont identiques pour des blocs différents")
	}

	// Ajouter une deuxième transaction
	tx2 := Transaction{
		ID:        "transaction2",
		Timestamp: time.Now(),
		Sender:    "sender2",
		Recipient: "recipient2",
		Amount:    20.0,
		Fee:       0.2,
	}

	blockWithTx.AddTransaction(tx2)

	// Calculer la racine de Merkle pour un bloc avec deux transactions
	rootWithTwoTx := blockWithTx.CalculateMerkleRoot()

	// Vérifier que la racine a changé
	if rootWithTx == rootWithTwoTx {
		t.Error("La racine de Merkle n'a pas changé après ajout d'une transaction")
	}
}

func TestBlockProofOfWork(t *testing.T) {
	block := Block{
		Index:        1,
		Timestamp:    time.Now(),
		Transactions: []Transaction{},
		PrevHash:     "prevhash",
		Nonce:        0,
		Difficulty:   2, // Difficulté faible pour le test
	}

	// Exécuter la preuve de travail
	block.ProofOfWork(block.Difficulty)

	// Vérifier que le hash commence par le bon nombre de zéros
	prefix := "00"
	if block.Hash[:len(prefix)] != prefix {
		t.Errorf("Le hash %s ne commence pas par %s", block.Hash, prefix)
	}

	// Vérifier que le bloc est valide
	if !block.VerifyProofOfWork() {
		t.Error("La vérification de la preuve de travail a échoué")
	}

	// Modifier le bloc et vérifier que la validation échoue
	originalHash := block.Hash
	block.Nonce = block.Nonce + 1
	if block.VerifyProofOfWork() {
		t.Error("La vérification de la preuve de travail a réussi après modification du bloc")
	}

	// Restaurer le hash et vérifier que la validation réussit à nouveau
	block.Hash = originalHash
	if !block.VerifyProofOfWork() {
		t.Error("La vérification de la preuve de travail a échoué après restauration du hash")
	}
}

func TestBlockAddTransaction(t *testing.T) {
	block := Block{
		Index:        1,
		Timestamp:    time.Now(),
		Transactions: []Transaction{},
		PrevHash:     "prevhash",
		Nonce:        0,
		Difficulty:   4,
	}

	// Vérifier qu'il n'y a pas de transactions initialement
	if len(block.Transactions) != 0 {
		t.Error("Le bloc contient des transactions initialement")
	}

	// Ajouter une transaction
	tx := Transaction{
		ID:        "transaction1",
		Timestamp: time.Now(),
		Sender:    "sender",
		Recipient: "recipient",
		Amount:    10.0,
		Fee:       0.1,
	}

	block.AddTransaction(tx)

	// Vérifier qu'il y a maintenant une transaction
	if len(block.Transactions) != 1 {
		t.Error("La transaction n'a pas été ajoutée au bloc")
	}

	// Vérifier que la transaction ajoutée est correcte
	if block.Transactions[0].ID != tx.ID {
		t.Error("La transaction ajoutée n'est pas celle attendue")
	}

	// Vérifier que la racine de Merkle a été recalculée
	if block.MerkleRoot == "" {
		t.Error("La racine de Merkle n'a pas été calculée après ajout de la transaction")
	}
}

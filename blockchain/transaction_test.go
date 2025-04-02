package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"strings"
	"testing"
	"time"
)

func TestTransactionComputeHash(t *testing.T) {
	// Créer une transaction
	tx := Transaction{
		Timestamp: time.Now(),
		Sender:    "sender",
		Recipient: "recipient",
		Amount:    10.0,
		Fee:       0.1,
	}

	// Calculer le hash
	hash := tx.ComputeHash()

	// Vérifier que le hash n'est pas vide
	if hash == "" {
		t.Error("Le hash calculé est vide")
	}

	// Recalculer le hash et vérifier qu'il est identique
	hash2 := tx.ComputeHash()
	if hash != hash2 {
		t.Error("Le hash recalculé est différent")
	}

	// Modifier la transaction et vérifier que le hash change
	tx.Amount = 20.0
	hash3 := tx.ComputeHash()
	if hash == hash3 {
		t.Error("Le hash n'a pas changé après modification de la transaction")
	}
}

func TestTransactionSign(t *testing.T) {
	// Générer une paire de clés
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Erreur lors de la génération de la clé privée: %v", err)
	}

	// Créer une transaction
	tx := Transaction{
		Timestamp: time.Now(),
		Sender:    "sender",
		Recipient: "recipient",
		Amount:    10.0,
		Fee:       0.1,
	}
	tx.ID = tx.ComputeHash()

	// Vérifier qu'il n'y a pas de signature initialement
	if len(tx.Signature) != 0 {
		t.Error("La transaction a une signature initialement")
	}

	// Signer la transaction
	err = tx.Sign(privateKey)
	if err != nil {
		t.Fatalf("Erreur lors de la signature: %v", err)
	}

	// Vérifier qu'il y a maintenant une signature
	if len(tx.Signature) == 0 {
		t.Error("La transaction n'a pas été signée")
	}

	// Vérifier qu'il y a une clé publique
	if len(tx.PublicKey) == 0 {
		t.Error("La clé publique n'a pas été stockée")
	}
}

func TestTransactionVerify(t *testing.T) {
	// Générer une paire de clés
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Erreur lors de la génération de la clé privée: %v", err)
	}

	// Créer une transaction
	tx := Transaction{
		Timestamp: time.Now(),
		Sender:    "sender",
		Recipient: "recipient",
		Amount:    10.0,
		Fee:       0.1,
	}
	tx.ID = tx.ComputeHash()

	// Signer la transaction
	err = tx.Sign(privateKey)
	if err != nil {
		t.Fatalf("Erreur lors de la signature: %v", err)
	}

	// Vérifier la signature
	valid := tx.Verify()
	if !valid {
		t.Error("La vérification de la signature a échoué")
	}

	// Modifier la transaction et vérifier que la validation échoue
	tx.Amount = 20.0
	valid = tx.Verify()
	if valid {
		t.Error("La vérification de la signature a réussi après modification de la transaction")
	}

	// Tester une transaction système (sans signature)
	systemTx := Transaction{
		Timestamp: time.Now(),
		Sender:    "system",
		Recipient: "recipient",
		Amount:    10.0,
		Fee:       0.0,
	}
	systemTx.ID = systemTx.ComputeHash()

	// Les transactions système sont toujours valides
	valid = systemTx.Verify()
	if !valid {
		t.Error("La vérification d'une transaction système a échoué")
	}
}

func TestNewTransaction(t *testing.T) {
	// Créer une nouvelle transaction
	sender := "sender"
	recipient := "recipient"
	amount := 10.0
	fee := 0.1

	tx := NewTransaction(sender, recipient, amount, fee)

	// Vérifier les valeurs
	if tx.Sender != sender {
		t.Errorf("Expéditeur incorrect: attendu %s, obtenu %s", sender, tx.Sender)
	}
	if tx.Recipient != recipient {
		t.Errorf("Destinataire incorrect: attendu %s, obtenu %s", recipient, tx.Recipient)
	}
	if tx.Amount != amount {
		t.Errorf("Montant incorrect: attendu %.2f, obtenu %.2f", amount, tx.Amount)
	}
	if tx.Fee != fee {
		t.Errorf("Frais incorrects: attendu %.2f, obtenu %.2f", fee, tx.Fee)
	}
	if tx.ID == "" {
		t.Error("L'ID de transaction est vide")
	}

	// Vérifier que l'horodatage est proche de maintenant
	now := time.Now()
	diff := now.Sub(tx.Timestamp)
	if diff > time.Second {
		t.Errorf("Horodatage trop ancien: %v", diff)
	}
}

func TestTransactionString(t *testing.T) {
	// Créer une transaction
	tx := Transaction{
		ID:        "txid",
		Timestamp: time.Now(),
		Sender:    "sender",
		Recipient: "recipient",
		Amount:    10.0,
		Fee:       0.1,
	}

	// Obtenir la représentation en chaîne
	str := tx.String()

	// Vérifier que la chaîne contient les informations attendues
	if !strings.Contains(str, "txid") {
		t.Error("La représentation en chaîne ne contient pas l'ID")
	}
	if !strings.Contains(str, "sender") {
		t.Error("La représentation en chaîne ne contient pas l'expéditeur")
	}
	if !strings.Contains(str, "recipient") {
		t.Error("La représentation en chaîne ne contient pas le destinataire")
	}
	if !strings.Contains(str, "10") {
		t.Error("La représentation en chaîne ne contient pas le montant")
	}
	if !strings.Contains(str, "0.1") {
		t.Error("La représentation en chaîne ne contient pas les frais")
	}
}

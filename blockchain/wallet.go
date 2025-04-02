package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
)

// Wallet représente un portefeuille numérique
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey `json:"-"` // Ne pas sérialiser directement
	PublicKey  []byte            `json:"publicKey"`
	Address    string            `json:"address"`
}

// NewWallet crée un nouveau portefeuille avec une nouvelle paire de clés
func NewWallet() (*Wallet, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)

	wallet := &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    generateAddress(publicKey),
	}

	return wallet, nil
}

// generateAddress génère une adresse à partir de la clé publique
func generateAddress(publicKey []byte) string {
	// Étape 1: Appliquer SHA-256 à la clé publique
	publicKeyHash := sha256.Sum256(publicKey)

	// Étape 2: Appliquer RIPEMD-160 au résultat de l'étape 1
	// Note: Dans une implémentation complète, utiliser RIPEMD-160
	// Ici nous utilisons SHA-256 une seconde fois pour simplifier
	ripeMD := sha256.Sum256(publicKeyHash[:])

	// Étape 3: Préfixe de version (0x00 pour mainnet)
	versionedRipeMD := append([]byte{0x00}, ripeMD[:20]...)

	// Étape 4: Double SHA-256 pour le checksum
	firstSHA := sha256.Sum256(versionedRipeMD)
	secondSHA := sha256.Sum256(firstSHA[:])

	// Étape 5: Ajouter les 4 premiers octets comme checksum
	addressBytes := append(versionedRipeMD, secondSHA[:4]...)

	// Étape 6: Encoder en base58 (ici en hex pour simplifier)
	address := hex.EncodeToString(addressBytes)

	return address
}

// CreateTransaction crée une nouvelle transaction signée
func (w *Wallet) CreateTransaction(recipient string, amount, fee float64) (*Transaction, error) {
	if amount <= 0 {
		return nil, errors.New("le montant doit être positif")
	}

	tx := NewTransaction(w.Address, recipient, amount, fee)

	err := tx.Sign(w.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la signature de la transaction: %v", err)
	}

	return tx, nil
}

// SaveToFile sauvegarde le portefeuille dans un fichier
func (w *Wallet) SaveToFile(filename string) error {
	// Créer une structure pour la sérialisation
	type walletExport struct {
		PrivateKeyD *big.Int `json:"privateKeyD"`
		PrivateKeyX *big.Int `json:"privateKeyX"`
		PrivateKeyY *big.Int `json:"privateKeyY"`
		PublicKey   []byte   `json:"publicKey"`
		Address     string   `json:"address"`
	}

	export := walletExport{
		PrivateKeyD: w.PrivateKey.D,
		PrivateKeyX: w.PrivateKey.PublicKey.X,
		PrivateKeyY: w.PrivateKey.PublicKey.Y,
		PublicKey:   w.PublicKey,
		Address:     w.Address,
	}

	walletJSON, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return err
	}

	// Assurez-vous que le répertoire existe
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	return ioutil.WriteFile(filename, walletJSON, 0600)
}

// LoadFromFile charge un portefeuille depuis un fichier
func LoadWalletFromFile(filename string) (*Wallet, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Structure pour la désérialisation
	type walletExport struct {
		PrivateKeyD *big.Int `json:"privateKeyD"`
		PrivateKeyX *big.Int `json:"privateKeyX"`
		PrivateKeyY *big.Int `json:"privateKeyY"`
		PublicKey   []byte   `json:"publicKey"`
		Address     string   `json:"address"`
	}

	var export walletExport
	if err := json.Unmarshal(data, &export); err != nil {
		return nil, err
	}

	// Reconstruire la clé privée
	privateKey := &ecdsa.PrivateKey{
		D: export.PrivateKeyD,
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     export.PrivateKeyX,
			Y:     export.PrivateKeyY,
		},
	}

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  export.PublicKey,
		Address:    export.Address,
	}, nil
}

// LoadWalletFromString charge un portefeuille depuis une clé privée (simplifiée pour l'exemple)
func LoadWalletFromString(privateKeyStr string) (*Wallet, error) {
	// Dans un système réel, vous déchiffreriez ici la clé privée
	// depuis le format fourni (PEM, etc.)
	// Ici, nous créons simplement un nouveau portefeuille pour la démonstration

	if privateKeyStr == "" {
		return nil, errors.New("clé privée vide")
	}

	// Pour les besoins de la démo, si le format est "DEMO_ONLY_<address>"
	// nous créons un portefeuille avec cette adresse
	address := privateKeyStr
	if len(privateKeyStr) > 10 && privateKeyStr[:10] == "DEMO_ONLY_" {
		address = privateKeyStr[10:]
	}

	wallet, err := NewWallet()
	if err != nil {
		return nil, err
	}

	// Normalement, on reconstruirait ici le portefeuille à partir de la clé
	// Pour la démo, on modifie juste l'adresse
	wallet.Address = address

	return wallet, nil
}

package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Méthodes d'extension à ajouter à la structure Blockchain existante

// SaveContract sauvegarde un contrat dans la blockchain
func (bc *Blockchain) SaveContract(contract *SmartContract) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Vérifier l'unicité du contrat
	if existing := bc.GetContractByID(contract.ID); existing != nil {
		return errors.New("un contrat avec cet ID existe déjà")
	}

	// Vérifier si le créateur a suffisamment de fonds
	if contract.CreatedBy != "system" {
		balance := bc.GetBalance(contract.CreatedBy)
		if balance < contract.Amount+contract.Fee {
			return errors.New("fonds insuffisants pour créer ce contrat")
		}
	}

	// Dans une implémentation complète, vous stockeriez le contrat dans une base de données
	// Pour cette démo, nous allons l'ajouter à un fichier JSON
	contractsDir := filepath.Join("data", "contracts")
	os.MkdirAll(contractsDir, 0755)

	// Sauvegarder le contrat dans un fichier
	contractFile := filepath.Join(contractsDir, contract.ID+".json")
	contractData, err := json.MarshalIndent(contract, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur lors de la sérialisation du contrat: %w", err)
	}

	if err := os.WriteFile(contractFile, contractData, 0644); err != nil {
		return fmt.Errorf("erreur lors de l'écriture du fichier contrat: %w", err)
	}

	// Émettre un événement (pourrait être utilisé pour un système d'événements)
	event := map[string]interface{}{
		"type":        "contract_created",
		"contract_id": contract.ID,
		"timestamp":   time.Now(),
	}

	_ = event // À utiliser plus tard

	return nil
}

// UpdateContract met à jour un contrat existant
func (bc *Blockchain) UpdateContract(contract *SmartContract) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Vérifier que le contrat existe
	if bc.GetContractByID(contract.ID) == nil {
		return errors.New("contrat non trouvé")
	}

	// Sauvegarder le contrat mis à jour
	contractsDir := filepath.Join("data", "contracts")
	contractFile := filepath.Join(contractsDir, contract.ID+".json")

	contractData, err := json.MarshalIndent(contract, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur lors de la sérialisation du contrat: %w", err)
	}

	if err := os.WriteFile(contractFile, contractData, 0644); err != nil {
		return fmt.Errorf("erreur lors de l'écriture du fichier contrat: %w", err)
	}

	// Si le contrat est exécuté, créer une transaction dans la blockchain
	if contract.Status == ContractStatusExecuted && contract.TxID == "" {
		// Créer une transaction pour ce contrat
		tx := &Transaction{
			Sender:    contract.CreatedBy,
			Recipient: contract.Recipient,
			Amount:    contract.Amount,
			Fee:       contract.Fee,
			Timestamp: time.Now(),
			Data:      fmt.Sprintf("Exécution du contrat %s", contract.ID),
		}

		// Générer l'ID de la transaction
		txData, _ := json.Marshal(tx)
		hash := sha256.Sum256(txData)
		tx.ID = hex.EncodeToString(hash[:])

		// Ajouter la transaction au mempool
		if err := bc.AddTransaction(tx); err != nil {
			return fmt.Errorf("erreur lors de l'ajout de la transaction: %w", err)
		}

		// Mettre à jour l'ID de transaction dans le contrat
		contract.TxID = tx.ID

		// Sauvegarder à nouveau le contrat avec l'ID de transaction
		contractData, _ = json.MarshalIndent(contract, "", "  ")
		os.WriteFile(contractFile, contractData, 0644)
	}

	return nil
}

// GetContractByID récupère un contrat par son ID
func (bc *Blockchain) GetContractByID(contractID string) *SmartContract {
	// Rechercher dans le système de fichiers
	contractFile := filepath.Join("data", "contracts", contractID+".json")
	if _, err := os.Stat(contractFile); os.IsNotExist(err) {
		return nil
	}

	contractData, err := os.ReadFile(contractFile)
	if err != nil {
		return nil
	}

	var contract SmartContract
	if err := json.Unmarshal(contractData, &contract); err != nil {
		return nil
	}

	return &contract
}

// GetContractsForUser récupère tous les contrats impliquant un utilisateur
func (bc *Blockchain) GetContractsForUser(username string) []*SmartContract {
	contractsDir := filepath.Join("data", "contracts")

	// S'assurer que le répertoire existe
	if _, err := os.Stat(contractsDir); os.IsNotExist(err) {
		return []*SmartContract{}
	}

	// Lire tous les fichiers de contrats
	files, err := os.ReadDir(contractsDir)
	if err != nil {
		return []*SmartContract{}
	}

	var userContracts []*SmartContract

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		contractFile := filepath.Join(contractsDir, file.Name())
		contractData, err := os.ReadFile(contractFile)
		if err != nil {
			continue
		}

		var contract SmartContract
		if err := json.Unmarshal(contractData, &contract); err != nil {
			continue
		}

		// Vérifier si l'utilisateur est impliqué dans ce contrat
		if contract.CreatedBy == username || contract.Recipient == username {
			userContracts = append(userContracts, &contract)
			continue
		}

		// Vérifier si l'utilisateur est un participant
		for _, participant := range contract.Participants {
			if participant == username {
				userContracts = append(userContracts, &contract)
				break
			}
		}
	}

	return userContracts
}

// UpdateContractToBlockchain met à jour un contrat et crée une transaction dans la blockchain
func (bc *Blockchain) UpdateContractToBlockchain(contract *SmartContract) (string, error) {
	// Vérifier que le contrat est dans un état exécutable
	if contract.Status != ContractStatusExecuted {
		return "", errors.New("le contrat n'est pas dans un état exécutable")
	}

	// Créer une transaction pour ce contrat
	tx := &Transaction{
		Sender:    contract.CreatedBy,
		Recipient: contract.Recipient,
		Amount:    contract.Amount,
		Fee:       contract.Fee,
		Timestamp: time.Now(),
		Data:      fmt.Sprintf("Exécution du contrat %s", contract.ID),
	}

	// Générer l'ID de la transaction
	txData, _ := json.Marshal(tx)
	hash := sha256.Sum256(txData)
	tx.ID = hex.EncodeToString(hash[:])

	// Ajouter la transaction au mempool
	if err := bc.AddTransaction(tx); err != nil {
		return "", fmt.Errorf("erreur lors de l'ajout de la transaction: %w", err)
	}

	// Mettre à jour le contrat avec l'ID de transaction
	contract.TxID = tx.ID

	// Sauvegarder le contrat mis à jour
	if err := bc.UpdateContract(contract); err != nil {
		return "", fmt.Errorf("erreur lors de la mise à jour du contrat: %w", err)
	}

	// Forcer la création d'un nouveau bloc pour inclure cette transaction
	// Dans une blockchain réelle, cela serait fait par le processus de minage normal
	bc.CreateBlock(contract.CreatedBy)

	return tx.ID, nil
}

// ProcessPendingContracts traite les contrats en attente (vérification expiration, exécution automatique)
func (bc *Blockchain) ProcessPendingContracts() {
	contractsDir := filepath.Join("data", "contracts")

	// S'assurer que le répertoire existe
	if _, err := os.Stat(contractsDir); os.IsNotExist(err) {
		return
	}

	// Lire tous les fichiers de contrats
	files, err := os.ReadDir(contractsDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		contractFile := filepath.Join(contractsDir, file.Name())
		contractData, err := os.ReadFile(contractFile)
		if err != nil {
			continue
		}

		var contract SmartContract
		if err := json.Unmarshal(contractData, &contract); err != nil {
			continue
		}

		// Traiter uniquement les contrats en attente
		if contract.Status != ContractStatusPending {
			continue
		}

		// Vérifier si le contrat a expiré
		if contract.CheckExpiration() {
			bc.UpdateContract(&contract)
			continue
		}

		// Vérifier si le contrat peut être exécuté automatiquement
		if contract.CanExecute() {
			tx, err := contract.ExecuteContract(bc)
			if err == nil && tx != nil {
				// Mise à jour réussie, ajouter à la blockchain
				bc.UpdateContractToBlockchain(&contract)
			}
		}
	}
}

package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ContractType représente le type de contrat intelligent
type ContractType string

const (
	// Types de contrats supportés
	ContractTransfer  ContractType = "TRANSFER"    // Transfert simple
	ContractMultiSig  ContractType = "MULTISIG"    // Transaction multi-signature
	ContractTimeLock  ContractType = "TIMELOCK"    // Transfert avec délai
	ContractCondition ContractType = "CONDITIONAL" // Transfert conditionnel
	ContractEscrow    ContractType = "ESCROW"      // Dépôt fiduciaire
)

// ContractStatus représente l'état d'un contrat
type ContractStatus string

const (
	ContractPending   ContractStatus = "PENDING"   // En attente
	ContractExecuted  ContractStatus = "EXECUTED"  // Exécuté
	ContractCancelled ContractStatus = "CANCELLED" // Annulé
	ContractExpired   ContractStatus = "EXPIRED"   // Expiré
)

// SmartContract représente un contrat intelligent
type SmartContract struct {
	ID                string            `json:"id"`
	Type              ContractType      `json:"type"`
	CreatedBy         string            `json:"createdBy"`
	CreatedAt         time.Time         `json:"createdAt"`
	ExpiresAt         time.Time         `json:"expiresAt"`
	Status            ContractStatus    `json:"status"`
	Conditions        map[string]string `json:"conditions"`
	Participants      []string          `json:"participants"`
	Approvals         map[string]bool   `json:"approvals"`
	RequiredApprovals int               `json:"requiredApprovals"`
	Amount            float64           `json:"amount"`
	Fee               float64           `json:"fee"`
	Recipient         string            `json:"recipient"`
	Data              string            `json:"data"`
	ExecutedAt        time.Time         `json:"executedAt"`
	ResultData        string            `json:"resultData"`
	TxID              string            `json:"txId,omitempty"`
}

// ComputeContractHash calcule le hash d'un contrat
func (sc *SmartContract) ComputeContractHash() string {
	contractData, _ := json.Marshal(map[string]interface{}{
		"type":         sc.Type,
		"createdBy":    sc.CreatedBy,
		"createdAt":    sc.CreatedAt,
		"expiresAt":    sc.ExpiresAt,
		"conditions":   sc.Conditions,
		"participants": sc.Participants,
		"amount":       sc.Amount,
		"recipient":    sc.Recipient,
		"data":         sc.Data,
	})

	hash := sha256.Sum256(contractData)
	return hex.EncodeToString(hash[:])
}

// NewSmartContract crée un nouveau contrat intelligent
func NewSmartContract(
	contractType ContractType,
	creator string,
	participants []string,
	requiredApprovals int,
	amount float64,
	fee float64,
	recipient string,
	data string,
	expiresIn time.Duration,
	conditions map[string]string,
) (*SmartContract, error) {

	// Validation de base
	if creator == "" {
		return nil, errors.New("le créateur du contrat ne peut pas être vide")
	}

	if amount <= 0 {
		return nil, errors.New("le montant doit être supérieur à zéro")
	}

	if recipient == "" && contractType != ContractEscrow {
		return nil, errors.New("le destinataire ne peut pas être vide")
	}

	// Vérifier que le nombre d'approbations requises est cohérent
	if requiredApprovals <= 0 || requiredApprovals > len(participants) {
		return nil, fmt.Errorf("le nombre d'approbations requises doit être entre 1 et %d", len(participants))
	}

	createdAt := time.Now()
	expiresAt := createdAt.Add(expiresIn)

	// Initialiser les approbations
	approvals := make(map[string]bool)
	for _, participant := range participants {
		approvals[participant] = false
	}

	contract := &SmartContract{
		Type:              contractType,
		CreatedBy:         creator,
		CreatedAt:         createdAt,
		ExpiresAt:         expiresAt,
		Status:            ContractPending,
		Conditions:        conditions,
		Participants:      participants,
		Approvals:         approvals,
		RequiredApprovals: requiredApprovals,
		Amount:            amount,
		Fee:               fee,
		Recipient:         recipient,
		Data:              data,
	}

	// Générer l'ID du contrat
	contract.ID = contract.ComputeContractHash()

	return contract, nil
}

// ApproveContract approuve un contrat par un participant
func (sc *SmartContract) ApproveContract(participant string) error {
	// Vérifier si le participant fait partie du contrat
	found := false
	for _, p := range sc.Participants {
		if p == participant {
			found = true
			break
		}
	}

	if !found {
		return errors.New("participant non autorisé à approuver ce contrat")
	}

	// Vérifier si le contrat est toujours en attente
	if sc.Status != ContractPending {
		return fmt.Errorf("le contrat ne peut pas être approuvé dans l'état %s", sc.Status)
	}

	// Vérifier si le contrat n'est pas expiré
	if time.Now().After(sc.ExpiresAt) {
		sc.Status = ContractExpired
		return errors.New("le contrat a expiré")
	}

	// Approuver le contrat
	sc.Approvals[participant] = true

	return nil
}

// CanExecute vérifie si un contrat peut être exécuté
func (sc *SmartContract) CanExecute() bool {
	// Vérifier si le contrat est en attente
	if sc.Status != ContractPending {
		return false
	}

	// Vérifier si le contrat n'est pas expiré
	if time.Now().After(sc.ExpiresAt) {
		sc.Status = ContractExpired
		return false
	}

	// Compter les approbations
	approvalCount := 0
	for _, approved := range sc.Approvals {
		if approved {
			approvalCount++
		}
	}

	// Vérifier si le nombre d'approbations est suffisant
	return approvalCount >= sc.RequiredApprovals
}

// ExecuteContract exécute le contrat et retourne une transaction
func (sc *SmartContract) ExecuteContract(blockchain *Blockchain) (*Transaction, error) {
	// Vérifier si le contrat peut être exécuté
	if !sc.CanExecute() {
		return nil, errors.New("le contrat ne peut pas être exécuté")
	}

	// Créer la transaction en fonction du type de contrat
	var tx *Transaction

	switch sc.Type {
	case ContractTransfer:
		// Contrat de transfert simple
		tx = NewTransaction(sc.CreatedBy, sc.Recipient, sc.Amount, sc.Fee)

	case ContractMultiSig:
		// Contrat multi-signature
		tx = NewTransaction(sc.CreatedBy, sc.Recipient, sc.Amount, sc.Fee)

	case ContractTimeLock:
		// Vérifier si le délai est respecté
		if time.Now().Before(sc.ExpiresAt) {
			return nil, errors.New("le délai d'attente n'est pas encore écoulé")
		}
		tx = NewTransaction(sc.CreatedBy, sc.Recipient, sc.Amount, sc.Fee)

	case ContractCondition:
		// Contrat conditionnel (implémentation simplifiée)
		tx = NewTransaction(sc.CreatedBy, sc.Recipient, sc.Amount, sc.Fee)

	case ContractEscrow:
		// Contrat de dépôt fiduciaire
		if sc.Recipient == "" {
			return nil, errors.New("le destinataire n'est pas défini pour ce contrat d'entiercement")
		}
		tx = NewTransaction(sc.CreatedBy, sc.Recipient, sc.Amount, sc.Fee)

	default:
		return nil, fmt.Errorf("type de contrat non supporté: %s", sc.Type)
	}

	// Marquer le contrat comme exécuté
	sc.Status = ContractExecuted
	sc.ExecutedAt = time.Now()
	sc.TxID = tx.ID

	return tx, nil
}

// CancelContract annule un contrat
func (sc *SmartContract) CancelContract(canceller string) error {
	// Vérifier si le contrat est toujours en attente
	if sc.Status != ContractPending {
		return fmt.Errorf("le contrat ne peut pas être annulé dans l'état %s", sc.Status)
	}

	// Vérifier si le demandeur est autorisé à annuler
	if sc.CreatedBy != canceller {
		isParticipant := false
		for _, p := range sc.Participants {
			if p == canceller {
				isParticipant = true
				break
			}
		}

		if !isParticipant {
			return errors.New("non autorisé à annuler ce contrat")
		}
	}

	sc.Status = ContractCancelled
	return nil
}

// UpdateContractToBlockchain intègre les contrats intelligents à la blockchain
func (bc *Blockchain) UpdateContractToBlockchain(contract *SmartContract) (string, error) {
	// Vérifier que le contrat est valide
	if contract.Status != ContractPending && contract.Status != ContractExecuted {
		return "", fmt.Errorf("contrat dans un état non valide pour la mise à jour: %s", contract.Status)
	}

	// Si le contrat est exécuté, ajouter la transaction à la blockchain
	if contract.Status == ContractExecuted && contract.TxID != "" {
		// Créer une nouvelle transaction basée sur le contrat
		tx := NewTransaction(contract.CreatedBy, contract.Recipient, contract.Amount, contract.Fee)
		tx.ID = contract.TxID

		// Ajouter la transaction au mempool
		if err := bc.AddTransaction(tx); err != nil {
			return "", err
		}

		return tx.ID, nil
	}

	// Sinon, on ne fait rien et on retourne juste l'ID du contrat
	return contract.ID, nil
}

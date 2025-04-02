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

// Types de contrats prédéfinis
const (
	ContractTransfer  ContractType = "TRANSFER"  // Transfert simple
	ContractMultiSig  ContractType = "MULTISIG"  // Signature multiple
	ContractTimeLock  ContractType = "TIMELOCK"  // Temporisé
	ContractCondition ContractType = "CONDITION" // Conditionnel
	ContractEscrow    ContractType = "ESCROW"    // Dépôt fiduciaire
)

// ContractStatus représente le statut du contrat
type ContractStatus string

// Statuts possibles pour un contrat
const (
	ContractStatusPending   ContractStatus = "PENDING"   // En attente
	ContractStatusApproved  ContractStatus = "APPROVED"  // Approuvé mais pas exécuté
	ContractStatusExecuted  ContractStatus = "EXECUTED"  // Exécuté
	ContractStatusCancelled ContractStatus = "CANCELLED" // Annulé
	ContractStatusExpired   ContractStatus = "EXPIRED"   // Expiré
	ContractStatusFailed    ContractStatus = "FAILED"    // Échec
)

// SmartContract représente un contrat intelligent dans la blockchain
type SmartContract struct {
	ID                string                `json:"id"`                // Identifiant unique du contrat
	Type              ContractType          `json:"type"`              // Type de contrat
	CreatedBy         string                `json:"createdBy"`         // Adresse du créateur
	CreatedAt         time.Time             `json:"createdAt"`         // Date de création
	Participants      []string              `json:"participants"`      // Adresses des participants
	RequiredApprovals int                   `json:"requiredApprovals"` // Nombre d'approbations requises
	Approvals         map[string]bool       `json:"approvals"`         // Approbations par participant
	Amount            float64               `json:"amount"`            // Montant du contrat
	Fee               float64               `json:"fee"`               // Frais de traitement
	Recipient         string                `json:"recipient"`         // Adresse du destinataire
	Data              string                `json:"data"`              // Données supplémentaires
	ExpiresAt         time.Time             `json:"expiresAt"`         // Date d'expiration
	Status            ContractStatus        `json:"status"`            // Statut du contrat
	ExecutedAt        time.Time             `json:"executedAt"`        // Date d'exécution
	TxID              string                `json:"txId"`              // ID de la transaction d'exécution
	Conditions        map[string]string     `json:"conditions"`        // Conditions pour l'exécution
	StateLog          []ContractStateChange `json:"stateLog"`          // Journal des changements d'état
}

// ContractStateChange représente un changement d'état du contrat
type ContractStateChange struct {
	Timestamp time.Time      `json:"timestamp"` // Moment du changement
	OldStatus ContractStatus `json:"oldStatus"` // Statut précédent
	NewStatus ContractStatus `json:"newStatus"` // Nouveau statut
	Actor     string         `json:"actor"`     // Qui a effectué le changement
	Reason    string         `json:"reason"`    // Raison du changement
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
	if amount <= 0 {
		return nil, errors.New("le montant doit être positif")
	}

	if fee < 0 {
		return nil, errors.New("les frais ne peuvent pas être négatifs")
	}

	if creator == "" {
		return nil, errors.New("créateur requis")
	}

	if recipient == "" {
		return nil, errors.New("destinataire requis")
	}

	// Validation spécifique au type de contrat
	switch contractType {
	case ContractMultiSig:
		if len(participants) < 2 {
			return nil, errors.New("au moins 2 participants sont requis pour un contrat multi-signatures")
		}
		if requiredApprovals < 1 || requiredApprovals > len(participants) {
			return nil, errors.New("nombre d'approbations requis invalide")
		}

	case ContractEscrow:
		if len(participants) < 2 {
			return nil, errors.New("au moins 2 participants sont requis pour un contrat de dépôt fiduciaire")
		}
		if requiredApprovals < 1 || requiredApprovals > len(participants) {
			return nil, errors.New("nombre d'approbations requis invalide")
		}
		if len(conditions) == 0 {
			return nil, errors.New("conditions requises pour un contrat de dépôt fiduciaire")
		}

	case ContractCondition:
		if len(conditions) == 0 {
			return nil, errors.New("conditions requises pour un contrat conditionnel")
		}
	}

	// Création du contrat
	now := time.Now()
	contract := &SmartContract{
		Type:              contractType,
		CreatedBy:         creator,
		CreatedAt:         now,
		Participants:      participants,
		RequiredApprovals: requiredApprovals,
		Approvals:         make(map[string]bool),
		Amount:            amount,
		Fee:               fee,
		Recipient:         recipient,
		Data:              data,
		ExpiresAt:         now.Add(expiresIn),
		Status:            ContractStatusPending,
		Conditions:        conditions,
		StateLog:          []ContractStateChange{},
	}

	// Générer l'ID du contrat
	contractJSON, _ := json.Marshal(map[string]interface{}{
		"type":              contract.Type,
		"createdBy":         contract.CreatedBy,
		"createdAt":         contract.CreatedAt,
		"participants":      contract.Participants,
		"requiredApprovals": contract.RequiredApprovals,
		"amount":            contract.Amount,
		"fee":               contract.Fee,
		"recipient":         contract.Recipient,
		"data":              contract.Data,
		"expiresAt":         contract.ExpiresAt,
	})

	hash := sha256.Sum256(contractJSON)
	contract.ID = hex.EncodeToString(hash[:])

	// Ajouter l'entrée initiale dans le journal d'état
	contract.logStateChange(ContractStatusPending, creator, "Création du contrat")

	// Pour les contrats de transfert simple, le créateur approuve automatiquement
	if contractType == ContractTransfer {
		contract.Approvals[creator] = true
	}

	return contract, nil
}

// ApproveContract permet à un participant d'approuver un contrat
func (c *SmartContract) ApproveContract(participant string) error {
	// Vérifier si le contrat est dans un état valide pour l'approbation
	if c.Status != ContractStatusPending {
		return fmt.Errorf("impossible d'approuver un contrat avec le statut %s", c.Status)
	}

	// Vérifier si l'utilisateur est un participant valide
	isParticipant := false
	for _, p := range c.Participants {
		if p == participant {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		return errors.New("seuls les participants désignés peuvent approuver ce contrat")
	}

	// Vérifier si le participant a déjà approuvé
	if approved, exists := c.Approvals[participant]; exists && approved {
		return errors.New("ce participant a déjà approuvé le contrat")
	}

	// Approuver le contrat
	c.Approvals[participant] = true

	// Enregistrer le changement d'état
	c.logStateChange(c.Status, participant, "Approbation du contrat")

	return nil
}

// CancelContract annule un contrat
func (c *SmartContract) CancelContract(actor string) error {
	// Vérifier si le contrat est dans un état valide pour l'annulation
	if c.Status != ContractStatusPending {
		return fmt.Errorf("impossible d'annuler un contrat avec le statut %s", c.Status)
	}

	// Vérifier si l'utilisateur est autorisé à annuler
	isCreator := c.CreatedBy == actor
	isParticipant := false
	for _, p := range c.Participants {
		if p == actor {
			isParticipant = true
			break
		}
	}

	if !isCreator && !isParticipant {
		return errors.New("seul le créateur ou un participant peut annuler ce contrat")
	}

	// Annuler le contrat
	oldStatus := c.Status
	c.Status = ContractStatusCancelled

	// Enregistrer le changement d'état
	c.logStateChange(oldStatus, actor, "Annulation du contrat")

	return nil
}

// ExecuteContract exécute un contrat
func (c *SmartContract) ExecuteContract(bc *Blockchain) (*Transaction, error) {
	// Vérifier si le contrat peut être exécuté
	if !c.CanExecute() {
		return nil, errors.New("le contrat ne peut pas être exécuté actuellement")
	}

	// Vérifier si le contrat a expiré
	if time.Now().After(c.ExpiresAt) {
		c.Status = ContractStatusExpired
		c.logStateChange(ContractStatusPending, "system", "Contrat expiré")
		return nil, errors.New("le contrat a expiré")
	}

	// Créer la transaction
	tx := &Transaction{
		Timestamp: time.Now(),
		Sender:    c.CreatedBy,
		Recipient: c.Recipient,
		Amount:    c.Amount,
		Fee:       c.Fee,
		Data:      c.Data,
	}

	// Générer l'ID de la transaction
	txData, _ := json.Marshal(tx)
	hash := sha256.Sum256(txData)
	tx.ID = hex.EncodeToString(hash[:])

	// Mettre à jour le statut du contrat
	oldStatus := c.Status
	c.Status = ContractStatusExecuted
	c.ExecutedAt = tx.Timestamp

	// Enregistrer le changement d'état
	c.logStateChange(oldStatus, "system", "Exécution automatique du contrat")

	return tx, nil
}

// CanExecute vérifie si un contrat peut être exécuté
func (c *SmartContract) CanExecute() bool {
	// Vérifier le statut
	if c.Status != ContractStatusPending {
		return false
	}

	// Vérifier l'expiration
	if time.Now().After(c.ExpiresAt) {
		return false
	}

	// Vérifier les approbations selon le type de contrat
	switch c.Type {
	case ContractTransfer:
		// Un contrat de transfert simple nécessite seulement l'approbation du créateur
		return c.Approvals[c.CreatedBy]

	case ContractMultiSig, ContractEscrow:
		// Compter les approbations
		approvalCount := 0
		for _, approved := range c.Approvals {
			if approved {
				approvalCount++
			}
		}
		return approvalCount >= c.RequiredApprovals

	case ContractTimeLock:
		// Un contrat temporisé nécessite que la date de déverrouillage soit atteinte
		lockTime, exists := c.Conditions["unlock_time"]
		if !exists {
			return false
		}

		unlockTime, err := time.Parse(time.RFC3339, lockTime)
		if err != nil {
			return false
		}

		return time.Now().After(unlockTime) && c.Approvals[c.CreatedBy]

	case ContractCondition:
		// Pour un contrat conditionnel, les conditions doivent être vérifiées ailleurs
		// (dans une implémentation réelle, un oracle externe pourrait vérifier ces conditions)
		return c.Approvals[c.CreatedBy]
	}

	return false
}

// logStateChange enregistre un changement d'état dans le journal du contrat
func (c *SmartContract) logStateChange(oldStatus ContractStatus, actor, reason string) {
	stateChange := ContractStateChange{
		Timestamp: time.Now(),
		OldStatus: oldStatus,
		NewStatus: c.Status,
		Actor:     actor,
		Reason:    reason,
	}

	c.StateLog = append(c.StateLog, stateChange)
}

// GetApprovalCount compte le nombre d'approbations reçues
func (c *SmartContract) GetApprovalCount() int {
	count := 0
	for _, approved := range c.Approvals {
		if approved {
			count++
		}
	}
	return count
}

// IsParticipant vérifie si une adresse est un participant du contrat
func (c *SmartContract) IsParticipant(address string) bool {
	for _, p := range c.Participants {
		if p == address {
			return true
		}
	}
	return false
}

// CheckExpiration vérifie si le contrat a expiré et met à jour son statut
func (c *SmartContract) CheckExpiration() bool {
	if c.Status == ContractStatusPending && time.Now().After(c.ExpiresAt) {
		c.Status = ContractStatusExpired
		c.logStateChange(ContractStatusPending, "system", "Contrat expiré")
		return true
	}
	return false
}

// ToJSON convertit le contrat en JSON
func (c *SmartContract) ToJSON() (string, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
cursor blockchain/blockchain_extensions.go
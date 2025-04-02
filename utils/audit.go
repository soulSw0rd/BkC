package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// EventType représente le type d'événement d'audit
type EventType string

// Types d'événements d'audit prédéfinis
const (
	// Événements système
	EventTypeServerStarted EventType = "SERVER_STARTED"
	EventTypeServerStopped EventType = "SERVER_STOPPED"
	EventTypeConfigChanged EventType = "CONFIG_CHANGED"

	// Événements d'authentification
	EventTypeLogin          EventType = "LOGIN"
	EventTypeLogout         EventType = "LOGOUT"
	EventTypeLoginFailed    EventType = "LOGIN_FAILED"
	EventTypePasswordChange EventType = "PASSWORD_CHANGE"
	EventTypeUserCreated    EventType = "USER_CREATED"

	// Événements de blockchain
	EventTypeBlockCreated EventType = "BLOCK_CREATED"
	EventTypeBlockMined   EventType = "BLOCK_MINED"

	// Événements de transactions
	EventTypeTransactionAdded  EventType = "TRANSACTION_ADDED"
	EventTypeTransactionFailed EventType = "TRANSACTION_FAILED"

	// Événements de contrats
	EventTypeContractCreated   EventType = "CONTRACT_CREATED"
	EventTypeContractApproved  EventType = "CONTRACT_APPROVED"
	EventTypeContractCancelled EventType = "CONTRACT_CANCELLED"
	EventTypeContractExecuted  EventType = "CONTRACT_EXECUTED"
	EventTypeContractListed    EventType = "CONTRACT_LISTED"
	EventTypeContractViewed    EventType = "CONTRACT_VIEWED"
	EventTypeContractAction    EventType = "CONTRACT_ACTION"

	// Événements d'interface utilisateur
	EventTypeUIAccess EventType = "UI_ACCESS"

	// Événements de sécurité
	EventTypeSecurityAlert EventType = "SECURITY_ALERT"
)

// RiskLevel représente un niveau de risque pour un événement
type RiskLevel int

// Niveaux de risque prédéfinis
const (
	RiskLow      RiskLevel = 1
	RiskMedium   RiskLevel = 2
	RiskHigh     RiskLevel = 3
	RiskCritical RiskLevel = 4
)

// AuditEntry représente une entrée dans la piste d'audit
type AuditEntry struct {
	Timestamp   time.Time   `json:"timestamp"`
	Type        EventType   `json:"type"`
	UserID      string      `json:"user_id"`
	IPAddress   string      `json:"ip_address"`
	Description string      `json:"description"`
	Data        interface{} `json:"data,omitempty"`
	RiskLevel   RiskLevel   `json:"risk_level"`
	PrevHash    string      `json:"prev_hash"`
	Hash        string      `json:"hash"`
}

// AuditTrail représente la piste d'audit complète
type AuditTrail struct {
	Entries  []*AuditEntry `json:"entries"`
	filePath string
	mu       sync.RWMutex
}

// Variables globales
var (
	auditTrail *AuditTrail
	auditMutex sync.RWMutex
)

// InitAuditTrail initialise la piste d'audit
func InitAuditTrail(filePath string) error {
	auditMutex.Lock()
	defer auditMutex.Unlock()

	// S'assurer que le répertoire existe
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("impossible de créer le répertoire d'audit: %w", err)
	}

	// Vérifier si le fichier existe déjà
	var trail *AuditTrail
	if _, err := os.Stat(filePath); err == nil {
		// Lire le fichier existant
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("erreur lors de la lecture du fichier de piste d'audit: %w", err)
		}

		trail = &AuditTrail{filePath: filePath}
		if err := json.Unmarshal(data, &trail.Entries); err != nil {
			// Si erreur lors de la désérialisation, créer une nouvelle piste
			trail = &AuditTrail{
				Entries:  []*AuditEntry{},
				filePath: filePath,
			}
		}
	} else {
		// Créer une nouvelle piste d'audit
		trail = &AuditTrail{
			Entries:  []*AuditEntry{},
			filePath: filePath,
		}
	}

	// Si la piste est vide, ajouter un événement de génération
	if len(trail.Entries) == 0 {
		genesisEntry := &AuditEntry{
			Timestamp:   time.Now(),
			Type:        "GENESIS",
			UserID:      "system",
			IPAddress:   "0.0.0.0",
			Description: "Création de la piste d'audit",
			RiskLevel:   RiskLow,
			PrevHash:    "",
		}
		genesisEntry.Hash = genesisEntry.ComputeHash()
		trail.Entries = append(trail.Entries, genesisEntry)
	}

	// Vérifier l'intégrité de la chaîne
	if valid, err := trail.ValidateChain(); !valid {
		return fmt.Errorf("la piste d'audit est corrompue: %w", err)
	}

	auditTrail = trail
	return auditTrail.Save()
}

// AddEntry ajoute une nouvelle entrée à la piste d'audit
func (trail *AuditTrail) AddEntry(eventType EventType, userID, ipAddress, description string, riskLevel RiskLevel, data interface{}) *AuditEntry {
	trail.mu.Lock()
	defer trail.mu.Unlock()

	var prevHash string
	if len(trail.Entries) > 0 {
		prevHash = trail.Entries[len(trail.Entries)-1].Hash
	}

	entry := &AuditEntry{
		Timestamp:   time.Now(),
		Type:        eventType,
		UserID:      userID,
		IPAddress:   ipAddress,
		Description: description,
		Data:        data,
		RiskLevel:   riskLevel,
		PrevHash:    prevHash,
	}

	// Calculer le hash de l'entrée
	entry.Hash = entry.ComputeHash()

	// Ajouter l'entrée à la piste d'audit
	trail.Entries = append(trail.Entries, entry)

	// Sauvegarder la piste d'audit périodiquement (tous les 10 événements)
	if len(trail.Entries)%10 == 0 {
		go trail.Save()
	}

	return entry
}

// ComputeHash calcule le hash d'une entrée d'audit
func (e *AuditEntry) ComputeHash() string {
	// Ne pas inclure le hash lui-même dans le calcul
	record := fmt.Sprintf(
		"%s:%s:%s:%s:%s:%v:%d:%s",
		e.Timestamp.Format(time.RFC3339Nano),
		e.Type,
		e.UserID,
		e.IPAddress,
		e.Description,
		e.Data,
		e.RiskLevel,
		e.PrevHash,
	)

	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// ValidateChain vérifie l'intégrité de la piste d'audit
func (trail *AuditTrail) ValidateChain() (bool, error) {
	trail.mu.RLock()
	defer trail.mu.RUnlock()

	if len(trail.Entries) == 0 {
		return true, nil
	}

	for i := 1; i < len(trail.Entries); i++ {
		currentEntry := trail.Entries[i]
		previousEntry := trail.Entries[i-1]

		// Vérifier que le hash précédent correspond
		if currentEntry.PrevHash != previousEntry.Hash {
			return false, fmt.Errorf("hash précédent invalide pour l'entrée %d", i)
		}

		// Recalculer le hash pour vérifier
		calculatedHash := currentEntry.ComputeHash()
		if calculatedHash != currentEntry.Hash {
			return false, fmt.Errorf("hash invalide pour l'entrée %d", i)
		}
	}

	return true, nil
}

// Save sauvegarde la piste d'audit dans un fichier
func (trail *AuditTrail) Save() error {
	trail.mu.RLock()
	defer trail.mu.RUnlock()

	data, err := json.MarshalIndent(trail.Entries, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur lors de la sérialisation de la piste d'audit: %w", err)
	}

	if err := ioutil.WriteFile(trail.filePath, data, 0644); err != nil {
		return fmt.Errorf("erreur lors de l'écriture du fichier de piste d'audit: %w", err)
	}

	return nil
}

// LogAuditEvent enregistre un événement d'audit
func LogAuditEvent(eventType EventType, userID, ipAddress, description string, riskLevel RiskLevel, data interface{}) {
	auditMutex.RLock()
	defer auditMutex.RUnlock()

	if auditTrail != nil {
		auditTrail.AddEntry(eventType, userID, ipAddress, description, riskLevel, data)
	}
}

// GetAuditEntries récupère les entrées d'audit
func GetAuditEntries(limit, offset int) []*AuditEntry {
	auditMutex.RLock()
	defer auditMutex.RUnlock()

	if auditTrail == nil {
		return []*AuditEntry{}
	}

	auditTrail.mu.RLock()
	defer auditTrail.mu.RUnlock()

	// Calculer les indices de début et de fin
	start := len(auditTrail.Entries) - offset - limit
	if start < 0 {
		start = 0
	}
	end := len(auditTrail.Entries) - offset
	if end > len(auditTrail.Entries) {
		end = len(auditTrail.Entries)
	}
	if end < 0 {
		end = 0
	}

	// Créer une copie des entrées
	result := make([]*AuditEntry, end-start)
	copy(result, auditTrail.Entries[start:end])

	return result
}

// GetAuditEntriesByUser récupère les entrées d'audit pour un utilisateur spécifique
func GetAuditEntriesByUser(userID string, limit int) []*AuditEntry {
	auditMutex.RLock()
	defer auditMutex.RUnlock()

	if auditTrail == nil {
		return []*AuditEntry{}
	}

	auditTrail.mu.RLock()
	defer auditTrail.mu.RUnlock()

	result := []*AuditEntry{}
	for i := len(auditTrail.Entries) - 1; i >= 0 && len(result) < limit; i-- {
		if auditTrail.Entries[i].UserID == userID {
			result = append(result, auditTrail.Entries[i])
		}
	}

	return result
}

// GetAuditEntriesByType récupère les entrées d'audit pour un type spécifique
func GetAuditEntriesByType(eventType EventType, limit int) []*AuditEntry {
	auditMutex.RLock()
	defer auditMutex.RUnlock()

	if auditTrail == nil {
		return []*AuditEntry{}
	}

	auditTrail.mu.RLock()
	defer auditTrail.mu.RUnlock()

	result := []*AuditEntry{}
	for i := len(auditTrail.Entries) - 1; i >= 0 && len(result) < limit; i-- {
		if auditTrail.Entries[i].Type == eventType {
			result = append(result, auditTrail.Entries[i])
		}
	}

	return result
}

// GetAuditEntriesByRiskLevel récupère les entrées d'audit pour un niveau de risque spécifique
func GetAuditEntriesByRiskLevel(riskLevel RiskLevel, limit int) []*AuditEntry {
	auditMutex.RLock()
	defer auditMutex.RUnlock()

	if auditTrail == nil {
		return []*AuditEntry{}
	}

	auditTrail.mu.RLock()
	defer auditTrail.mu.RUnlock()

	result := []*AuditEntry{}
	for i := len(auditTrail.Entries) - 1; i >= 0 && len(result) < limit; i-- {
		if auditTrail.Entries[i].RiskLevel >= riskLevel {
			result = append(result, auditTrail.Entries[i])
		}
	}

	return result
}

// GetAuditTrailValidity vérifie et retourne l'état de validité de la piste d'audit
func GetAuditTrailValidity() (bool, error) {
	auditMutex.RLock()
	defer auditMutex.RUnlock()

	if auditTrail == nil {
		return false, fmt.Errorf("piste d'audit non initialisée")
	}

	return auditTrail.ValidateChain()
}

// GetEventTypeDescription retourne une description pour un type d'événement
func GetEventTypeDescription(eventType EventType) string {
	descriptions := map[EventType]string{
		// Événements système
		EventTypeServerStarted: "Démarrage du serveur",
		EventTypeServerStopped: "Arrêt du serveur",
		EventTypeConfigChanged: "Modification de la configuration",

		// Événements d'authentification
		EventTypeLogin:          "Connexion",
		EventTypeLogout:         "Déconnexion",
		EventTypeLoginFailed:    "Échec de connexion",
		EventTypePasswordChange: "Changement de mot de passe",
		EventTypeUserCreated:    "Création d'utilisateur",

		// Événements de blockchain
		EventTypeBlockCreated: "Création de bloc",
		EventTypeBlockMined:   "Minage de bloc",

		// Événements de transactions
		EventTypeTransactionAdded:  "Ajout de transaction",
		EventTypeTransactionFailed: "Échec de transaction",

		// Événements de contrats
		EventTypeContractCreated:   "Création de contrat",
		EventTypeContractApproved:  "Approbation de contrat",
		EventTypeContractCancelled: "Annulation de contrat",
		EventTypeContractExecuted:  "Exécution de contrat",
		EventTypeContractListed:    "Liste des contrats",
		EventTypeContractViewed:    "Visualisation de contrat",
		EventTypeContractAction:    "Action sur contrat",

		// Événements d'interface utilisateur
		EventTypeUIAccess: "Accès interface utilisateur",

		// Événements de sécurité
		EventTypeSecurityAlert: "Alerte de sécurité",
	}

	if desc, ok := descriptions[eventType]; ok {
		return desc
	}
	return "Type d'événement inconnu"
}

// GetRiskLevelDescription retourne une description pour un niveau de risque
func GetRiskLevelDescription(level RiskLevel) string {
	descriptions := map[RiskLevel]string{
		RiskLow:      "Faible - Information normale",
		RiskMedium:   "Moyen - Attention requise",
		RiskHigh:     "Élevé - Potentiellement dangereux",
		RiskCritical: "Critique - Action immédiate nécessaire",
	}

	if desc, ok := descriptions[level]; ok {
		return desc
	}
	return "Niveau de risque inconnu"
}

// ExportAuditTrail exporte la piste d'audit au format JSON pour une période donnée
func ExportAuditTrail(startTime, endTime time.Time, filePath string) error {
	auditMutex.RLock()
	defer auditMutex.RUnlock()

	if auditTrail == nil {
		return fmt.Errorf("piste d'audit non initialisée")
	}

	auditTrail.mu.RLock()
	defer auditTrail.mu.RUnlock()

	// Filtrer les entrées par période
	var filteredEntries []*AuditEntry
	for _, entry := range auditTrail.Entries {
		if (entry.Timestamp.Equal(startTime) || entry.Timestamp.After(startTime)) &&
			(entry.Timestamp.Equal(endTime) || entry.Timestamp.Before(endTime)) {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	// Sérialiser
	data, err := json.MarshalIndent(filteredEntries, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur lors de la sérialisation des entrées d'audit: %w", err)
	}

	// Écrire dans le fichier
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("erreur lors de l'écriture du fichier d'export: %w", err)
	}

	return nil
}

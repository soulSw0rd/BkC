package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RiskFactor représente un facteur de risque pour l'évaluation de sécurité
type RiskFactor string

// Facteurs de risque prédéfinis
const (
	FactorLoginAttempts     RiskFactor = "LOGIN_ATTEMPTS"
	FactorLoginFails        RiskFactor = "LOGIN_FAILS"
	FactorUnusualIP         RiskFactor = "UNUSUAL_IP"
	FactorUnusualTime       RiskFactor = "UNUSUAL_TIME"
	FactorTransactionVolume RiskFactor = "TRANSACTION_VOLUME"
	FactorLargeTransactions RiskFactor = "LARGE_TRANSACTIONS"
	FactorSessionDuration   RiskFactor = "SESSION_DURATION"
	FactorApiUsage          RiskFactor = "API_USAGE"
	FactorBlockMining       RiskFactor = "BLOCK_MINING"
)

// SecurityScore représente le score de sécurité global d'un utilisateur
type SecurityScore struct {
	Score           float64                `json:"score"`
	LastUpdated     time.Time              `json:"last_updated"`
	Factors         map[RiskFactor]float64 `json:"factors"`
	RecentEvents    []string               `json:"recent_events"`
	AnomalyDetected bool                   `json:"anomaly_detected"`
	ThreatLevel     RiskLevel              `json:"threat_level"`
}

// UserSecurityProfile représente le profil de sécurité complet d'un utilisateur
type UserSecurityProfile struct {
	UserID            string               `json:"user_id"`
	IPAddresses       map[string]time.Time `json:"ip_addresses"`
	LastLogin         time.Time            `json:"last_login"`
	LoginTimes        []time.Time          `json:"login_times"`
	FailedLogins      int                  `json:"failed_logins"`
	Transactions      int                  `json:"transactions"`
	LargeTransactions int                  `json:"large_transactions"`
	CurrentScore      SecurityScore        `json:"current_score"`
	HistoricalScores  []SecurityScore      `json:"historical_scores"`
	Alerts            []SecurityAlert      `json:"alerts"`
}

// SecurityAlert représente une alerte de sécurité
type SecurityAlert struct {
	Timestamp    time.Time `json:"timestamp"`
	Level        RiskLevel `json:"level"`
	Message      string    `json:"message"`
	RelatedEvent string    `json:"related_event"`
	Resolved     bool      `json:"resolved"`
}

// SecurityRiskAssessment gère l'évaluation des risques de sécurité
type SecurityRiskAssessment struct {
	Profiles      map[string]*UserSecurityProfile `json:"profiles"`
	GlobalThreats map[string]int                  `json:"global_threats"`
	FactorWeights map[RiskFactor]float64          `json:"factor_weights"`
	filePath      string
	mu            sync.RWMutex
}

// Variables globales
var (
	securityAssessment *SecurityRiskAssessment
	securityMutex      sync.RWMutex
)

// InitSecurityRiskAssessment initialise le système d'évaluation des risques
func InitSecurityRiskAssessment(filePath string) error {
	securityMutex.Lock()
	defer securityMutex.Unlock()

	// S'assurer que le répertoire existe
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("impossible de créer le répertoire d'évaluation des risques: %w", err)
	}

	// Vérifier si le fichier existe déjà
	var assessment *SecurityRiskAssessment
	if _, err := os.Stat(filePath); err == nil {
		// Lire le fichier existant
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("erreur lors de la lecture du fichier d'évaluation des risques: %w", err)
		}

		assessment = &SecurityRiskAssessment{filePath: filePath}
		if err := json.Unmarshal(data, assessment); err != nil {
			// Si erreur lors de la désérialisation, créer une nouvelle évaluation
			assessment = newSecurityRiskAssessment(filePath)
		}
	} else {
		// Créer une nouvelle évaluation des risques
		assessment = newSecurityRiskAssessment(filePath)
	}

	securityAssessment = assessment
	return securityAssessment.Save()
}

// newSecurityRiskAssessment crée une nouvelle instance d'évaluation des risques
func newSecurityRiskAssessment(filePath string) *SecurityRiskAssessment {
	return &SecurityRiskAssessment{
		Profiles:      make(map[string]*UserSecurityProfile),
		GlobalThreats: make(map[string]int),
		FactorWeights: map[RiskFactor]float64{
			FactorLoginAttempts:     0.2,
			FactorLoginFails:        0.3,
			FactorUnusualIP:         0.25,
			FactorUnusualTime:       0.15,
			FactorTransactionVolume: 0.1,
			FactorLargeTransactions: 0.2,
			FactorSessionDuration:   0.05,
			FactorApiUsage:          0.1,
			FactorBlockMining:       0.05,
		},
		filePath: filePath,
	}
}

// Save sauvegarde l'évaluation des risques dans un fichier
func (sra *SecurityRiskAssessment) Save() error {
	sra.mu.RLock()
	defer sra.mu.RUnlock()

	data, err := json.MarshalIndent(sra, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur lors de la sérialisation de l'évaluation des risques: %w", err)
	}

	if err := ioutil.WriteFile(sra.filePath, data, 0644); err != nil {
		return fmt.Errorf("erreur lors de l'écriture du fichier d'évaluation des risques: %w", err)
	}

	return nil
}

// GetUserProfile récupère le profil de sécurité d'un utilisateur, en le créant si nécessaire
func (sra *SecurityRiskAssessment) GetUserProfile(userID string) *UserSecurityProfile {
	sra.mu.Lock()
	defer sra.mu.Unlock()

	profile, exists := sra.Profiles[userID]
	if !exists {
		// Créer un nouveau profil
		profile = &UserSecurityProfile{
			UserID:      userID,
			IPAddresses: make(map[string]time.Time),
			LoginTimes:  []time.Time{},
			CurrentScore: SecurityScore{
				Score:           100, // Commence avec un score parfait
				LastUpdated:     time.Now(),
				Factors:         make(map[RiskFactor]float64),
				RecentEvents:    []string{},
				AnomalyDetected: false,
				ThreatLevel:     RiskLow,
			},
			HistoricalScores: []SecurityScore{},
			Alerts:           []SecurityAlert{},
		}
		sra.Profiles[userID] = profile
	}

	return profile
}

// RecordLogin enregistre une connexion réussie
func (sra *SecurityRiskAssessment) RecordLogin(userID, ipAddress string) {
	profile := sra.GetUserProfile(userID)

	sra.mu.Lock()
	defer sra.mu.Unlock()

	now := time.Now()

	// Enregistrer l'adresse IP
	profile.IPAddresses[ipAddress] = now

	// Mettre à jour les dates de connexion
	profile.LastLogin = now
	profile.LoginTimes = append(profile.LoginTimes, now)

	// Limiter l'historique des connexions aux 100 dernières
	if len(profile.LoginTimes) > 100 {
		profile.LoginTimes = profile.LoginTimes[len(profile.LoginTimes)-100:]
	}

	// Réinitialiser les échecs de connexion
	profile.FailedLogins = 0

	// Réévaluer le score de sécurité
	sra.EvaluateUserSecurity(userID, ipAddress)

	// Sauvegarder périodiquement
	go sra.Save()
}

// RecordFailedLogin enregistre un échec de connexion
func (sra *SecurityRiskAssessment) RecordFailedLogin(userID, ipAddress string) {
	profile := sra.GetUserProfile(userID)

	sra.mu.Lock()
	defer sra.mu.Unlock()

	// Incrémenter le compteur d'échecs de connexion
	profile.FailedLogins++

	// Vérifier s'il faut générer une alerte
	if profile.FailedLogins >= 3 {
		alert := SecurityAlert{
			Timestamp:    time.Now(),
			Level:        RiskMedium,
			Message:      fmt.Sprintf("Plusieurs échecs de connexion détectés (%d)", profile.FailedLogins),
			RelatedEvent: "LOGIN_FAILED",
			Resolved:     false,
		}
		profile.Alerts = append(profile.Alerts, alert)

		// Mettre à jour le niveau de menace
		if profile.FailedLogins >= 5 {
			profile.CurrentScore.ThreatLevel = RiskHigh
		} else {
			profile.CurrentScore.ThreatLevel = RiskMedium
		}
	}

	// Réévaluer le score de sécurité
	sra.EvaluateUserSecurity(userID, ipAddress)

	// Sauvegarder
	go sra.Save()
}

// RecordTransaction enregistre une transaction
func (sra *SecurityRiskAssessment) RecordTransaction(userID string, amount float64) {
	profile := sra.GetUserProfile(userID)

	sra.mu.Lock()
	defer sra.mu.Unlock()

	// Incrémenter le compteur de transactions
	profile.Transactions++

	// Déterminer si c'est une transaction importante
	if amount > 100.0 { // Seuil arbitraire
		profile.LargeTransactions++
	}

	// Réévaluer le score de sécurité
	sra.EvaluateUserSecurity(userID, "")

	// Sauvegarder périodiquement
	if profile.Transactions%10 == 0 {
		go sra.Save()
	}
}

// IsIPUnusual détermine si une adresse IP est inhabituelle pour un utilisateur
func (sra *SecurityRiskAssessment) IsIPUnusual(userID, ipAddress string) bool {
	profile := sra.GetUserProfile(userID)

	sra.mu.RLock()
	defer sra.mu.RUnlock()

	// Si c'est la première connexion de l'utilisateur, ce n'est pas inhabituel
	if len(profile.IPAddresses) == 0 {
		return false
	}

	// Vérifier si l'adresse IP a déjà été utilisée
	_, exists := profile.IPAddresses[ipAddress]
	return !exists
}

// IsLoginTimeUnusual détermine si l'heure de connexion est inhabituelle
func (sra *SecurityRiskAssessment) IsLoginTimeUnusual(userID string) bool {
	profile := sra.GetUserProfile(userID)

	sra.mu.RLock()
	defer sra.mu.RUnlock()

	// Si moins de 5 connexions historiques, ce n'est pas inhabituel
	if len(profile.LoginTimes) < 5 {
		return false
	}

	// Obtenir l'heure actuelle
	now := time.Now()
	hour := now.Hour()

	// Compter combien de connexions passées ont eu lieu dans la même tranche horaire
	sameHourLogins := 0
	for _, loginTime := range profile.LoginTimes {
		if loginTime.Hour() == hour {
			sameHourLogins++
		}
	}

	// Si moins de 10% des connexions passées ont eu lieu à cette heure, c'est inhabituel
	threshold := float64(len(profile.LoginTimes)) * 0.1
	return float64(sameHourLogins) < threshold
}

// EvaluateUserSecurity évalue le score de sécurité d'un utilisateur
func (sra *SecurityRiskAssessment) EvaluateUserSecurity(userID, ipAddress string) {
	profile := sra.GetUserProfile(userID)

	// Vérifier si nous avons besoin de l'adresse IP
	if ipAddress == "" && len(profile.IPAddresses) > 0 {
		// Utiliser la dernière adresse IP connue
		for ip := range profile.IPAddresses {
			ipAddress = ip
			break
		}
	}

	// Calculer les scores des facteurs individuels
	factors := make(map[RiskFactor]float64)

	// 1. Facteur: tentatives de connexion
	loginAttemptsScore := 100.0
	if profile.FailedLogins > 0 {
		// Réduire le score en fonction du nombre d'échecs
		loginAttemptsScore = math.Max(0, 100-float64(profile.FailedLogins)*20)
	}
	factors[FactorLoginFails] = loginAttemptsScore

	// 2. Facteur: IP inhabituelle
	unusualIPScore := 100.0
	if sra.IsIPUnusual(userID, ipAddress) {
		unusualIPScore = 50.0 // Score réduit pour une IP inhabituelle
	}
	factors[FactorUnusualIP] = unusualIPScore

	// 3. Facteur: heure inhabituelle
	unusualTimeScore := 100.0
	if sra.IsLoginTimeUnusual(userID) {
		unusualTimeScore = 70.0 // Score réduit pour une heure inhabituelle
	}
	factors[FactorUnusualTime] = unusualTimeScore

	// 4. Facteur: volume de transactions
	transactionVolumeScore := 100.0
	if profile.Transactions > 100 { // Seuil arbitraire
		// Réduire légèrement le score pour un volume élevé
		transactionVolumeScore = 90.0
	}
	factors[FactorTransactionVolume] = transactionVolumeScore

	// 5. Facteur: transactions importantes
	largeTransactionsScore := 100.0
	if profile.LargeTransactions > 10 { // Seuil arbitraire
		// Réduire le score en fonction du nombre de transactions importantes
		largeTransactionsScore = math.Max(60, 100-float64(profile.LargeTransactions)*2)
	}
	factors[FactorLargeTransactions] = largeTransactionsScore

	// Calculer le score global pondéré
	var totalScore float64 = 0
	var totalWeight float64 = 0

	for factor, score := range factors {
		weight := sra.FactorWeights[factor]
		totalScore += score * weight
		totalWeight += weight
	}

	// Normaliser le score
	if totalWeight > 0 {
		totalScore = totalScore / totalWeight
	} else {
		totalScore = 100 // Par défaut, score parfait si aucun facteur
	}

	// Déterminer le niveau de menace
	var threatLevel RiskLevel = RiskLow
	switch {
	case totalScore < 50:
		threatLevel = RiskCritical
	case totalScore < 70:
		threatLevel = RiskHigh
	case totalScore < 85:
		threatLevel = RiskMedium
	default:
		threatLevel = RiskLow
	}

	// Vérifier les anomalies (changement significatif de score)
	anomalyDetected := false
	if len(profile.HistoricalScores) > 0 {
		lastScore := profile.HistoricalScores[len(profile.HistoricalScores)-1].Score
		if math.Abs(lastScore-totalScore) > 15 {
			anomalyDetected = true
		}
	}

	// Mettre à jour le score
	newScore := SecurityScore{
		Score:           totalScore,
		LastUpdated:     time.Now(),
		Factors:         factors,
		RecentEvents:    []string{}, // À remplir si nécessaire
		AnomalyDetected: anomalyDetected,
		ThreatLevel:     threatLevel,
	}

	// Archiver l'ancien score
	profile.HistoricalScores = append(profile.HistoricalScores, profile.CurrentScore)

	// Limiter l'historique des scores aux 50 derniers
	if len(profile.HistoricalScores) > 50 {
		profile.HistoricalScores = profile.HistoricalScores[len(profile.HistoricalScores)-50:]
	}

	// Mettre à jour le score courant
	profile.CurrentScore = newScore

	// Vérifier si une alerte doit être générée en raison d'une anomalie
	if anomalyDetected {
		alert := SecurityAlert{
			Timestamp:    time.Now(),
			Level:        RiskMedium,
			Message:      "Changement significatif dans le score de sécurité détecté",
			RelatedEvent: "SECURITY_SCORE_CHANGE",
			Resolved:     false,
		}
		profile.Alerts = append(profile.Alerts, alert)
	}
}

// GetUserSecurityScore récupère le score de sécurité actuel d'un utilisateur
func GetUserSecurityScore(userID string) SecurityScore {
	securityMutex.RLock()
	defer securityMutex.RUnlock()

	if securityAssessment == nil {
		return SecurityScore{
			Score:           100,
			LastUpdated:     time.Now(),
			Factors:         make(map[RiskFactor]float64),
			RecentEvents:    []string{},
			AnomalyDetected: false,
			ThreatLevel:     RiskLow,
		}
	}

	profile := securityAssessment.GetUserProfile(userID)
	return profile.CurrentScore
}

// GetUserAlerts récupère les alertes actives pour un utilisateur
func GetUserAlerts(userID string) []SecurityAlert {
	securityMutex.RLock()
	defer securityMutex.RUnlock()

	if securityAssessment == nil {
		return []SecurityAlert{}
	}

	profile := securityAssessment.GetUserProfile(userID)

	// Filtrer pour n'obtenir que les alertes non résolues
	activeAlerts := []SecurityAlert{}
	for _, alert := range profile.Alerts {
		if !alert.Resolved {
			activeAlerts = append(activeAlerts, alert)
		}
	}

	return activeAlerts
}

// ResolveAlert marque une alerte comme résolue
func ResolveAlert(userID string, timestamp time.Time) {
	securityMutex.Lock()
	defer securityMutex.Unlock()

	if securityAssessment == nil {
		return
	}

	profile := securityAssessment.GetUserProfile(userID)

	// Rechercher et marquer l'alerte comme résolue
	for i, alert := range profile.Alerts {
		if alert.Timestamp.Equal(timestamp) {
			profile.Alerts[i].Resolved = true
			break
		}
	}

	// Sauvegarder les changements
	go securityAssessment.Save()
}

// RegisterGlobalThreat enregistre une menace globale
func RegisterGlobalThreat(threatType string) {
	securityMutex.Lock()
	defer securityMutex.Unlock()

	if securityAssessment == nil {
		return
	}

	// Incrémenter le compteur pour ce type de menace
	securityAssessment.GlobalThreats[threatType]++

	// Journaliser l'événement
	LogAuditEvent(
		EventSecurityAlert,
		"system",
		"0.0.0.0",
		fmt.Sprintf("Menace globale détectée: %s", threatType),
		RiskHigh,
		map[string]interface{}{
			"threat_type": threatType,
			"count":       securityAssessment.GlobalThreats[threatType],
		},
	)

	// Sauvegarder les changements
	go securityAssessment.Save()
}

// GetSecurityFactorDescription retourne une description pour un facteur de risque
func GetSecurityFactorDescription(factor RiskFactor) string {
	descriptions := map[RiskFactor]string{
		FactorLoginAttempts:     "Nombre de tentatives de connexion récentes",
		FactorLoginFails:        "Échecs de connexion récents",
		FactorUnusualIP:         "Connexion depuis une adresse IP inhabituelle",
		FactorUnusualTime:       "Connexion à une heure inhabituelle",
		FactorTransactionVolume: "Volume global de transactions",
		FactorLargeTransactions: "Transactions de montant important",
		FactorSessionDuration:   "Durée des sessions de l'utilisateur",
		FactorApiUsage:          "Utilisation de l'API",
		FactorBlockMining:       "Activité de minage de blocs",
	}

	if desc, ok := descriptions[factor]; ok {
		return desc
	}
	return "Facteur de risque inconnu"
}

// GetRiskLevelDescription retourne une description pour un niveau de risque
func GetRiskLevelDescription(level RiskLevel) string {
	descriptions := map[RiskLevel]string{
		RiskLow:      "Faible - Aucune action requise",
		RiskMedium:   "Moyen - Surveillance recommandée",
		RiskHigh:     "Élevé - Attention requise",
		RiskCritical: "Critique - Action immédiate nécessaire",
	}

	if desc, ok := descriptions[level]; ok {
		return desc
	}
	return "Niveau de risque inconnu"
}

// IsIPUnusual version globale pour l'appel externe
func IsIPUnusual(userID, ipAddress string) bool {
	securityMutex.RLock()
	defer securityMutex.RUnlock()

	if securityAssessment == nil {
		return false
	}

	return securityAssessment.IsIPUnusual(userID, ipAddress)
}

// IsLoginTimeUnusual version globale pour l'appel externe
func IsLoginTimeUnusual(userID string) bool {
	securityMutex.RLock()
	defer securityMutex.RUnlock()

	if securityAssessment == nil {
		return false
	}

	return securityAssessment.IsLoginTimeUnusual(userID)
}

// RecordFailedLogin version globale pour l'appel externe
func RecordFailedLogin(userID, ipAddress string) {
	securityMutex.Lock()
	defer securityMutex.Unlock()

	if securityAssessment == nil {
		return
	}

	securityAssessment.RecordFailedLogin(userID, ipAddress)
}

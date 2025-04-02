package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// PerformanceMetrics représente les métriques de performance du système
type PerformanceMetrics struct {
	Timestamp      time.Time `json:"timestamp"`
	CPUUsage       float64   `json:"cpu_usage"`       // Pourcentage d'utilisation CPU
	MemoryUsage    uint64    `json:"memory_usage"`    // Utilisation mémoire en bytes
	GoroutineCount int       `json:"goroutine_count"` // Nombre de goroutines
	RequestCount   int64     `json:"request_count"`   // Nombre de requêtes traitées
	ResponseTime   float64   `json:"response_time"`   // Temps de réponse moyen en ms
	ErrorCount     int64     `json:"error_count"`     // Nombre d'erreurs
	TxPoolSize     int       `json:"tx_pool_size"`    // Taille du pool de transactions
	BlockTime      float64   `json:"block_time"`      // Temps moyen de création de bloc en s
	HashRate       float64   `json:"hash_rate"`       // Taux de hachage estimé en H/s
}

// PerformanceMonitor surveille les performances du système
type PerformanceMonitor struct {
	metrics        []PerformanceMetrics
	requestCount   int64
	errorCount     int64
	responseTimes  []float64
	lastSampleTime time.Time
	outputFile     string
	mutex          sync.RWMutex
	isRunning      bool
	stopChan       chan struct{}
}

// Singleton global
var (
	performanceMonitor *PerformanceMonitor
	perfMutex          sync.Mutex
)

// InitPerformanceMonitor initialise le moniteur de performance
func InitPerformanceMonitor(outputFile string) {
	perfMutex.Lock()
	defer perfMutex.Unlock()

	if performanceMonitor != nil {
		return
	}

	performanceMonitor = &PerformanceMonitor{
		metrics:        make([]PerformanceMetrics, 0),
		responseTimes:  make([]float64, 0),
		lastSampleTime: time.Now(),
		outputFile:     outputFile,
		isRunning:      false,
		stopChan:       make(chan struct{}),
	}

	// Créer le répertoire si nécessaire
	dir := filepath.Dir(outputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		LogToConsole("Erreur lors de la création du répertoire pour les métriques: %v", err)
		return
	}

	// Démarrer le monitoring
	go performanceMonitor.Start()
}

// Start démarre le monitoring périodique
func (pm *PerformanceMonitor) Start() {
	pm.mutex.Lock()
	if pm.isRunning {
		pm.mutex.Unlock()
		return
	}
	pm.isRunning = true
	pm.mutex.Unlock()

	// Échantillonnage toutes les 60 secondes
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pm.collectMetrics(nil)
		case <-pm.stopChan:
			return
		}
	}
}

// Stop arrête le monitoring
func (pm *PerformanceMonitor) Stop() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if !pm.isRunning {
		return
	}

	pm.isRunning = false
	close(pm.stopChan)
	pm.saveMetrics()
}

// collectMetrics collecte les métriques de performance actuelles
func (pm *PerformanceMonitor) collectMetrics(txPoolSize *int) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Calculer l'utilisation CPU (simplifié)
	cpuUsage := 0.0 // Dans une implémentation réelle, utilisez un package comme "github.com/shirou/gopsutil"

	// Calculer le temps de réponse moyen
	avgResponseTime := 0.0
	if len(pm.responseTimes) > 0 {
		sum := 0.0
		for _, t := range pm.responseTimes {
			sum += t
		}
		avgResponseTime = sum / float64(len(pm.responseTimes))
	}

	// Créer la métrique
	metric := PerformanceMetrics{
		Timestamp:      time.Now(),
		CPUUsage:       cpuUsage,
		MemoryUsage:    memStats.Alloc,
		GoroutineCount: runtime.NumGoroutine(),
		RequestCount:   pm.requestCount,
		ResponseTime:   avgResponseTime,
		ErrorCount:     pm.errorCount,
	}

	// Ajouter la taille du pool de transactions si fournie
	if txPoolSize != nil {
		metric.TxPoolSize = *txPoolSize
	}

	// Ajouter aux métriques
	pm.metrics = append(pm.metrics, metric)

	// Réinitialiser les compteurs
	pm.requestCount = 0
	pm.errorCount = 0
	pm.responseTimes = make([]float64, 0)

	// Enregistrer périodiquement
	if len(pm.metrics) >= 60 { // Sauvegarder après 1 heure (60 échantillons)
		pm.saveMetrics()
		pm.metrics = pm.metrics[:0] // Vider le tableau
	}
}

// saveMetrics sauvegarde les métriques dans un fichier
func (pm *PerformanceMonitor) saveMetrics() {
	if len(pm.metrics) == 0 {
		return
	}

	// Ouvrir le fichier en append ou le créer
	file, err := os.OpenFile(pm.outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		LogToConsole("Erreur lors de l'ouverture du fichier de métriques: %v", err)
		return
	}
	defer file.Close()

	// Écrire chaque métrique sur une ligne
	for _, metric := range pm.metrics {
		data, err := json.Marshal(metric)
		if err != nil {
			LogToConsole("Erreur lors de la sérialisation de la métrique: %v", err)
			continue
		}

		if _, err := file.Write(append(data, '\n')); err != nil {
			LogToConsole("Erreur lors de l'écriture dans le fichier de métriques: %v", err)
			return
		}
	}
}

// RecordRequest enregistre une requête
func RecordRequest() {
	if performanceMonitor == nil {
		return
	}

	performanceMonitor.mutex.Lock()
	defer performanceMonitor.mutex.Unlock()

	performanceMonitor.requestCount++
}

// RecordError enregistre une erreur
func RecordError() {
	if performanceMonitor == nil {
		return
	}

	performanceMonitor.mutex.Lock()
	defer performanceMonitor.mutex.Unlock()

	performanceMonitor.errorCount++
}

// RecordResponseTime enregistre un temps de réponse
func RecordResponseTime(duration time.Duration) {
	if performanceMonitor == nil {
		return
	}

	performanceMonitor.mutex.Lock()
	defer performanceMonitor.mutex.Unlock()

	performanceMonitor.responseTimes = append(performanceMonitor.responseTimes, float64(duration.Milliseconds()))
}

// UpdateBlockchainMetrics met à jour les métriques spécifiques à la blockchain
func UpdateBlockchainMetrics(txPoolSize int, blockTime, hashRate float64) {
	if performanceMonitor == nil {
		return
	}

	performanceMonitor.mutex.Lock()
	defer performanceMonitor.mutex.Unlock()

	// Mettre à jour la dernière métrique si elle existe
	if len(performanceMonitor.metrics) > 0 {
		lastIndex := len(performanceMonitor.metrics) - 1
		performanceMonitor.metrics[lastIndex].TxPoolSize = txPoolSize
		performanceMonitor.metrics[lastIndex].BlockTime = blockTime
		performanceMonitor.metrics[lastIndex].HashRate = hashRate
	}
}

// GetCurrentMetrics retourne les métriques actuelles
func GetCurrentMetrics() *PerformanceMetrics {
	if performanceMonitor == nil {
		return nil
	}

	performanceMonitor.mutex.RLock()
	defer performanceMonitor.mutex.RUnlock()

	if len(performanceMonitor.metrics) == 0 {
		return nil
	}

	// Copier la dernière métrique
	lastMetric := performanceMonitor.metrics[len(performanceMonitor.metrics)-1]
	return &lastMetric
}

// LogToConsole écrit un message dans la console
func LogToConsole(format string, args ...interface{}) {
	fmt.Printf("[Performance Monitor] "+format+"\n", args...)
}

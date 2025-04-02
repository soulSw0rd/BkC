package utils

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Metrics représente les métriques d'application pour Prometheus
type Metrics struct {
	// Compteurs généraux
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	TotalTransactions  int64
	PendingTxCount     int64
	BlocksCount        int64
	NodesCount         int64
	ActiveUsers        int64

	// Distributions de temps
	RequestDurations   map[string][]float64 // Histogramme des temps de réponse par endpoint
	BlockTimeDurations []float64            // Temps de génération des blocs

	// Durées moyennes
	AverageRequestTime     float64
	AverageBlockTime       float64
	AverageTransactionTime float64

	// État du système
	SystemStartTime time.Time
	LastBlockTime   time.Time
	CurrentUptime   time.Duration

	// Verrouillage
	mu sync.RWMutex
}

// Instance singleton des métriques
var (
	metricsInstance *Metrics
	metricsOnce     sync.Once
)

// GetMetrics obtient l'instance singleton des métriques
func GetMetrics() *Metrics {
	metricsOnce.Do(func() {
		metricsInstance = &Metrics{
			RequestDurations:   make(map[string][]float64),
			BlockTimeDurations: make([]float64, 0),
			SystemStartTime:    time.Now(),
		}
	})
	return metricsInstance
}

// RecordRequest enregistre une requête
func (m *Metrics) RecordRequest(path string, duration time.Duration, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests++
	if success {
		m.SuccessfulRequests++
	} else {
		m.FailedRequests++
	}

	// Enregistrer la durée par endpoint
	durationMs := float64(duration) / float64(time.Millisecond)
	if _, exists := m.RequestDurations[path]; !exists {
		m.RequestDurations[path] = make([]float64, 0)
	}
	m.RequestDurations[path] = append(m.RequestDurations[path], durationMs)

	// Limiter la taille des tableaux
	if len(m.RequestDurations[path]) > 1000 {
		m.RequestDurations[path] = m.RequestDurations[path][1:]
	}

	// Recalculer le temps moyen
	var totalDuration float64
	var count int
	for _, durations := range m.RequestDurations {
		for _, d := range durations {
			totalDuration += d
			count++
		}
	}
	if count > 0 {
		m.AverageRequestTime = totalDuration / float64(count)
	}

	// Mettre à jour l'uptime
	m.CurrentUptime = time.Since(m.SystemStartTime)
}

// RecordBlock enregistre un nouveau bloc
func (m *Metrics) RecordBlock(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.BlocksCount++
	m.LastBlockTime = time.Now()

	// Enregistrer la durée
	durationSec := float64(duration) / float64(time.Second)
	m.BlockTimeDurations = append(m.BlockTimeDurations, durationSec)

	// Limiter la taille du tableau
	if len(m.BlockTimeDurations) > 100 {
		m.BlockTimeDurations = m.BlockTimeDurations[1:]
	}

	// Recalculer le temps moyen
	var totalDuration float64
	for _, d := range m.BlockTimeDurations {
		totalDuration += d
	}
	if len(m.BlockTimeDurations) > 0 {
		m.AverageBlockTime = totalDuration / float64(len(m.BlockTimeDurations))
	}
}

// RecordTransaction enregistre une transaction
func (m *Metrics) RecordTransaction(pending bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalTransactions++
	if pending {
		m.PendingTxCount++
	}
}

// SetNodesCount définit le nombre de nœuds
func (m *Metrics) SetNodesCount(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.NodesCount = count
}

// SetActiveUsers définit le nombre d'utilisateurs actifs
func (m *Metrics) SetActiveUsers(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveUsers = count
}

// SetPendingTxCount définit le nombre de transactions en attente
func (m *Metrics) SetPendingTxCount(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PendingTxCount = count
}

// PrometheusHandler génère les métriques au format Prometheus
func PrometheusHandler(w http.ResponseWriter, r *http.Request) {
	metrics := GetMetrics()
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()

	// Définir le type de contenu
	w.Header().Set("Content-Type", "text/plain")

	// Métriques de compteur
	w.Write([]byte("# HELP bkc_total_requests Total number of HTTP requests\n"))
	w.Write([]byte("# TYPE bkc_total_requests counter\n"))
	w.Write([]byte("bkc_total_requests " + strconv.FormatInt(metrics.TotalRequests, 10) + "\n"))

	w.Write([]byte("# HELP bkc_successful_requests Total number of successful HTTP requests\n"))
	w.Write([]byte("# TYPE bkc_successful_requests counter\n"))
	w.Write([]byte("bkc_successful_requests " + strconv.FormatInt(metrics.SuccessfulRequests, 10) + "\n"))

	w.Write([]byte("# HELP bkc_failed_requests Total number of failed HTTP requests\n"))
	w.Write([]byte("# TYPE bkc_failed_requests counter\n"))
	w.Write([]byte("bkc_failed_requests " + strconv.FormatInt(metrics.FailedRequests, 10) + "\n"))

	w.Write([]byte("# HELP bkc_total_transactions Total number of transactions\n"))
	w.Write([]byte("# TYPE bkc_total_transactions counter\n"))
	w.Write([]byte("bkc_total_transactions " + strconv.FormatInt(metrics.TotalTransactions, 10) + "\n"))

	w.Write([]byte("# HELP bkc_pending_transactions Current number of pending transactions\n"))
	w.Write([]byte("# TYPE bkc_pending_transactions gauge\n"))
	w.Write([]byte("bkc_pending_transactions " + strconv.FormatInt(metrics.PendingTxCount, 10) + "\n"))

	w.Write([]byte("# HELP bkc_blocks_count Total number of blocks\n"))
	w.Write([]byte("# TYPE bkc_blocks_count counter\n"))
	w.Write([]byte("bkc_blocks_count " + strconv.FormatInt(metrics.BlocksCount, 10) + "\n"))

	w.Write([]byte("# HELP bkc_nodes_count Current number of connected nodes\n"))
	w.Write([]byte("# TYPE bkc_nodes_count gauge\n"))
	w.Write([]byte("bkc_nodes_count " + strconv.FormatInt(metrics.NodesCount, 10) + "\n"))

	w.Write([]byte("# HELP bkc_active_users Current number of active users\n"))
	w.Write([]byte("# TYPE bkc_active_users gauge\n"))
	w.Write([]byte("bkc_active_users " + strconv.FormatInt(metrics.ActiveUsers, 10) + "\n"))

	// Métriques de distribution
	w.Write([]byte("# HELP bkc_average_request_time_ms Average HTTP request processing time in milliseconds\n"))
	w.Write([]byte("# TYPE bkc_average_request_time_ms gauge\n"))
	w.Write([]byte("bkc_average_request_time_ms " + strconv.FormatFloat(metrics.AverageRequestTime, 'f', 2, 64) + "\n"))

	w.Write([]byte("# HELP bkc_average_block_time_seconds Average block generation time in seconds\n"))
	w.Write([]byte("# TYPE bkc_average_block_time_seconds gauge\n"))
	w.Write([]byte("bkc_average_block_time_seconds " + strconv.FormatFloat(metrics.AverageBlockTime, 'f', 2, 64) + "\n"))

	// Métriques par endpoint
	w.Write([]byte("# HELP bkc_endpoint_request_count Request count by endpoint\n"))
	w.Write([]byte("# TYPE bkc_endpoint_request_count counter\n"))
	for endpoint, durations := range metrics.RequestDurations {
		safeEndpoint := `endpoint="` + endpoint + `"`
		w.Write([]byte("bkc_endpoint_request_count{" + safeEndpoint + "} " + strconv.Itoa(len(durations)) + "\n"))
	}

	// Uptime
	w.Write([]byte("# HELP bkc_uptime_seconds System uptime in seconds\n"))
	w.Write([]byte("# TYPE bkc_uptime_seconds counter\n"))
	w.Write([]byte("bkc_uptime_seconds " + strconv.FormatFloat(metrics.CurrentUptime.Seconds(), 'f', 1, 64) + "\n"))
}

// MetricsMiddleware est un middleware HTTP qui enregistre les métriques de requête
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrapper de ResponseWriter pour capturer le code de statut
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Défaut
		}

		// Traiter la requête
		next.ServeHTTP(rw, r)

		// Calculer la durée
		duration := time.Since(start)

		// Enregistrer la requête
		metrics := GetMetrics()
		success := rw.statusCode < 400
		metrics.RecordRequest(r.URL.Path, duration, success)
	})
}

// utils/performance_monitor.go
package utils

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// PerformanceStats contient les statistiques de performance
type PerformanceStats struct {
	CPUUsage     float64
	MemoryUsage  uint64
	NumGoroutine int
	LastUpdate   time.Time
	Uptime       time.Duration
	StartTime    time.Time
}

// PerformanceMonitor surveille les performances du système
type PerformanceMonitor struct {
	stats    PerformanceStats
	interval time.Duration
	quit     chan struct{}
	mu       sync.RWMutex
}

// NewPerformanceMonitor crée un nouveau moniteur de performances
func NewPerformanceMonitor(interval time.Duration) *PerformanceMonitor {
	return &PerformanceMonitor{
		stats: PerformanceStats{
			StartTime: time.Now(),
		},
		interval: interval,
		quit:     make(chan struct{}),
	}
}

// Start démarre la surveillance des performances
func (pm *PerformanceMonitor) Start() {
	go func() {
		ticker := time.NewTicker(pm.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				pm.updateStats()
			case <-pm.quit:
				return
			}
		}
	}()
}

// Stop arrête la surveillance des performances
func (pm *PerformanceMonitor) Stop() {
	close(pm.quit)
}

// updateStats met à jour les statistiques de performance
func (pm *PerformanceMonitor) updateStats() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Mettre à jour le nombre de goroutines
	pm.stats.NumGoroutine = runtime.NumGoroutine()

	// Mettre à jour l'usage mémoire
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	pm.stats.MemoryUsage = memStats.Alloc

	// Mettre à jour l'horodatage
	pm.stats.LastUpdate = time.Now()
	pm.stats.Uptime = time.Since(pm.stats.StartTime)

	// Note: La mesure précise du CPU nécessiterait une bibliothèque externe
	// Ceci est une approximation simplifiée
	pm.stats.CPUUsage = float64(pm.stats.NumGoroutine) / 100.0

	if pm.stats.MemoryUsage > 500*1024*1024 { // 500 MB
		Warning("Usage mémoire élevé: %.2f MB", float64(pm.stats.MemoryUsage)/1024/1024)
	}
}

// GetStats récupère les statistiques de performance actuelles
func (pm *PerformanceMonitor) GetStats() PerformanceStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.stats
}

// String retourne une représentation textuelle des statistiques
func (ps PerformanceStats) String() string {
	return fmt.Sprintf(
		"CPU: %.2f%%, Mémoire: %.2f MB, Goroutines: %d, Uptime: %s",
		ps.CPUUsage*100,
		float64(ps.MemoryUsage)/1024/1024,
		ps.NumGoroutine,
		ps.Uptime.Round(time.Second),
	)
}

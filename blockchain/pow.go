package blockchain

import (
	"runtime"
	"strings"
	"time"
)

// ProofOfWork effectue une preuve de travail (PoW)
func (b *Block) ProofOfWork(difficulty int) time.Duration {
	prefix := strings.Repeat("0", difficulty)
	b.Difficulty = difficulty

	startTime := time.Now()

	// Calculer le hash du bloc et vérifier s'il commence par le nombre requis de zéros
	for {
		b.Hash = b.ComputeHash()
		if strings.HasPrefix(b.Hash, prefix) {
			break
		}
		b.Nonce++

		// Toutes les 100 000 itérations, permettre à d'autres goroutines de s'exécuter
		// Cela évite de surcharger le CPU
		if b.Nonce%100000 == 0 {
			runtime.Gosched()
		}
	}

	duration := time.Since(startTime)
	b.MiningTime = duration
	return duration
}

// ValidateProofOfWork vérifie si le hash du bloc est valide selon la difficulté
func (b *Block) ValidateProofOfWork() bool {
	prefix := strings.Repeat("0", b.Difficulty)
	return strings.HasPrefix(b.Hash, prefix)
}

// EstimateHashRate estime le nombre de hachages par seconde effectués pendant le minage
func (b *Block) EstimateHashRate() float64 {
	if b.MiningTime.Seconds() == 0 {
		return 0
	}

	// Estimation grossière basée sur le nonce et le temps de minage
	// Dans une implémentation réelle, on suivrait précisément le nombre de hachages
	return float64(b.Nonce) / b.MiningTime.Seconds()
}

// CalculateDifficulty calcule la difficulté recommandée basée sur le temps de minage souhaité
// et le temps réel des blocs précédents
func CalculateDifficulty(previousDifficulty int, targetTime, actualTime time.Duration) int {
	// Si le minage était trop rapide, augmenter la difficulté
	if actualTime < targetTime/2 {
		return previousDifficulty + 1
	}

	// Si le minage était trop lent, diminuer la difficulté (mais jamais en dessous de 1)
	if actualTime > targetTime*2 && previousDifficulty > 1 {
		return previousDifficulty - 1
	}

	// Maintenir la difficulté actuelle
	return previousDifficulty
}

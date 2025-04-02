package blockchain

import (
	"math"
	"strings"
	"time"
)

// ProofOfWork effectue une preuve de travail (PoW)
func (b *Block) ProofOfWork(difficulty int) {
	b.Difficulty = difficulty
	prefix := strings.Repeat("0", difficulty)

	startTime := time.Now()
	for !strings.HasPrefix(b.Hash, prefix) {
		b.Nonce++
		b.Hash = b.ComputeHash()
	}

	// Pour le débogage: calculer le temps de minage et le hashrate
	duration := time.Since(startTime)
	hashrate := float64(b.Nonce) / duration.Seconds()

	// Logguez ces informations dans un système de journalisation réel
	_ = hashrate // Utilisez cette variable dans votre système de log
}

// VerifyProofOfWork vérifie que le bloc a été correctement miné
func (b *Block) VerifyProofOfWork() bool {
	prefix := strings.Repeat("0", b.Difficulty)
	return strings.HasPrefix(b.ComputeHash(), prefix)
}

// AdjustDifficulty ajuste la difficulté en fonction du temps de minage cible
func AdjustDifficulty(blockchain []*Block, targetTime time.Duration) int {
	if len(blockchain) < 2 {
		return 4 // Difficulté par défaut
	}

	// Considérer seulement les 10 derniers blocs pour l'ajustement (ou moins s'il y en a moins)
	count := int(math.Min(float64(len(blockchain)), 10))
	lastBlock := blockchain[len(blockchain)-1]
	firstBlock := blockchain[len(blockchain)-count]

	// Calculer le temps moyen de minage
	totalTime := lastBlock.Timestamp.Sub(firstBlock.Timestamp)
	averageTime := totalTime / time.Duration(count-1)

	// Ajuster la difficulté
	if averageTime < targetTime/2 {
		// Le minage est trop rapide, augmenter la difficulté
		return lastBlock.Difficulty + 1
	} else if averageTime > targetTime*2 {
		// Le minage est trop lent, réduire la difficulté (mais jamais en dessous de 1)
		return int(math.Max(1, float64(lastBlock.Difficulty-1)))
	}

	// Maintenir la difficulté actuelle
	return lastBlock.Difficulty
}

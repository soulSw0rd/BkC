package blockchain

import "strings"

// ProofOfWork effectue une preuve de travail (PoW)
func (b *Block) ProofOfWork(difficulty int) {
	prefix := strings.Repeat("0", difficulty)
	for !strings.HasPrefix(b.Hash, prefix) {
		b.Nonce++
		b.Hash = b.ComputeHash()
	}
}

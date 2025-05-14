package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Message représente un message envoyé entre utilisateurs
type Message struct {
	ID          string    `json:"id"`
	Sender      string    `json:"sender"`
	Recipient   string    `json:"recipient"`
	Content     string    `json:"content"`
	ContentHash string    `json:"content_hash"`
	Timestamp   time.Time `json:"timestamp"`
}

// CreateMessage crée un nouveau message
func CreateMessage(sender, recipient, content string) Message {
	contentHash := hashContent(sender, recipient, content, time.Now())
	return Message{
		ID:          generateMessageID(sender, recipient, time.Now()),
		Sender:      sender,
		Recipient:   recipient,
		Content:     content,
		ContentHash: contentHash,
		Timestamp:   time.Now(),
	}
}

// hashContent crée un hash du contenu du message
func hashContent(sender, recipient, content string, timestamp time.Time) string {
	data := fmt.Sprintf("%s%s%s%d", sender, recipient, content, timestamp.UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// generateMessageID génère un ID unique pour le message
func generateMessageID(sender, recipient string, timestamp time.Time) string {
	data := fmt.Sprintf("%s%s%d", sender, recipient, timestamp.UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16] // Utiliser seulement les 16 premiers caractères pour l'ID
}

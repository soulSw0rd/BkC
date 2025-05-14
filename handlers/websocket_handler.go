package handlers

import (
	"BkC/blockchain"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Autoriser toutes les origines pour le dev
	},
}

// ClientMessage représente un message du client au serveur
type ClientMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

// ServerMessage représente un message du serveur au client
type ServerMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Time    time.Time   `json:"time"`
	Success bool        `json:"success"`
}

// WebSocketClient gère une connexion WebSocket avec un client
type WebSocketClient struct {
	conn          *websocket.Conn
	bc            *blockchain.Blockchain
	username      string
	updates       chan blockchain.BlockUpdate
	send          chan ServerMessage
	lastMessageID string
	mutex         sync.Mutex
}

// WebSocketHandler gère les connexions WebSocket
func WebSocketHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Vérifier si l'utilisateur est connecté
		username, ok := getLoggedInUser(r)
		if !ok {
			http.Error(w, "Non autorisé", http.StatusUnauthorized)
			return
		}

		// Mettre à niveau la connexion HTTP vers WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Erreur lors de la mise à niveau de la connexion: %v", err)
			return
		}

		// Créer un client WebSocket
		client := &WebSocketClient{
			conn:     conn,
			bc:       bc,
			username: username,
			updates:  bc.Subscribe(), // S'abonner aux mises à jour de la blockchain
			send:     make(chan ServerMessage, 256),
		}

		// Démarrer les goroutines pour gérer la lecture et l'écriture
		go client.readPump()
		go client.writePump()
		go client.listenForUpdates()

		// Envoyer un message initial
		client.send <- ServerMessage{
			Type:    "connected",
			Data:    "Connecté au système de blockchain en temps réel",
			Time:    time.Now(),
			Success: true,
		}
	}
}

// readPump lit les messages WebSocket du client
func (c *WebSocketClient) readPump() {
	defer func() {
		c.bc.Unsubscribe(c.updates)
		c.conn.Close()
	}()

	// Configurer les paramètres de lecture
	c.conn.SetReadLimit(512) // Limite à 512 octets
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Erreur de lecture WebSocket: %v", err)
			}
			break
		}

		// Décoder le message
		var clientMsg ClientMessage
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			log.Printf("Erreur de décodage JSON: %v", err)
			continue
		}

		// Traiter le message en fonction de son type
		switch clientMsg.Type {
		case "send_message":
			// Traiter l'envoi d'un message
			var messageData struct {
				Recipient string `json:"recipient"`
				Content   string `json:"content"`
			}
			if err := json.Unmarshal([]byte(clientMsg.Data), &messageData); err != nil {
				log.Printf("Erreur lors du décodage des données du message: %v", err)
				continue
			}

			// Créer et ajouter le message à la blockchain de manière asynchrone
			msg := blockchain.CreateMessage(c.username, messageData.Recipient, messageData.Content)
			c.bc.AddMessageBlockAsync(msg, 4)

			// Sauvegarder l'ID du dernier message envoyé pour le tracking
			c.mutex.Lock()
			c.lastMessageID = msg.ID
			c.mutex.Unlock()

			// Confirmer l'envoi
			c.send <- ServerMessage{
				Type:    "message_sent",
				Data:    msg.ID,
				Time:    time.Now(),
				Success: true,
			}
		}
	}
}

// writePump envoie des messages WebSocket au client
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(30 * time.Second) // Ping toutes les 30 secondes
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Le canal a été fermé
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Encoder et envoyer le message
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("Erreur d'encodage JSON: %v", err)
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(data)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// listenForUpdates écoute les mises à jour de la blockchain
func (c *WebSocketClient) listenForUpdates() {
	for update := range c.updates {
		// Ne traiter que les mises à jour de type "new"
		if update.Type == "new" {
			// Vérifier si c'est un message
			var message blockchain.Message
			if err := json.Unmarshal([]byte(update.Block.Data), &message); err == nil && message.ID != "" {
				// Vérifier si le message concerne l'utilisateur actuel
				if message.Sender == c.username || message.Recipient == c.username {
					// Vérifier que ce n'est pas le message que l'utilisateur vient d'envoyer
					c.mutex.Lock()
					isSameMessage := message.ID == c.lastMessageID
					c.mutex.Unlock()

					if !isSameMessage {
						// C'est un nouveau message destiné à cet utilisateur
						c.send <- ServerMessage{
							Type: "new_message",
							Data: map[string]interface{}{
								"id":           message.ID,
								"sender":       message.Sender,
								"recipient":    message.Recipient,
								"content":      message.Content,
								"content_hash": message.ContentHash,
								"timestamp":    message.Timestamp,
							},
							Time:    time.Now(),
							Success: true,
						}
					}
				}
			}

			// Envoyer une mise à jour générale de la blockchain
			c.send <- ServerMessage{
				Type: "blockchain_update",
				Data: map[string]interface{}{
					"block_index": update.Block.Index,
					"block_hash":  update.Block.Hash,
					"timestamp":   update.Block.Timestamp,
				},
				Time:    time.Now(),
				Success: true,
			}
		}
	}
}

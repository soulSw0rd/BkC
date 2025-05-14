package handlers

import (
	"BkC/blockchain"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

// Pour l'affichage dans l'interface utilisateur
type MessageView struct {
	ID            string
	Sender        string
	Recipient     string
	Content       string
	ContentHash   string
	Timestamp     time.Time
	FormattedTime string
}

// Pour l'affichage des conversations
type Conversation struct {
	Username    string
	LastMessage string
	LastTime    time.Time
}

// MessageData est la structure pour les données de message en JSON
type MessageData struct {
	Recipient string `json:"recipient"`
	Content   string `json:"content"`
}

// MessagePageData est la structure pour la page de messages
type MessagePageData struct {
	Username         string
	CurrentRecipient string
	Messages         []MessageView
	Conversations    []Conversation
	BlockCount       int
	LastHash         string
}

// MessagesHandler gère l'affichage de la page des messages
func MessagesHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Vérifier si l'utilisateur est connecté
		username, ok := getLoggedInUser(r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Récupérer le destinataire actuel si présent dans l'URL
		recipient := r.URL.Query().Get("recipient")

		// Récupérer tous les messages de l'utilisateur
		allMessages := bc.GetUserMessages(username)

		// Préparer les données pour la page
		pageData := MessagePageData{
			Username:         username,
			CurrentRecipient: recipient,
			BlockCount:       len(bc.Blocks),
		}

		// Récupérer le dernier hash
		if len(bc.Blocks) > 0 {
			pageData.LastHash = bc.Blocks[len(bc.Blocks)-1].Hash
		}

		// Construire la liste des conversations
		conversations := make(map[string]Conversation)
		for _, msg := range allMessages {
			// Identifier le partenaire de conversation
			var partner string
			if msg.Sender == username {
				partner = msg.Recipient
			} else {
				partner = msg.Sender
			}

			// Mettre à jour ou créer la conversation
			conv, exists := conversations[partner]
			if !exists || msg.Timestamp.After(conv.LastTime) {
				conversations[partner] = Conversation{
					Username:    partner,
					LastMessage: msg.Content,
					LastTime:    msg.Timestamp,
				}
			}
		}

		// Convertir la map en slice pour le template
		for _, conv := range conversations {
			pageData.Conversations = append(pageData.Conversations, conv)
		}

		// Si un destinataire est spécifié, afficher les messages de cette conversation
		if recipient != "" {
			for _, msg := range allMessages {
				// Ne garder que les messages entre utilisateur courant et destinataire
				if (msg.Sender == username && msg.Recipient == recipient) ||
					(msg.Sender == recipient && msg.Recipient == username) {
					pageData.Messages = append(pageData.Messages, MessageView{
						ID:            msg.ID,
						Sender:        msg.Sender,
						Recipient:     msg.Recipient,
						Content:       msg.Content,
						ContentHash:   msg.ContentHash,
						Timestamp:     msg.Timestamp,
						FormattedTime: msg.Timestamp.Format("02/01/2006 15:04"),
					})
				}
			}
		}

		// Afficher la page
		tmpl, err := template.ParseFiles("templates/messages.html")
		if err != nil {
			http.Error(w, "Erreur lors du chargement de la page de messages", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, pageData)
	}
}

// APIMessagesHandler gère l'envoi de nouveaux messages via API
func APIMessagesHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Vérifier si l'utilisateur est connecté
		username, ok := getLoggedInUser(r)
		if !ok {
			http.Error(w, "Utilisateur non connecté", http.StatusUnauthorized)
			return
		}

		if r.Method == "POST" {
			// Décoder le corps de la requête
			var messageData MessageData
			err := json.NewDecoder(r.Body).Decode(&messageData)
			if err != nil {
				http.Error(w, "Format de message invalide", http.StatusBadRequest)
				return
			}

			// Valider les données
			if messageData.Recipient == "" || messageData.Content == "" {
				http.Error(w, "Destinataire ou contenu manquant", http.StatusBadRequest)
				return
			}

			// Créer le message
			message := blockchain.CreateMessage(username, messageData.Recipient, messageData.Content)

			// Ajouter à la blockchain de manière asynchrone
			bc.AddMessageBlockAsync(message, 4) // Difficulté 4 comme standard

			// Répondre avec succès
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `{"status":"success","message":"Message envoyé"}`)
		} else {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	}
}

// getLoggedInUser vérifie si l'utilisateur est connecté et renvoie son nom d'utilisateur
func getLoggedInUser(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", false
	}

	mu.Lock()
	session, exists := sessions[cookie.Value]
	mu.Unlock()

	if !exists || session == nil {
		return "", false
	}

	return session.Username, true
}

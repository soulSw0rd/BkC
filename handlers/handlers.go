package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"BkC/blockchain"
	"BkC/utils"
)

// Variables globales pour les statistiques
var (
	visitorMutex   sync.Mutex
	uniqueVisitors = make(map[string]bool) // Dictionnaire pour suivre les visiteurs uniques
	visitorCount   int
	activeSessions int // Remplace `sessions` pour simplifier
)

// Stats représente les statistiques à afficher
type Stats struct {
	VisitorCount       int
	ActiveSessionCount int
	LastBlock          *blockchain.Block
}

// HomeHandler gère la page d'accueil
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	utils.LogRequest(r)
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Charge le template de la page d'accueil
	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement de la page d'accueil", http.StatusInternalServerError)
		return
	}

	// Exécute le template
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Erreur lors du rendu de la page", http.StatusInternalServerError)
	}
}

// StatsHandler gère l'affichage des statistiques
func StatsHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.LogRequest(r)
		visitorIP := utils.GetVisitorIP(r)

		// Gestion des visiteurs uniques
		visitorMutex.Lock()
		if !uniqueVisitors[visitorIP] {
			uniqueVisitors[visitorIP] = true
			visitorCount++
		}
		visitorMutex.Unlock()

		// Générer les statistiques
		stats := generateStats(bc)

		// Charge le template des statistiques
		tmpl, err := template.ParseFiles("templates/stats.html")
		if err != nil {
			http.Error(w, "Erreur lors du chargement des statistiques", http.StatusInternalServerError)
			return
		}

		// Formater le dernier bloc
		lastBlockHTML := formatBlock(stats.LastBlock)

		// Données à passer au template
		data := struct {
			VisitorCount       int
			ActiveSessionCount int
			LastBlock          string
		}{
			VisitorCount:       stats.VisitorCount,
			ActiveSessionCount: stats.ActiveSessionCount,
			LastBlock:          lastBlockHTML,
		}

		// Exécute le template avec les données
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Erreur lors du rendu des statistiques", http.StatusInternalServerError)
		}
	}
}

// generateStats génère les statistiques actuelles
func generateStats(bc *blockchain.Blockchain) Stats {
	bc.Lock()
	defer bc.Unlock()

	var lastBlock *blockchain.Block
	if len(bc.Blocks) > 0 {
		lastBlock = bc.Blocks[len(bc.Blocks)-1]
	}

	visitorMutex.Lock()
	defer visitorMutex.Unlock()

	return Stats{
		VisitorCount:       visitorCount,
		ActiveSessionCount: activeSessions,
		LastBlock:          lastBlock,
	}
}

// formatBlock formate les données d'un bloc pour le rendu HTML
func formatBlock(block *blockchain.Block) string {
	if block == nil {
		return "<p>Aucun bloc disponible.</p>"
	}

	return fmt.Sprintf(`
<p>Index : %d</p>
<p>Timestamp : %s</p>
<p>Données : %s</p>
<p>Hash : %s</p>
`, block.Index, block.Timestamp, block.Data, block.Hash)
}

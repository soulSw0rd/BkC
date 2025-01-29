package handlers

import (
	"fmt"
	"mon-projet/blockchain"
	"mon-projet/utils"
	"net/http"
	"sync"
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

	w.Header().Set("Content-Type", "text/html")
	html := `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <title>CryptoChain Go</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #1a1a1a;
            color: white;
            margin: 0;
            padding: 0;
        }
        header {
            background-color: #2b2b2b;
            padding: 20px;
            text-align: center;
            border-bottom: 2px solid #5d1f8e;
        }
        h1 {
            font-size: 2.5em;
        }
        main {
            padding: 20px;
            text-align: center;
        }
        nav a {
            color: #5d1f8e;
            font-size: 1.2em;
            text-decoration: none;
            margin: 10px;
        }
    </style>
</head>
<body>
    <header>
        <h1>Bienvenue sur CryptoChain Go</h1>
    </header>
    <main>
        <nav>
            <a href="/blockchain">Voir la Blockchain</a>
            <a href="/stats">Statistiques</a>
        </nav>
    </main>
</body>
</html>`
	fmt.Fprint(w, html)
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

		// Rendu HTML
		w.Header().Set("Content-Type", "text/html")
		html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <title>Statistiques</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #1a1a1a;
            color: white;
            margin: 0;
            padding: 0;
        }
        h1, h2 {
            text-align: center;
            color: #5d1f8e;
        }
        p {
            font-size: 1.2em;
        }
        a {
            color: #5d1f8e;
            text-decoration: none;
        }
    </style>
</head>
<body>
    <h1>Statistiques</h1>
    <p>Visiteurs uniques : %d</p>
    <p>Sessions actives : %d</p>
    <h2>Dernier bloc</h2>
    %s
    <p><a href="/">Retour à l'accueil</a></p>
</body>
</html>
`, stats.VisitorCount, stats.ActiveSessionCount, formatBlock(stats.LastBlock))
		fmt.Fprint(w, html)
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

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Block représente un bloc dans la blockchain
type Block struct {
	Index     int    `json:"index"`
	Timestamp string `json:"timestamp"`
	Data      string `json:"data"`
	PrevHash  string `json:"prev_hash"`
	Hash      string `json:"hash"`
	Nonce     int    `json:"nonce"`
}

// Blockchain représente la chaîne de blocs
type Blockchain struct {
	Blocks []*Block
	mu     sync.Mutex
}

// UserSession pour suivre les sessions utilisateur
type UserSession struct {
	IP        string
	StartTime time.Time
	LastSeen  time.Time
}

// Variables globales
var (
	visitorCount        int
	uniqueVisitors      = make(map[string]bool)
	logFile             *os.File
	sessions            = make(map[string]*UserSession)
	sessionMutex        sync.Mutex
	difficulty          = 4
	sessionHashDuration = 1 * time.Minute
)

// ComputeHash calcule le hash d'un bloc
func (b *Block) ComputeHash() string {
	record := fmt.Sprintf("%d%s%s%s%d", b.Index, b.Timestamp, b.Data, b.PrevHash, b.Nonce)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// ProofOfWork effectue une preuve de travail (PoW)
func (b *Block) ProofOfWork(difficulty int) {
	prefix := strings.Repeat("0", difficulty)
	for !strings.HasPrefix(b.Hash, prefix) {
		b.Nonce++
		b.Hash = b.ComputeHash()
	}
}

// CreateGenesisBlock crée le premier bloc (genesis block)
func CreateGenesisBlock() *Block {
	block := &Block{
		Index:     0,
		Timestamp: time.Now().String(),
		Data:      "Genesis Block",
		PrevHash:  "",
		Nonce:     0,
	}
	block.Hash = block.ComputeHash()
	block.ProofOfWork(difficulty)
	return block
}

// NewBlockchain initialise une nouvelle blockchain
func NewBlockchain() *Blockchain {
	return &Blockchain{
		Blocks: []*Block{CreateGenesisBlock()},
	}
}

// AddBlock ajoute un nouveau bloc à la blockchain
func (bc *Blockchain) AddBlock(data string, difficulty int) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := &Block{
		Index:     len(bc.Blocks),
		Timestamp: time.Now().String(),
		Data:      data,
		PrevHash:  prevBlock.Hash,
		Nonce:     0,
	}
	newBlock.ProofOfWork(difficulty)
	bc.Blocks = append(bc.Blocks, newBlock)
}

// Fonction pour enregistrer les journaux
func logRequest(r *http.Request) {
	if logFile == nil {
		return
	}
	clientIP := getVisitorIP(r)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	method := r.Method
	url := r.URL.Path
	logLine := fmt.Sprintf("%s - %s %s %s\n", timestamp, clientIP, method, url)
	logFile.WriteString(logLine)
}

// Fonction pour obtenir l'adresse IP du visiteur
func getVisitorIP(r *http.Request) string {
	remoteAddr := r.RemoteAddr
	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		remoteAddr = remoteAddr[:idx]
	}
	return remoteAddr
}

// Page d'accueil HTML
func homeHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	html := `<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <title>CryptoChain Go</title>
   <style>
        /* Styles généraux */
        body {
            font-family: Arial, sans-serif;
            background-color: #1a1a1a; /* Fond noir foncé */
            color: white; /* Texte en blanc */
            margin: 0;
            padding: 0;
        }

        header {
            background-color: #2b2b2b; /* Couleur plus claire pour l'en-tête */
            padding: 20px;
            text-align: center;
            border-bottom: 2px solid #5d1f8e; /* Ligne violette foncée */
        }

        header h1 {
            font-size: 2.5em;
            margin: 0;
        }

        main {
            padding: 20px;
        }

        h2 {
            font-size: 1.8em;
            margin-bottom: 20px;
            color: #5d1f8e; /* Violet foncé pour les sous-titres */
        }

        nav {
            display: flex;
            justify-content: space-around;
            margin-top: 20px;
        }

		h2 {
			text-align : center
		}
	
        nav a {
            color: #5d1f8e; /* Violet foncé pour les liens */
            font-size: 1.2em;
            text-decoration: none;
            padding: 10px 20px;
            border-radius: 5px;
            transition: background-color 0.3s ease;
        }

        nav a:hover {
            background-color: #5d1f8e; /* Violet foncé au survol */
            color: white;
        }

        footer {
            background-color: #2b2b2b;
            color: white;
            padding: 10px;
            text-align: center;
            margin-top: 20px;
            border-top: 2px solid #5d1f8e;
        }

        .block-info {
            background-color: #2b2b2b;
            padding: 15px;
            margin-top: 20px;
            border-radius: 5px;
            border: 1px solid #5d1f8e;
        }

        .block-info p {
            font-size: 1.2em;
        }

        /* Responsive */
        @media (max-width: 768px) {
            header h1 {
                font-size: 2em;
            }

            nav {
                flex-direction: column;
                align-items: center;
            }

            nav a {
                margin: 5px 0;
            }
        }
    </style>
</head>
<body>
    <header>
        <h1>Bienvenue sur CryptoChain Go</h1>
    </header>
    <main>
        <h2>Ajoutez un bloc ou consultez la Blockchain</h2>
        <nav>
            <a href="/blockchain">Voir la Blockchain</a>
            <a href="/stats">Statistiques</a>
        </nav>
    </main>
</body>
</html>`
	fmt.Fprint(w, html)
}

// Handler pour afficher la blockchain
func blockchainHandler(bc *Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		clientIP := getVisitorIP(r)

		// Gestion des sessions
		sessionMutex.Lock()
		session, exists := sessions[clientIP]
		now := time.Now()

		if !exists {
			session = &UserSession{
				IP:        clientIP,
				StartTime: now,
				LastSeen:  now,
			}
			sessions[clientIP] = session
			fmt.Println("Nouvelle session créée pour IP:", clientIP) // Debug log
		} else {
			// Vérification de la durée de session pour générer un bloc
			sessionDuration := now.Sub(session.StartTime)
			if sessionDuration >= 5*time.Minute { // Increased duration
				sessionData := fmt.Sprintf("Session from %s started at %v (duration: %v)",
					clientIP, session.StartTime, sessionDuration)

				fmt.Println("Génération de nouveau bloc pour la session:", sessionData) // Debug log
				bc.AddBlock(sessionData, difficulty)
				session.StartTime = now
			}
			session.LastSeen = now
		}
		sessionMutex.Unlock()

		// Logique de gestion des requêtes blockchain
		if r.Method == "POST" {
			var input struct {
				Data string `json:"data"`
			}
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			bc.AddBlock(input.Data, difficulty)
			fmt.Fprintf(w, "Nouveau bloc ajouté !")
		} else if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(bc.Blocks)
		} else {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	}
}

// Handler pour les statistiques
func statsHandler(bc *Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		visitorIP := getVisitorIP(r)
		if !uniqueVisitors[visitorIP] {
			uniqueVisitors[visitorIP] = true
			visitorCount++
		}

		bc.mu.Lock()
		lastBlock := bc.Blocks[len(bc.Blocks)-1]
		bc.mu.Unlock()

		sessionMutex.Lock()
		activeSessionCount := len(sessions)
		sessionMutex.Unlock()

		w.Header().Set("Content-Type", "text/html")
		html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <title>Statistiques</title>
    <style>
        /* Styles généraux */
        body {
            font-family: Arial, sans-serif;
            background-color: #1a1a1a; /* Fond noir foncé */
            color: white; /* Texte en blanc */
            margin: 0;
            padding: 0;
        }

        header {
            background-color: #2b2b2b; /* Couleur plus claire pour l'en-tête */
            padding: 20px;
            text-align: center;
            border-bottom: 2px solid #5d1f8e; /* Ligne violette foncée */
        }

        header h1 {
            font-size: 2.5em;
            margin: 0;
        }

        main {
            padding: 20px;
        }

        h2 {
            font-size: 1.8em;
            margin-bottom: 20px;
            color: #5d1f8e; /* Violet foncé pour les sous-titres */
        }

        nav {
            display: flex;
            justify-content: space-around;
            margin-top: 20px;
        }

        nav a {
            color: #5d1f8e; /* Violet foncé pour les liens */
            font-size: 1.2em;
            text-decoration: none;
            padding: 10px 20px;
            border-radius: 5px;
            transition: background-color 0.3s ease;
        }

        nav a:hover {
            background-color: #5d1f8e; /* Violet foncé au survol */
            color: white;
        }

        footer {
            background-color: #2b2b2b;
            color: white;
            padding: 10px;
            text-align: center;
            margin-top: 20px;
            border-top: 2px solid #5d1f8e;
        }

        .block-info {
            background-color: #2b2b2b;
            padding: 15px;
            margin-top: 20px;
            border-radius: 5px;
            border: 1px solid #5d1f8e;
        }

        .block-info p {
            font-size: 1.2em;
        }

        /* Responsive */
        @media (max-width: 768px) {
            header h1 {
                font-size: 2em;
            }

            nav {
                flex-direction: column;
                align-items: center;
            }

            nav a {
                margin: 5px 0;
            }
        }
    </style>
</head>
<body>
    <h1>Statistiques</h1>
    <p>Visiteurs uniques : %d</p>
    <p>Sessions actives : %d</p>
    <h2>Dernier bloc</h2>
    <p>Index : %d</p>
    <p>Timestamp : %s</p>
    <p>Données : %s</p>
    <p>Hash : %s</p>
    <p><a href="/">Retour à l'accueil</a></p>
</body>
</html>
`, visitorCount, activeSessionCount, lastBlock.Index, lastBlock.Timestamp, lastBlock.Data, lastBlock.Hash)
		fmt.Fprint(w, html)
	}
}

// Nettoyage des sessions
func cleanupSessions() {
	for {
		time.Sleep(10 * time.Minute)
		now := time.Now()
		sessionMutex.Lock()
		for ip, session := range sessions {
			if now.Sub(session.LastSeen) > 30*time.Minute {
				delete(sessions, ip)
			}
		}
		sessionMutex.Unlock()
	}
}

func main() {
	var err error
	logFile, err = os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Erreur lors de l'ouverture du fichier log : %v", err)
	}
	defer logFile.Close()

	bc := NewBlockchain()

	// Démarrage du nettoyage des sessions
	go cleanupSessions()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/blockchain", blockchainHandler(bc))
	http.HandleFunc("/stats", statsHandler(bc))

	fmt.Println("Serveur lancé sur : http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur : %v", err)
	}
}

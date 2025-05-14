package main

import (
	"BkC/blockchain"
	"BkC/handlers"
	"BkC/utils"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

// openBrowser ouvre le navigateur par défaut avec l'URL spécifiée.
func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	default:
		log.Println("⚠️ Système non supporté pour l'ouverture automatique du navigateur")
		return
	}

	args = append(args, url)
	exec.Command(cmd, args...).Start()
}

func main() {
	var err error
	utils.LogFile, err = os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Erreur lors de l'ouverture du fichier log : %v", err)
	}
	defer utils.LogFile.Close()

	// Initialisation de la blockchain.
	bc := blockchain.NewBlockchain()

	// Initialiser la référence globale
	handlers.InitGlobalBC(bc)

	// Route par défaut : affiche la page d'accueil (acceuil.html)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/acceuil.html")
		if err != nil {
			http.Error(w, "Erreur lors du chargement de la page d'accueil", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	// Routes pour l'authentification
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/login-submit", handlers.LoginSubmitHandler)
	http.HandleFunc("/signin", handlers.SigninHandler)
	http.HandleFunc("/signin-submit", handlers.SigninSubmitHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)

	// Route d'accueil après connexion
	http.HandleFunc("/home", handlers.HomeHandler)

	// Route pour la messagerie
	http.HandleFunc("/messages", handlers.MessagesHandler(bc))
	http.HandleFunc("/api/messages", handlers.APIMessagesHandler(bc))

	// Route pour le minage de blocs
	http.HandleFunc("/mine-block", handlers.MineBlockHandler(bc))

	// Route WebSocket pour les mises à jour en temps réel
	http.HandleFunc("/ws", handlers.WebSocketHandler(bc))

	// Route pour la blockchain
	http.HandleFunc("/blockchain", handlers.BlockchainHandler(bc))

	// Route pour les statistiques des mineurs
	http.HandleFunc("/miners-stats", handlers.MinersStatsHandler(bc))

	// Route pour les statistiques
	http.HandleFunc("/stats", handlers.StatsHandler(bc))

	// Servir les fichiers statiques
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Ouvre le navigateur automatiquement
	go func() {
		log.Println("🌍 Ouverture du navigateur...")
		openBrowser("http://localhost:8080")
	}()

	fmt.Println("🚀 Serveur lancé sur : http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("❌ Erreur lors du démarrage du serveur : %v", err)
	}
}

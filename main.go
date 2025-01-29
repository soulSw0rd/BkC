package main

import (
	"BkC/blockchain"
	"BkC/handlers"
	"BkC/utils"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

// Fonction pour ouvrir automatiquement le navigateur
func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin": // MacOS
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

	bc := blockchain.NewBlockchain()

	// Routes publiques
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/login-submit", handlers.LoginSubmitHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)

	// Routes protégées
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/blockchain", handlers.StatsHandler(bc)) // Protéger l'accès

	// Ouvrir le navigateur après le démarrage du serveur
	go func() {
		log.Println("🌍 Ouverture du navigateur...")
		openBrowser("http://localhost:8080")
	}()

	fmt.Println("🚀 Serveur lancé sur : http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("❌ Erreur lors du démarrage du serveur : %v", err)
	}
}

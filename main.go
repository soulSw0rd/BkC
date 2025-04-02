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

	// Créer les répertoires nécessaires s'ils n'existent pas
	os.MkdirAll("logs", 0755)
	os.MkdirAll("wallets", 0755)
	os.MkdirAll("data", 0755)

	utils.LogFile, err = os.OpenFile("logs/server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Erreur lors de l'ouverture du fichier log : %v", err)
	}
	defer utils.LogFile.Close()

	// Initialisation de la blockchain.
	bc := blockchain.NewBlockchain()

	// Initialisation des utilisateurs par défaut
	handlers.InitSampleUsers()

	// Fichiers statiques
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Route par défaut : affiche la page d'accueil (acceuil.html)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		tmpl, err := template.ParseFiles("templates/acceuil.html")
		if err != nil {
			http.Error(w, "Erreur lors du chargement de la page d'accueil", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	// Routes d'authentification
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/login-submit", handlers.LoginSubmitHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/register-submit", handlers.RegisterSubmitHandler)

	// Routes sécurisées
	http.HandleFunc("/home", handlers.HomeHandler)
	http.HandleFunc("/profile", handlers.ProfileHandler)
	http.HandleFunc("/admin", handlers.AdminHandler(bc))

	// Routes de la blockchain
	http.HandleFunc("/blockchain", handlers.BlockchainHandler(bc))
	http.HandleFunc("/transactions", handlers.TransactionHandler(bc))
	http.HandleFunc("/wallets", handlers.WalletHandler("wallets"))
	http.HandleFunc("/stats", handlers.StatsHandler(bc))

	// Ouvre le navigateur automatiquement.
	go func() {
		log.Println("🌍 Ouverture du navigateur...")
		openBrowser("http://localhost:8080")
	}()

	fmt.Println("🚀 Serveur lancé sur : http://localhost:8080")
	fmt.Println("👤 Utilisateurs par défaut :")
	fmt.Println("   - Admin: admin/admin")
	fmt.Println("   - User:  user/user")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("❌ Erreur lors du démarrage du serveur : %v", err)
	}
}

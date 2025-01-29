package main

import (
	"BkC/blockchain" // Package blockchain personnalisé
	"BkC/handlers"   // Package pour les gestionnaires de requêtes HTTP
	"BkC/utils"      // Package utilitaire pour les fonctions communes
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Ouverture ou création du fichier de log
	var err error
	utils.LogFile, err = os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Erreur lors de l'ouverture du fichier log : %v", err)
	}
	defer utils.LogFile.Close()

	// Initialisation de la blockchain
	bc := blockchain.NewBlockchain()

	// Définition des routes
	http.HandleFunc("/", handlers.HomeHandler)                     // Page d'accueil
	http.HandleFunc("/blockchain", handlers.BlockchainHandler(bc)) // Gestion de la blockchain
	http.HandleFunc("/stats", handlers.StatsHandler(bc))           // Statistiques

	// Démarrage du serveur HTTP
	fmt.Println("Serveur lancé sur : http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur : %v", err)
	}
}

package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Configuration contient tous les paramètres de configuration de l'application
type Configuration struct {
	// Configuration serveur
	ServerPort int    `json:"server_port"`
	LogsDir    string `json:"logs_dir"`
	DataDir    string `json:"data_dir"`

	// Configuration de la blockchain
	BlockchainFile  string        `json:"blockchain_file"`
	BlockDifficulty int           `json:"block_difficulty"`
	MiningReward    float64       `json:"mining_reward"`
	MaxBlockSize    int           `json:"max_block_size"`
	BlockTimeTarget time.Duration `json:"block_time_target"`

	// Configuration de sécurité
	SessionTimeout  time.Duration `json:"session_timeout"`
	CsrfKey         string        `json:"csrf_key"`
	AutoOpenBrowser bool          `json:"auto_open_browser"`

	// Configuration de développement
	DevMode  bool `json:"dev_mode"`
	DebugLog bool `json:"debug_log"`
}

// Config est l'instance globale de configuration
var Config Configuration

// DefaultConfig retourne une configuration par défaut
func DefaultConfig() Configuration {
	return Configuration{
		// Configuration serveur
		ServerPort: 8080,
		LogsDir:    "logs",
		DataDir:    "data",

		// Configuration de la blockchain
		BlockchainFile:  "data/blockchain.json",
		BlockDifficulty: 4,
		MiningReward:    10.0,
		MaxBlockSize:    10,
		BlockTimeTarget: 30 * time.Second,

		// Configuration de sécurité
		SessionTimeout:  30 * time.Minute,
		CsrfKey:         "32-byte-long-auth-key-change-in-prod", // À changer en production
		AutoOpenBrowser: true,

		// Configuration de développement
		DevMode:  false,
		DebugLog: true,
	}
}

// LoadConfig charge la configuration depuis un fichier
func LoadConfig(configFile string) error {
	// Utiliser la configuration par défaut
	Config = DefaultConfig()

	// Si le fichier n'existe pas, créer un fichier avec la configuration par défaut
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		saveDefaultConfig(configFile)
		return nil
	}

	// Lire le fichier
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	// Désérialiser
	if err := json.Unmarshal(data, &Config); err != nil {
		return err
	}

	return nil
}

// saveDefaultConfig sauvegarde la configuration par défaut dans un fichier
func saveDefaultConfig(configFile string) {
	// Créer le répertoire parent si nécessaire
	dir := filepath.Dir(configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Erreur lors de la création du répertoire de configuration: %v", err)
		return
	}

	// Sérialiser la configuration par défaut
	data, err := json.MarshalIndent(DefaultConfig(), "", "  ")
	if err != nil {
		log.Printf("Erreur lors de la sérialisation de la configuration: %v", err)
		return
	}

	// Écrire dans le fichier
	if err := ioutil.WriteFile(configFile, data, 0644); err != nil {
		log.Printf("Erreur lors de l'écriture du fichier de configuration: %v", err)
		return
	}
}

// Initialize initialise la configuration
func InitializeConfig() {
	configFile := "config.json"

	// Essayer de charger la configuration
	if err := LoadConfig(configFile); err != nil {
		log.Printf("Erreur lors du chargement de la configuration: %v", err)
		log.Println("Utilisation de la configuration par défaut")
		Config = DefaultConfig()
	}

	// Créer les répertoires nécessaires
	if err := os.MkdirAll(Config.LogsDir, 0755); err != nil {
		log.Printf("Erreur lors de la création du répertoire des logs: %v", err)
	}

	if err := os.MkdirAll(Config.DataDir, 0755); err != nil {
		log.Printf("Erreur lors de la création du répertoire des données: %v", err)
	}
}

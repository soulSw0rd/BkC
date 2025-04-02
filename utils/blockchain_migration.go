package utils

import (
	"BkC/blockchain"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MigrationVersion représente la version d'une migration
type MigrationVersion struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

// String retourne la version sous forme de chaîne
func (v MigrationVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// IsGreaterThan vérifie si cette version est supérieure à une autre
func (v MigrationVersion) IsGreaterThan(other MigrationVersion) bool {
	if v.Major > other.Major {
		return true
	}
	if v.Major == other.Major && v.Minor > other.Minor {
		return true
	}
	if v.Major == other.Major && v.Minor == other.Minor && v.Patch > other.Patch {
		return true
	}
	return false
}

// BlockchainMigration représente une migration de la blockchain
type BlockchainMigration struct {
	FromVersion MigrationVersion          `json:"from_version"`
	ToVersion   MigrationVersion          `json:"to_version"`
	Description string                    `json:"description"`
	Timestamp   time.Time                 `json:"timestamp"`
	Migrate     func(*blockchain.Blockchain) error `json:"-"`
}

// MigrationManager gère les migrations de la blockchain
type MigrationManager struct {
	Migrations        []BlockchainMigration `json:"migrations"`
	CurrentVersion    MigrationVersion      `json:"current_version"`
	MigrationHistory  []BlockchainMigration `json:"migration_history"`
	MigrationFilePath string                `json:"-"`
}

// NewMigrationManager crée un nouveau gestionnaire de migrations
func NewMigrationManager(migrationFilePath string) (*MigrationManager, error) {
	manager := &MigrationManager{
		Migrations:        []BlockchainMigration{},
		CurrentVersion:    MigrationVersion{1, 0, 0}, // Version par défaut
		MigrationHistory:  []BlockchainMigration{},
		MigrationFilePath: migrationFilePath,
	}

	// Créer le répertoire si nécessaire
	dir := filepath.Dir(migrationFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("impossible de créer le répertoire des migrations: %w", err)
	}

	// Essayer de charger l'historique des migrations
	if _, err := os.Stat(migrationFilePath); err == nil {
		// Le fichier existe, essayer de le charger
		data, err := os.ReadFile(migrationFilePath)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture du fichier de migrations: %w", err)
		}

		var migrationData struct {
			CurrentVersion   MigrationVersion      `json:"current_version"`
			MigrationHistory []BlockchainMigration `json:"migration_history"`
		}

		if err := json.Unmarshal(data, &migrationData); err != nil {
			// Nouvelle installation ou format incompatible
			Info("Aucun historique de migration trouvé ou format incompatible. Utilisation de la version par défaut 1.0.0")
		} else {
			manager.CurrentVersion = migrationData.CurrentVersion
			manager.MigrationHistory = migrationData.MigrationHistory
			Info("Version actuelle de la blockchain: %s", manager.CurrentVersion.String())
		}
	} else {
		Info("Aucun historique de migration trouvé. Utilisation de la version par défaut 1.0.0")
	}

	// Ajouter les migrations disponibles
	manager.RegisterMigrations()

	return manager, nil
}

// RegisterMigrations enregistre toutes les migrations disponibles
func (mm *MigrationManager) RegisterMigrations() {
	// Exemple de migration de 1.0.0 à 1.1.0
	mm.Migrations = append(mm.Migrations, BlockchainMigration{
		FromVersion: MigrationVersion{1, 0, 0},
		ToVersion:   MigrationVersion{1, 1, 0},
		Description: "Ajout du champ MerkleRoot aux blocs",
		Timestamp:   time.Now(),
		Migrate:     migrateToV1_1_0,
	})

	// Exemple de migration de 1.1.0 à 1.2.0
	mm.Migrations = append(mm.Migrations, BlockchainMigration{
		FromVersion: MigrationVersion{1, 1, 0},
		ToVersion:   MigrationVersion{1, 2, 0},
		Description: "Ajout du support des contrats intelligents",
		Timestamp:   time.Now(),
		Migrate:     migrateToV1_2_0,
	})

	// Ajoutez d'autres migrations au besoin
}

// MigrateBlockchain effectue toutes les migrations nécessaires
func (mm *MigrationManager) MigrateBlockchain(bc *blockchain.Blockchain) error {
	// Vérifier s'il y a des migrations à faire
	var migrationsToApply []BlockchainMigration
	for _, migration := range mm.Migrations {
		if migration.FromVersion.String()
		// MigrateBlockchain effectue toutes les migrations nécessaires
func (mm *MigrationManager) MigrateBlockchain(bc *blockchain.Blockchain) error {
	// Vérifier s'il y a des migrations à faire
	var migrationsToApply []BlockchainMigration
	for _, migration := range mm.Migrations {
		if migration.FromVersion.String() == mm.CurrentVersion.String() && 
		   migration.ToVersion.IsGreaterThan(mm.CurrentVersion) {
			migrationsToApply = append(migrationsToApply, migration)
		}
	}

	if len(migrationsToApply) == 0 {
		Info("La blockchain est à jour (version %s). Aucune migration nécessaire.", mm.CurrentVersion.String())
		return nil
	}

	// Appliquer les migrations dans l'ordre
	for _, migration := range migrationsToApply {
		Info("Migration de la blockchain de la version %s vers %s: %s", 
			migration.FromVersion.String(), 
			migration.ToVersion.String(), 
			migration.Description)

		// Exécuter la migration
		if err := migration.Migrate(bc); err != nil {
			Error("Échec de la migration %s -> %s: %v", 
				migration.FromVersion.String(), 
				migration.ToVersion.String(), 
				err)
			return fmt.Errorf("échec de la migration %s -> %s: %w", 
				migration.FromVersion.String(), 
				migration.ToVersion.String(), 
				err)
		}

		// Mettre à jour la version courante
		mm.CurrentVersion = migration.ToVersion

		// Ajouter à l'historique
		mm.MigrationHistory = append(mm.MigrationHistory, migration)

		// Sauvegarder l'état des migrations
		if err := mm.SaveMigrationState(); err != nil {
			Warning("Impossible de sauvegarder l'état des migrations: %v", err)
		}

		Info("Migration %s -> %s appliquée avec succès", 
			migration.FromVersion.String(), 
			migration.ToVersion.String())
	}

	return nil
}

// SaveMigrationState sauvegarde l'état des migrations
func (mm *MigrationManager) SaveMigrationState() error {
	// Préparer les données à sauvegarder
	data := struct {
		CurrentVersion   MigrationVersion      `json:"current_version"`
		MigrationHistory []BlockchainMigration `json:"migration_history"`
	}{
		CurrentVersion:   mm.CurrentVersion,
		MigrationHistory: mm.MigrationHistory,
	}

	// Sérialiser les données
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur lors de la sérialisation des données de migration: %w", err)
	}

	// Écrire dans le fichier
	if err := os.WriteFile(mm.MigrationFilePath, jsonData, 0644); err != nil {
		return fmt.Errorf("erreur lors de l'écriture du fichier de migrations: %w", err)
	}

	return nil
}

// GetAvailableMigrations retourne les migrations disponibles
func (mm *MigrationManager) GetAvailableMigrations() []BlockchainMigration {
	return mm.Migrations
}

// GetCurrentVersion retourne la version actuelle
func (mm *MigrationManager) GetCurrentVersion() MigrationVersion {
	return mm.CurrentVersion
}

// GetMigrationHistory retourne l'historique des migrations
func (mm *MigrationManager) GetMigrationHistory() []BlockchainMigration {
	return mm.MigrationHistory
}

// Implémentations des migrations

// migrateToV1_1_0 migre la blockchain vers la version 1.1.0
func migrateToV1_1_0(bc *blockchain.Blockchain) error {
	Info("Mise à jour des blocs pour ajouter la racine de Merkle...")
	
	// Parcourir tous les blocs et recalculer la racine de Merkle
	for i := range bc.Blocks {
		block := &bc.Blocks[i]
		// Recalculer la racine de Merkle
		merkleRoot := block.CalculateMerkleRoot()
		// Mettre à jour le bloc
		block.MerkleRoot = merkleRoot
		// Recalculer le hash du bloc
		block.Hash = block.ComputeHash()
	}

	return nil
}

// migrateToV1_2_0 migre la blockchain vers la version 1.2.0
func migrateToV1_2_0(bc *blockchain.Blockchain) error {
	Info("Ajout du support des contrats intelligents...")
	
	// Initialiser le module de contrats intelligents s'il n'existe pas déjà
	if bc.SmartContracts == nil {
		bc.SmartContracts = make(map[string]*blockchain.SmartContract)
	}

	// Ajouter les contrats systèmes si nécessaire
	// Par exemple, un contrat de création de token
	systemContract := blockchain.NewSmartContract(
		blockchain.ContractTransfer,
		"system", 
		[]string{"system"}, 
		1, 
		0.0, 
		0.0, 
		"system",
		"Contrat système initial", 
		365*24*time.Hour, // Expire dans 1 an
		make(map[string]string),
	)
	bc.SmartContracts[systemContract.ID] = systemContract

	return nil
}
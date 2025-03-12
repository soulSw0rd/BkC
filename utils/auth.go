// utils/auth.go
package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/crypto/argon2"
)

// Structure pour le stockage des utilisateurs
var (
	// Map des utilisateurs, protégée par un mutex
	Users = struct {
		sync.RWMutex
		data map[string]string // username -> hashed password
	}{data: make(map[string]string)}
)

// PasswordConfig contient les paramètres pour le hachage des mots de passe
type PasswordConfig struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
}

// DefaultPasswordConfig contient les paramètres par défaut pour le hachage
var DefaultPasswordConfig = &PasswordConfig{
	Time:    1,
	Memory:  64 * 1024,
	Threads: 4,
	KeyLen:  32,
}

// InitDefaultUsers initialise les utilisateurs par défaut
func InitDefaultUsers() {
	// Ajouter l'utilisateur admin par défaut (admin/admin)
	hashedPassword, err := HashPassword("admin")
	if err != nil {
		// En cas d'erreur, utiliser un mot de passe pré-haché (moins sécurisé mais fonctionnel)
		hashedPassword = "$argon2id$v=19$m=65536,t=1,p=4$bNVwSVMpq3sDUxH4TaKbxw$oPwjSXaXfy1aYHUeX4SMR8CWYoPU9BaQlCWGjDt0xvQ"
	}

	Users.Lock()
	Users.data["admin"] = hashedPassword
	Users.Unlock()
}

// HashPassword génère un hash sécurisé du mot de passe
func HashPassword(password string) (string, error) {
	c := DefaultPasswordConfig

	// Générer un sel aléatoire
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Dériver la clé du mot de passe avec Argon2id
	hash := argon2.IDKey([]byte(password), salt, c.Time, c.Memory, c.Threads, c.KeyLen)

	// Encoder en base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=1,p=4$<sel>$<hash>
	encodedHash := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		c.Memory, c.Time, c.Threads, b64Salt, b64Hash)

	return encodedHash, nil
}

// VerifyPassword vérifie si un mot de passe correspond au hash stocké
func VerifyPassword(password, encodedHash string) (bool, error) {
	// Extraire les paramètres du hash
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return false, errors.New("format de hash invalide")
	}

	var version int
	if _, err := fmt.Sscanf(vals[2], "v=%d", &version); err != nil {
		return false, err
	}
	if version != 19 {
		return false, errors.New("version non supportée")
	}

	var c PasswordConfig
	if _, err := fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d",
		&c.Memory, &c.Time, &c.Threads); err != nil {
		return false, err
	}

	// Décoder le sel et le hash
	salt, err := base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return false, err
	}

	c.KeyLen = uint32(len(decodedHash))

	// Calculer le hash du mot de passe avec les mêmes paramètres
	calculatedHash := argon2.IDKey([]byte(password), salt, c.Time, c.Memory, c.Threads, c.KeyLen)

	// Comparaison en temps constant pour éviter les attaques par timing
	return subtle.ConstantTimeCompare(calculatedHash, decodedHash) == 1, nil
}

// AuthenticateUser vérifie les identifiants de l'utilisateur
func AuthenticateUser(username, password string) bool {
	// Récupérer le hash du mot de passe stocké
	Users.RLock()
	storedHash, exists := Users.data[username]
	Users.RUnlock()

	if !exists {
		return false
	}

	// Vérifier le mot de passe
	match, err := VerifyPassword(password, storedHash)
	if err != nil {
		return false
	}

	return match
}

// AddUser ajoute un nouvel utilisateur
func AddUser(username, password string) error {
	// Vérifier si l'utilisateur existe déjà
	Users.RLock()
	_, exists := Users.data[username]
	Users.RUnlock()

	if exists {
		return errors.New("l'utilisateur existe déjà")
	}

	// Hasher le mot de passe
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}

	// Ajouter l'utilisateur
	Users.Lock()
	Users.data[username] = hashedPassword
	Users.Unlock()

	return nil
}

// RemoveUser supprime un utilisateur
func RemoveUser(username string) {
	Users.Lock()
	delete(Users.data, username)
	Users.Unlock()
}

// ChangePassword modifie le mot de passe d'un utilisateur
func ChangePassword(username, newPassword string) error {
	// Vérifier si l'utilisateur existe
	Users.RLock()
	_, exists := Users.data[username]
	Users.RUnlock()

	if !exists {
		return errors.New("utilisateur non trouvé")
	}

	// Hasher le nouveau mot de passe
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Mettre à jour le mot de passe
	Users.Lock()
	Users.data[username] = hashedPassword
	Users.Unlock()

	return nil
}

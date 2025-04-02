package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Paramètres recommandés pour Argon2id
const (
	saltLength  = 16
	keyLength   = 32
	iterations  = 3
	memory      = 64 * 1024
	parallelism = 4
)

// HashPassword génère un hash sécurisé à partir d'un mot de passe en utilisant Argon2id
func HashPassword(password string) (string, error) {
	// Générer un sel aléatoire
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Hacher le mot de passe avec Argon2id
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		uint8(parallelism), // Converti en uint8
		keyLength,
	)

	// Encoder le résultat au format base64
	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)

	// Retourner le hash au format "argon2id$iterations$memory$parallelism$saltB64$hashB64"
	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, iterations, parallelism, saltB64, hashB64,
	)

	return encodedHash, nil
}

// VerifyPassword vérifie un mot de passe par rapport à son hash
func VerifyPassword(password, encodedHash string) (bool, error) {
	// Extraire les paramètres et les valeurs du hash encodé
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("format de hash invalide")
	}

	// Vérifier que le format est bien argon2id
	if parts[1] != "argon2id" {
		return false, errors.New("algorithme non supporté")
	}

	// Extraire les paramètres
	var version int
	var memory, iterations, parallelism uint32
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}

	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, err
	}

	// Décoder le sel et le hash
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	// Calculer le hash avec le mot de passe fourni
	newHash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		uint8(parallelism), // Converti en uint8
		uint32(len(hash)),
	)

	// Comparer les deux hash de manière sécurisée (temps constant)
	return subtle.ConstantTimeCompare(hash, newHash) == 1, nil
}

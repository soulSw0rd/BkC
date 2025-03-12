# BkC - CryptoChain Go

![Version](https://img.shields.io/badge/version-1.2.0-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)
![License](https://img.shields.io/badge/license-GPL%20v3-green.svg)

CryptoChain Go est une implémentation moderne de blockchain en Go, combinée avec une interface utilisateur web élégante permettant de visualiser et interagir avec la blockchain. Cette application offre une plateforme complète pour créer des blocs, effectuer des transactions et suivre les activités dans un environnement sécurisé.

![CryptoChain Screenshot](https://via.placeholder.com/800x400?text=CryptoChain+Go+Screenshot)

## 🚀 Fonctionnalités

- **Blockchain complète** avec preuve de travail (PoW) et ajustement automatique de difficulté
- **Système de transactions** permettant d'échanger des tokens entre utilisateurs
- **Explorateur de blockchain** pour visualiser tous les blocs et transactions
- **Dashboard interactif** avec statistiques en temps réel
- **Persistance des données** avec sauvegarde et chargement automatique de la chaîne
- **Interface utilisateur moderne** avec design responsive et thèmes clair/sombre
- **Authentification sécurisée** utilisant Argon2 pour le hachage des mots de passe
- **API RESTful** pour l'intégration avec d'autres applications

## 📋 Prérequis

- [Go](https://golang.org/dl/) version 1.21 ou supérieure
- Navigateur web moderne (Chrome, Firefox, Edge, Safari)
- Connexion Internet pour charger les bibliothèques CSS/JS via CDN
- Minimum 100 Mo d'espace disque libre pour la blockchain et les logs

### Dépendances Go requises

```
github.com/gorilla/mux v1.8.1
github.com/gorilla/csrf v1.7.2
golang.org/x/crypto v0.36.0+
```

## 🔧 Installation

### Option 1 : Installation depuis le code source

1. Clonez le dépôt:

```bash
git clone https://github.com/soulSw0rd/BkC.git
cd BkC
```

2. Installez les dépendances:

```bash
go mod tidy
```

3. Compilez l'application:

```bash
go build
```

### Option 2 : Installation avec Go install

```bash
go install github.com/soulSw0rd/BkC@latest
```

## 🏃‍♂️ Démarrage

1. Créez les dossiers nécessaires:

```bash
mkdir -p data logs
```

2. Lancez l'application:

```bash
./BkC
```

Ou directement sans compilation préalable:

```bash
go run main.go
```

3. Accédez à l'application dans votre navigateur:

```
http://localhost:8080
```

4. Connectez-vous avec les identifiants par défaut:
   - Nom d'utilisateur: `admin`
   - Mot de passe: `admin`

## 📊 Utilisation

### Exploration de la blockchain

- Accédez à la page **Blockchain** pour voir tous les blocs minés
- Consultez les détails de chaque bloc, y compris ses transactions et hash
- Utilisez la barre de recherche pour trouver des blocs ou transactions spécifiques

### Gestion des transactions

- Créez de nouvelles transactions en spécifiant un destinataire et un montant
- Visualisez l'historique de vos transactions, tant confirmées qu'en attente
- Recevez des notifications lorsque vos transactions sont confirmées

### Minage

- Minez de nouveaux blocs et recevez des récompenses en tokens
- Observez l'ajustement automatique de la difficulté en fonction du temps de minage
- Suivez vos récompenses de minage dans votre historique de transactions

### Tableau de bord

- Consultez les statistiques en temps réel de la blockchain
- Visualisez les graphiques d'activité, de transactions et de difficulté
- Surveillez votre solde et vos transactions récentes

## 🛠️ Configuration avancée

Un fichier de configuration `config.json` peut être créé à la racine du projet pour personnaliser divers paramètres:

```json
{
  "server_port": 8080,
  "logs_dir": "logs",
  "data_dir": "data",
  "blockchain_file": "data/blockchain.json",
  "block_difficulty": 4,
  "mining_reward": 10.0,
  "dev_mode": false
}
```

## 🔄 API

L'application expose une API RESTful pour l'intégration avec d'autres systèmes:

- `GET /api/blockchain` - Récupère la blockchain complète
- `GET /api/stats` - Obtient les statistiques de la blockchain
- `POST /api/mining` - Mine un nouveau bloc
- `GET /api/wallet` - Récupère les informations de portefeuille
- `POST /api/transactions` - Crée une nouvelle transaction

## 👩‍💻 Développement

Pour les développeurs souhaitant contribuer au projet:

1. Fork le dépôt
2. Créez une branche pour votre fonctionnalité (`git checkout -b feature/amazing-feature`)
3. Commit vos changements (`git commit -m 'Add some amazing feature'`)
4. Push vers la branche (`git push origin feature/amazing-feature`)
5. Ouvrez une Pull Request

## 🔍 Dépannage

### Problèmes courants

- **Erreur de port déjà utilisé**: Modifiez le port dans le fichier `config.json`
- **Erreurs de dépendances**: Exécutez `go mod tidy` pour mettre à jour les dépendances
- **Problèmes de performance**: Ajustez la difficulté de minage dans `config.json`

### Journaux

Les journaux sont stockés dans le répertoire `logs` et suivent le format `server_YYYY-MM-DD.log`.

## 📜 Licence

Ce projet est sous licence [GNU General Public License v3.0](LICENSE) - voir le fichier LICENSE pour plus de détails.

## 🙏 Remerciements

- L'équipe Go pour leur langage excellent
- La communauté blockchain pour leur inspiration et documentation
- Tous les contributeurs qui ont aidé à améliorer ce projet

---

Développé avec ❤️ par [SoulSw0rd](https://github.com/soulSw0rd)
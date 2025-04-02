<<<<<<< HEAD
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

Hébergé sur le repository de [SoulSw0rd](https://github.com/soulSw0rd)
Dévellopé par l'équipe de BKC
=======
# CryptoChain Go

CryptoChain Go est une implémentation simple de blockchain en Go, combinée avec un serveur HTTP pour afficher des informations sur la blockchain et ses statistiques d'utilisation. L'application permet de créer des blocs, de les ajouter à la chaîne et de suivre les sessions des utilisateurs.

## Fonctionnalités

- **Blockchain** : Crée et gère une blockchain avec un mécanisme de preuve de travail (Proof of Work).
- **Sessions utilisateur** : Suivi des sessions avec les visiteurs et génération de nouveaux blocs à intervalles réguliers.
- **Statistiques** : Affiche des statistiques sur le nombre de visiteurs uniques, les sessions actives, et les informations sur le dernier bloc.
- **Serveur HTTP** : Fournit une interface web simple pour interagir avec la blockchain et consulter les statistiques.

## Installation

### Prérequis

- [Go](https://golang.org/dl/) (version 1.16 ou supérieure)
- Un éditeur de texte ou un IDE compatible avec Go

### Étapes

1. Clonez ce dépôt :

   ```bash
   git clone https://github.com/votre-utilisateur/cryptochain-go.git
   cd cryptochain-go
   ```

2. Compilez et lancez le serveur Go :

   ```bash
   go run main.go
   ```

3. Le serveur sera accessible à l'adresse suivante : `http://localhost:8080`.

## Fonctionnement

### Blockchain

L'application implémente une blockchain simple avec un mécanisme de preuve de travail (PoW). Chaque bloc contient :
- **Index** : L'index du bloc dans la chaîne.
- **Timestamp** : L'heure de création du bloc.
- **Data** : Les données du bloc (dans cet exemple, des informations sur la session de l'utilisateur).
- **PrevHash** : Le hash du bloc précédent.
- **Hash** : Le hash du bloc actuel.
- **Nonce** : Un nombre utilisé pour la preuve de travail.

### Sessions Utilisateur

Le serveur suit les sessions des utilisateurs en fonction de leur adresse IP. Si un utilisateur reste actif pendant une durée suffisante (5 minutes dans ce cas), un nouveau bloc est ajouté à la blockchain avec les informations sur cette session.

### Statistiques

Le serveur affiche des statistiques sur le nombre de visiteurs uniques, le nombre de sessions actives, et les informations sur le dernier bloc de la blockchain.

### Routes du serveur

- **`/`** : Page d'accueil avec des liens vers la blockchain et les statistiques.
- **`/blockchain`** : Affiche la blockchain sous forme de JSON ou permet d'ajouter un nouveau bloc via une requête POST.
- **`/stats`** : Affiche les statistiques, y compris le nombre de visiteurs uniques et les détails du dernier bloc.

## Configuration

- **Difficulté de la preuve de travail** : La difficulté est définie à `4` dans le code (le nombre de zéros à ajouter au début du hash pour qu'il soit valide). Vous pouvez modifier cette valeur dans le code pour ajuster la difficulté de la preuve de travail.

- **Durée de session** : Les sessions sont vérifiées toutes les 5 minutes. Si un utilisateur reste inactif plus longtemps, un nouveau bloc est ajouté à la blockchain.

- **Fichier de log** : Les requêtes HTTP sont enregistrées dans un fichier `server.log`.

## Structure du code

- **main.go** : Le fichier principal contenant la logique de la blockchain, du serveur HTTP et des sessions utilisateur.
- **Block** : La structure représentant un bloc dans la blockchain.
- **Blockchain** : La structure représentant la chaîne de blocs.
- **UserSession** : La structure représentant la session d'un utilisateur.

## Tests

Les tests manuels peuvent être effectués en interagissant avec l'interface web à l'adresse `http://localhost:8080`. Pour tester les fonctionnalités de la blockchain, vous pouvez :
- Ajouter un bloc via l'interface web en envoyant des données.
- Consulter la blockchain en accédant à `/blockchain`.
- Vérifier les statistiques via `/stats`.

## Licence

Ce projet est sous licence MIT. Consultez le fichier LICENSE pour plus de détails.

---

Cela résume bien les fonctionnalités et fournit des instructions claires pour démarrer avec le projet. Si tu as des ajouts ou des ajustements, n’hésite pas à me le dire !
>>>>>>> 4d2c62fda33a4aae6364412e44c2c556c211d0df

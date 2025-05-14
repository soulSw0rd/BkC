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

# BkC - CryptoChain Go

![Version](https://img.shields.io/badge/version-1.2.0-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)
![License](https://img.shields.io/badge/license-GPL%20v3-green.svg)

CryptoChain Go is a modern blockchain implementation in Go, combined with an elegant web user interface for visualizing and interacting with the blockchain. This application offers a complete platform for creating blocks, making transactions, and tracking activities in a secure environment.

![CryptoChain Screenshot](https://via.placeholder.com/800x400?text=CryptoChain+Go+Screenshot)

## 🚀 Features

- **Complete blockchain** with proof of work (PoW) and automatic difficulty adjustment
- **Transaction system** allowing token exchange between users
- **Blockchain explorer** to visualize all blocks and transactions
- **Interactive dashboard** with real-time statistics
- **Data persistence** with automatic blockchain saving and loading
- **Modern user interface** with responsive design and light/dark themes
- **Secure authentication** using Argon2 for password hashing
- **RESTful API** for integration with other applications

## 📋 Prerequisites

- [Go](https://golang.org/dl/) version 1.21 or higher
- Modern web browser (Chrome, Firefox, Edge, Safari)
- Internet connection to load CSS/JS libraries via CDN
- Minimum 100 MB of free disk space for blockchain and logs

### Required Go dependencies

```
github.com/gorilla/mux v1.8.1
github.com/gorilla/csrf v1.7.2
golang.org/x/crypto v0.36.0+
```

## 🔧 Installation

### Option 1: Installation from source code

1. Clone the repository:

```bash
git clone https://github.com/soulSw0rd/BkC.git
cd BkC
```

2. Install dependencies:

```bash
go mod tidy
```

3. Compile the application:

```bash
go build
```

### Option 2: Installation with Go install

```bash
go install github.com/soulSw0rd/BkC@latest
```

## 🏃‍♂️ Getting Started

1. Create necessary directories:

```bash
mkdir -p data logs
```

2. Launch the application:

```bash
./BkC
```

Or directly without prior compilation:

```bash
go run main.go
```

3. Access the application in your browser:

```
http://localhost:8080
```

4. Log in with default credentials:
   - Username: `admin`
   - Password: `admin`

## 📊 Usage

### Exploring the blockchain

- Go to the **Blockchain** page to see all mined blocks
- View details of each block, including its transactions and hash
- Use the search bar to find specific blocks or transactions

### Managing transactions

- Create new transactions by specifying a recipient and amount
- View your transaction history, both confirmed and pending
- Receive notifications when your transactions are confirmed

### Mining

- Mine new blocks and receive token rewards
- Observe automatic difficulty adjustment based on mining time
- Track your mining rewards in your transaction history

### Dashboard

- View real-time blockchain statistics
- Visualize activity, transaction, and difficulty graphs
- Monitor your balance and recent transactions

## 🛠️ Advanced Configuration

A `config.json` configuration file can be created at the project root to customize various parameters:

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

The application exposes a RESTful API for integration with other systems:

- `GET /api/blockchain` - Retrieves the complete blockchain
- `GET /api/stats` - Gets blockchain statistics
- `POST /api/mining` - Mines a new block
- `GET /api/wallet` - Retrieves wallet information
- `POST /api/transactions` - Creates a new transaction

## 👩‍💻 Development

For developers wishing to contribute to the project:

1. Fork the repository
2. Create a branch for your feature (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 🔍 Troubleshooting

### Common issues

- **Port already in use error**: Modify the port in the `config.json` file
- **Dependency errors**: Run `go mod tidy` to update dependencies
- **Performance issues**: Adjust mining difficulty in `config.json`

### Logs

Logs are stored in the `logs` directory and follow the format `server_YYYY-MM-DD.log`.

## 📜 License

This project is licensed under the [GNU General Public License v3.0](LICENSE) - see the LICENSE file for details.

## 🙏 Acknowledgments

- The Go team for their excellent language
- The blockchain community for their inspiration and documentation
- All contributors who helped improve this project

---

Hosted on [SoulSw0rd](https://github.com/soulSw0rd)'s repository
Developed by the BKC team
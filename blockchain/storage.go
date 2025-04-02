package blockchain

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// Storage interface defines the methods that any storage implementation must provide
type Storage interface {
	// Block operations
	SaveBlock(block *Block) error
	GetBlock(hash string) (*Block, error)
	GetBlockByHeight(height int) (*Block, error)
	GetLatestBlockHeight() (int, error)

	// Transaction operations
	SaveTransaction(tx *Transaction, blockHash string) error
	GetTransaction(txID string) (*Transaction, error)
	GetTransactionsByAddress(address string) ([]*Transaction, error)

	// Mempool operations
	SaveMempool(transactions map[string]*Transaction) error
	GetMempool() (map[string]*Transaction, error)

	// Smart contract operations
	SaveContract(contract *SmartContract) error
	GetContract(contractID string) (*SmartContract, error)
	GetContractsByAddress(address string) ([]*SmartContract, error)

	// Wallet operations
	SaveAccountBalance(address string, balance float64) error
	GetAccountBalance(address string) (float64, error)

	// Utility
	Close() error
}

// LevelDBStorage implements the Storage interface using LevelDB
type LevelDBStorage struct {
	db    *leveldb.DB
	mutex sync.RWMutex
}

// NewLevelDBStorage creates a new LevelDB storage instance
func NewLevelDBStorage(dbPath string) (*LevelDBStorage, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open the database with appropriate options
	options := &opt.Options{
		BlockCacheCapacity:  32 * 1024 * 1024, // 32 MB cache
		WriteBuffer:         16 * 1024 * 1024, // 16 MB write buffer
		CompactionTableSize: 2 * 1024 * 1024,  // 2 MB
	}

	db, err := leveldb.OpenFile(dbPath, options)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &LevelDBStorage{
		db: db,
	}, nil
}

// Key prefixes for different types of data
const (
	blockPrefix       = "b:"      // b:hash -> block data
	blockHeightPrefix = "bh:"     // bh:height -> block hash
	txPrefix          = "tx:"     // tx:id -> transaction data
	txBlockPrefix     = "txb:"    // txb:id -> block hash
	addressTxPrefix   = "atx:"    // atx:address:txid -> 1
	mempoolKey        = "mempool" // mempool -> serialized mempool
	contractPrefix    = "c:"      // c:id -> contract data
	addressContPrefix = "ac:"     // ac:address:contractid -> 1
	balancePrefix     = "bal:"    // bal:address -> balance
	latestHeightKey   = "height"  // height -> latest block height
)

// SaveBlock saves a block to the database
func (s *LevelDBStorage) SaveBlock(block *Block) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Serialize the block
	blockData, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to serialize block: %w", err)
	}

	// Start a batch
	batch := new(leveldb.Batch)

	// Add block by hash
	blockKey := fmt.Sprintf("%s%s", blockPrefix, block.Hash)
	batch.Put([]byte(blockKey), blockData)

	// Add block by height
	heightKey := fmt.Sprintf("%s%d", blockHeightPrefix, block.Index)
	batch.Put([]byte(heightKey), []byte(block.Hash))

	// Update latest height
	batch.Put([]byte(latestHeightKey), []byte(fmt.Sprintf("%d", block.Index)))

	// Write the batch
	if err := s.db.Write(batch, nil); err != nil {
		return fmt.Errorf("failed to write block batch: %w", err)
	}

	return nil
}

// GetBlock retrieves a block by its hash
func (s *LevelDBStorage) GetBlock(hash string) (*Block, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Get block data
	blockKey := fmt.Sprintf("%s%s", blockPrefix, hash)
	data, err := s.db.Get([]byte(blockKey), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, fmt.Errorf("block not found: %s", hash)
		}
		return nil, fmt.Errorf("failed to retrieve block: %w", err)
	}

	// Deserialize the block
	var block Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, fmt.Errorf("failed to deserialize block: %w", err)
	}

	return &block, nil
}

// GetBlockByHeight retrieves a block by its height
func (s *LevelDBStorage) GetBlockByHeight(height int) (*Block, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Get block hash
	heightKey := fmt.Sprintf("%s%d", blockHeightPrefix, height)
	hashBytes, err := s.db.Get([]byte(heightKey), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, fmt.Errorf("block at height %d not found", height)
		}
		return nil, fmt.Errorf("failed to retrieve block hash: %w", err)
	}

	// Get block data
	return s.GetBlock(string(hashBytes))
}

// GetLatestBlockHeight retrieves the height of the latest block
func (s *LevelDBStorage) GetLatestBlockHeight() (int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Get latest height
	heightBytes, err := s.db.Get([]byte(latestHeightKey), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return -1, nil // No blocks yet
		}
		return -1, fmt.Errorf("failed to retrieve latest height: %w", err)
	}

	// Parse height
	var height int
	if _, err := fmt.Sscanf(string(heightBytes), "%d", &height); err != nil {
		return -1, fmt.Errorf("failed to parse height: %w", err)
	}

	return height, nil
}

// SaveTransaction saves a transaction to the database
func (s *LevelDBStorage) SaveTransaction(tx *Transaction, blockHash string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Serialize the transaction
	txData, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Start a batch
	batch := new(leveldb.Batch)

	// Add transaction
	txKey := fmt.Sprintf("%s%s", txPrefix, tx.ID)
	batch.Put([]byte(txKey), txData)

	// Add block reference if provided
	if blockHash != "" {
		txBlockKey := fmt.Sprintf("%s%s", txBlockPrefix, tx.ID)
		batch.Put([]byte(txBlockKey), []byte(blockHash))
	}

	// Add address indices
	if tx.Sender != "" {
		senderTxKey := fmt.Sprintf("%s%s:%s", addressTxPrefix, tx.Sender, tx.ID)
		batch.Put([]byte(senderTxKey), []byte{1})
	}

	if tx.Recipient != "" {
		recipientTxKey := fmt.Sprintf("%s%s:%s", addressTxPrefix, tx.Recipient, tx.ID)
		batch.Put([]byte(recipientTxKey), []byte{1})
	}

	// Write the batch
	if err := s.db.Write(batch, nil); err != nil {
		return fmt.Errorf("failed to write transaction batch: %w", err)
	}

	return nil
}

// GetTransaction retrieves a transaction by its ID
func (s *LevelDBStorage) GetTransaction(txID string) (*Transaction, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Get transaction data
	txKey := fmt.Sprintf("%s%s", txPrefix, txID)
	data, err := s.db.Get([]byte(txKey), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, fmt.Errorf("transaction not found: %s", txID)
		}
		return nil, fmt.Errorf("failed to retrieve transaction: %w", err)
	}

	// Deserialize the transaction
	var tx Transaction
	if err := json.Unmarshal(data, &tx); err != nil {
		return nil, fmt.Errorf("failed to deserialize transaction: %w", err)
	}

	return &tx, nil
}

// GetTransactionsByAddress retrieves all transactions for an address
func (s *LevelDBStorage) GetTransactionsByAddress(address string) ([]*Transaction, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// List to store results
	var transactions []*Transaction

	// Create a prefix iterator for the address
	prefix := fmt.Sprintf("%s%s:", addressTxPrefix, address)
	iter := s.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	defer iter.Release()

	// Iterate over transactions for this address
	for iter.Next() {
		// Extract transaction ID from the key
		key := string(iter.Key())
		parts := filepath.SplitList(key)
		if len(parts) < 3 {
			log.Printf("Invalid address transaction key: %s", key)
			continue
		}

		txID := parts[2]

		// Get the transaction
		tx, err := s.GetTransaction(txID)
		if err != nil {
			log.Printf("Error retrieving transaction %s: %v", txID, err)
			continue
		}

		transactions = append(transactions, tx)
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}

// SaveMempool saves the mempool to the database
func (s *LevelDBStorage) SaveMempool(transactions map[string]*Transaction) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Serialize the mempool
	data, err := json.Marshal(transactions)
	if err != nil {
		return fmt.Errorf("failed to serialize mempool: %w", err)
	}

	// Save to database
	if err := s.db.Put([]byte(mempoolKey), data, nil); err != nil {
		return fmt.Errorf("failed to save mempool: %w", err)
	}

	return nil
}

// GetMempool retrieves the mempool from the database
func (s *LevelDBStorage) GetMempool() (map[string]*Transaction, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Get mempool data
	data, err := s.db.Get([]byte(mempoolKey), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return make(map[string]*Transaction), nil // Empty mempool
		}
		return nil, fmt.Errorf("failed to retrieve mempool: %w", err)
	}

	// Deserialize the mempool
	var mempool map[string]*Transaction
	if err := json.Unmarshal(data, &mempool); err != nil {
		return nil, fmt.Errorf("failed to deserialize mempool: %w", err)
	}

	return mempool, nil
}

// SaveContract saves a smart contract to the database
func (s *LevelDBStorage) SaveContract(contract *SmartContract) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Serialize the contract
	contractData, err := json.Marshal(contract)
	if err != nil {
		return fmt.Errorf("failed to serialize contract: %w", err)
	}

	// Start a batch
	batch := new(leveldb.Batch)

	// Add contract
	contractKey := fmt.Sprintf("%s%s", contractPrefix, contract.ID)
	batch.Put([]byte(contractKey), contractData)

	// Add address indices
	if contract.CreatedBy != "" {
		creatorKey := fmt.Sprintf("%s%s:%s", addressContPrefix, contract.CreatedBy, contract.ID)
		batch.Put([]byte(creatorKey), []byte{1})
	}

	if contract.Recipient != "" {
		recipientKey := fmt.Sprintf("%s%s:%s", addressContPrefix, contract.Recipient, contract.ID)
		batch.Put([]byte(recipientKey), []byte{1})
	}

	// Add participant indices
	for _, participant := range contract.Participants {
		participantKey := fmt.Sprintf("%s%s:%s", addressContPrefix, participant, contract.ID)
		batch.Put([]byte(participantKey), []byte{1})
	}

	// Write the batch
	if err := s.db.Write(batch, nil); err != nil {
		return fmt.Errorf("failed to write contract batch: %w", err)
	}

	return nil
}

// GetContract retrieves a smart contract by its ID
func (s *LevelDBStorage) GetContract(contractID string) (*SmartContract, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Get contract data
	contractKey := fmt.Sprintf("%s%s", contractPrefix, contractID)
	data, err := s.db.Get([]byte(contractKey), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, fmt.Errorf("contract not found: %s", contractID)
		}
		return nil, fmt.Errorf("failed to retrieve contract: %w", err)
	}

	// Deserialize the contract
	var contract SmartContract
	if err := json.Unmarshal(data, &contract); err != nil {
		return nil, fmt.Errorf("failed to deserialize contract: %w", err)
	}

	return &contract, nil
}

// GetContractsByAddress retrieves all contracts for an address
func (s *LevelDBStorage) GetContractsByAddress(address string) ([]*SmartContract, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// List to store results
	var contracts []*SmartContract

	// Create a prefix iterator for the address
	prefix := fmt.Sprintf("%s%s:", addressContPrefix, address)
	iter := s.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	defer iter.Release()

	// Iterate over contracts for this address
	for iter.Next() {
		// Extract contract ID from the key
		key := string(iter.Key())
		parts := filepath.SplitList(key)
		if len(parts) < 3 {
			log.Printf("Invalid address contract key: %s", key)
			continue
		}

		contractID := parts[2]

		// Get the contract
		contract, err := s.GetContract(contractID)
		if err != nil {
			log.Printf("Error retrieving contract %s: %v", contractID, err)
			continue
		}

		contracts = append(contracts, contract)
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("error iterating contracts: %w", err)
	}

	return contracts, nil
}

// SaveAccountBalance saves an account's balance to the database
func (s *LevelDBStorage) SaveAccountBalance(address string, balance float64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Format balance as string
	balanceStr := fmt.Sprintf("%f", balance)

	// Save to database
	balanceKey := fmt.Sprintf("%s%s", balancePrefix, address)
	if err := s.db.Put([]byte(balanceKey), []byte(balanceStr), nil); err != nil {
		return fmt.Errorf("failed to save balance: %w", err)
	}

	return nil
}

// GetAccountBalance retrieves an account's balance from the database
func (s *LevelDBStorage) GetAccountBalance(address string) (float64, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Get balance data
	balanceKey := fmt.Sprintf("%s%s", balancePrefix, address)
	data, err := s.db.Get([]byte(balanceKey), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return 0, nil // No balance yet
		}
		return 0, fmt.Errorf("failed to retrieve balance: %w", err)
	}

	// Parse balance
	var balance float64
	if _, err := fmt.Sscanf(string(data), "%f", &balance); err != nil {
		return 0, fmt.Errorf("failed to parse balance: %w", err)
	}

	return balance, nil
}

// Close closes the database connection
func (s *LevelDBStorage) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.db.Close()
}

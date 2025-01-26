package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"q2p/blockchain/types"
	"time"

	"github.com/dgraph-io/badger/v3"
)

type Blockchain struct {
	db       *badger.DB
	lastHash []byte
	txPool   []types.Transaction
}

// NewBlockchain initializes or loads an existing blockchain
func NewBlockchain(dbPath string) (*Blockchain, error) {
	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	bc := &Blockchain{
		db:     db,
		txPool: make([]types.Transaction, 0),
	}

	// Load the last hash from DB
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err == badger.ErrKeyNotFound {
			// Initialize genesis block if blockchain is new
			genesis := bc.CreateGenesisBlock()
			return bc.AddBlock(genesis)
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			bc.lastHash = append([]byte{}, val...)
			return nil
		})
	})

	return bc, err
}

// CreateGenesisBlock creates the first block in the chain
func (bc *Blockchain) CreateGenesisBlock() *types.Block {
	genesisBlock := &types.Block{
		Hash:          []byte("genesis-block-hash"), // Add initial hash
		Transactions:  []types.Transaction{},
		PrevBlockHash: []byte("0"), // Add initial prev hash
		Timestamp:     time.Now().Unix(),
		Nonce:         0,
	}

	// Calculate proper hash for genesis block
	genesisBlock.Hash = bc.CalculateHash(genesisBlock)

	return genesisBlock
}

// AddTransaction adds a new transaction to the pool
func (bc *Blockchain) AddTransaction(tx types.Transaction) {
	// Add timestamp if not set
	if tx.Timestamp == 0 {
		tx.Timestamp = time.Now().Unix()
	}
	bc.txPool = append(bc.txPool, tx)
}

// CreateBlock creates a new block with pending transactions
func (bc *Blockchain) CreateBlock() (*types.Block, error) {
	// Create block with transactions from pool
	if len(bc.txPool) == 0 {
		return nil, fmt.Errorf("no transactions to create block")
	}

	block := &types.Block{
		Transactions:  bc.txPool,
		PrevBlockHash: bc.lastHash,
		Timestamp:     time.Now().Unix(),
		Nonce:         0,
	}

	// Calculate block hash
	block.Hash = bc.CalculateHash(block)

	// Validate block

	// Persist block
	if err := bc.AddBlock(block); err != nil {
		return nil, err
	}

	// Clear transaction pool
	bc.txPool = make([]types.Transaction, 0)

	return block, nil
}

// ValidateBlock validates a block's integrity
func (bc *Blockchain) ValidateBlock(block *types.Block) error {
	// Verify previous hash matches our last block
	if !bytes.Equal(block.PrevBlockHash, bc.lastHash) {
		return fmt.Errorf("invalid previous hash")
	}

	// Verify block hash
	expectedHash := bc.CalculateHash(block)
	if !bytes.Equal(block.Hash, expectedHash) {
		return fmt.Errorf("invalid block hash")
	}

	// Verify transactions (basic validation)
	for _, tx := range block.Transactions {
		if tx.ID == "" || tx.Timestamp == 0 {
			return fmt.Errorf("invalid transaction in block")
		}
	}

	return nil
}

// GetLastBlock retrieves the last block in the chain
func (bc *Blockchain) GetLastBlock() (*types.Block, error) {
	// Retrieve last block from DB
	if bc.lastHash == nil {
		return nil, fmt.Errorf("blockchain is empty")
	}

	var block types.Block
	err := bc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(bc.lastHash)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &block)
		})
	})

	if err != nil {
		return nil, err
	}
	return &block, nil
}

// AddBlock persists a block to the database
func (bc *Blockchain) AddBlock(block *types.Block) error {
	blockData, err := json.Marshal(block)
	if err != nil {
		return err
	}

	return bc.db.Update(func(txn *badger.Txn) error {
		// Store block
		if err := txn.Set(block.Hash, blockData); err != nil {
			return err
		}
		// Update last hash
		if err := txn.Set([]byte("lh"), block.Hash); err != nil {
			return err
		}
		bc.lastHash = block.Hash
		return nil
	})
}

// CalculateHash calculates the hash of a block
func (bc *Blockchain) CalculateHash(block *types.Block) []byte {
	blockData, _ := json.Marshal(struct {
		PrevHash     []byte
		Transactions []types.Transaction
		Timestamp    int64
		Nonce        int
	}{
		PrevHash:     block.PrevBlockHash,
		Transactions: block.Transactions,
		Timestamp:    block.Timestamp,
		Nonce:        block.Nonce,
	})

	hash := sha256.Sum256(blockData)
	return hash[:]
}

// ValidateTransaction validates a single transaction
func (bc *Blockchain) ValidateTransaction(tx *types.Transaction) error {
	if tx.ID == "" || tx.Timestamp == 0 {
		return fmt.Errorf("invalid transaction")
	}
	// Add more validation logic as needed
	return nil
}

// AddToTxPool adds a transaction to the memory pool
func (bc *Blockchain) AddToTxPool(tx *types.Transaction) error {
	if err := bc.ValidateTransaction(tx); err != nil {
		return err
	}
	bc.txPool = append(bc.txPool, *tx)
	return nil
}

// DB returns the underlying database
func (bc *Blockchain) DB() *badger.DB {
	return bc.db
}

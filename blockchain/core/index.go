package core

import (
	"q2p/blockchain/types"

	"github.com/dgraph-io/badger/v3"
)

type Blockchain struct {
	db       *badger.DB
	lastHash []byte
	txPool   []types.Transaction
}

func NewBlockchain(dbPath string) (*Blockchain, error) {
	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &Blockchain{
		db:     db,
		txPool: make([]types.Transaction, 0),
	}, nil
}

func (bc *Blockchain) AddTransaction(tx types.Transaction) {
	bc.txPool = append(bc.txPool, tx)
}

func (bc *Blockchain) CreateBlock() (*types.Block, error) {
	// Create block with transactions from pool
	// Calculate block hash
	// Validate block
	// Persist to BadgerDB
	return nil, nil
}

func (bc *Blockchain) ValidateBlock(block *types.Block) error {
	// Verify previous hash
	// Verify transactions
	// Verify block hash
	return nil
}

func (bc *Blockchain) GetLastBlock() (*types.Block, error) {
	// Retrieve last block from DB
	return nil, nil
}

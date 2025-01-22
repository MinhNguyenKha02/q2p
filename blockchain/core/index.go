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
	// TODO: Implement block creation
	return nil, nil
}

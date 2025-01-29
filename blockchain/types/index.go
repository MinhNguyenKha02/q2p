package types

import "github.com/dgraph-io/badger/v3"

// Block represents a block in the blockchain
type Block struct {
	Hash          []byte
	Transactions  []*Transaction
	PrevBlockHash []byte
	Timestamp     int64
	Nonce         int
}

// Transaction represents a single transaction
type Transaction struct {
	ID        string
	Timestamp int64
	Amount    float64
}
type Blockchain struct {
	db       *badger.DB
	lastHash []byte
}

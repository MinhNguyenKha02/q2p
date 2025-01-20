package types

import "github.com/dgraph-io/badger/v3"

type Block struct {
	Hash          []byte
	Transactions  []Transaction
	PrevBlockHash []byte
	Timestamp     int64
	Nonce         int
}

type Transaction struct {
	ID        string
	Timestamp int64
	Amount    float64
}

type Blockchain struct {
	db       *badger.DB
	lastHash []byte
}

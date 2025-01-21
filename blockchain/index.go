package blockchain

import (
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/yourusername/yourproject/blockchain/core"
	"github.com/yourusername/yourproject/blockchain/p2p"
)

// Node represents a blockchain node in the P2P network
type Node struct {
	chain *core.Blockchain
	p2p   *p2p.Service
}

// NewNode creates a new blockchain node
func NewNode(h host.Host) (*Node, error) {
	chain, err := core.NewBlockchain("./data/blockchain")
	if err != nil {
		return nil, err
	}

	return &Node{
		chain: chain,
		p2p:   p2p.NewService(h),
	}, nil
}

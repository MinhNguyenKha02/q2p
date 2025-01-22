package blockchain

import (
	"log"


	"q2p/blockchain/core"
	"q2p/blockchain/p2p"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
)

// Node represents a blockchain node in the P2P network
type Node struct {
	chain *core.Blockchain
	p2p   *p2p.Service
	host  host.Host
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
		host:  h,
	}, nil
}

// Start initializes and starts the blockchain node
func (n *Node) Start() {
	log.Println("Starting blockchain node...")

	// Set up message handlers for blockchain-related messages
	n.host.SetStreamHandler("/blockchain/1.0", n.handleBlockchainStream)

	// Start syncing with peers
	go n.syncWithPeers()

	log.Println("Blockchain node started successfully")
}

// handleBlockchainStream handles incoming blockchain-related messages
func (n *Node) handleBlockchainStream(stream network.Stream) {
	// TODO: Implement message handling logic
	log.Printf("Received new stream from peer: %s", stream.Conn().RemotePeer())
}

// syncWithPeers attempts to sync blockchain state with connected peers
func (n *Node) syncWithPeers() {
	// Request latest blocks
	// Validate and add missing blocks
	log.Println("Starting blockchain sync with peers...")
}
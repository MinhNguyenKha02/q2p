package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"q2p/blockchain/core"
	"q2p/blockchain/p2p"
	"q2p/blockchain/types"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

func main() {
	// Command line flags
	listenPort := flag.Int("port", 9000, "node listen port")
	dbPath := flag.String("db", "./data/node", "path to blockchain database")
	connectPeer := flag.String("peer", "", "peer address to connect to")
	flag.Parse()

	// Create libp2p host
	h, err := createHost(*listenPort)
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	// Print node address
	log.Printf("Node address: %s/p2p/%s\n", h.Addrs()[0], h.ID())

	// Initialize blockchain
	bc, err := core.NewBlockchain(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer bc.DB().Close()

	// Create P2P service
	p2pService := p2p.NewService(h, bc)

	// Connect to peer if specified
	if *connectPeer != "" {
		if err := p2pService.ConnectToPeer(*connectPeer); err != nil {
			log.Printf("Failed to connect to peer: %v", err)
		}
	}

	// Example: Create and broadcast a transaction every 10 seconds
	go func() {
		for {
			tx := &types.Transaction{
				ID:        fmt.Sprintf("tx-%d", time.Now().Unix()),
				Timestamp: time.Now().Unix(),
				Amount:    1.0,
			}

			if err := p2pService.BroadcastTransaction(tx); err != nil {
				log.Printf("Failed to broadcast transaction: %v", err)
			}

			time.Sleep(10 * time.Second)
		}
	}()

	// Example: Create a new block every 30 seconds if there are transactions
	go func() {
		for {
			block, err := bc.CreateBlock()
			if err != nil {
				if err.Error() != "no transactions to create block" {
					log.Printf("Failed to create block: %v", err)
				}
			} else {
				if err := p2pService.BroadcastBlock(block); err != nil {
					log.Printf("Failed to broadcast block: %v", err)
				}
			}

			time.Sleep(30 * time.Second)
		}
	}()

	// Print connected peers every 5 seconds
	go func() {
		for {
			peers := p2pService.GetPeers()
			log.Printf("Connected peers: %d", len(peers))
			time.Sleep(5 * time.Second)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down node...")
}

func createHost(port int) (host.Host, error) {
	// Create libp2p host with TCP transport
	return libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)),
	)
}

// discoveryNotifee gets notified when we find a new peer
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound connects to peers discovered via mDNS
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("discovered new peer %s\n", pi.ID.String())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID.String(), err)
	}
}

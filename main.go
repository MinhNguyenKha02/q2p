package main

import (
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
)

type Metrics struct {
	TxCount        int64
	BlockCount     int64
	ConnectedPeers int
	LastBlockTime  time.Time
}

// func monitorMetrics(bc *core.Blockchain, p2p *p2p.Service) {
// 	metrics := &Metrics{}
// 	// Update and log metrics
// }

type Config struct {
	P2P struct {
		Port           int
		BootstrapPeers []string
		MaxPeers       int
	}
	Blockchain struct {
		DBPath        string
		MaxTxPoolSize int
		BlockInterval time.Duration
		MinTxPerBlock int
	}
}

func main() {
	// Set up logging
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Println("Starting application...")

	// Command line flags
	listenPort := flag.Int("port", 9000, "node listen port")
	dbPath := flag.String("db", "./data/node", "path to blockchain database")
	connectPeer := flag.String("peer", "", "peer address to connect to")
	flag.Parse()

	log.Printf("Flags parsed - Port: %d, DB Path: %s, Peer: %s", *listenPort, *dbPath, *connectPeer)

	// Create libp2p host
	log.Println("Creating libp2p host...")
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *listenPort)),
	)
	if err != nil {
		log.Fatalf("Failed to create host: %v", err)
	}
	defer h.Close()
	log.Printf("Host created successfully. ID: %s", h.ID().String())

	// Print node address for other peers to connect
	log.Printf("Node address for other peers: /ip4/127.0.0.1/tcp/%d/p2p/%s\n",
		*listenPort, h.ID().String())

	// Ensure database directory exists
	if err := os.MkdirAll(*dbPath, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	// Initialize blockchain
	log.Println("Initializing blockchain...")
	bc, err := core.NewBlockchain(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize blockchain: %v", err)
	}
	defer func() {
		log.Println("Closing blockchain database...")
		bc.DB().Close()
	}()
	log.Println("Blockchain initialized successfully")

	// Create P2P service
	log.Println("Creating P2P service...")
	p2pService := p2p.NewService(h, bc)
	log.Println("P2P service created successfully")

	// Connect to peer if specified
	if *connectPeer != "" {
		log.Printf("Attempting to connect to peer: %s", *connectPeer)
		// Add retry logic for peer connection
		for i := 0; i < 3; i++ {
			if err := p2pService.ConnectToPeer(*connectPeer); err != nil {
				log.Printf("Attempt %d: Failed to connect to peer: %v", i+1, err)
				time.Sleep(time.Second * 2)
				continue
			}
			log.Println("Successfully connected to peer")
			break
		}
	}

	// Transaction creation goroutine
	go func() {
		log.Println("Starting transaction creation routine...")
		for {
			time.Sleep(10 * time.Second)

			tx := &types.Transaction{
				ID:        fmt.Sprintf("tx-%d", time.Now().Unix()),
				Timestamp: time.Now().Unix(),
				Amount:    1.0,
			}

			log.Printf("Creating transaction: %s", tx.ID)
			if err := p2pService.BroadcastTransaction(tx); err != nil {
				log.Printf("ERROR: Failed to broadcast transaction: %v", err)
				continue
			}
			log.Printf("Transaction broadcast successful: %s", tx.ID)
		}
	}()

	// Block creation goroutine
	go func() {
		log.Println("Starting block creation routine...")
		for {
			time.Sleep(10 * time.Second)
			if len(bc.GetTxPool()) > 10 {
				log.Println("Attempting to create new block...")
				block, err := bc.CreateBlock()
				if err != nil {
					if err.Error() != "no transactions to create block" {
						log.Printf("ERROR: Failed to create block: %v", err)
					} else {
						log.Println("No transactions available for new block")
					}
					continue
				}

				log.Printf("Broadcasting new block with %d transactions", len(block.Transactions))
				if err := p2pService.BroadcastBlock(block); err != nil {
					log.Printf("ERROR: Failed to broadcast block: %v", err)
					continue
				}
				log.Println("Block broadcast successful")
			}
		}
	}()

	// Peer monitoring goroutine
	go func() {
		log.Println("Starting peer monitoring routine...")
		for {
			time.Sleep(5 * time.Second)
			peers := p2pService.GetPeers()
			log.Printf("Connected peers: %d", len(peers))
		}
	}()

	// Add transaction pool cleanup routine
	go func() {
		log.Println("Starting transaction pool cleanup routine...")
		for {
			time.Sleep(time.Minute) // Clean pool every minute
			poolSize := len(bc.GetTxPool())
			bc.CleanTxPool()
			newSize := len(bc.GetTxPool())
			if poolSize != newSize {
				log.Printf("Cleaned transaction pool: %d -> %d transactions",
					poolSize, newSize)
			}
		}
	}()

	// Add this to periodic monitoring in main.go
	go func() {
		for {
			time.Sleep(5 * time.Second)
			log.Printf("Transaction pool status: %s", bc.GetTxPoolStatus())
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	log.Println("Node is running. Press Ctrl+C to shut down...")
	<-sigChan

	log.Println("Shutdown signal received")
	log.Println("Shutting down node...")
}

package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"q2p/blockchain/core"
	"q2p/blockchain/types"

	"bytes"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	BlockProtocolID = protocol.ID("/blockchain/blocks/1.0.0")
	TxProtocolID    = protocol.ID("/blockchain/tx/1.0.0")
	SyncProtocolID  = protocol.ID("/blockchain/sync/1.0.0")
)

type SyncMessage struct {
	Type     string // "HEIGHT_REQ", "HEIGHT_RESP", "BLOCK_REQ", "BLOCK_RESP"
	Height   int
	Block    *types.Block
	LastHash []byte
}

type Service struct {
	host host.Host
	bc   *core.Blockchain
}

func NewService(h host.Host, bc *core.Blockchain) *Service {
	s := &Service{
		host: h,
		bc:   bc,
	}
	s.initProtocols()
	return s
}

func (s *Service) initProtocols() {
	s.host.SetStreamHandler(BlockProtocolID, s.handleBlockStream)
	s.host.SetStreamHandler(TxProtocolID, s.handleTxStream)
	s.host.SetStreamHandler(SyncProtocolID, s.handleSyncStream)
}

func (s *Service) BroadcastBlock(block *types.Block) error {
	log.Printf("Broadcasting block with %d transactions. Pool before: %s",
		len(block.Transactions),
		s.bc.GetTxPoolStatus())

	successCount := 0
	for _, peer := range s.host.Network().Peers() {
		if err := s.sendBlockToPeer(peer, block); err != nil {
			log.Printf("Failed to send block to peer %s: %v", peer, err)
			continue
		}
		successCount++
	}
	if successCount == 0 {
		return fmt.Errorf("failed to broadcast block to any peers")
	}

	log.Printf("Block broadcast complete. Pool after: %s",
		s.bc.GetTxPoolStatus())

	return nil
}

func (s *Service) BroadcastTransaction(tx *types.Transaction) error {
	// First add to local pool before broadcasting
	if err := s.bc.AddToTxPool(tx); err != nil {
		return fmt.Errorf("failed to add to local pool: %v", err)
	}

	// Then broadcast to peers
	txData, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to serialize transaction: %v", err)
	}

	for _, peer := range s.host.Network().Peers() {
		stream, err := s.host.NewStream(context.Background(), peer, TxProtocolID)
		if err != nil {
			continue
		}
		defer stream.Close()

		_, err = stream.Write(txData)
		if err != nil {
			continue
		}
	}
	return nil
}

func (s *Service) handleBlockStream(stream network.Stream) {
	defer stream.Close()
	// Read block data
	data, err := io.ReadAll(stream)
	if err != nil {
		return
	}
	// Deserialize block
	var block types.Block
	if err := json.Unmarshal(data, &block); err != nil {
		return
	}
	// Validate block
	if err := s.bc.ValidateBlock(&block); err != nil {
		return
	}
	// Add to blockchain
	if err := s.bc.AddBlock(&block); err != nil {
		return
	}
	// Relay to other peers
	s.BroadcastBlock(&block)
}

func (s *Service) handleTxStream(stream network.Stream) {
	defer stream.Close()
	// Read transaction data
	data, err := io.ReadAll(stream)
	if err != nil {
		return
	}

	// Deserialize block
	var tx types.Transaction
	if err := json.Unmarshal(data, &tx); err != nil {
		return
	}
	// Validate transaction
	if err := s.bc.ValidateTransaction(&tx); err != nil {
		return
	}
	// Add to pool
	if err := s.bc.AddToTxPool(&tx); err != nil {
		return
	}
	// Relay to other peers
	s.BroadcastTransaction(&tx)
}

func (s *Service) handleSyncStream(stream network.Stream) {
	defer stream.Close()

	for {
		var msg SyncMessage
		if err := json.NewDecoder(stream).Decode(&msg); err != nil {
			if err == io.EOF {
				return
			}
			log.Printf("Error decoding: %v", err)
			return
		}

		switch msg.Type {
		case "HEIGHT_REQ":
			lastBlock, err := s.bc.GetLastBlock()
			if err != nil {
				log.Printf("Error getting last block: %v", err)
				return
			}
			resp := SyncMessage{
				Type:     "HEIGHT_RESP",
				Height:   1, // TODO: Implement proper height tracking
				LastHash: lastBlock.Hash,
			}
			if err := json.NewEncoder(stream).Encode(resp); err != nil {
				log.Printf("Error sending height response: %v", err)
				return
			}
		case "BLOCK_REQ":
			// TODO: Implement block fetching and sending
			// For now, just send the last block
			lastBlock, err := s.bc.GetLastBlock()
			if err != nil {
				log.Printf("Error getting last block: %v", err)
				return
			}
			resp := SyncMessage{
				Type:  "BLOCK_RESP",
				Block: lastBlock,
			}
			if err := json.NewEncoder(stream).Encode(resp); err != nil {
				log.Printf("Error sending block response: %v", err)
				return
			}
		}
	}
}

func (s *Service) ConnectToPeer(peerAddr string) error {
	peerInfo, err := peer.AddrInfoFromString(peerAddr)
	if err != nil {
		return fmt.Errorf("invalid peer address: %v", err)
	}

	err = s.host.Connect(context.Background(), *peerInfo)
	if err != nil {
		return fmt.Errorf("failed to connect to peer: %v", err)
	}

	return nil
}

func (s *Service) GetPeers() []peer.ID {
	return s.host.Network().Peers()
}

func (s *Service) SyncWithPeer(peer peer.ID) error {
	log.Printf("Starting sync with peer: %s", peer.String())

	// Open stream for sync
	stream, err := s.host.NewStream(context.Background(), peer, SyncProtocolID)
	if err != nil {
		return fmt.Errorf("failed to open sync stream: %v", err)
	}
	defer stream.Close()

	// Request peer's height
	heightReq := SyncMessage{Type: "HEIGHT_REQ"}
	if err := json.NewEncoder(stream).Encode(heightReq); err != nil {
		return fmt.Errorf("failed to send height request: %v", err)
	}

	// Get peer's response
	var heightResp SyncMessage
	if err := json.NewDecoder(stream).Decode(&heightResp); err != nil {
		return fmt.Errorf("failed to receive height response: %v", err)
	}

	// Compare heights and request missing blocks
	lastBlock, err := s.bc.GetLastBlock()
	if err != nil {
		return fmt.Errorf("failed to get last block: %v", err)
	}

	if heightResp.Height > 0 && !bytes.Equal(lastBlock.Hash, heightResp.LastHash) {
		// Request blocks
		blockReq := SyncMessage{
			Type:     "BLOCK_REQ",
			LastHash: lastBlock.Hash,
		}
		if err := json.NewEncoder(stream).Encode(blockReq); err != nil {
			return fmt.Errorf("failed to send block request: %v", err)
		}

		// Process incoming blocks
		for {
			var blockResp SyncMessage
			if err := json.NewDecoder(stream).Decode(&blockResp); err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("failed to receive block: %v", err)
			}

			if blockResp.Type == "BLOCK_RESP" && blockResp.Block != nil {
				if err := s.bc.AddBlock(blockResp.Block); err != nil {
					log.Printf("Failed to add synced block: %v", err)
					continue
				}
				log.Printf("Added synced block with %d transactions",
					len(blockResp.Block.Transactions))
			}
		}
	}

	log.Printf("Sync complete with peer: %s", peer.String())
	return nil
}

func (s *Service) sendBlockToPeer(peer peer.ID, block *types.Block) error {
	stream, err := s.host.NewStream(context.Background(), peer, BlockProtocolID)
	if err != nil {
		return err
	}
	defer stream.Close()

	blockData, err := json.Marshal(block)
	if err != nil {
		return err
	}

	_, err = stream.Write(blockData)
	return err
}

package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"q2p/blockchain/core"
	"q2p/blockchain/types"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	BlockProtocolID = protocol.ID("/blockchain/blocks/1.0.0")
	TxProtocolID    = protocol.ID("/blockchain/tx/1.0.0")
)

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
}

func (s *Service) BroadcastBlock(block *types.Block) error {
	// Serialize block
	blockData, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to serialize block: %v", err)
	}
	// Send to all peers
	for _, peer := range s.host.Network().Peers() {
		stream, err := s.host.NewStream(context.Background(), peer, BlockProtocolID)
		if err != nil {
			continue
		}
		defer stream.Close()

		_, err = stream.Write(blockData)
		if err != nil {
			continue
		}
	}
	return nil
}

func (s *Service) BroadcastTransaction(tx *types.Transaction) error {
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

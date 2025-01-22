package p2p

import (
	"q2p/blockchain/types"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	BlockProtocolID = protocol.ID("/blockchain/blocks/1.0.0")
	TxProtocolID    = protocol.ID("/blockchain/tx/1.0.0")
)

type Service struct {
	host host.Host
}

func NewService(h host.Host) *Service {
	s := &Service{host: h}
	s.initProtocols()
	return s
}

func (s *Service) initProtocols() {
	s.host.SetStreamHandler(BlockProtocolID, s.handleBlockStream)
	s.host.SetStreamHandler(TxProtocolID, s.handleTxStream)
}

func (s *Service) BroadcastBlock(block *types.Block) {
	// TODO: Implement block broadcasting
}

func (s *Service) BroadcastTransaction(tx *types.Transaction) {
	// TODO: Implement transaction broadcasting
}

func (s *Service) handleBlockStream(stream network.Stream) {
	// TODO: Implement block receiving and validation
}

func (s *Service) handleTxStream(stream network.Stream) {
	// TODO: Implement transaction receiving and validation
}

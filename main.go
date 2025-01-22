package main

import (
	"context"
	"fmt"
	"log"

	"q2p/blockchain"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

func main() {
	// Create a new libp2p host
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.EnableRelay(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	// Create a new mDNS service
	ser := mdns.NewMdnsService(h, "p2p-demo", &discoveryNotifee{h: h})
	if err := ser.Start(); err != nil {
		log.Fatal(err)
	}

	// Print the node's addresses
	fmt.Println("Host ID:", h.ID())
	fmt.Println("Host Addresses:", h.Addrs())

	// Initialize blockchain node
	node, err := blockchain.NewNode(h)
	if err != nil {
		log.Fatal(err)
	}
	node.Start()

	// Keep the program running
	select {}
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

package main

import (
	"context"
	"flag"
	"log"
	"strings"

	"github.com/genovatix/trustmesh/networking/p2pnetwork"
)

func main() {
	var peersFlag string
	flag.StringVar(&peersFlag, "peer", "", "Bootstrap peers to connect to")
	flag.Parse()

	ctx := context.Background()

	// Example listen addresses and protocol configuration
	listenAddrs := []string{"/ip4/0.0.0.0/tcp/0"}
	rendezvousString := "trust-mesh-rendezvous"
	protocolID := "/trustmesh/1.0.0"

	// Initialize and start the node
	node, err := p2pnetwork.NewNode(ctx, listenAddrs, rendezvousString, protocolID)
	if err != nil {
		log.Fatal("Failed to create a new Node:", err)
	}

	// Parse peersFlag and connect to bootstrap peers
	if peersFlag != "" {
		peers := strings.Split(peersFlag, ",")
		if err := node.ConnectToBootstrapPeers(ctx, peers); err != nil {
			log.Fatal("Failed to connect to bootstrap peers:", err)
		}
	}

	if err := node.Start(ctx); err != nil {
		log.Fatal("Failed to start the Node:", err)
	}

	// Keep the application running
	<-make(chan struct{})
}

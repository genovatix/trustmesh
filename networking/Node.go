package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"

	"github.com/multiformats/go-multiaddr"
)

var logger = log.Default()

const RendezvousString = "my-app-rendezvous"

// Node represents a node in the network.
type Node struct {
	Host       host.Host
	DHT        *dht.IpfsDHT
	Rendezvous string
	ProtocolID string
	Discovery  *drouting.RoutingDiscovery
}

type Message struct {
}

// NodeInterface defines the operations for a peer node in the network.
type NodeInterface interface {
	Start() error
	ConnectToPeer(ctx context.Context, peerAddr string) error
	HandleStream(stream network.Stream)
}

func (n *Node) ConnectToPeer(ctx context.Context, peerID peer.ID) error {
	fmt.Println("Connecting to peer:", peerID)

	stream, err := n.Host.NewStream(ctx, peerID, protocol.ID(n.ProtocolID))
	if err != nil {
		return fmt.Errorf("stream creation error to peer %s: %s", peerID, err)
	}

	// Handle stream in a separate goroutine
	go n.HandleStream(stream)
	return nil
}

func (n *Node) HandleStream(stream network.Stream) {
	fmt.Println("New stream established:", stream.Conn().RemotePeer())
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	// Example read-write operations (similar to the provided readData and writeData functions)
	go readData(rw)
	go writeData(rw)
}

func (n *Node) Start(ctx context.Context) error {
	// Bootstrap the DHT to build a routing table
	if err := n.DHT.Bootstrap(ctx); err != nil {
		return fmt.Errorf("DHT bootstrap error: %s", err)
	}

	// Announce ourself using the rendezvous string to make the node discoverable
	n.Discovery = drouting.NewRoutingDiscovery(n.DHT)
	dutil.Advertise(ctx, n.Discovery, n.Rendezvous)

	fmt.Println("Successfully announced on the rendezvous point:", n.Rendezvous)

	// Start listening for peers contacting us through the rendezvous point
	peerChan, err := n.Discovery.FindPeers(ctx, n.Rendezvous)
	if err != nil {
		return fmt.Errorf("Discovery find peers error: %s", err)
	}

	go func() {
		for peer := range peerChan {
			log.Println(peer.ID)
			if peer.ID == n.Host.ID() {
				continue // skip self
			}
			fmt.Println("Discovered new peer:", peer.ID)
			err := n.ConnectToPeer(ctx, peer.ID)
			if err != nil {
				panic(err)
			}
		}
	}()

	return nil
}

// NewNode creates a new Node instance with the provided configuration.
func NewNode(ctx context.Context, listenAddrs []string, rendezvous string, protocolID string) (*Node, error) {
	// Convert listen addresses to multiaddr format
	var maddrs []multiaddr.Multiaddr
	for _, addrStr := range listenAddrs {
		maddr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			return nil, err
		}
		maddrs = append(maddrs, maddr)
	}

	// Create a libp2p host
	host, err := libp2p.New(libp2p.ListenAddrs(maddrs...))

	if err != nil {
		return nil, err
	}

	// Set up a DHT for peer discovery
	kademliaDHT, err := dht.New(ctx, host)
	if err != nil {
		return nil, err
	}

	node := &Node{
		Host:       host,
		DHT:        kademliaDHT,
		Rendezvous: rendezvous,
		ProtocolID: protocolID,
	}

	return node, nil
}

func handleStream(stream network.Stream) {

	// Create a buffer stream for non-blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw)
	go writeData(rw)

	// 'stream' will stay open until you close it (or the other side closes it).
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer")
			panic(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}

// Add this method to your Node struct
func (n *Node) ConnectToBootstrapPeers(ctx context.Context, peers []string) error {
	for _, peerAddrStr := range peers {
		peerAddr, err := multiaddr.NewMultiaddr(peerAddrStr)
		if err != nil {
			fmt.Println("Error parsing multiaddr:", err)
			continue
		}

		peerinfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
		if err != nil {
			fmt.Println("Error creating peer info from multiaddr:", err)
			continue
		}

		if err := n.Host.Connect(ctx, *peerinfo); err != nil {
			fmt.Println("Error connecting to peer:", err)
			continue
		}

		fmt.Println("Connected to bootstrap peer:", peerinfo.ID)
	}

	return nil
}

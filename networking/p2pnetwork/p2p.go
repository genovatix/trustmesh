package p2pnetwork

// Message represents a message that can be gossiped or published in the network.
type Message struct {
	ID        string
	Type      string
	Payload   string
	Timestamp int64 // Consider using int64 for Unix timestamp representation
}

// GossipProtocol defines the methods for gossiping messages across the network.
type GossipProtocol interface {
	GossipMessage(message Message)
	ReceiveMessage(message Message)
}

// PublishSubscribeProtocol defines the methods for a pub/sub system within the network.
type PublishSubscribeProtocol interface {
	Publish(topic string, message Message)
	Subscribe(topic string, listener Listener)
	Unsubscribe(topic string, listener Listener)
}

// Listener is an interface for handling received messages.
type Listener interface {
	OnMessageReceived(message Message)
}

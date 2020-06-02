package ring

// NodeInfo contains the basic information of a node, and it serves as the
// identifier of the node to be passed over the RPCs
type NodeInfo struct {
	IP string
	ID uint64
}

// NodeEntry defines all RPC interfaces of a node
type NodeEntry interface {
	// FindSuccessor is called to find the successor of a given id
	FindSuccessor(id uint64, found *NodeInfo) error
	// Notify is called when the given node thinks it might be our predecessor
	Notify(node *NodeInfo, ok *bool) error
	// Next returns the successor, or returns error if has no successor
	Next(id uint64, next *NodeInfo) error
	// Previous returns the predecessor, or returns error if has no predecessor
	Previous(id uint64, next *NodeInfo) error
	// Get returns the value of a particular k
	Get(k string, v *string) error
	// Set returns true if the kv is successfully set in the storage
	Set(kv db.KV, ok *bool) error
	// Keys
    Keys(p Pattern, list *List) error
}

// Node is used as the entity of a remote chord node
type Node interface {
	// GetID returns the id of this node
	GetID() uint64
	// GetIP returns the ip of this node
	GetIP() string
	NodeEntry
}
package pkg

import (
	"fmt"
	"sync"
	"time"

	"github.com/EshaanAgg/dis/raft/pkg/logger"
	"github.com/EshaanAgg/dis/raft/rpc"
)

type RaftNode struct {
	NodeID int64

	// Internal variables used for synchronization and communication between nodes
	mu      sync.Mutex
	clients []PeerClient
	l       logger.Logger
	cfg     *Config

	// Stores if the node has been disconnected from the network. This would not be
	// present in an actual implementation, and we would not need to have checks related
	// to being disconnected in the source code. But we need to have them here, so that
	// we can simulate network partitions and failures in the tests.
	disconnected bool

	// Persistent state
	currentTerm int64
	votedFor    *int64
	log         []LogEntry

	// Transient state
	role NodeRole
	// For each client, mapping to the index of the next
	// log entry to send to that server (Initialized to leader last log index + 1)
	nextIndex []int
	// For each client, mapping to the index of the highest log entry known
	// to be replicated on that server (Initialized to 0, increases monotonically)
	matchIndex       []int
	nextElectionTime time.Time

	// Channels for stopping election and heartbeat goroutines
	stopElectionCh  chan any
	stopHeartbeatCh chan any

	rpc.UnimplementedNodeServer
}

func NewRaftNode(nodeID int64, peers []Peer, cfg *Config) *RaftNode {
	l := logger.NewLogger(fmt.Sprintf("[Node %d] ", nodeID), nodeID)
	l.Printf("Starting Raft node with ID %d", nodeID)

	clients := getPeerClients(peers, l)
	l.Printf("Connected to %d peers", len(clients))

	rn := &RaftNode{
		NodeID:       nodeID,
		l:            *l,
		clients:      clients,
		mu:           sync.Mutex{},
		cfg:          cfg,
		disconnected: false,

		// Persistent state
		// TODO: Inialize properly from stable storage
		currentTerm: 0,
		votedFor:    nil,
		log:         make([]LogEntry, 0),

		// Transient state
		role:       Follower,
		nextIndex:  make([]int, len(clients)),
		matchIndex: make([]int, len(clients)),
	}

	// Start a goroutine to periodically check if an election needs to be conducted
	rn.updateElectionTime()
	go rn.checkElection()

	return rn
}

// Returns the term of the last log entry, and 0 if there are no logs
// The caller MUST hold the lock on the mutex before calling it.
func (rn *RaftNode) getLastLogTerm() int64 {
	lastTerm := int64(0)
	if len(rn.log) > 0 {
		lastTerm = rn.log[len(rn.log)-1].Term
	}
	return lastTerm
}

// Returns the currentTerm and if the current node is the leader
func (rn *RaftNode) GetState() (int64, bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return rn.currentTerm, rn.role == Leader
}

// Shutdown stops the node
func (rn *RaftNode) Shutdown() {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	close(rn.stopElectionCh)
	close(rn.stopHeartbeatCh)
	rn.disconnected = true

	// Close client connections
	for _, client := range rn.clients {
		client.conn.Close()
	}
}

// Sets the disconnected state of the node
func (rn *RaftNode) SetDisconnected(disconnected bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	rn.disconnected = disconnected
}

// Returns whether the node is disconnected
func (rn *RaftNode) IsDisconnected() bool {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return rn.disconnected
}

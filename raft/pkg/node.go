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

	rpc.UnimplementedNodeServer
}

func NewRaftNode(nodeID int64, peers []Peer, cfg *Config) *RaftNode {
	l := logger.NewLogger(fmt.Sprintf("[Node %d] ", nodeID), nodeID)
	l.Printf("Starting Raft node with ID %d", nodeID)

	clients := getPeerClients(peers, l)
	l.Printf("Connected to [%d] peers", len(clients))

	rn := &RaftNode{
		NodeID:           nodeID,
		l:                *l,
		clients:          clients,
		mu:               sync.Mutex{},
		cfg:              cfg,
		nextElectionTime: time.Now().Add(cfg.GetElectionTimeout()),

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

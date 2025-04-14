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

func NewRaftNode(nodeID int64, peers []string, cfg *Config) *RaftNode {
	l := logger.NewLogger(fmt.Sprintf("[Node %d] ", nodeID), nodeID)
	l.Printf("Starting Raft node with ID %d", nodeID)

	clients := getPeerClients(peers, l)
	l.Printf("Connected to [%d] peers", len(clients))

	return &RaftNode{
		NodeID:  nodeID,
		l:       *l,
		clients: clients,
		mu:      sync.Mutex{},
		cfg:     cfg,

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
}

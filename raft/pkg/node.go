package pkg

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EshaanAgg/dis/raft/pkg/logger"
	"github.com/EshaanAgg/dis/raft/rpc"
)

type RaftNode struct {
	NodeID int64

	// Internal variables used for synchronization and communication between nodes
	mu           sync.Mutex
	peers        []Peer
	clients      []PeerClient
	l            logger.Logger
	cfg          *Config
	disconnected bool // Used to mock disconnections from the network

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

	// Cancel functions to manage go-routines
	stopHeartbeatCancel *context.CancelFunc
	stopElectionCancel  *context.CancelFunc

	rpc.UnimplementedNodeServer
}

func NewRaftNode(nodeID int64, peers []Peer, cfg *Config, logLevel logger.LogLevel) *RaftNode {
	l := logger.NewLogger(fmt.Sprintf("[Node %d] ", nodeID), nodeID, &logLevel)
	l.Info("Starting Raft node with ID %d", nodeID)

	rn := &RaftNode{
		NodeID:       nodeID,
		l:            *l,
		mu:           sync.Mutex{},
		cfg:          cfg,
		disconnected: false,
		peers:        peers,
	}

	rn.setPeerClients()
	l.Info("Connected to %d peers", len(rn.clients))

	rn.resetState()
	return rn
}

// Initializes the state of the node when it is first starting up after a crash.
// Responsible for handling loading of state from the stable medium, and kicking off the
// appropiate election go-routines.
// The caller must hold the lock on the mutex (if needed).
func (rn *RaftNode) resetState() {
	// TODO: Inialize properly from stable storage
	// Persistent state
	rn.currentTerm = 0
	rn.votedFor = nil
	rn.log = make([]LogEntry, 0)

	// Transient state
	rn.role = Follower
	rn.nextIndex = make([]int, len(rn.clients))
	rn.matchIndex = make([]int, len(rn.clients))

	// Cancel fucntions
	rn.stopElectionCancel = nil
	rn.stopHeartbeatCancel = nil

	rn.updateElectionTime()
	go rn.checkElection()
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

// Sets the disconnected state of the node
func (rn *RaftNode) SetDisconnected(shouldDisconnect bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if rn.disconnected == shouldDisconnect {
		return
	}

	if shouldDisconnect {
		// Need to disconnect by shutting down the node
		rn.disconnected = true
		rn.revertToFollower()
		(*rn.stopElectionCancel)()
	} else {
		// Need to re-initialize the node as if the same is joining the network
		rn.disconnected = false
		rn.resetState()
	}
}

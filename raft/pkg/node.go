package pkg

import (
	"fmt"
	"sync"

	"github.com/EshaanAgg/dis/raft/pkg/logger"
	"github.com/EshaanAgg/dis/raft/rpc"
)

type RaftNode struct {
	NodeID int64

	// Internal variables used for synchronization and communication between nodes
	mu      sync.Mutex
	clients map[string]PeerClient
	l       logger.Logger

	// Persistent state
	state NodeState

	// Transient state

	rpc.UnimplementedNodeServer
}

func NewRaftNode(nodeID int64, peers []string) *RaftNode {
	rn := &RaftNode{
		NodeID:  nodeID,
		clients: make(map[string]PeerClient),
		state:   Follower,
		l:       *logger.NewLogger(fmt.Sprintf("[Node %d] ", nodeID), nodeID),
	}

	go rn.ConnectToPeers(peers)
	return rn
}

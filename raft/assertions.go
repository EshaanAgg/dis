package raft

import (
	"testing"
	"time"
)

// Gracefully stop all gRPC servers
func (r *Raft) Cleanup() {
	for _, server := range r.servers {
		server.GracefulStop()
	}
}

// Check that there is exactly one leader among the nodes.
// Returns the leader node ID.
func (r *Raft) AssertOneLeader(t *testing.T) int {
	for range 10 {
		time.Sleep(500 * time.Millisecond)

		leaders := make(map[int64][]int)
		for i, node := range r.nodes {
			term, isLeader := node.GetState()
			if isLeader {
				leaders[term] = append(leaders[term], i)
			}
		}

		lastTermWithLeader := int64(-1)
		for term, ids := range leaders {
			if len(ids) > 1 {
				t.Errorf("Term %d has %d (>1) leaders: %v", term, len(ids), ids)
			}
			if term > lastTermWithLeader {
				lastTermWithLeader = term
			}
		}

		if len(leaders) != 0 {
			return leaders[lastTermWithLeader][0]
		}
	}

	t.Errorf("Expected one leader, got none")
	return -1
}

package raft

import (
	"testing"
	"time"
)

type RaftTest struct {
	*Raft
	t *testing.T
}

// Check that there is exactly one leader among the nodes.
// Returns the leader node ID.
func (rt *RaftTest) AssertOneLeader() int {
	for range 10 {
		time.Sleep(500 * time.Millisecond)

		leaders := make(map[int64][]int)
		for i, node := range rt.nodes {
			term, isLeader := node.GetState()
			if isLeader {
				leaders[term] = append(leaders[term], i)
			}
		}

		lastTermWithLeader := int64(-1)
		for term, ids := range leaders {
			if len(ids) > 1 {
				rt.t.Errorf("Term %d has %d (>1) leaders: %v", term, len(ids), ids)
			}
			if term > lastTermWithLeader {
				lastTermWithLeader = term
			}
		}

		if len(leaders) != 0 {
			return leaders[lastTermWithLeader][0]
		}
	}

	rt.t.Errorf("Expected one leader, got none")
	return -1
}

func (rt *RaftTest) AssertOneLeaderExcept(nodes ...int) int {
	leader := rt.AssertOneLeader()
	for _, n := range nodes {
		if n == leader {
			rt.t.Fatalf("Node %d is leader, but nodes %v can't be the leader", n, nodes)
		}
	}
	return leader
}

func (rt *RaftTest) AssertOneLeaderIn(nodes ...int) int {
	leader := rt.AssertOneLeader()
	for _, n := range nodes {
		if n == leader {
			return leader
		}
	}
	rt.t.Fatalf("The leader is %d, which not is the provided nodes %v", leader, nodes)
	return -1
}

// Checks that everyone agrees on the term.
// Returns the term if successful.
func (rt *RaftTest) CheckTerms() int64 {
	term := int64(-1)

	for i := range rt.cfg.NumberOfNodes {
		if rt.connected[i] {
			xterm, _ := rt.nodes[i].GetState()
			if term == -1 {
				term = xterm
			} else if term != xterm {
				rt.t.Fatalf("servers disagree on term")
			}
		}
	}

	return term
}

// Checks that there's no leader.
func (rt *RaftTest) AssertNoLeader() {
	for i := range rt.cfg.NumberOfNodes {
		if rt.connected[i] {
			_, isLeader := rt.nodes[i].GetState()
			if isLeader {
				rt.t.Fatalf("expected no leader, but %v claims to be leader", i)
			}
		}
	}
}

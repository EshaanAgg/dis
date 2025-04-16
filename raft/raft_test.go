package raft

import (
	"testing"
	"time"
)

// The tester generously allows solutions to complete elections in one second
// (much more than the paper's range of timeouts).
const RaftElectionTimeout = 1000 * time.Millisecond

func TestInitialElection2A(t *testing.T) {
	raft := NewRaft(&Config{
		NumberOfNodes: 3,
		StartPort:     5000,
	})
	defer raft.Cleanup()

	t.Logf("Test A: Initial election...\n")

	// Is a leader elected?
	raft.AssertOneLeader(t)

	// does the leader+term stay the same if there is no network failure?
	// term1 := cfg.checkTerms()
	// time.Sleep(2 * RaftElectionTimeout)
	// term2 := cfg.checkTerms()
	// if term1 != term2 {
	// 	fmt.Printf("warning: term changed even though there were no failures")
	// }
	t.Logf("...passed\n")
}

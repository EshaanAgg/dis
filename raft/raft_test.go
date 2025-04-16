package raft

import (
	"testing"
	"time"

	"github.com/EshaanAgg/dis/raft/pkg/logger"
)

// The tester generously allows solutions to complete elections in one second
// (much more than the paper's range of timeouts).
const RaftElectionTimeout = 1000 * time.Millisecond

var usedPort = 5000

// Creates a new RaftTest that can be used for testing.
// Ensures that the ports are not shared between various tests to avoid port already in use issues.
func NewRaftTest(serverCnt int, t *testing.T) RaftTest {
	raft := NewRaft(
		&Config{
			NumberOfNodes: serverCnt,
			StartPort:     usedPort,
		}, logger.Info)

	usedPort += serverCnt
	return RaftTest{
		raft,
		t,
	}
}

func TestInitialElection(t *testing.T) {
	r := NewRaftTest(3, t)
	defer r.Cleanup()

	t.Logf("Test A: Initial election...\n")

	// Is a leader elected?
	r.AssertOneLeader()

	// Does the leader & term stay the same if there is no network failure?
	term1 := r.CheckTerms()
	time.Sleep(2 * RaftElectionTimeout)
	term2 := r.CheckTerms()
	if term1 != term2 {
		t.Fatalf("term changed even though there were no failures")
	}

	t.Log("...passed")
}

func TestReElection(t *testing.T) {
	r := NewRaftTest(3, t)
	defer r.Cleanup()

	t.Logf("Test B: Election after network failure...\n")

	// Check for a leader
	leader1 := r.AssertOneLeader()

	// If the leader disconnects, a new one should be elected
	t.Log("...disconnecting leader")
	r.Disconnect(leader1)
	r.AssertOneLeaderExcept(leader1)

	// If the old leader rejoins, that shouldn't disturb the old leader
	t.Log("...reconnecting old leader")
	r.Connect(leader1)
	r.AssertOneLeader()

	// If there's no quorum, no leader should be elected
	t.Log("...disconnecting all but server node 2")
	r.Disconnect(0)
	r.Disconnect(1)
	time.Sleep(2 * RaftElectionTimeout)
	r.AssertNoLeader()

	// If a quorum arises, it should elect a leader
	t.Log("...reconnecting one server")
	r.Connect(0)
	r.AssertOneLeaderIn(0, 2)

	// Re-joining of last node shouldn't prevent leader from existing
	t.Log("...reconnecting the final node")
	r.Connect(1)
	r.AssertOneLeader()

	t.Log("...passed")
}

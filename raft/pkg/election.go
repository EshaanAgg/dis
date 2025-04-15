package pkg

import (
	"context"
	"sync"
	"time"

	"github.com/EshaanAgg/dis/raft/rpc"
)

// A goroutine that periodically checks if an election needs to be conducted
func (rn *RaftNode) checkElection() {
	for {
		rn.mu.Lock()

		if rn.role == Follower && time.Now().After(rn.nextElectionTime) {
			rn.mu.Unlock()
			rn.conductElection()
		} else {
			rn.mu.Unlock()
		}
	}
}

func (rn *RaftNode) conductElection() {
	rn.mu.Lock()

	// Check role and time
	if rn.role != Follower || time.Now().Before(rn.nextElectionTime) {
		rn.mu.Unlock()
		return
	}

	rn.l.Printf("Starting election for term %d", rn.currentTerm+1)

	// Transition to candidate state
	rn.currentTerm++
	rn.votedFor = &rn.NodeID
	rn.role = Candidate
	rn.updateElectionTime()

	// Initialize the voting procedure
	voteMutex := sync.Mutex{}
	cond := sync.NewCond(&voteMutex)
	voteCount := 1
	totalCount := 1

	logLen := int64(len(rn.log))
	logTerm := rn.getLastLogTerm()

	for _, peer := range rn.clients {

		go func(nodeId int64, currentTerm int64, logLength int64, lastLogTerm int64, peer *PeerClient) {
			reply, err := peer.RequestVote(context.Background(), &rpc.RequestVoteInput{
				Term:        currentTerm,
				CandidateId: nodeId,
				LogLength:   logLength,
				LastLogTerm: lastLogTerm,
			})

			if err != nil {
				rn.l.Printf("Failed to connect Node %d via RequestVote RPC: (%v)", nodeId, err)
			} else {
				// Update the current term from the response
				// and state
				rn.mu.Lock()
				if rn.currentTerm != reply.Term {
					rn.currentTerm = reply.Term
					rn.role = Follower
				}
				rn.mu.Unlock()
			}

			voteMutex.Lock()
			defer voteMutex.Unlock()
			if err == nil && reply.VoteGranted {
				voteCount++
			}
			totalCount++
			cond.Broadcast()

		}(peer.ID, rn.currentTerm, logLen, logTerm, &peer)

	}

	rn.mu.Unlock()

	voteMutex.Lock()
	defer voteMutex.Unlock()

	for 2*voteCount <= len(rn.clients) && totalCount != len(rn.clients) {
		cond.Wait()
	}

	rn.mu.Lock()

	if 2*voteCount > len(rn.clients) && rn.role == Candidate {
		rn.l.Printf("Became the leader: Recieved %d/%d", voteCount, totalCount)
		rn.role = Leader
		// Send AppendEntries messages to all clients to establish myself as leader
		rn.sendAppendEntriesMesssage()
	} else {
		rn.l.Printf("Lost the election: Recieved %d/%d", voteCount, totalCount)
		rn.role = Follower
	}

	rn.mu.Unlock()
}

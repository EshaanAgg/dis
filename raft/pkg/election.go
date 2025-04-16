package pkg

import (
	"context"
	"sync"
	"time"

	"github.com/EshaanAgg/dis/raft/rpc"
)

// A goroutine that periodically checks if an election needs to be conducted
func (rn *RaftNode) checkElection() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	ctx, cancel := context.WithCancel(context.Background())

	rn.mu.Lock()
	rn.stopElectionCancel = &cancel
	rn.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rn.mu.Lock()
			if rn.role == Follower && time.Now().After(rn.nextElectionTime) {
				rn.mu.Unlock()
				rn.conductElection()
			} else {
				rn.mu.Unlock()
			}
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
	rn.l.Debug("Starting election for term %d", rn.currentTerm+1)

	// Transition to candidate state
	rn.currentTerm += 1
	rn.votedFor = &rn.NodeID
	rn.role = Candidate

	electionTerm := rn.currentTerm
	nodesCnt := len(rn.clients) + 1

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
				rn.l.Debug("Failed to connect Node %d via RequestVote RPC: (%v)", nodeId, err)
			} else {
				// Update the current term from the response and state
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

		}(peer.ID, electionTerm, logLen, logTerm, &peer)
	}

	rn.mu.Unlock()

	voteMutex.Lock()
	defer voteMutex.Unlock()
	for 2*voteCount <= nodesCnt && totalCount != nodesCnt {
		cond.Wait()
	}

	rn.mu.Lock()
	if 2*voteCount > nodesCnt && rn.role == Candidate {
		rn.l.Info("Became the leader: Recieved %d/%d for term %d", voteCount, totalCount, electionTerm)
		rn.role = Leader
		go rn.leaderHeartbeat()
	} else {
		rn.l.Info("Lost the election: Recieved %d/%d for term %d", voteCount, totalCount, electionTerm)
		rn.role = Follower
	}

	// Update the election time only at the end when the election process has completed
	rn.updateElectionTime()

	rn.mu.Unlock()
}

// Continuously send heartbeats when in leader state.
func (rn *RaftNode) leaderHeartbeat() {
	ticker := time.NewTicker(rn.cfg.HeartbeatTimeout)
	defer ticker.Stop()

	rn.mu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	rn.stopHeartbeatCancel = &cancel
	rn.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rn.mu.Lock()
			if rn.role == Leader {
				rn.sendAppendEntriesMesssage()
			}
			rn.mu.Unlock()
		}
	}
}

// Step down as leader/candidate and become a follower.
// The caller must hold the lock on the mutex before calling this function.
func (rn *RaftNode) revertToFollower() {
	if rn.role == Leader {
		rn.l.Info("Stepped down as leader")
		if rn.stopHeartbeatCancel != nil {
			(*rn.stopHeartbeatCancel)()
		}
	}

	rn.votedFor = nil
	rn.role = Follower
	rn.updateElectionTime()
}

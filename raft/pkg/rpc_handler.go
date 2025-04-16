package pkg

import (
	"context"
	"fmt"

	"github.com/EshaanAgg/dis/raft/rpc"
)

func (rn *RaftNode) RequestVote(ctx context.Context, args *rpc.RequestVoteInput) (*rpc.RequestVoteOutput, error) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	// Manually disconnect the node from the network
	if rn.disconnected {
		return nil, fmt.Errorf("node %d is disconnected from the network", rn.NodeID)
	}

	// If the message is from a newer term, update the current term and reset the votedFor field
	if args.Term > rn.currentTerm {
		rn.currentTerm = args.Term
		rn.stepDownAsLeader()
	}

	// Check that the candidate's log is at least as up-to-date as the receiver's log
	lastTerm := rn.getLastLogTerm()
	logOk := (args.LastLogTerm > lastTerm) || (args.LastLogTerm == lastTerm && args.LogLength >= int64(len(rn.log)))

	// Check the term, votedFor, and logOk conditions to finally grant the vote
	if (args.Term == rn.currentTerm) && (rn.votedFor == nil || *rn.votedFor == args.CandidateId) && logOk {
		// Grant the vote
		rn.votedFor = &args.CandidateId
		return &rpc.RequestVoteOutput{
			Term:        rn.currentTerm,
			VoteGranted: true,
		}, nil
	} else {
		// Deny the vote
		return &rpc.RequestVoteOutput{
			Term:        rn.currentTerm,
			VoteGranted: false,
		}, nil
	}
}

func (rn *RaftNode) AppendEntries(ctx context.Context, args *rpc.AppendEntryInput) (*rpc.AppendEntryOutput, error) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	// Manually disconnect the node from the network
	if rn.disconnected {
		return nil, fmt.Errorf("node %d is disconnected from the network", rn.NodeID)
	}

	// If from a previous term, deny the request
	if args.Term < rn.currentTerm {
		return &rpc.AppendEntryOutput{
			Term:    rn.currentTerm,
			Success: false,
		}, nil
	}

	// Become a follower since the leader is known and update the timeout
	rn.role = Follower
	rn.votedFor = nil
	rn.currentTerm = args.Term
	rn.updateElectionTime()

	// TODO: Add actual logic of appending the log entries
	return &rpc.AppendEntryOutput{
		Term:    rn.currentTerm,
		Success: true,
	}, nil
}

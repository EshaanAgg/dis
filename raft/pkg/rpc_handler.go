package pkg

import (
	"context"

	"github.com/EshaanAgg/dis/raft/rpc"
)

func (rn *RaftNode) RequestVote(ctx context.Context, args *rpc.RequestVoteInput) (*rpc.RequestVoteOutput, error) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	// If the message is from a newer term, update the current term and reset the votedFor field
	if args.Term > rn.currentTerm {
		rn.currentTerm = args.Term
		rn.votedFor = nil
		rn.role = Follower
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

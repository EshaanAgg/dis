package pkg

import (
	"context"

	"github.com/EshaanAgg/dis/raft/pkg/logger"
	"github.com/EshaanAgg/dis/raft/rpc"
)

// Sends the AppendEntires message to all the connected peers.
// The lock must be held by the caller while calling it.
func (rn *RaftNode) sendAppendEntriesMesssage() {
	// TODO: Properly construct the args
	args := &rpc.AppendEntryInput{
		Term:     rn.currentTerm,
		LeaderId: rn.NodeID,
	}

	for _, peer := range rn.clients {
		go func(args *rpc.AppendEntryInput, peer *PeerClient, l *logger.Logger) {
			resp, err := peer.AppendEntries(context.Background(), args)
			if err != nil {
				l.Printf("Failed to contact Node %d via AppendEntries RPC (%v)", peer.ID, err)
				return
			}

			rn.mu.Lock()
			defer rn.mu.Unlock()

			// Handle the case the node is disconnected
			if rn.disconnected {
				return
			}

			if rn.currentTerm == resp.Term && rn.role == Leader {
				// TODO: Handle the reply
			} else if rn.currentTerm < resp.Term {
				rn.currentTerm = resp.Term
				rn.stepDownAsLeader()
			}

		}(args, &peer, &rn.l)
	}
}

package pkg

import "github.com/EshaanAgg/dis/raft/rpc"

type NodeRole int

const (
	Leader NodeRole = iota
	Candidate
	Follower
)

type LogEntry struct {
	Term int64
	rpc.LogEntry
}

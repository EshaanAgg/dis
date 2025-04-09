package pkg

type NodeState int

const (
	Leader NodeState = iota
	Candidate
	Follower
)

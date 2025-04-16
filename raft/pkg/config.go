package pkg

import (
	"math/rand"
	"time"
)

type Config struct {
	MinimumElectionTimeout time.Duration
	MaximumElectionTimeout time.Duration
	HeartbeatTimeout       time.Duration
}

func NewConfig() *Config {
	return &Config{
		MinimumElectionTimeout: 300 * time.Millisecond,
		MaximumElectionTimeout: 400 * time.Millisecond,
		HeartbeatTimeout:       50 * time.Millisecond,
	}
}

func (c *Config) GetElectionTimeout() time.Duration {
	// Randomly select a timeout between the minimum and maximum election timeout
	// so that all the nodes do not timeout at the same time.
	electionRange := c.MaximumElectionTimeout - c.MinimumElectionTimeout
	return c.MinimumElectionTimeout + time.Duration(rand.Int63n(int64(electionRange)))
}

// Updates the next election time for the node.
// The lock on the node must be held by the caller.
func (rn *RaftNode) updateElectionTime() {
	rn.nextElectionTime = time.Now().Add(rn.cfg.GetElectionTimeout())
}

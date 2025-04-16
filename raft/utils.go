package raft

import "log"

// Disconnects a server from the network
func (r *Raft) Disconnect(i int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.connected[i] = false
	r.nodes[i].SetDisconnected(true)

	log.Printf("Disconnected server %d", i)
}

// Connects a server to the network
func (r *Raft) Connect(i int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.connected[i] = true
	r.nodes[i].SetDisconnected(false)

	log.Printf("Connected server %d", i)
}

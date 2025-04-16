package pkg

import (
	"sync"

	"github.com/EshaanAgg/dis/raft/pkg/logger"
	"github.com/EshaanAgg/dis/raft/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PeerClient struct {
	rpc.NodeClient

	ID   int64
	addr string
	conn *grpc.ClientConn
}

type Peer struct {
	ID   int64
	Addr string
}

func getPeerClients(peers []Peer, l *logger.Logger) []PeerClient {
	// This function connects to all peers in the cluster and initializes the gRPC clients for communication
	// and returns a slice of PeerClient structs.
	clients := make([]PeerClient, 0)
	mtx := sync.Mutex{}
	wg := sync.WaitGroup{}

	for _, peer := range peers {
		// Use a goroutine to connect to each peer concurrently
		wg.Add(1)

		go func(peer Peer) {
			defer wg.Done()

			// Create a new gRPC client connection to the peer
			conn, err := grpc.NewClient(peer.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				l.Error("Failed to connect to peer '%s': %v", peer.Addr, err)
				return
			}
			client := rpc.NewNodeClient(conn)
			l.Info("Connected to peer '%s'", peer.Addr)

			mtx.Lock() // Lock the mutex to safely append to the clients slice
			clients = append(clients, PeerClient{
				addr:       peer.Addr,
				conn:       conn,
				NodeClient: client,
				ID:         peer.ID,
			})
			mtx.Unlock()
		}(peer)
	}

	wg.Wait()
	return clients
}

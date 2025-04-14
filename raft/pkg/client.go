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

	addr string
	conn *grpc.ClientConn
}

func getPeerClients(peers []string, l *logger.Logger) []PeerClient {
	// This function connects to all peers in the cluster and initializes the gRPC clients for communication
	// and returns a slice of PeerClient structs.
	clients := make([]PeerClient, 0)
	mtx := sync.Mutex{}
	wg := sync.WaitGroup{}

	for _, addr := range peers {
		// Use a goroutine to connect to each peer concurrently
		wg.Add(1)

		go func(peerAddr string) {
			defer wg.Done()

			// Create a new gRPC client connection to the peer
			conn, err := grpc.NewClient(peerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				l.Printf("Failed to connect to peer '%s': %v", peerAddr, err)
				return
			}
			client := rpc.NewNodeClient(conn)
			l.Printf("Connected to peer '%s'", peerAddr)

			mtx.Lock() // Lock the mutex to safely append to the clients slice
			clients = append(clients, PeerClient{addr: peerAddr, conn: conn, NodeClient: client})
			mtx.Unlock()
		}(addr)
	}

	wg.Wait()
	return clients
}

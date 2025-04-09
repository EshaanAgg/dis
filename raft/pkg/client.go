package pkg

import (
	"github.com/EshaanAgg/dis/raft/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PeerClient struct {
	rpc.NodeClient

	addr string
	conn *grpc.ClientConn
}

func (rn *RaftNode) ConnectToPeers(peers []string) {
	// This function connects to all peers in the cluster and initializes the gRPC clients for communication
	rn.mu.Lock()
	defer rn.mu.Unlock()

	for _, addr := range peers {
		go func(peerAddr string) {
			conn, err := grpc.NewClient(peerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				rn.l.Printf("Failed to connect to peer '%s': %v", peerAddr, err)
				return
			}
			client := rpc.NewNodeClient(conn)

			rn.l.Printf("Connected to peer '%s'", peerAddr)
			rn.clients[peerAddr] = PeerClient{addr: peerAddr, conn: conn, NodeClient: client}
		}(addr)
	}
}

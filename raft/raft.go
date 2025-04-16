package raft

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/EshaanAgg/dis/raft/pkg"
	"github.com/EshaanAgg/dis/raft/pkg/logger"
	"github.com/EshaanAgg/dis/raft/rpc"
	"google.golang.org/grpc"
)

type Config struct {
	StartPort     int
	NumberOfNodes int
}

type Raft struct {
	cfg       *Config
	servers   []*grpc.Server
	nodes     []*pkg.RaftNode
	connected []bool

	mu sync.Mutex
}

func NewRaft(config *Config, logLevel logger.LogLevel) *Raft {
	if config.NumberOfNodes <= 0 {
		log.Fatalf("Number of nodes must be greater than 0")
	}
	if config.StartPort <= 0 {
		log.Fatalf("Starting port must be greater than 0")
	}

	serverCh := make(chan *grpc.Server, config.NumberOfNodes)
	nodeCh := make(chan *pkg.RaftNode, config.NumberOfNodes)
	var wg sync.WaitGroup

	for nodeID := range config.NumberOfNodes {
		wg.Add(1)
		port := config.StartPort + nodeID

		peers := make([]pkg.Peer, 0, config.NumberOfNodes-1)
		for i := range config.NumberOfNodes {
			if i == nodeID {
				continue
			}
			peers = append(peers, pkg.Peer{
				Addr: fmt.Sprintf("localhost:%d", config.StartPort+i),
				ID:   int64(i),
			})
		}

		go func(nodeID, port int, peers []pkg.Peer) {
			defer wg.Done()
			server, node := startNode(nodeID, port, peers, logLevel)
			serverCh <- server
			nodeCh <- node
		}(nodeID, port, peers)
	}

	// Wait for all servers to be added
	wg.Wait()
	close(serverCh)
	close(nodeCh)

	servers := make([]*grpc.Server, 0, config.NumberOfNodes)
	for server := range serverCh {
		servers = append(servers, server)
	}

	nodes := make([]*pkg.RaftNode, 0, config.NumberOfNodes)
	for n := range nodeCh {
		nodes = append(nodes, n)
	}

	return &Raft{
		cfg:       config,
		servers:   servers,
		nodes:     nodes,
		connected: make([]bool, config.NumberOfNodes),
	}
}

func startNode(nodeID, port int, peers []pkg.Peer, logLevel logger.LogLevel) (*grpc.Server, *pkg.RaftNode) {
	rn := pkg.NewRaftNode(int64(nodeID), peers, pkg.NewConfig(), logLevel)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	rpc.RegisterNodeServer(grpcServer, rn)

	log.Printf("Starting RAFT node %d on port %d", nodeID, port)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	return grpcServer, rn
}

// Gracefully stop all gRPC servers
func (r *Raft) Cleanup() {
	for _, server := range r.servers {
		server.GracefulStop()
	}
}

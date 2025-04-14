package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/EshaanAgg/dis/raft/pkg"
	"github.com/EshaanAgg/dis/raft/rpc"
	"google.golang.org/grpc"
)

type Config struct {
	StartPort     int
	NumberOfNodes int
}

func main() {
	var config Config
	flag.IntVar(&config.StartPort, "start-port", 5000, "Starting port for the nodes")
	flag.IntVar(&config.NumberOfNodes, "number-of-nodes", 3, "Number of nodes in the cluster")
	flag.Parse()

	if config.NumberOfNodes <= 0 {
		log.Fatalf("Number of nodes must be greater than 0")
	}
	if config.StartPort <= 0 {
		log.Fatalf("Starting port must be greater than 0")
	}

	for nodeID := range config.NumberOfNodes {
		port := config.StartPort + nodeID

		peers := make([]string, 0, config.NumberOfNodes-1)
		for i := range config.NumberOfNodes {
			if i == nodeID {
				continue
			}
			peers = append(peers, fmt.Sprintf("localhost:%d", config.StartPort+i))
		}
		go startNode(nodeID, port, peers)
	}

	// Keep the main function running
	select {}
}

func startNode(nodeID, port int, peers []string) {
	rn := pkg.NewRaftNode(int64(nodeID), peers, pkg.NewConfig())

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	rpc.RegisterNodeServer(grpcServer, rn)

	log.Printf("Starting RAFT node %d on port %d", nodeID, port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}

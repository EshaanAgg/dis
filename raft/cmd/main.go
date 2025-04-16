package main

import (
	"flag"

	"github.com/EshaanAgg/dis/raft"
	"github.com/EshaanAgg/dis/raft/pkg/logger"
)

type Config struct {
	StartPort     int
	NumberOfNodes int
}

func main() {
	var config raft.Config
	flag.IntVar(&config.StartPort, "start-port", 5000, "Starting port for the nodes")
	flag.IntVar(&config.NumberOfNodes, "number-of-nodes", 3, "Number of nodes in the cluster")
	flag.Parse()

	raft.NewRaft(&config, logger.Info)

	// Keep the main loop running
	select {}
}

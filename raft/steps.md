# Implementation Log 

- Defined the RPC contracts in [`raft.proto`](../rpc/raft.proto) and used `gRPC` to generate the procedures for the same. 
- Define a basic `Node` type that will represent a server in the cluster. Embed the `rpc.UnimplementedNodeServer` type in the `Node` struct to implement the RPC methods.
- Added the initial methods to connect to all the peers in the cluster, and added a `main`/`cmd` package so that we can start the cluster via CLI.
- Implemented a simple logger class that prints the logs to the STDOUT, but with node information and custom colours for better visibility. 

Commit: [`282770c3b97d8c681d3244ec56f80fde39d6f56a`](https://github.com/EshaanAgg/dis/tree/282770c3b97d8c681d3244ec56f80fde39d6f56a/raft)

- Changed the `rn.clients` to be a slice instead of a map. This would allow direct mapping via index when working with the `nectIndex` and `matchIndex` slices.
- Added a `Config` struct to the package that can be used to configure the timeouts. This should probably be helpful when we want to test the implementation. 
- Implemented the `RequestVote` RPC method, as it is the first step in the RAFT algorithm. Implemented the logic for term validation, log comparison, and the voting process.
- Made minor refactors to the variables, as well as the documentation.
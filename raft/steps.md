# Implementation Log 

- Defined the RPC contracts in [`raft.proto`](../rpc/raft.proto) and used `gRPC` to generate the procedures for the same. 
- Define a basic `Node` type that will represent a server in the cluster. Embed the `rpc.UnimplementedNodeServer` type in the `Node` struct to implement the RPC methods.
- Added the initial methods to connect to all the peers in the cluster, and added a `main`/`cmd` package so that we can start the cluster via CLI.
- Implemented a simple logger class that prints the logs to the STDOUT, but with node information for better visibility. 

Commit: [`282770c3b97d8c681d3244ec56f80fde39d6f56a`](https://github.com/EshaanAgg/dis/tree/282770c3b97d8c681d3244ec56f80fde39d6f56a/raft)

- Changed the `rn.clients` to be a slice instead of a map. This would allow direct mapping via index when working with the `nectIndex` and `matchIndex` slices.
- Added a `Config` struct to the package that can be used to configure the timeouts. This should probably be helpful when we want to test the implementation. 
- Implemented the `RequestVote` RPC method, as it is the first step in the RAFT algorithm. Implemented the logic for term validation, log comparison, and the voting process.
- Made minor refactors to the variables, as well as the documentation.

Commit: [`16fa3c3fd6d05016a3809c06c87ac5dc41042f11`](https://github.com/EshaanAgg/dis/tree/16fa3c3fd6d05016a3809c06c87ac5dc41042f11/raft)

- Implemented the logic to start elections and request for votes using the `RequestVote` RPC.
- Added go routine to kick of elections at random delays for repeated elections.
- Implemented the logic to ensure that the leader is recognised by all the other clients based on the `AppendEntries` RPC.
- Made use of context to kill the goroutines of the elections and the heartbeats when the node is stopped, and implemented a `Shutdown` method on the nodes. Also used a ticker to improve the performance of the elections.
- Added some utils in the `Raft` structure to simulate network paritions and disconnects, and used them to initialise a basic testing suite. 
- Implemented the logic for sending periodic heartbeats, with proper logic for starting and killing the go-routine for the same. 
- Added testing infrastructure, and added some basic test cases to ensure the correctness of leader election.
- Changed the logger to support multiple levels of logging for more structured output. 
  
Commit: [`6d9c1631877d3d0717b17aae408ffca4626690dd`](https://github.com/EshaanAgg/dis/tree/6d9c1631877d3d0717b17aae408ffca4626690dd/raft)
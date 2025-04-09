# Theory

- A RAFT cluster contains several servers; 5 is a typical number which allows the system to tolerate upto 2 failures. 

- At any given time each server is in one of three states:
    - Leader 
    - Follower
    - Candidate

- In normal operation there is exactly one leader and all of the other servers are followers. 

- Followers are passive: they issue no requests on their own but simply respond to requests from leaders and candidates. The leader handles all client requests (if a client contacts a follower, the follower redirects it to the leader). The third state, candidate, is used to elect a new leader through elections.

- Raft divides time into terms of arbitrary length. Terms are numbered with consecutive integers. Each term begins with an election, in which one or more candidates attempt to become leader. 

- If a candidate wins the election, then it serves as leader for the rest of the term. In some situations an election will result in a split vote. In this case the term will end with no leader; a new term (with a new election) will be initiated soon.

- Different servers may observe the transitions between terms at different times, and in some situations a server may not observe an election or even entire terms. Terms act as a logical clock in Raft, and they allow servers to detect obsolete information such as stale leaders. 

- Each server stores a current term number, which increases monotonically over time. Current terms are exchanged whenever servers communicate; if one server’s current term is smaller than the other’s, then it updates its current term to the larger value. If a candidate or leader discovers that its term is out of date, it immediately reverts to follower state. If a server receives a request with a stale term number, it rejects the request.

- Each server has some volatile state that can be lost without harming correctness, and some persistent state which must be updated on the stable storage before responding to RPCs. 

- Raft servers communicate using remote procedure calls (RPCs), and the basic consensus algorithm requires only two types of RPCs. 
  - `RequestVote` RPCs are initiated by candidates during elections.
  - `AppendEntries` RPCs are initiated by leaders to replicate log entries and to provide a form of heartbeat. 
  - Sometimes a third RPC for transferring snapshots between servers is also used. 

- Servers retry RPCs if they do not receive a response in a timely manner, and they issue RPCs in parallel for best performance.


## Implementation steps 

1. Defined the RPC contracts in [`raft.proto`](../rpc/raft.proto) and used `gRPC` to generate the procedures for the same. 
2. Define a basic `Node` type that will represent a server in the cluster. Embed the `rpc.UnimplementedNodeServer` type in the `Node` struct to implement the RPC methods.
3. Added the initial methods to connect to all the peers in the cluster, and added a `main`/`cmd` package so that we can start the cluster via CLI.
4. Implemented a simple logger class that prints the logs to the STDOUT, but with node information and custom colours for better visibility. 

Commit: [`282770c3b97d8c681d3244ec56f80fde39d6f56a`](https://github.com/EshaanAgg/dis/commit/282770c3b97d8c681d3244ec56f80fde39d6f56a)
package main

import (
	"github.com/google/uuid"
	"log"
	"math/rand"
	"net/http"
	"net/rpc"
	"time"
)

type NodeState string

const (
	// NodeStateFollower is the state of a node that is not the leader
	Follower NodeState = "Follower"
	// NodeStateCandidate is the state of a node that is running for leader
	Candidate NodeState = "Candidate"
	// NodeStateLeader is the state of a node that is the leader
	Leader NodeState = "Leader"
)

// Peer represents a connection to another node
type Peer struct {
	Connected bool
	Address   string
}

// NewPeer creates a new Peer
func NewPeer(address string) *Peer {
	return &Peer{
		Connected: false,
		Address:   address,
	}
}

// NodeChannels is a struct that contains all channels used by a node
type NodeChannels struct {
	HeartbeatRequest  chan HeartbeatRequest
	HeartbeatResponse chan HeartbeatResponse
	VoteResponse chan VoteResponse
}

// NewNodeChannels creates a new NodeChannels
func NewNodeChannels() NodeChannels {
	return NodeChannels{
		HeartbeatRequest:  make(chan HeartbeatRequest),
		HeartbeatResponse: make(chan HeartbeatResponse),
		VoteResponse:      make(chan VoteResponse, 3),
	}
}

// Node represents a single node in the cluster
type Node struct {
	ID          uuid.UUID
	Peer_ID     int
	LeaderID    uuid.UUID
	CurrentTerm int
	VotedFor    uuid.UUID
	VotedCount  int
	State       NodeState
	Peers       map[int]*Peer
	Channels    NodeChannels
}

// NewNode creates a new node
func NewNode(peer_ID int, peers map[int]*Peer) *Node {
	return &Node{
		ID:          uuid.New(),
		Peer_ID:     peer_ID,
		LeaderID:    uuid.Nil,
		CurrentTerm: 0,
		VotedFor:    uuid.Nil,
		VotedCount:  0,
		State:       Follower,
		Peers:       peers,
		Channels:    NewNodeChannels(),
	}
}

// StepFollower is the state of a node that is not the leader
func (n *Node) stepFollower() {
	select {
	case <-n.Channels.HeartbeatRequest:
		log.Printf("Node %d [%s]: received heartbeat\n", n.Peer_ID, n.State)
	case <-time.After(time.Duration(rand.Intn(200)+300) * time.Millisecond):
		log.Printf("Node %d [%s]: timeout -> change State to Candidate\n", n.Peer_ID, n.State)
		n.State = Candidate
	}
}

// StepCandidate is the state of a node that is running for leader
func (n *Node) stepCandidate() {
	log.Printf("Node %d [%s]: I'm candidate !\n", n.Peer_ID, n.State)
	n.CurrentTerm++
	n.VotedFor = n.ID
	n.VotedCount = 1
	go n.broadcastRequestVotes()

	select {
	case res := <-n.Channels.VoteResponse:
		if res.Term > n.CurrentTerm {
			n.CurrentTerm = res.Term
			n.State = Follower
			n.VotedFor = uuid.Nil
			return
		}
		if res.VoteGranted {
			n.VotedCount++
		}
		if n.VotedCount >= len(n.Peers)/2+1 {
			log.Printf("Node %d [%s]: I'm the new leader !\n", n.Peer_ID, n.State)
			n.State = Leader
		}
	case <-time.After(time.Duration(rand.Intn(200)+300) * time.Millisecond):
		log.Printf("Node %d [%s]: timeout -> change State to Follower\n", n.Peer_ID, n.State)
		n.State = Follower
	}
}

// StepLeader is the state of a node that is the leader
func (n *Node) stepLeader() {
	select {
	case heartbeatResponse := <-n.Channels.HeartbeatResponse:
		if heartbeatResponse.Success {
			// TODO
		} else {
			// TODO
		}
	default:
		n.broadcastHeartbeat()
		time.Sleep(100 * time.Millisecond)
	}
}

func (n *Node) Step() {
	for {
		switch n.State {
		case Follower:
			n.stepFollower()
		case Candidate:
			n.stepCandidate()
		case Leader:
			n.stepLeader()
		}
	}

}

func (n *Node) Start() {
	rand.Seed(time.Now().UnixNano())
	go n.Step()
}

func (n *Node) Stop() {
	// TODO
}

func (n *Node) startRpc(port string) {
	rpc.Register(n)
	rpc.HandleHTTP()
	log.Printf("Node %d [%s] : now listening on %s\n", n.Peer_ID, n.State, port)
	go func() {
		err := http.ListenAndServe(port, nil)
		if err != nil {
			log.Fatalf("Node %d [%s] : Listen error: %s", n.Peer_ID, n.State, err)
		}
	}()
}

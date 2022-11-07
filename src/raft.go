package raft 

import (
	"github.com/google/uuid"
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

type Peer struct {
	Connected bool
	Address string
}

func NewPeer(address string) *Peer {
	return &Peer{
		Connected: false,
		Address: address,
	}
}

// Node represents a single node in the cluster
type Node struct {
	ID uuid.UUID
	LeaderID uuid.UUID
	CurrentTerm int
	VotedFor uuid.UUID
	State NodeState
	Peers map[int] * Peer
	Channels NodeChannels
}

// NewNode creates a new node
func NewNode(peers map[int] * Peer) *Node {
	return &Node{
		ID: uuid.New(),
		LeaderID: uuid.Nil,
		CurrentTerm: 0,
		VotedFor: uuid.Nil,
		State: Follower,
		Peers: peers,
	}
}

// StepFollower is the state of a node that is not the leader
func (n *Node) stepFollower() {
	select {
		// TODO handle messages
		case <-time.After(100 * time.Millisecond):
			n.State = Candidate
	}
}

// StepCandidate is the state of a node that is running for leader
func (n *Node) stepCandidate() {
	select {
		case <-time.After(100 * time.Millisecond):
			n.State = Follower
	}
}

// StepLeader is the state of a node that is the leader
func (n *Node) stepLeader() {
	// TODO
}

func (n *Node) Step() {
	switch n.State {
	case Follower:
		n.stepFollower()
	case Candidate:
		n.stepCandidate()
	case Leader:
		n.stepLeader()
	}
}
 
func (n *Node) Start() {
	n.Step()
}

func (n *Node) Stop() {
	// TODO
}
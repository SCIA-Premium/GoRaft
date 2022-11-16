package main

import (
	"log"
	"math/rand"
	"net/http"
	"net/rpc"
	"time"

	"github.com/google/uuid"
)

type NodeState string

const (
	// NodeStateDead is the state of a node that is dead
	Dead NodeState = "Dead"
	// NodeStateFollower is the state of a node that is not the leader
	Follower NodeState = "Follower"
	// NodeStateCandidate is the state of a node that is running for leader
	Candidate NodeState = "Candidate"
	// NodeStateLeader is the state of a node that is the leader
	Leader NodeState = "Leader"
)

type NodeSpeed struct {
	key   string
	value int
}

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
	AppendEntriesRequest  chan AppendEntriesRequest
	AppendEntriesResponse chan AppendEntriesResponse
	VoteResponse          chan VoteResponse
}

// NewNodeChannels creates a new NodeChannels
func NewNodeChannels() NodeChannels {
	return NodeChannels{
		AppendEntriesRequest:  make(chan AppendEntriesRequest),
		AppendEntriesResponse: make(chan AppendEntriesResponse),
		VoteResponse:          make(chan VoteResponse),
	}
}

// Node represents a single node in the cluster
type Node struct {
	PeerID int

	PeerUID       uuid.UUID
	PeerAddress   string
	LeaderUID     uuid.UUID
	LeaderAddress string

	CurrentTerm int
	VotedFor    uuid.UUID
	VotedCount  int

	commitIndex int
	lastApplied int

	nextIndex  []int
	matchIndex []int

	State    NodeState
	Peers    []*Peer
	Channels NodeChannels

	Log   []LogEntry
	Speed NodeSpeed
}

// NewNode creates a new node
func NewNode(peerID int, peer_address string, peers []*Peer) *Node {
	return &Node{
		PeerID: peerID,

		PeerUID:       uuid.New(),
		PeerAddress:   peer_address,
		LeaderUID:     uuid.Nil,
		LeaderAddress: "",

		CurrentTerm: 0,
		VotedFor:    uuid.Nil,
		VotedCount:  0,

		commitIndex: 0,
		lastApplied: 0,

		nextIndex:  make([]int, len(peers)),
		matchIndex: make([]int, len(peers)),

		State:    Follower,
		Peers:    peers,
		Channels: NewNodeChannels(),

		Log:   []LogEntry{},
		Speed: NodeSpeed{"medium", 600},
	}
}

func get_sleep_duration(n *Node) time.Duration {
	return time.Duration((rand.Intn(200)+n.Speed.value)*10) * time.Millisecond
}

// StepFollower is the state of a node that is not the leader
func (n *Node) stepFollower() {
	select {
	case req := <-n.Channels.AppendEntriesRequest:
		if len(req.Entries) == 0 {
			log.Printf("Node %d [%s]: received heartbeat\n", n.PeerID, n.State)
		}
	case <-time.After(get_sleep_duration(n)):
		log.Printf("Node %d [%s]: timeout -> change State to Candidate\n", n.PeerID, n.State)
		n.State = Candidate

		n.CurrentTerm++
		n.VotedFor = n.PeerUID
		n.VotedCount = 1
	}
}

// StepCandidate is the state of a node that is running for leader
func (n *Node) stepCandidate() {
	log.Printf("Node %d [%s]: I'm candidate !\n", n.PeerID, n.State)
	n.CurrentTerm++
	n.VotedFor = n.PeerUID
	n.VotedCount = 1
	go n.broadcastRequestVotes()

	for {
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
			if n.VotedCount >= (len(n.Peers)+1)/2+1 {
				log.Printf("Node %d [%s]: I'm the new leader !\n", n.PeerID, n.State)
				n.State = Leader

				n.LeaderUID = n.PeerUID
				n.LeaderAddress = n.PeerAddress

				for i := 0; i < len(n.Peers); i++ {
					n.nextIndex[i] = len(n.Log) + 1
					n.matchIndex[i] = 0
				}

				return
			}
		case <-time.After(get_sleep_duration(n)):
			log.Printf("Node %d [%s]: timeout -> change State to Follower\n", n.PeerID, n.State)
			n.State = Follower
		}
	}
}

// StepLeader is the state of a node that is the leader
func (n *Node) stepLeader() {
	select {
	case res := <-n.Channels.AppendEntriesResponse:
		if res.Success {
			// TODO
		} else {
			// TODO
		}
	default:
		n.broadcastAppendEntries()
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

func (n *Node) startRpc(port string) {
	rpc.Register(n)
	rpc.HandleHTTP()
	log.Printf("Node %d [%s] : now listening on %s\n", n.PeerID, n.State, port)
	go func() {
		err := http.ListenAndServe(port, nil)
		if err != nil {
			log.Fatalf("Node %d [%s] : Listen error: %s", n.PeerID, n.State, err)
		}
	}()
}

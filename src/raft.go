package main

import (
	"log"
	"math/rand"
	"net/http"
	"net/rpc"
	"strings"
	"time"

	"github.com/google/uuid"
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

type SpeedState struct {
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

	CommitIndex int
	LastApplied int

	NextIndex  []int
	MatchIndex []int

	State    NodeState
	Peers    []*Peer
	Channels NodeChannels

	SpeedState SpeedState
	Alive      bool

	Log []LogEntry

	RegisteredFiles map[uuid.UUID]string
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

		CommitIndex: 0,
		LastApplied: 0,

		NextIndex:  make([]int, len(peers)),
		MatchIndex: make([]int, len(peers)),

		State:    Follower,
		Peers:    peers,
		Channels: NewNodeChannels(),

		SpeedState: SpeedState{"medium", 600},
		Alive:      true,

		Log: []LogEntry{},

		RegisteredFiles: make(map[uuid.UUID]string),
	}
}

func get_sleep_duration(n *Node) time.Duration {
	return time.Duration((rand.Intn(200)+n.SpeedState.value)*10) * time.Millisecond
}

func (n *Node) executeCommandFollower(command string) {
	log.Printf("[T%d][%s]: executing command: %s\n", n.CurrentTerm, n.State, command)
	splited := strings.Split(command, " ")
	switch splited[0] {
	case "LOAD":
		file_uid, _ := uuid.Parse(splited[2])
		n.RegisteredFiles[file_uid] = splited[1]
	case "DELETE":
		file_uid, err := uuid.Parse(splited[1])
		if err != nil {
			return
		}
		delete(n.RegisteredFiles, file_uid)
	case "APPEND":
	}
}

// StepFollower is the state of a node that is not the leader
func (n *Node) stepFollower() {
	select {
	case req := <-n.Channels.AppendEntriesRequest:
		if len(req.Entries) == 0 {
			log.Printf("[T%d][%s]: received heartbeat\n", n.CurrentTerm, n.State)
			if req.LeaderCommit > n.CommitIndex {
				for i := n.LastApplied; i < req.LeaderCommit; i++ {
					n.Log[i].Committed = true
					n.executeCommandFollower(n.Log[i].Command)
				}
				n.LastApplied = req.LeaderCommit
				n.CommitIndex = req.LeaderCommit
			}
		} else {
			log.Printf("[T%d][%s]: received AppendEntriesRequest : %d\n", n.CurrentTerm, n.State, len(req.Entries))
		}
	case <-time.After(get_sleep_duration(n)):
		log.Printf("[T%d][%s]: timeout -> change State to Candidate\n", n.CurrentTerm, n.State)
		n.State = Candidate
	}
}

// StepCandidate is the state of a node that is running for leader
func (n *Node) stepCandidate() {
	log.Printf("[T%d][%s]: I'm candidate !\n", n.CurrentTerm, n.State)
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
				log.Printf("[T%d][%s]: I'm the new leader !\n", n.CurrentTerm, n.State)
				n.State = Leader

				n.LeaderUID = n.PeerUID
				n.LeaderAddress = n.PeerAddress

				for i := 0; i < len(n.Peers); i++ {
					n.NextIndex[i] = len(n.Log) + 1
					n.MatchIndex[i] = 0
				}

				return
			}
		case <-time.After(get_sleep_duration(n)):
			log.Printf("[T%d][%s]: timeout -> change State to Follower\n", n.CurrentTerm, n.State)
			n.State = Follower
		}
	}
}

// StepLeader is the state of a node that is the leader
func (n *Node) stepLeader() {
	select {
	case res := <-n.Channels.AppendEntriesResponse:
		log.Printf("[T%d][%s]: received AppendEntriesResponse from %d with nextIndex %d\n", n.CurrentTerm, n.State, res.NodeRelativeID, res.RequestID)
		if res.Success && res.RequestID != 0 {
			for i := n.MatchIndex[res.NodeRelativeID]; i < res.RequestID; i++ {
				n.Log[i].Count += 1
				if !n.Log[i].Committed && (n.Log[i].Count > (len(n.Peers))/2) {
					n.Log[i].Committed = true
					n.CommitIndex = i + 1
					n.LastApplied = i
				}
			}

			n.MatchIndex[res.NodeRelativeID] = res.RequestID
			n.NextIndex[res.NodeRelativeID] = res.RequestID + 1

		} else {
			log.Printf("[T%d][%s]: received a failed AppendEntriesResponse from %d\n", n.CurrentTerm, n.State, res.NodeRelativeID)
			n.NextIndex[res.NodeRelativeID] -= 1
		}
	default:
		log.Printf("[T%d][%s]: broadcast AppendEntriesRequest\n", n.CurrentTerm, n.State)
		n.broadcastAppendEntries()
		time.Sleep(1000 * time.Millisecond)
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
	log.Printf("[T%d][%s] : now listening on %s\n", n.CurrentTerm, n.State, port)
	go func() {
		err := http.ListenAndServe(port, nil)
		if err != nil {
			log.Fatalf("[T%d][%s] : Listen error: %s", n.PeerID, n.State, err)
		}
	}()
}

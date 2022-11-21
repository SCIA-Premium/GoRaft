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
	Answered  bool
}

// NewPeer creates a new Peer
func NewPeer(address string) *Peer {
	return &Peer{
		Connected: false,
		Address:   address,
		Answered:  false,
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
	Started    bool

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

		CommitIndex: -1,
		LastApplied: -1,

		NextIndex:  make([]int, len(peers)),
		MatchIndex: make([]int, len(peers)),

		State:    Follower,
		Peers:    peers,
		Channels: NewNodeChannels(),

		SpeedState: SpeedState{"medium", 600},
		Alive:      true,
		Started:    false,

		Log: []LogEntry{},

		RegisteredFiles: make(map[uuid.UUID]string),
	}
}

// StepFollower is the state of a node that is not the leader
func (n *Node) stepFollower() {
	log.Printf("[T%d][%s]: Waiting heartbeat\n", n.CurrentTerm, n.State)

	select {
	case req := <-n.Channels.AppendEntriesRequest:
		log.Printf("[T%d][%s]: Received heartbeat\n", n.CurrentTerm, n.State)

		if req.LeaderCommit >= n.CommitIndex && req.LeaderCommit != -1 {
			// Execute all commands in the log up to the leader's commit index
			for i := n.LastApplied + 1; i <= req.LeaderCommit; i++ {
				n.Log[i].Committed = true
				n.ExecuteCommand(n.Log[i].Command)
			}
			n.LastApplied = req.LeaderCommit
			n.CommitIndex = req.LeaderCommit
		}
	case <-time.After(time.Duration((rand.Intn(200)+600)*10) * time.Millisecond):
		if n.Alive {
			log.Printf("[T%d][%s]: Timeout -> Change State to Candidate\n", n.CurrentTerm, n.State)
			n.State = Candidate
		}
	}
}

// StepCandidate is the state of a node that is running for leader
func (n *Node) stepCandidate() {
	log.Printf("[T%d][%s]: Starting Leader Election\n", n.CurrentTerm, n.State)
	log.Printf("[T%d][%s]: I'm candidate !\n", n.CurrentTerm, n.State)

	n.CurrentTerm++
	n.VotedFor = n.PeerUID
	n.VotedCount = 1

	for _, peer := range n.Peers {
		peer.Answered = false
	}

	go func() {
		for {
			select {
			case res := <-n.Channels.VoteResponse:
				if res.Term == -2 {
					return
				}

				// If the response is later than the current term, update the current term and change state to follower
				if res.Term > n.CurrentTerm {
					n.CurrentTerm = res.Term
					n.State = Follower
					n.VotedFor = uuid.Nil
					return
				}

				if n.Peers[res.NodeRelativeID].Answered {
					continue
				}

				n.Peers[res.NodeRelativeID].Answered = true

				if !res.VoteGranted {
					continue
				}

				n.VotedCount++

				// If the node has received a majority of votes, it becomes the leader
				if n.VotedCount < (len(n.Peers)+1)/2+1 {
					continue
				}

				log.Printf("[T%d][%s]: I'm the new leader !\n", n.CurrentTerm, n.State)
				n.State = Leader

				n.LeaderUID = n.PeerUID
				n.LeaderAddress = n.PeerAddress
				n.VotedFor = uuid.Nil

				// Update the next index and match index of all peers
				for i := 0; i < len(n.Peers); i++ {
					n.NextIndex[i] = n.LastApplied + 1
					n.MatchIndex[i] = n.LastApplied
				}

				return
			}
		}
	}()

	// Election timeout
	// 10 * 100 * time.Millisecond = 1 second
	for i := 0; i < 10; i++ {
		n.broadcastRequestVotes()
		time.Sleep(100 * time.Millisecond)

		if n.State != Candidate {
			break
		}
	}

	if n.State == Candidate {
		var stopResponse = VoteResponse{
			Term: -2,
		}
		n.Channels.VoteResponse <- stopResponse
	}
}

// StepLeader is the state of a node that is the leader
func (n *Node) stepLeader() {

	// Send heartbeat to all peers
	n.broadcastAppendEntries()

	go func(n *Node) {
		for {
			select {
			case res := <-n.Channels.AppendEntriesResponse:
				if res.Term == -2 {
					return
				}

				n.Started = n.Started || res.Started

				if res.Success {
					for i := n.MatchIndex[res.NodeRelativeID] + 1; i < res.NodeRelativeNextIndex; i++ {
						log.Printf("[T%d][%s]: Received a heartbeat answer from %d with up-to-date log %d\n", n.CurrentTerm, n.State, res.NodeRelativeID, i)
						n.Log[i].Count += 1
						if !n.Log[i].Committed && (n.Log[i].Count >= (len(n.Peers))/2+1) {
							log.Printf("[T%d][%s]: Committing log %d\n", n.CurrentTerm, n.State, i)
							n.Log[i].Committed = true
							n.CommitIndex = i
							n.LastApplied = i
						}
					}

					n.MatchIndex[res.NodeRelativeID] = res.NodeRelativeNextIndex - 1
					n.NextIndex[res.NodeRelativeID] = res.NodeRelativeNextIndex
					continue
				}

				log.Printf("[T%d][%s]: Received a failed heartbeat from %d\n", n.CurrentTerm, n.State, res.NodeRelativeID)

				if res.Term > n.CurrentTerm {
					log.Printf("[T%d][%s]: Term has changed to term %d -> Change state to Follower\n", n.CurrentTerm, n.State, res.Term)
					n.CurrentTerm = res.Term
					n.State = Follower
					n.VotedFor = uuid.Nil
					return
				}

				n.NextIndex[res.NodeRelativeID] -= 1
				if n.NextIndex[res.NodeRelativeID] < 0 {
					n.NextIndex[res.NodeRelativeID] = 0
				}
			default:
				if !n.Alive {
					return
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}(n)

	// Leader timer for next heartbeat
	// 20 * 50 * timeMillisecond = 1 second
	for i := 0; i < 20; i++ {
		time.Sleep(50 * time.Millisecond)
		if n.Alive && n.State != Leader {
			break
		}
	}

	if n.Alive && n.State == Leader {
		var stopResponse AppendEntriesResponse
		stopResponse.Term = -2
		n.Channels.AppendEntriesResponse <- stopResponse
	}
}

// Step is the main function of the node
func (n *Node) Step() {
	for {
		if n.Alive {
			switch n.State {
			case Follower:
				n.stepFollower()
			case Candidate:
				n.stepCandidate()
			case Leader:
				n.stepLeader()
			}
		} else {
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// Start starts the node and set random seed
func (n *Node) Start() {
	rand.Seed(time.Now().UnixNano())
	go n.Step()
}

// startRpc starts the RPC server and register the node
func (n *Node) startRpc(port string) {
	rpc.Register(n)
	rpc.HandleHTTP()
	log.Printf("[T%d][%s]: Now listening on port %s\n", n.CurrentTerm, n.State, port)
	go func() {
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			log.Fatalf("[T%d][%s]: Listen error: %s", n.PeerID, n.State, err)
		}
	}()
}

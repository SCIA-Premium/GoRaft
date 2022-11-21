package main

import (
	"errors"
	"log"
	"net/rpc"
	"time"

	"github.com/google/uuid"
)

// LogEntry represents a single entry in the log
type LogEntry struct {
	Term    int
	Index   int
	Command string

	Count     int  `default: 0`
	Committed bool `default: false`
}

// AppendEntriesRequest is the request sent to append entries to the log
type AppendEntriesRequest struct {
	Term          int
	LeaderUID     uuid.UUID
	LeaderAddress string
	PrevLogIndex  int
	PrevLogTerm   int
	Entries       []LogEntry
	LeaderCommit  int

	Started bool
}

// AppendEntriesResponse is the response sent after appending entries to the log
type AppendEntriesResponse struct {
	NodeRelativeNextIndex int
	NodeRelativeID        int
	Term                  int
	Success               bool

	Started bool
}

// VoteRequest is the request sent to vote for a candidate
type VoteRequest struct {
	Term         int
	CandidateID  uuid.UUID
	LastLogIndex int
	LastLogTerm  int
}

// VoteResponse is the response sent after voting for a candidate
type VoteResponse struct {
	NodeRelativeID int
	Term           int
	VoteGranted    bool
}

func (n *Node) waitComputeTime() {
	time.Sleep(time.Duration(n.SpeedState.value) * time.Millisecond)
}

// RequestVotes is the RPC method to request votes
func (n *Node) RequestVotes(req VoteRequest, res *VoteResponse) error {
	if !n.Alive {
		return errors.New("Node is not alive")
	}

	res.VoteGranted = false
	res.Term = n.CurrentTerm

	if req.Term < n.CurrentTerm || n.State == Leader && req.Term == n.CurrentTerm {
		return nil
	}

	if (n.VotedFor == uuid.Nil || n.VotedFor == req.CandidateID) &&
		(len(n.Log) == 0 || req.LastLogTerm > n.Log[len(n.Log)-1].Term || (req.LastLogTerm == n.Log[len(n.Log)-1].Term && req.LastLogIndex >= len(n.Log)-1)) {
		n.VotedFor = req.CandidateID
		res.VoteGranted = true
	}

	return nil
}

// broadcastVoteRequest sends a vote request to all peers
func (n *Node) broadcastRequestVotes() {
	req := VoteRequest{
		Term:         n.CurrentTerm,
		CandidateID:  n.PeerUID,
		LastLogIndex: len(n.Log) - 1,
		LastLogTerm:  -1,
	}

	if len(n.Log) > 0 {
		req.LastLogTerm = n.Log[req.LastLogIndex].Term
	}

	for i, peer := range n.Peers {
		if n.Peers[i].Answered {
			continue
		}

		log.Printf("[T%d][%s]: Requesting vote from %s\n", n.CurrentTerm, n.State, peer.Address)
		go func(i int, peer *Peer) {
			client, err := rpc.DialHTTP("tcp", peer.Address)
			if err != nil {
				log.Println(err)
				return
			}

			var res VoteResponse
			res.NodeRelativeID = i
			err = client.Call("Node.RequestVotes", req, &res)
			if err != nil {
				if err.Error() != "Node is not alive" {
					log.Println(err)
				}
				return
			}
			n.Channels.VoteResponse <- res
		}(i, peer)
	}
}

// AppendEntries is the RPC method to append entries to the log
func (n *Node) AppendEntries(req AppendEntriesRequest, res *AppendEntriesResponse) error {
	n.waitComputeTime()

	if !n.Alive {
		return errors.New("Node is not alive")
	}

	n.Started = req.Started || n.Started
	res.Started = n.Started

	if req.Term < n.CurrentTerm {
		res.Term = n.CurrentTerm
		res.Success = false
		return nil
	}

	if req.PrevLogIndex != -1 && len(n.Log) > req.PrevLogIndex && n.Log[req.PrevLogIndex].Term != req.PrevLogTerm {
		log.Printf("[T%d][%s]: Erasing bad logs\n", n.CurrentTerm, n.State)
		if req.PrevLogIndex == -1 {
			n.Log = []LogEntry{}
		} else {
			n.Log = n.Log[:req.PrevLogIndex]
		}
		res.Term = n.CurrentTerm
		res.Success = false
		return nil
	}

	if req.Term > n.CurrentTerm {
		log.Printf("[T%d][%s]: Term has changed to term %d -> Change state to Follower\n", n.CurrentTerm, n.State, req.Term)
		n.State = Follower
		n.CurrentTerm = req.Term
		n.VotedFor = uuid.Nil

		if req.LeaderCommit == -1 {
			n.Log = make([]LogEntry, 0)
		} else if req.LeaderCommit < n.CommitIndex {
			n.Log = n.Log[:req.LeaderCommit]
		}
	}

	n.VotedFor = uuid.Nil
	n.LeaderUID = req.LeaderUID
	n.LeaderAddress = req.LeaderAddress
	n.State = Follower

	res.Term = n.CurrentTerm
	res.Success = true

	if len(req.Entries) != 0 {
		log.Printf("[T%d][%s]: Receiving %d entries\n", n.CurrentTerm, n.State, len(req.Entries))
		if req.PrevLogIndex == -1 {
			n.Log = req.Entries
		} else {
			to_not_append := len(n.Log) - 1 - req.PrevLogIndex
			log.Printf("[T%d][%s]: Appending %d entries\n", n.CurrentTerm, n.State, len(req.Entries)-to_not_append)
			if to_not_append != len(req.Entries) {
				n.Log = append(n.Log, req.Entries[to_not_append:]...)
			}
		}

		n.CommitIndex = len(n.Log) - 1
		if req.LeaderCommit > n.CommitIndex {
			n.CommitIndex = req.LeaderCommit
		}
	}

	res.NodeRelativeNextIndex = len(n.Log)

	n.Channels.AppendEntriesRequest <- req

	return nil
}

// broadCastAppendEntries sends an append entries request to all peers
func (n *Node) broadcastAppendEntries() {
	log.Printf("[T%d][%s]: broadcasting\n", n.CurrentTerm, n.State)
	for i, peer := range n.Peers {
		go func(peer *Peer, i int) {
			client, err := rpc.DialHTTP("tcp", peer.Address)
			if err != nil {
				log.Println(err)
				return
			}

			req := AppendEntriesRequest{
				Term:          n.CurrentTerm,
				LeaderUID:     n.PeerUID,
				LeaderAddress: n.PeerAddress,
				LeaderCommit:  n.CommitIndex,

				Started: n.Started,
			}

			if len(n.Log) == 0 {
				req.PrevLogIndex = 0
				req.PrevLogTerm = 0
			} else {
				if n.NextIndex[i] == 0 {
					req.PrevLogIndex = -1
					req.PrevLogTerm = -1
					req.Entries = n.Log
				} else {
					req.PrevLogIndex = n.NextIndex[i] - 1
					req.PrevLogTerm = n.Log[req.PrevLogIndex].Term
					if n.NextIndex[i] < len(n.Log) {
						req.Entries = n.Log[n.NextIndex[i]:]
					}
				}
			}

			var res AppendEntriesResponse
			res.NodeRelativeID = i
			err = client.Call("Node.AppendEntries", req, &res)
			if err != nil {
				if err.Error() != "Node is not alive" {
					log.Println(err)
				}
				return
			}

			n.Channels.AppendEntriesResponse <- res
		}(peer, i)
	}
}

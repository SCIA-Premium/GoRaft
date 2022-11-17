package main

import (
	"errors"
	"log"
	"net/rpc"

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
}

// AppendEntriesResponse is the response sent after appending entries to the log
type AppendEntriesResponse struct {
	RequestID      int
	NodeRelativeID int
	Term           int
	Success        bool
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
	Term        int
	VoteGranted bool
}

// RequestVotes is the RPC method to request votes
func (n *Node) RequestVotes(req VoteRequest, res *VoteResponse) error {
	if req.Term < n.CurrentTerm {
		res.Term = n.CurrentTerm
		res.VoteGranted = false
		return nil
	}

	if (n.VotedFor == uuid.Nil || n.VotedFor == req.CandidateID) && req.Term >= n.CurrentTerm {
		n.CurrentTerm = req.Term
		n.VotedFor = req.CandidateID
		res.Term = req.Term
		res.VoteGranted = true
	}

	return nil
}

// broadcastVoteRequest sends a vote request to all peers
func (n *Node) broadcastRequestVotes() {
	req := VoteRequest{
		Term:        n.CurrentTerm,
		CandidateID: n.PeerUID,
	}
	log.Printf("[T%d][%s]: Starting Leader Election\n", n.CurrentTerm, n.State)
	for _, peer := range n.Peers {
		log.Printf("[T%d][%s]: Requesting vote from %s\n", n.CurrentTerm, n.State, peer.Address)
		go func(peer *Peer) {
			client, err := rpc.DialHTTP("tcp", peer.Address)
			if err != nil {
				log.Println(err)
				return
			}

			var res VoteResponse
			err = client.Call("Node.RequestVotes", req, &res)
			if err != nil {
				if err.Error() != "Node is not alive" {
					log.Println(err)
				}
				return
			}
			n.Channels.VoteResponse <- res
		}(peer)
	}
}

// AppendEntries is the RPC method to append entries to the log
func (n *Node) AppendEntries(req AppendEntriesRequest, res *AppendEntriesResponse) error {
	if !n.Alive {
		return errors.New("Node is not alive")
	}

	if req.Term < n.CurrentTerm {
		res.Term = n.CurrentTerm
		res.Success = false
		return nil
	}

	if len(n.Log) > req.PrevLogIndex && n.Log[req.PrevLogIndex].Term != req.PrevLogTerm {
		res.Term = n.CurrentTerm
		res.Success = false
		return nil
	}

	if req.Term > n.CurrentTerm {
		log.Printf("[T%d][%s]: term has changed to term %d -> Change state to Follower\n", n.CurrentTerm, n.State, req.Term)
		n.State = Follower
		n.CurrentTerm = req.Term
		n.VotedFor = uuid.Nil

		if req.LeaderCommit == 0 {
			n.Log = make([]LogEntry, 0)
		} else {
			n.Log = n.Log[:req.LeaderCommit]
		}
	}

	n.VotedFor = uuid.Nil
	n.LeaderUID = req.LeaderUID
	n.LeaderAddress = req.LeaderAddress
	n.State = Follower

	res.Term = n.CurrentTerm
	res.Success = true

	if len(req.Entries) == 0 {
		res.RequestID = len(n.Log)
	} else {
		log.Printf("[T%d][%s]: len(req.Entries) %d and len(n.log) %d\n", n.CurrentTerm, n.State, len(req.Entries), len(n.Log))
		n.Log = append(n.Log, req.Entries...)
		res.RequestID = len(n.Log)

		n.CommitIndex = len(n.Log) - 1
		if req.LeaderCommit > n.CommitIndex {
			n.CommitIndex = req.LeaderCommit
		}
	}

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
			}

			if len(n.Log) == 0 {
				req.PrevLogIndex = 0
				req.PrevLogTerm = 0
			} else {
				if n.NextIndex[i] == 0 {
					req.PrevLogIndex = 0
					req.PrevLogTerm = 0
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

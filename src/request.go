package main

import (
	"log"
	"net/rpc"
	"github.com/google/uuid"
)

// LogEntry represents a single entry in the log
type LogEntry struct {
	Term    int
	Index   int
	Command string
}

// AppendEntriesRequest is the request sent to append entries to the log
type AppendEntriesRequest struct {
	Term         int
	LeaderID     uuid.UUID
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

// AppendEntriesResponse is the response sent after appending entries to the log
type AppendEntriesResponse struct {
	Term    int
	Success bool
}

// VoteRequest is the request sent to vote for a candidate
type VoteRequest struct {
	Term        int
	CandidateID uuid.UUID
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

	if n.VotedFor == uuid.Nil {
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
		CandidateID: n.ID,
	}
	log.Printf("Node %d [%s]: Starting Leader Election\n", n.Peer_ID, n.State)
	for _, peer := range n.Peers {
		log.Printf("Node %d [%s]: Requesting vote from %s", n.Peer_ID, n.State, peer.Address)
		go func(peer *Peer) {
			client, err := rpc.DialHTTP("tcp", peer.Address)
			if err != nil {
				log.Println(err)
				return
			}

			var res VoteResponse
			err = client.Call("Node.RequestVotes", req, &res)
			if err != nil {
				log.Println(err)
				return
			}
			n.Channels.VoteResponse <- res
		}(peer)
	}
}

// AppendEntries is the RPC method to append entries to the log
func (n *Node) AppendEntries(req AppendEntriesRequest, res* AppendEntriesResponse) error {

	if req.Term < n.CurrentTerm {
		res.Term = n.CurrentTerm
		res.Success = false
		return nil
	}

	n.Channels.AppendEntriesRequest <- req
	if len(req.Entries) == 0 {
		res.Term = n.CurrentTerm
		res.Success = true
		return nil
	}

	// TODO : add entries to log

	res.Success = true
	res.Term = n.CurrentTerm
	return nil
}

// broadCastAppendEntries sends an append entries request to all peers
func (n *Node) broadcastAppendEntries() {
	req := AppendEntriesRequest{
		Term:     n.CurrentTerm,
		LeaderID: n.ID,
	}
	for _, peer := range n.Peers {
		go func(peer *Peer) {
			client, err := rpc.DialHTTP("tcp", peer.Address)
			if err != nil {
				log.Println(err)
				return
			}

			var res AppendEntriesResponse
			err = client.Call("Node.AppendEntries", req, &res)
			if err != nil {
				log.Println(err)
				return
			}
			n.Channels.AppendEntriesResponse <- res
		}(peer)
	}
}

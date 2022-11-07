package raft

import (
	"github.com/google/uuid"
)

// LogEntry represents a single entry in the log
type LogEntry struct {
	Term int
	Index int
	Command string
}

type NodeChannels struct {
	AppendEntriesRequest chan AppendEntriesRequest
	AppendEntriesResponse chan AppendEntriesResponse
	RequestVoteRequest chan RequestVoteRequest
	RequestVoteResponse chan RequestVoteResponse
}

// AppendEntriesRequest is the request sent to append entries to the log
type AppendEntriesRequest struct {
	Term int
	LeaderID uuid.UUID
	PrevLogIndex int
	PrevLogTerm int
	Entries []LogEntry
	LeaderCommit int
}

// AppendEntriesResponse is the response sent after appending entries to the log
type AppendEntriesResponse struct {
	Term int
	Success bool
}

// RequestVoteRequest is the request sent to vote for a candidate
type RequestVoteRequest struct {
	Term int
	CandidateID uuid.UUID
	LastLogIndex int
	LastLogTerm int
}

// RequestVoteResponse is the response sent after voting for a candidate
type RequestVoteResponse struct {
	Term int
	VoteGranted bool
}

func (n *Node) broadcastAppendEntriesRequest() {
	// TODO
}
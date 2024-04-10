package core

import (
	"fmt"
	"strconv"

	miniraft "github.com/lsig/Raft/server/pb"
)

func (s *Server) CreateVoteRequest() *miniraft.RequestVoteRequest {
	lastIndex := len(s.Raft.Log) - 1
	cn := strconv.Itoa(s.Info.Id)

	if lastIndex == -1 {
		details := &miniraft.RequestVoteRequest{
			Term:          s.Raft.CurrentTerm,
			CandidateName: cn,
			LastLogIndex:  uint64(0),
			LastLogTerm:   0,
		}
		return details
	} else {
		details := &miniraft.RequestVoteRequest{
			Term:          s.Raft.CurrentTerm,
			CandidateName: cn,
			LastLogIndex:  uint64(lastIndex),
			LastLogTerm:   s.Raft.Log[lastIndex].Term,
		}
		return details
	}
}

func (s *Server) ChangeState(newState State) {
	var newStateName string
	switch newState {
	case Candidate:
		newStateName = "candidate"
	case Failed:
		newStateName = "failed"
	case Follower:
		newStateName = "follower"
	case Leader:
		newStateName = "leader"
	}

	// only announce the new state if its different than the old
	if newState != s.State {
		fmt.Printf("New state: %v\n", newStateName)
	}

	s.State = newState
}

func (s *Server) UpdateTerm(newTerm uint64, newVote int) {
	// update current term
	s.Raft.CurrentTerm = newTerm

	s.Raft.VotedFor = newVote

	// This might only be useful if the server is the leader
	for idx := range s.Info.Address {
		s.Raft.NextIndex[idx] = len(s.Raft.Log) + 1
	}
}

package core

import (
	"fmt"

	miniraft "github.com/lsig/Raft/server/pb"
)

func (s *Server) CreateVoteRequest() *miniraft.RequestVoteRequest {
	lastIndex := len(s.Log) - 1

	if lastIndex == -1 {
		details := &miniraft.RequestVoteRequest{
			Term:          s.CurrentTerm,
			CandidateName: s.Nodes[s.Address],
			LastLogIndex:  uint64(0),
			LastLogTerm:   0,
		}
		return details
	} else {
		details := &miniraft.RequestVoteRequest{
			Term:          s.CurrentTerm,
			CandidateName: s.Nodes[s.Address],
			LastLogIndex:  uint64(lastIndex),
			LastLogTerm:   s.Log[lastIndex].Term,
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

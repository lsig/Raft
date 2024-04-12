package core

import (
	"fmt"
	"strconv"

	miniraft "github.com/lsig/Raft/server/pb"
	"github.com/lsig/Raft/server/util"
)

func (s *Server) CreateVoteRequest() *miniraft.RequestVoteRequest {
	lastIndex := len(s.Raft.Logs) - 1
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
			LastLogTerm:   s.Raft.Logs[lastIndex].Term,
		}
		return details
	}
}

func (s *Server) ChangeState(newState State) {
	// var newStateName string
	// switch newState {
	// case Candidate:
	// 	newStateName = "candidate"
	// case Failed:
	// 	newStateName = "failed"
	// case Follower:
	// 	newStateName = "follower"
	// case Leader:
	// 	newStateName = "leader"
	// }

	// // only announce the new state if its different than the old
	// if newState != s.State {
	// 	fmt.Printf("New state: %v\n", newStateName)
	// }

	s.State = newState
}

func (s *Server) UpdateTerm(newTerm uint64, newVote int) {
	// updating terms means resetting timer
	s.Timer.Reset(util.GetRandomTimeout())

	// update current term
	s.Raft.CurrentTerm = newTerm

	s.Raft.VotedFor = newVote
	s.Raft.LeaderId = -1 // we don't have a leader at this point

	// This might only be useful if the server is the leader
	for idx := range s.Nodes.Addresses {
		s.Raft.NextIndex[idx] = len(s.Raft.Logs)
	}
}

func (s *Server) isCandidateValid(message *miniraft.RequestVoteRequest) (bool, error) {
	// Candidate's term can't be lower than ours
	if message.Term < s.Raft.CurrentTerm {
		return false, fmt.Errorf("candidate's term is lower")
	}

	// We can only vote for one candidate per term
	if message.Term == s.Raft.CurrentTerm {
		// We can resend the validation to the correct candidate if it didn't arrive
		if message.CandidateName == fmt.Sprint(s.Raft.VotedFor) {
			return true, nil
		} else {
			return false, fmt.Errorf("already voted this term")
		}
	}

	// candidate can't have fewer logs than us
	if len(s.Raft.Logs)-1 > int(message.LastLogIndex) {
		return false, fmt.Errorf("candidate is missing logs")
	}

	// candidate can't have its oldest log be less than ours
	var lastLogTerm uint64 = 0
	if len(s.Raft.Logs) > 0 {
		lastLogTerm = uint64(s.Raft.Logs[len(s.Raft.Logs)-1].Term)
	}
	if lastLogTerm > message.LastLogTerm {
		return false, fmt.Errorf("candidate has older logs")
	}

	return true, nil
}

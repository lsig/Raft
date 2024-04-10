package core

import (
	"fmt"
	"strconv"

	miniraft "github.com/lsig/Raft/server/pb"
	"github.com/lsig/Raft/server/util"
)

func (s *Server) HandleClientCommand(address string, cmd string) {
	fmt.Printf("Received command: %s", cmd)

	msg := &miniraft.Raft_CommandName{
		CommandName: fmt.Sprintf("Thank you for the command: %s", cmd),
	}

	packet := &miniraft.Raft{
		Message: msg,
	}

	s.SendMessage(address, packet)
}

func (s *Server) HandleVoteRequest(address string, message *miniraft.RequestVoteRequest) {
	fmt.Printf("received RVR - ")

	var granted bool
	// Only grant votes if the received term is new
	// or if the original term winner resends a request
	if fmt.Sprint(s.Raft.VotedFor) != message.CandidateName && s.Raft.CurrentTerm >= message.Term {
		fmt.Printf("Already voted to %v...\n", s.Raft.VotedFor)
		granted = false
	} else {
		// TODO check whether requesting candiate is up-to-date

		// Become a follower
		if s.State == Leader || s.State == Candidate {
			s.State = Follower
		}

		// this term is strictly higher than the server's current term, due to the check above
		newTerm := message.Term
		newVote, _ := strconv.Atoi(message.CandidateName)
		s.Timer.Reset(util.GetRandomTimeout())

		s.UpdateTerm(newTerm, newVote)
		s.Raft.VotedFor = newVote
		fmt.Printf("Entered term %d! Voted casted to %v!\n", newTerm, newVote)
		granted = true
	}

	s.sendVoteResponse(address, granted)
}

func (s *Server) HandleVoteResponse(address string, message *miniraft.RequestVoteResponse) {
	serverIndex := util.FindServerId(s.Nodes.Addresses, address)

	if message.VoteGranted {
		fmt.Printf("Vote granted by %s\n", address)

		s.Raft.Votes[s.Raft.CurrentTerm][serverIndex] = 1

		if util.ReceivedMajorityVotes(s.Raft.Votes[s.Raft.CurrentTerm]) {
			s.AnnounceLeadership()
		}

	} else {
		// TODO check whether the responding server's Term is higher...
		// if so, we must update the term and restart the timeout
		fmt.Printf("Vote NOT granted by %s\n", address)
		s.Raft.Votes[s.Raft.CurrentTerm][serverIndex] = -1
	}

	fmt.Printf("my votes: %v\n", s.Raft.Votes[s.Raft.CurrentTerm])

}

func (s *Server) HandleAppendEntriesRequest(address string, message *miniraft.AppendEntriesRequest) {
	s.TimeoutReset <- struct{}{}
	fmt.Printf("Received AER from: %s\n", address)

}

func (s *Server) sendVoteResponse(address string, granted bool) {

	message := &miniraft.Raft{Message: &miniraft.Raft_RequestVoteResponse{
		RequestVoteResponse: &miniraft.RequestVoteResponse{
			Term:        s.Raft.CurrentTerm,
			VoteGranted: granted,
		},
	}}

	fmt.Printf("sending VoteResponse to %s\n", address)
	s.SendMessage(address, message)
}

func (s *Server) HandleLogCommand() {
	for _, log := range s.Raft.Logs {
		fmt.Println(log.String())
	}
}

func (s *Server) HandlePrintCommand() {
	fmt.Printf("CurrentTerm: %d\n", s.Raft.CurrentTerm)
	fmt.Printf("VotedFor: %d\n", s.Raft.VotedFor)
	fmt.Printf("CommitIndex: %d\n", s.Raft.CommitIndex)
	fmt.Printf("LastAppliedIndex: %d\n", s.Raft.LastApplied)
	fmt.Printf("CurrentTerm: %d\n", s.Raft.CurrentTerm)
	fmt.Printf("NextIndex: %v\n", s.Raft.NextIndex)
	fmt.Printf("MatchIndex: %v\n", s.Raft.NextIndex)
}

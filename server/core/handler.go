package core

import (
	"fmt"
	"strconv"

	miniraft "github.com/lsig/Raft/server/pb"
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

func (s *Server) HandleVoteRequest(address string, message *miniraft.Raft_RequestVoteRequest) {
	fmt.Printf("received RVR - ")

	var granted bool
	if s.VotedFor != -1 && s.CurrentTerm >= message.RequestVoteRequest.Term {
		fmt.Printf("Already voted to %v...\n", s.VotedFor)
		granted = false
	} else {
		// TODO check whether requesting candiate is up-to-date
		// TODO check whether requesting candiate's term is new?

		// this term is strictly higher than the server's current term, due to the check above
		newTerm := message.RequestVoteRequest.Term
		newVote, _ := strconv.Atoi(message.RequestVoteRequest.CandidateName)
		s.CurrentTerm = newTerm
		s.VotedFor = newVote
		fmt.Printf("Entered term %d! Voted casted to %v!\n", newTerm, newVote)
		granted = true
	}

	s.SendVoteResponse(address, granted)
}

func (s *Server) SendVoteResponse(address string, granted bool) {

	message := &miniraft.Raft{Message: &miniraft.Raft_RequestVoteResponse{
		RequestVoteResponse: &miniraft.RequestVoteResponse{
			Term:        s.CurrentTerm,
			VoteGranted: true,
		},
	}}

	s.SendMessage(address, message)
}

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
	if s.VotedFor != -1 && s.CurrentTerm >= message.Term {
		fmt.Printf("Already voted to %v...\n", s.VotedFor)
		granted = false
	} else {
		// TODO check whether requesting candiate is up-to-date
		// TODO check whether requesting candiate's term is new?

		// this term is strictly higher than the server's current term, due to the check above
		newTerm := message.Term
		newVote, _ := strconv.Atoi(message.CandidateName)
		s.Timer.Reset(util.GetRandomTimeout())
		s.CurrentTerm = newTerm
		s.VotedFor = newVote
		fmt.Printf("Entered term %d! Voted casted to %v!\n", newTerm, newVote)
		granted = true
	}

	s.sendVoteResponse(address, granted)
}

func (s *Server) HandleVoteResponse(address string, message *miniraft.RequestVoteResponse) {

}

func (s *Server) sendVoteResponse(address string, granted bool) {

	message := &miniraft.Raft{Message: &miniraft.Raft_RequestVoteResponse{
		RequestVoteResponse: &miniraft.RequestVoteResponse{
			Term:        s.CurrentTerm,
			VoteGranted: granted,
		},
	}}

	s.SendMessage(address, message)
}

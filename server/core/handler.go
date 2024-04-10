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
	if s.VotedFor != -1 {
		fmt.Printf("Already voted to %v...\n", s.VotedFor)
		granted = false
	} else {
		// TODO check whether requesting candiate is up-to-date
		// TODO check whether requesting candiate's term is new?

		vote, _ := strconv.Atoi(message.RequestVoteRequest.CandidateName)
		s.VotedFor = vote
		fmt.Printf("Voted casted to %v!\n", vote)
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

package core

import (
	"fmt"
	"strconv"

	miniraft "github.com/lsig/Raft/server/pb"
	"github.com/lsig/Raft/server/util"
)

func (s *Server) HandleClientCommand(address string, cmd string) {
	fmt.Printf("Received command: %s\n", cmd)

	if s.State == Leader {
		// Command is either from server or client
		newLog := Log{Term: s.Raft.CurrentTerm, Index: len(s.Raft.Logs), Command: cmd}
		s.Raft.Logs = append(s.Raft.Logs, newLog)
		s.Raft.MatchIndex[s.Info.Id]++ // Increment our server's matchIndex

	} else {
		if s.Raft.LeaderId == -1 {
			fmt.Printf("No leader to send command to, aborting...\n")
			return
		}

		// Command must be from client, forward to leader
		leaderAddress := util.FindLeaderAddress(s.Nodes.Addresses, s.Raft.LeaderId)
		packet := &miniraft.Raft{Message: &miniraft.Raft_CommandName{CommandName: cmd}}

		s.SendMessage(leaderAddress, packet)
	}
}

func (s *Server) HandleVoteRequest(address string, message *miniraft.RequestVoteRequest) {
	fmt.Printf("received RVR - ")

	var granted bool
	// Only grant votes if the received term is new
	// or if the original term winner resends a request
	if fmt.Sprint(s.Raft.VotedFor) != message.CandidateName && s.Raft.CurrentTerm >= message.Term {
		fmt.Printf("Already voted to %v...\n", s.Raft.VotedFor)
		granted = false
	} else if fmt.Sprint(s.Raft.VotedFor) != message.CandidateName && len(s.Raft.Logs) > 0 && s.Raft.Logs[len(s.Raft.Logs)-1].Term > message.LastLogTerm {
		fmt.Printf("Already voted to %v...\n", s.Raft.VotedFor)
		granted = false
	} else if fmt.Sprint(s.Raft.VotedFor) != message.CandidateName && len(s.Raft.Logs)-1 > int(message.LastLogIndex) {
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
	// fmt.Printf("my votes: %v\n", s.Raft.Votes[s.Raft.CurrentTerm])
}

func (s *Server) HandleAppendEntriesRequest(address string, message *miniraft.AppendEntriesRequest) {
	s.TimeoutReset <- struct{}{}
	lId, _ := strconv.Atoi(message.LeaderId)
	s.Raft.LeaderId = lId

	// if len(message.Entries) > 0 {
	// 	fmt.Printf("leader msg: %v\n", message)
	// }

	// If the requesting leader's term is behind ours,
	if message.Term < s.Raft.CurrentTerm {
		fmt.Println("denying AER, leader's term is behind")
		s.sendAppendEntriesRes(address, false)
		return
	}

	// If my logs are more than the acceptable 1 behind the leader,
	if max(len(s.Raft.Logs)-1, 0) < int(message.PrevLogIndex) {
		fmt.Println("denying AER, my logs are too far behind")
		s.sendAppendEntriesRes(address, false)
		return
	}

	// if my last log's term is different from the leader's,
	var lastLogsTerm uint64
	if len(s.Raft.Logs) > int(message.PrevLogIndex) {
		lastLogsTerm = s.Raft.Logs[message.PrevLogIndex].Term
	}
	if lastLogsTerm != message.PrevLogTerm {
		fmt.Printf("denying AER, my last log's term (%v), is different than the leader's (%v)\n", lastLogsTerm, message.PrevLogTerm)
		s.sendAppendEntriesRes(address, false)
		return
	}

	// We have a successful request, and will respond successfully

	// copy the leader's commit index (-1 because uints)
	s.Raft.CommitIndex = int(message.LeaderCommit) - 1

	for _, entry := range message.Entries {
		fmt.Printf("received log: %v\n", entry)
		log := Log{}
		s.Raft.Logs = append(s.Raft.Logs, log.FromLogEntry(entry))
	}
	s.sendAppendEntriesRes(address, true)
}

func (s *Server) HandleAppendEntriesResponse(address string, message *miniraft.AppendEntriesResponse) {
	fmt.Printf("\nreceived AE-Response:\nSuccess: %v\nTerm: %v\n", message.Success, message.Term)

	sId := util.FindServerId(s.Nodes.Addresses, address)

	serverNextIndex := s.Raft.NextIndex[sId]
	// serverMatchIndex := s.Raft.MatchIndex[sId]
	logsLen := len(s.Raft.Logs)

	if message.Success {
		// Check whether we have to increment the responding server's nextIndex and matchIndex.

		if logsLen > serverNextIndex {
			// we had a log to send to the server, which it accepted
			s.Raft.NextIndex[sId]++
		}

		s.Raft.MatchIndex[sId] = s.Raft.NextIndex[sId] - 1
		s.checkCommits()

	} else {
		s.Raft.NextIndex[sId]--
		// Send another request with a decremented prevIndex and prevTerm.
	}
}

func (s *Server) checkCommits() {
	nextCommitIndex := s.Raft.CommitIndex + 1

	matched := 0
	total := len(s.Raft.MatchIndex)

	for _, matchIndex := range s.Raft.MatchIndex {
		if matchIndex >= nextCommitIndex {
			matched++
		}
	}

	// If the majority have matched the index
	if matched > total/2 {
		s.Raft.CommitIndex = nextCommitIndex
	}
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

func (s *Server) sendAppendEntriesRes(address string, success bool) {
	message := &miniraft.Raft{Message: &miniraft.Raft_AppendEntriesResponse{
		AppendEntriesResponse: &miniraft.AppendEntriesResponse{
			Term:    s.Raft.CurrentTerm,
			Success: success,
		},
	}}

	// fmt.Printf("Sending AppendEntriesResponse to %s\n", address)
	s.SendMessage(address, message)
}

func (s *Server) HandleLogCommand() {
	for _, log := range s.Raft.Logs {
		fmt.Println(log.String())
	}
}

func (s *Server) HandlePrintCommand() {
	fmt.Printf("State: %d\n", s.State)
	fmt.Printf("CurrentTerm: %d\n", s.Raft.CurrentTerm)
	fmt.Printf("VotedFor: %d\n", s.Raft.VotedFor)
	fmt.Printf("CommitIndex: %d\n", s.Raft.CommitIndex)
	fmt.Printf("LastAppliedIndex: %d\n", s.Raft.LastApplied)
	fmt.Printf("CurrentTerm: %d\n", s.Raft.CurrentTerm)
	fmt.Printf("NextIndex: %v\n", s.Raft.NextIndex)
	fmt.Printf("MatchIndex: %v\n", s.Raft.MatchIndex)
}

func (s *Server) HandleResumeCommand() {
	if s.State == Failed {
		s.ChangeState(Follower)
	}
}

func (s *Server) HandleSuspendCommand() {
	s.ChangeState(Failed)
}

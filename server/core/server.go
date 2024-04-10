package core

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	miniraft "github.com/lsig/Raft/server/pb"
	"github.com/lsig/Raft/server/util"
	"google.golang.org/protobuf/proto"
)

func (s *Server) SendMessage(addr string, message *miniraft.Raft) error {
	data, err := proto.Marshal(message)

	if err != nil {
		log.Println("Failed to marshal message")
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Println("Failed to resolve UDP address")
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	if _, err := s.Gateway.WriteTo(data, udpAddr); err != nil {
		log.Printf("Error sending UDP packet %s\n", err.Error())
		log.Printf("Error sending to UDP Address: %s\n", udpAddr.String())
		log.Printf("Error sending from UDP raw Address: %s\n", addr)
		return fmt.Errorf("error sending UDP packet: %w", err)
	}

	return nil
}

func (s *Server) ReceiveMessage() error {
	bs := make([]byte, 65536)

	n, addr, err := s.Gateway.ReadFromUDP(bs)
	if err != nil {
		log.Println("Failed to read from udp buffer")
		return fmt.Errorf("failed to read from udp buffer: %w", err)
	}
	packet := &Packet{
		Address: addr.String(),
		Content: &miniraft.Raft{},
	}

	err = proto.Unmarshal(bs[0:n], packet.Content)

	if err != nil {
		log.Println("Failed to unmarshal message")
		return fmt.Errorf("failed to read from udp buffer: %w", err)
	}

	s.Messages <- packet

	return nil
}

func (s *Server) Start() {
	go s.MessageProcessing()
	go s.CommandLineInterface()
	go s.StartTimeout()
	go s.WaitForTimeout()

	serverAddr, err := net.ResolveUDPAddr("udp", s.Info.Address)
	if err != nil {
		log.Fatal("Failed to resolve server address", err)
	}

	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		log.Fatal("Failed to listen to udp", err)
	}
	s.Gateway = conn
	defer conn.Close()

	log.Printf("Server listening on %s\n", s.Info.Address)
	for {
		s.ReceiveMessage()
	}
}

func (s *Server) MessageProcessing() {
	for packet := range s.Messages {
		switch msg := packet.Content.Message.(type) {
		case *miniraft.Raft_CommandName:
			s.HandleClientCommand(packet.Address, msg.CommandName)
		case *miniraft.Raft_RequestVoteRequest:
			if s.State == Failed {
				continue
			}
			s.HandleVoteRequest(packet.Address, msg.RequestVoteRequest)
		case *miniraft.Raft_RequestVoteResponse:
			// If I receive a positive vote, check how many votes I now have
			// If I have the majority of votes, become the leader
			if s.State == Failed {
				continue
			}
			s.HandleVoteResponse(packet.Address, msg.RequestVoteResponse)
		case *miniraft.Raft_AppendEntriesRequest:
			if s.State == Failed {
				continue
			}
			s.HandleAppendEntriesRequest(packet.Address, msg.AppendEntriesRequest)
		case *miniraft.Raft_AppendEntriesResponse:
			continue
		}
	}
}

func (s *Server) StartTimeout() {
	// need to check if server is leader
	for {
		select {
		case <-s.Timer.C:
			s.TimeoutDone <- struct{}{}
			s.Timer.Reset(util.GetRandomTimeout())
		case <-s.TimeoutReset:
			if !s.Timer.Stop() {
				<-s.Timer.C
			}
			s.Timer.Reset(util.GetRandomTimeout())
		}
	}
}

func (s *Server) CommandLineInterface() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := strings.TrimSpace(scanner.Text())
		switch {
		case command == "log":
			s.HandleLogCommand()
		case command == "print":
			s.HandlePrintCommand()
		case command == "resume":
			continue
		case command == "suspend":
			s.HandleSuspendCommand()
		}
	}
}

func (s *Server) WaitForTimeout() {
	for range s.TimeoutDone {
		// I'm the leader, timeouts don't affect me
		if s.State == Leader || s.State == Failed {
			continue
		}

		// Timed out, increment term
		s.UpdateTerm(s.Raft.CurrentTerm+1, s.Info.Id)

		fmt.Printf("Timed out - entering term %d\n", s.Raft.CurrentTerm)
		s.ChangeState(Candidate)

		s.Raft.Votes[s.Raft.CurrentTerm] = make([]int, s.Nodes.Len)
		s.Raft.Votes[s.Raft.CurrentTerm][uint64(s.Info.Id)] = 1

		details := s.CreateVoteRequest()
		message := &miniraft.Raft{
			Message: &miniraft.Raft_RequestVoteRequest{
				RequestVoteRequest: details,
			},
		}

		for _, addr := range s.Nodes.Addresses {
			if addr != s.Info.Address {
				s.SendMessage(addr, message)
			}
		}
	}
}

func (s *Server) AnnounceLeadership() {
	s.ChangeState(Leader)

	for s.State == Leader {
		for _, address := range s.Nodes.Addresses {
			// don't send to myself
			if address == s.Info.Address {
				continue
			}
			message := &miniraft.Raft{Message: &miniraft.Raft_AppendEntriesRequest{AppendEntriesRequest: &miniraft.AppendEntriesRequest{
				Term:         s.Raft.CurrentTerm,
				PrevLogIndex: 0,
				PrevLogTerm:  0,
				LeaderCommit: 0,
				LeaderId:     fmt.Sprint(s.Info.Id),
			}}}
			s.SendMessage(address, message)
		}

		// send heartbeats at an order of magnitude faster than timeouts
		time.Sleep(util.GetRandomTimeout() / 10)
	}
}

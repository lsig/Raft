package core

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"strconv"
	"time"

	miniraft "github.com/lsig/Raft/server/pb"
	"google.golang.org/protobuf/proto"
)

type State int

const (
	Leader State = iota
	Candidate
	Follower
	Failed
)

type Server struct {
	Address      string
	Messages     chan *Packet
	Commands     chan string
	TimeoutReset chan struct{}
	TimeoutDone  chan struct{}
	Nodes        map[string]string
	LeaderId     int
	Timeout      time.Duration
	State        State
	CurrentTerm  uint64
	VotedFor     int
	Log          []miniraft.LogEntry
	CommitIndex  int
	LastApplied  int
	NextIndex    []int
	MatchIndex   []int
}

type Packet struct {
	Address string
	Content *miniraft.Raft
}

func NewServer(address string, nodes []string) *Server {
	dict := map[string]string{}
	for idx, addr := range nodes {
		dict[addr] = strconv.Itoa(idx)
	}

	return &Server{
		Address:     address,
		Messages:    make(chan *Packet, 128),
		Commands:    make(chan string, 128),
		Nodes:       dict,
		LeaderId:    -1,
		Timeout:     time.Duration(rand.IntN(300)+150) * time.Millisecond,
		State:       Follower,
		CurrentTerm: 0,
		VotedFor:    -1,
		CommitIndex: 0,
		LastApplied: 0,
		NextIndex:   nil,
		MatchIndex:  nil,
	}
}

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

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Println("Failed to dial UDP")
		return fmt.Errorf("failed to dial UDP: %w", err)
	}
	defer conn.Close()

	if _, err := conn.Write(data); err != nil {
		log.Println("Error sending UDP packet")
		return fmt.Errorf("error sending UDP packet: %w", err)
	}

	return nil
}

func (s *Server) ReceiveMessage(conn *net.UDPConn) error {
	bs := make([]byte, 65536)
	n, addr, err := conn.ReadFromUDP(bs)
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
	go s.CommandProcessing()
	go s.StartTimeout()
	go s.WaitForTimeout()

	serverAddr, err := net.ResolveUDPAddr("udp", s.Address)

	if err != nil {
		log.Fatal("Failed to resolve server address", err)
	}

	conn, err := net.ListenUDP("udp", serverAddr)

	if err != nil {
		log.Fatal("Failed to listen to udp", err)
	}

	defer conn.Close()

	log.Printf("Server listening on %s\n", s.Address)
	for {
		s.ReceiveMessage(conn)
	}
}

func (s *Server) MessageProcessing() {
	for packet := range s.Messages {
		switch msg := packet.Content.Message.(type) {
		case *miniraft.Raft_CommandName:
			s.HandleClientCommand(packet.Address, msg.CommandName)
		case *miniraft.Raft_RequestVoteRequest:
			continue
		case *miniraft.Raft_RequestVoteResponse:
			continue
		case *miniraft.Raft_AppendEntriesRequest:
			s.TimeoutReset <- struct{}{}
		case *miniraft.Raft_AppendEntriesResponse:
			continue
		default:
			fmt.Println(msg)
		}
	}

}

func (s *Server) StartTimeout() {
	timer := time.NewTimer(s.Timeout)
	for {
		select {
		case <-timer.C:
			s.TimeoutDone <- struct{}{}
			timer.Reset(time.Duration(rand.IntN(300)+150) * time.Millisecond)
		case <-s.TimeoutReset:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(time.Duration(rand.IntN(300)+150) * time.Millisecond)
		}
	}
}

func (s *Server) CommandProcessing() {
	for command := range s.Commands {
		switch {
		case command == "log":
			continue
		case command == "print":
			continue
		case command == "resume":
			continue
		case command == "suspend":
			continue
		}
	}
}

func (s *Server) WaitForTimeout() {
	for range s.TimeoutDone {
		s.State = Candidate

		lastIndex := len(s.Log) - 1

		details := &miniraft.RequestVoteRequest{
			Term:          s.CurrentTerm + 1,
			CandidateName: s.Nodes[s.Address],
			LastLogIndex:  uint64(lastIndex),
			LastLogTerm:   s.Log[lastIndex].Term,
		}

		message := &miniraft.Raft{
			Message: &miniraft.Raft_RequestVoteRequest{
				RequestVoteRequest: details,
			},
		}

		for addr := range s.Nodes {
			s.SendMessage(addr, message)
		}
	}
}

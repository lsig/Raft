package core

import (
	"fmt"
	"log"
	"net"

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
	Address     string
	Messages    chan *Packet
	Commands    chan string
	State       State
	CurrentTerm int
	VotedFor    int
	Log         []miniraft.LogEntry
	CommitIndex int
	LastApplied int
	NextIndex   []int
	MatchIndex  []int
}

type Packet struct {
	Address string
	Content *miniraft.Raft
}

func NewServer(address string) *Server {
	return &Server{
		Address:     address,
		Messages:    make(chan *Packet, 128),
		Commands:    make(chan string, 128),
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

func (s *Server) ReceiveMessage(conn *net.UDPConn) (*Packet, error) {
	bs := make([]byte, 1024)
	n, addr, err := conn.ReadFromUDP(bs)
	if err != nil {
		log.Println("Failed to read from udp buffer")
		return &Packet{}, fmt.Errorf("failed to read from udp buffer: %w", err)
	}

	packet := &Packet{
		Address: addr.String(),
		Content: &miniraft.Raft{},
	}
	err = proto.Unmarshal(bs[0:n], packet.Content)

	return packet, nil
}

func (s *Server) MessageProcessing() {
	for packet := range s.Messages {
		switch msg := packet.Content.Message.(type) {
		case *miniraft.Raft_CommandName:
			continue
		case *miniraft.Raft_RequestVoteRequest:
			continue
		case *miniraft.Raft_RequestVoteResponse:
			continue
		case *miniraft.Raft_AppendEntriesRequest:
			continue
		case *miniraft.Raft_AppendEntriesResponse:
			continue
		default:
			fmt.Printf("msg: %v\n", msg)
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

func (s *Server) SendingMessages() {

}

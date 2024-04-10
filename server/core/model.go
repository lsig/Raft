package core

import (
	"fmt"
	"net"
	"time"

	miniraft "github.com/lsig/Raft/server/pb"
	"github.com/lsig/Raft/server/util"
)

type State int

const (
	Leader State = iota
	Candidate
	Follower
	Failed
)

type Nodes struct {
	Addresses []string
	Len       int
}

type Info struct {
	Address string
	Id      int
}

type Log struct {
	Term    uint64
	Index   int
	Command string
}

type Raft struct {
	LeaderId    int
	CurrentTerm uint64
	VotedFor    int
	Votes       map[uint64][]int
	Logs        []Log
	CommitIndex int
	LastApplied int
	NextIndex   []int
	MatchIndex  []int
	// Updating    []bool
}

type Server struct {
	Info         Info
	Messages     chan *Packet
	TimeoutReset chan struct{}
	TimeoutDone  chan struct{}
	Nodes        Nodes
	Gateway      *net.UDPConn
	Timer        *time.Timer
	State        State
	Raft         Raft
}

type Packet struct {
	Address string
	Content *miniraft.Raft
}

func NewServer(address string, nodes []string) *Server {
	var id int
	for idx, addr := range nodes {
		if addr == address {
			id = idx
		}
	}

	fmt.Printf("Number of addresses: %d\n", len(nodes))

	nodeInfo := Nodes{
		Addresses: nodes,
		Len:       len(nodes),
	}

	info := Info{
		Address: address,
		Id:      id,
	}

	raft := Raft{
		LeaderId:    -1,
		CurrentTerm: 0,
		VotedFor:    -1,
		Votes:       map[uint64][]int{},
		Logs:        []Log{},
		CommitIndex: 0,
		LastApplied: 0,
		NextIndex:   make([]int, nodeInfo.Len),
		MatchIndex:  make([]int, nodeInfo.Len),
	}

	return &Server{
		Info:         info,
		Messages:     make(chan *Packet, 128),
		TimeoutDone:  make(chan struct{}),
		TimeoutReset: make(chan struct{}),
		Nodes:        nodeInfo,
		Timer:        time.NewTimer(util.GetRandomTimeout()),
		State:        Follower,
		Raft:         raft,
	}
}

func (l *Log) String() string {
	return fmt.Sprintf("%d,%d,%s", l.Term, l.Index, l.Command)
}

func (l *Log) ToLogEntry() *miniraft.LogEntry {
	logEntry := &miniraft.LogEntry{Index: uint64(l.Index), Term: l.Term, CommandName: l.Command}
	return logEntry
}

func (l Log) FromLogEntry(entry *miniraft.LogEntry) Log {
	l.Index = int(entry.Index)
	l.Term = entry.Term
	l.Command = entry.CommandName
	return l
}

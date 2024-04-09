package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lsig/Raft/server/core"
	"github.com/lsig/Raft/server/util"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run raftserver.go server-host:server-port filename")
		return
	}

	address := os.Args[1]
	file := os.Args[2]

	nodes := util.FindNodesFromFile(file, address)
	for _, n := range nodes {
		fmt.Printf("node: %s\n", n)
	}

	err := util.AppendToFile(fmt.Sprintf("%s.log", address), address)

	if err != nil {
		log.Fatal("Error writing server config to file", err)
	}

	server := core.NewServer(address)
	server.Start()
}

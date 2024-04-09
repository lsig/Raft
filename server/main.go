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

	err := util.AppendToFile(file, address)

	if err != nil {
		log.Fatal("Error writing server config to file", err)
	}

	fmt.Println(address)
	fmt.Println(file)

	server := core.NewServer(address)
    server.Start()
}

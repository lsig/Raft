package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lsig/Raft/server/util"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run raftserver.go server-host:server-port filename")
		return
	}

	server := os.Args[1]
	file := os.Args[2]

	err := util.AppendToFile(file, server)

	if err != nil {
		log.Fatal("Error writing server config to file", err)
	}

	fmt.Println(server)
	fmt.Println(file)
}

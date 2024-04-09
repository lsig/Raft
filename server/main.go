package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run raftserver.go server-host:server-port filename")
		return
	}

	server := os.Args[1]
	file := os.Args[2]

	fmt.Println(server)
	fmt.Println(file)
}

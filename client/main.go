package main

import (
	"fmt"
	"os"

	client "github.com/lsig/Raft/client/core"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run . <server-host>:<server-port>")
		os.Exit(1)
	}

	server := os.Args[1]

	c := client.NewClient(server)

	message := "hello world"

	err := c.SendMessage(message)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
	} else {
		fmt.Printf("message sent: %s\n", message)
	}
}

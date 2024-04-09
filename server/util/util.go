package util

import (
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
	"time"
)

func FindNodesAndAddressFromArgs() (nodes []string, address string) {
	// handle invalid program arguments
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run raftserver.go server-host:server-port filename")
		os.Exit(1)
	}
	address = os.Args[1]
	file := os.Args[2]

	nodes = []string{}

	fileBytes, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}

	// src: https://golangdocs.com/golang-byte-array-to-string
	contents := string(fileBytes[:])

	servers := strings.Split(contents, "\n")
	for _, s := range servers {
		if s != address {
			nodes = append(nodes, s)
		}
	}

	return
}

func AppendToFile(filename, text string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(text + "\n")
	if err != nil {
		return err
	}
	return nil
}

func GetRandomTimeout() time.Duration {
	debugScale := 10
	return time.Duration(rand.IntN(300*debugScale)+150*debugScale) * time.Millisecond
}

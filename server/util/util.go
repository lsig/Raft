package util

import (
	"fmt"
	"math/rand/v2"
	"net"
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

	fileBytes, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}

	// src: https://golangdocs.com/golang-byte-array-to-string
	contents := string(fileBytes[:])
	contents = strings.Trim(contents, "\n")
	nodes = strings.Split(contents, "\n")

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
	debugScale := 1
	return time.Duration(rand.IntN(300*debugScale)+150*debugScale) * time.Millisecond
}

func FindServerId(addresses []string, address string) uint64 {
	var serverIndex uint64

	addr, _ := net.ResolveUDPAddr("udp", address)

	for idx, nodeAddr := range addresses {
		nAddr, _ := net.ResolveUDPAddr("udp", nodeAddr)
		if addr.String() == nAddr.String() {
			serverIndex = uint64(idx)
			break
		}
	}

	return serverIndex
}

func FindLeaderAddress(addresses []string, leaderId int) string {
	return addresses[leaderId]
}

func ReceivedMajorityVotes(votes []int) bool {
	total := len(votes)
	received := 0

	for _, v := range votes {
		if v == 1 {
			received++
		}
	}

	// a majority is reached when the number of received votes is strictly more than half of all nodes
	return received > total/2
}

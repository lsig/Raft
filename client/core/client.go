package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	miniraft "github.com/lsig/Raft/client/pb"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	ServerAddress string
}

func NewClient(serverAddress string) *Client {
	return &Client{ServerAddress: serverAddress}
}

func (c *Client) SendMessage(msg string) error {
	message := &miniraft.Raft{Message: &miniraft.Raft_CommandName{CommandName: msg}}

	data, err := proto.Marshal(message)

	if err != nil {
		log.Println("Failed to marshal message")
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", c.ServerAddress)

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

func (c *Client) HandleUserInput() {
	// use bufio instead of fmt.Scanln, as Scanln doesn't handle whitespace
	// ... Even though the client doesn't have to support whitespace in input
	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error reading stdin: %s\n", err.Error())
			continue
		}
		message := strings.Trim(input, "\n")

		if message == "exit" {
			break
		}

		err = c.SendMessage(message)
		if err != nil {
			fmt.Printf("error sending msg: %s\n", err.Error())
		}
	}
}

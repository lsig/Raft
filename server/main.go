package main

import (
	"fmt"
	"log"

	"github.com/lsig/Raft/server/core"
	"github.com/lsig/Raft/server/util"
)

func main() {
	nodes, address := util.FindNodesAndAddressFromArgs()

	err := util.AppendToFile(fmt.Sprintf("%s.log", address), address)
	if err != nil {
		log.Fatal("Error writing server config to file", err)
	}

	server := core.NewServer(address, nodes)
	server.Start()
}

package main

import (
	"github.com/lsig/Raft/server/core"
	"github.com/lsig/Raft/server/util"
)

func main() {
	nodes, address := util.FindNodesAndAddressFromArgs()

	server := core.NewServer(address, nodes)
	server.Start()
}

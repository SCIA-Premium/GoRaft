package main

import (
	"flag"
	"strings"
)

func main() {
	// Parse the command line arguments
	peer_ID := flag.Int("peer_id", 1, "Peer ID")
	peers_addrs := flag.String("peer", "127.0.0.1:10000", "Peers adresses")
	port := flag.String("port", ":10000", "Cluster port")
	flag.Parse()

	peers_list := strings.Split(*peers_addrs, ",")
	// Create the peers map
	peers := make(map[int]*Peer)
	for k, v := range peers_list {
		peers[k] = NewPeer(v)
	}

	// Create the node
	node := NewNode(*peer_ID, peers)
	// Register the node to RPC
	node.startRpc(*port)
	// Start the node
	node.Start()
	select {}
}

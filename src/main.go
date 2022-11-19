package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

func main() {
	// Parse the command line arguments
	peer_ID := flag.Int("peer_id", 1, "Peer ID")
	peers_addrs := flag.String("peer", "127.0.0.1:10000", "Peers adresses")
	port := flag.String("port", "10000", "Cluster port")
	flag.Parse()

	// Create the peers list
	peers_list := strings.Split(*peers_addrs, ",")
	peers := make([]*Peer, len(peers_list))
	for i, v := range peers_list {
		peers[i] = NewPeer(v)
	}

	// Create the node
	node := NewNode(*peer_ID, *port, peers)

	// Create output directory
	os.Mkdir("output", 0777)
	os.Mkdir("output/"+"node_"+strconv.Itoa(*peer_ID), 0777)
	// Create the log file
	log_file, _ := os.Create("output/" + "node_" + strconv.Itoa(*peer_ID) + "/log")
	log_file.Close()

	// Register the node to RPC
	node.startRpc(*port)
	// Start the node
	node.Start()
	select {}
}

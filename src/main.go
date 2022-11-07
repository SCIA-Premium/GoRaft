package raft

import (
	"flag"
	"strings"

)

func main() {
	peers_addrs := flag.String("peer", "127.0.0.1:10000", "Peers adresses")
	//port := flag.String("port", ":10000", "Cluster port")

	flag.Parse()
	peers_list := strings.Split(*peers_addrs, ",")

	peers := make(map[int]*Peer)
	for k, v := range peers_list{
		peers[k] = NewPeer(v)
	}
	
	node := NewNode(peers)
	node.Start()

}

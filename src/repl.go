package main

import (
	"log"
)

func (n *Node) Stop(false_arg string, false_res *string) error {
	n.State = Dead
	log.Printf("Node %d [%s]\n", n.PeerID, n.State)
	return nil
}

func (n *Node) ChangeSpeed(new_speed_string string, false_res *string) error {

	var new_speed NodeSpeed
	switch new_speed_string {
	case "high":
		new_speed = NodeSpeed{"high", 300}
	case "medium":
		new_speed = NodeSpeed{"medium", 600}
	case "low":
		new_speed = NodeSpeed{"low", 1000}
	default:
		log.Printf("ChangeSpeed : unknow speed specification : %s", new_speed_string)
		return nil
	}

	log.Printf("Node %d [%s]: old speed state %s -> new speed state %s \n", n.PeerID, n.State, n.Speed.key, new_speed.key)

	n.Speed = new_speed

	return nil
}

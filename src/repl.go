package main

import (
	"log"
)

func (n *Node) Crash(false_arg string, false_res *string) error {
	n.State = Dead
	log.Printf("[T%d][%s]\n", n.CurrentTerm, n.State)
	return nil
}

func (n *Node) Speed(new_speed_string string, false_res *string) error {

	var new_speed SpeedState
	switch new_speed_string {
	case "high":
		new_speed = SpeedState{"high", 300}
	case "medium":
		new_speed = SpeedState{"medium", 600}
	case "low":
		new_speed = SpeedState{"low", 1000}
	default:
		log.Printf("ChangeSpeed : unknow speed specification : %s", new_speed_string)
		return nil
	}

	log.Printf("[T%d][%s]: old speed state %s -> new speed state %s \n", n.CurrentTerm, n.State, n.Speed.key, new_speed.key)
	log.Printf("Node %d [%s]: old speed state %s -> new speed state %s \n", n.PeerID, n.State, n.SpeedState.key, new_speed.key)

	n.SpeedState = new_speed

	return nil
}

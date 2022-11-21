package main

import (
	"fmt"
)

func (n *Node) StartClsf(false_arg string, res *string) error {
	if n.Started {
		return fmt.Errorf("Client entries are already accepted")
	}

	n.Started = true
	*res = fmt.Sprintf("Client entries are now accepted\n")
	return nil
}

func (n *Node) Crash(false_arg string, res *string) error {
	if !n.Alive {
		return fmt.Errorf("Node%d has already crashed", n.PeerID)
	}

	n.Alive = false
	*res = fmt.Sprintf("Node%d has been crashed\n", n.PeerID)
	return nil
}

func (n *Node) Recovery(false_arg string, res *string) error {
	if n.Alive {
		return fmt.Errorf("Node%d is already alive", n.PeerID)
	}

	n.Alive = true
	*res = fmt.Sprintf("Node%d has been recovered\n", n.PeerID)
	return nil
}

func (n *Node) Speed(new_speed_string string, res *string) error {
	var new_speed SpeedState
	switch new_speed_string {
	case "high":
		new_speed = SpeedState{"high", 300}
	case "medium":
		new_speed = SpeedState{"medium", 600}
	case "low":
		new_speed = SpeedState{"low", 1200}
	default:
		return fmt.Errorf("Unknown speed specification : %s", new_speed_string)
	}

	*res = fmt.Sprintf("Node%d change from %s speed to %s speed\n", n.PeerID, n.SpeedState.key, new_speed.key)

	n.SpeedState = new_speed

	return nil
}

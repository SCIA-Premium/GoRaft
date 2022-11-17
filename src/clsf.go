package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (n *Node) List(args string, res *string) error {
	if n.State == Dead {
		return errors.New("Node is dead")
	}

	if n.LeaderUID == uuid.Nil {
		return errors.New("No leader")
	}

	if n.LeaderUID != n.PeerUID {
		return fmt.Errorf("The leader address is : %s", n.LeaderAddress)
	}

	*res = ""
	for uid, _ := range n.RegisteredFiles {
		*res += uid.String() + "\n"
	}

	return nil
}

func (n *Node) Load(args string, res *string) error {
	if n.State == Dead {
		return errors.New("Node is dead")
	}

	if n.LeaderUID == uuid.Nil {
		return errors.New("No leader")
	}

	if n.LeaderUID != n.PeerUID {
		return fmt.Errorf("The leader port is %s", n.LeaderAddress)
	}

	filename_uid := uuid.New()
	save_len_log := len(n.Log)
	n.Log = append(n.Log, LogEntry{n.CurrentTerm, save_len_log, "LOAD " + args + " " + filename_uid.String(), 0, false})

	for {
		if n.Log[save_len_log].Committed {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	n.RegisteredFiles[filename_uid] = args
	*res = filename_uid.String()

	return nil
}

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (n *Node) List(args string, res *string) error {
	if !n.Alive {
		return errors.New("Node is not alive")
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

func (n *Node) load(filename string, s_uuid string) error {
	file_uid, _ := uuid.Parse(s_uuid)

	n.RegisteredFiles[file_uid] = filename

	_, err := os.Create("output/node_" + strconv.Itoa(n.PeerID) + "/" + filename)
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) Load(filename string, res *string) error {
	if !n.Alive {
		return errors.New("Node is not alive")
	}

	if n.LeaderUID == uuid.Nil {
		return errors.New("No leader")
	}

	if n.LeaderUID != n.PeerUID {
		return fmt.Errorf("The leader port is %s", n.LeaderAddress)
	}

	save_len_log := len(n.Log)

	filename_uid := uuid.New()
	n.Log = append(n.Log, LogEntry{n.CurrentTerm, save_len_log, "LOAD " + filename + " " + filename_uid.String(), 1, false})

	for {
		if len(n.Log) <= save_len_log {
			return fmt.Errorf("Could not load file %s", filename)
		}

		if n.Log[save_len_log].Committed {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	err := n.ExecuteCommand(n.Log[save_len_log].Command)

	*res = filename_uid.String()

	return err
}

func (n *Node) delete(s_uuid string) error {
	file_uid, err := uuid.Parse(s_uuid)
	if err != nil {
		return err
	}

	if _, ok := n.RegisteredFiles[file_uid]; !ok {
		return errors.New("File not found")
	}

	// Remove file
	err = os.Remove("output/node_" + strconv.Itoa(n.PeerID) + "/" + n.RegisteredFiles[file_uid])
	delete(n.RegisteredFiles, file_uid)

	return err
}

func (n *Node) Delete(args string, res *string) error {
	if !n.Alive {
		return errors.New("Node is not alive")
	}

	if n.LeaderUID == uuid.Nil {
		return errors.New("No leader")
	}

	if n.LeaderUID != n.PeerUID {
		return fmt.Errorf("The leader port is %s", n.LeaderAddress)
	}

	save_len_log := len(n.Log)
	n.Log = append(n.Log, LogEntry{n.CurrentTerm, save_len_log, "DELETE " + args, 1, false})

	for {
		if len(n.Log) < save_len_log {
			return fmt.Errorf("Could not delete file %s", args)
		}

		if n.Log[save_len_log].Committed {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	err := n.ExecuteCommand(n.Log[save_len_log].Command)

	return err
}

func (n *Node) ExecuteCommand(command string) error {
	log.Printf("[T%d][%s]: executing command: %s\n", n.CurrentTerm, n.State, command)

	splited := strings.Split(command, " ")

	switch splited[0] {
	case "LOAD":
		return n.load(splited[1], splited[2])
	case "DELETE":
		return n.delete(splited[1])
	case "APPEND":
	}

	return nil
}

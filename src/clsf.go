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

// Handle error
func (n *Node) handle_error() error {
	if !n.Alive {
		return errors.New("Node is not alive")
	}

	if !n.Started {
		return errors.New("Client entries are not currently accepted")
	}

	if n.LeaderUID == uuid.Nil {
		return errors.New("No leader")
	}

	if n.LeaderUID != n.PeerUID {
		return fmt.Errorf("The leader port is %s", n.LeaderAddress)
	}

	return nil
}

// Wait for commit of the command at index
func (n *Node) wait_commit(index int, err error) error {
	for {
		if len(n.Log) <= index {
			return err
		}

		if n.Log[index].Committed {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// List all registered files
func (n *Node) List(args string, res *string) error {
	err := n.handle_error()
	if err != nil {
		return err
	}

	*res = ""
	for uid, _ := range n.RegisteredFiles {
		*res += uid.String() + "\n"
	}

	return nil
}

// Load create a file with uuid and register it
func (n *Node) load(filename string, s_uuid string) error {
	file_uid, _ := uuid.Parse(s_uuid)

	for _, f := range n.RegisteredFiles {
		if f == filename {
			return fmt.Errorf("File '%s' already loaded", filename)
		}
	}

	_, err := os.Create("output/node_" + strconv.Itoa(n.PeerID) + "/" + filename)
	if err != nil {
		return err
	}
	n.RegisteredFiles[file_uid] = filename

	return nil
}

// Add LOAD command in Log and wait for commit before executing it
func (n *Node) Load(filename string, res *string) error {
	err := n.handle_error()
	if err != nil {
		return err
	}

	save_len_log := len(n.Log)

	filename_uid := uuid.New()
	n.Log = append(n.Log, LogEntry{n.CurrentTerm, save_len_log, "LOAD " + filename + " " + filename_uid.String(), 1, false})

	err = n.wait_commit(save_len_log, fmt.Errorf("Could not load file %s", filename))
	if err != nil {
		return err
	}

	_, err = n.ExecuteCommand(n.Log[save_len_log].Command)
	if err != nil {
		return err
	}

	*res = filename_uid.String()

	return err
}

// Delete the file with uuid if registered
func (n *Node) delete(s_uuid string) (string, error) {
	file_uid, err := uuid.Parse(s_uuid)
	if err != nil {
		return "", err
	}

	if _, ok := n.RegisteredFiles[file_uid]; !ok {
		return "", errors.New("File not found")
	}

	filename := n.RegisteredFiles[file_uid]

	// Remove file
	err = os.Remove("output/node_" + strconv.Itoa(n.PeerID) + "/" + n.RegisteredFiles[file_uid])
	delete(n.RegisteredFiles, file_uid)

	return filename, err
}

// Add DELETE command in Log and wait for commit before executing it
func (n *Node) Delete(args string, res *string) error {
	err := n.handle_error()
	if err != nil {
		return err
	}

	save_len_log := len(n.Log)
	n.Log = append(n.Log, LogEntry{n.CurrentTerm, save_len_log, "DELETE " + args, 1, false})

	err = n.wait_commit(save_len_log, fmt.Errorf("Could not delete file %s", args))
	if err != nil {
		return err
	}

	*res, err = n.ExecuteCommand(n.Log[save_len_log].Command)

	return err
}

// Append content to file with uuid if registered
func (n *Node) append(s_uuid string, content string) (string, error) {
	file_uid, err := uuid.Parse(s_uuid)
	if err != nil {
		return "", err
	}

	if _, ok := n.RegisteredFiles[file_uid]; !ok {
		return "", errors.New("File not found")
	}

	// Append to file
	f, err := os.OpenFile("output/node_"+strconv.Itoa(n.PeerID)+"/"+n.RegisteredFiles[file_uid], os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}
	f.Write([]byte(content + "\n"))
	f.Close()

	return n.RegisteredFiles[file_uid], err
}

// Add APPEND command in Log and wait for commit before executing it
func (n *Node) Append(args string, res *string) error {
	err := n.handle_error()
	if err != nil {
		return err
	}

	save_len_log := len(n.Log)
	n.Log = append(n.Log, LogEntry{n.CurrentTerm, save_len_log, "APPEND " + args, 1, false})

	err = n.wait_commit(save_len_log, fmt.Errorf("Could not append in file", args))
	if err != nil {
		return err
	}

	*res, err = n.ExecuteCommand(n.Log[save_len_log].Command)

	return err
}

// ExecuteCommand executes a command for the node (Load, Delete, Append)
func (n *Node) ExecuteCommand(command string) (string, error) {
	log.Printf("[T%d][%s]: Executing command: %s\n", n.CurrentTerm, n.State, command)

	splited := strings.Split(command, " ")

	// Write log to file
	f, _ := os.OpenFile("output/node_"+strconv.Itoa(n.PeerID)+"/log", os.O_APPEND|os.O_WRONLY, 0600)
	f.Write([]byte(command + "\n"))
	f.Close()

	// Execute command
	switch splited[0] {
	case "LOAD":
		return "", n.load(splited[1], splited[2])
	case "DELETE":
		return n.delete(splited[1])
	case "APPEND":
		return n.append(splited[1], command[len(splited[0])+len(splited[1])+2:])
	}

	return "", nil
}

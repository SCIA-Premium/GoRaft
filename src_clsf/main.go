package main

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
)

// Get rpc client from address
func get_client(address string) *rpc.Client {
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Println(err)
		return nil
	}
	return client
}

// Call Node.List function on client and print the result
func list(client *rpc.Client) {
	var args string
	var res string
	err := client.Call("Node.List", args, &res)
	if err != nil {
		log.Println(err)
		return
	}

	if res == "" {
		res = "(No files)\n"
	}

	fmt.Printf("Available files uuid:\n%s", res)
}

// Call Node.Load function on client
func load(client *rpc.Client, filename string) {
	args := filename
	var res string

	err := client.Call("Node.Load", args, &res)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("File '%s' loaded with uuid '%s'\n", filename, res)
}

// Call Node.Delete function on client
func delete(client *rpc.Client, s_uuid string) {
	args := s_uuid
	var res string
	err := client.Call("Node.Delete", args, &res)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("File '%s' with uuid '%s' deleted\n", res, s_uuid)
}

// Call Node.Append function on client
func append(client *rpc.Client, s_uuid string, content string) {
	args := s_uuid + " " + content
	var res string
	err := client.Call("Node.Append", args, &res)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("Content '%s' added to file '%s' with uuid '%s'\n", content, res, s_uuid)
}

func main() {
	// Parse the command line arguments
	address := os.Args[1]
	args := os.Args[2:]

	client := get_client(address)
	if client == nil {
		return
	}

	switch args[0] {
	case "LIST":
		list(client)
	case "LOAD":
		load(client, args[1])
	case "DELETE":
		delete(client, args[1])
	case "APPEND":
		append(client, args[1], args[2])
	}
}

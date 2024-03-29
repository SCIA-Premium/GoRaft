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

// Call Node.StartClsf function on client to allow the ask of requests
func start(client *rpc.Client) {
	var args string
	var res string
	err := client.Call("Node.StartClsf", args, &res)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("%s", res)
}

// Call Node.Speed function on client and print the result
func speed(client *rpc.Client, new_speed_string string) {
	var res string
	err := client.Call("Node.Speed", new_speed_string, &res)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("%s", res)
}

// Call Node.Crash function on client and print the result
func crash(client *rpc.Client) {
	var args string
	var res string
	err := client.Call("Node.Crash", args, &res)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("%s", res)
}

// Call Node.Recovery function on client and print the result
func recovery(client *rpc.Client) {
	var args string
	var res string
	err := client.Call("Node.Recovery", args, &res)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("%s", res)
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
	case "START":
		start(client)
	case "SPEED":
		speed(client, args[1])
	case "CRASH":
		crash(client)
	case "RECOVERY":
		recovery(client)
	default:
		fmt.Printf("Unknown command: %s", args[0])
	}
}

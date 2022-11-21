package main

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
)

func get_client(address string) *rpc.Client {
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Println(err)
		return nil
	}
	return client
}

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

func speed(client *rpc.Client, new_speed_string string) {
	var res string
	err := client.Call("Node.Speed", new_speed_string, &res)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("%s", res)
}

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

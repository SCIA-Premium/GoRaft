package main

import (
	"flag"
	"log"
	"net/rpc"
	"strings"
)

func change_speed(to_change_speed string) {
	address_list := strings.Split(to_change_speed, ",")
	var new_speed_string = address_list[0]

	for _, address := range address_list[1:] {
		func(address string) {
			client, err := rpc.DialHTTP("tcp", address)
			if err != nil {
				log.Println(err)
				return
			}
			
			var res string
			err = client.Call("Node.ChangeSpeed", new_speed_string, &res)
			if err != nil {
				log.Println(err)
				return
			}
		}(address)
	}
}

func kill_node(to_kill string) {
	address_list := strings.Split(to_kill, ",")

	for _, address := range address_list {
		func(address string) {
			client, err := rpc.DialHTTP("tcp", address)
			if err != nil {
				log.Println(err)
				return
			}
			
			var args string
			var res string
			err = client.Call("Node.Stop", args, &res)
			if err != nil {
				log.Println(err)
				return
			}
		}(address)
	}
}

func main() {
	// Parse the command line arguments
	to_speed := flag.String("to_speed", "NONE", "Adresses of nodes to change speed")
	to_kill := flag.String("to_kill", "NONE", "Adresses of nodes to kill")
	flag.Parse()

	if (*to_speed != "NONE") {
		change_speed(*to_speed)
	}

	if (*to_kill != "NONE") {
		kill_node(*to_kill)
	}
}

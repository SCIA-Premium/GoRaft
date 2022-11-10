package main

import (
	"flag"
	"log"
	"net/rpc"
	"strings"
)

func main() {
	// Parse the command line arguments
	to_kill := flag.String("to_kill", "127.0.0.1:10000", "Adresses of nodes to kill")
	flag.Parse()

	address_list := strings.Split(*to_kill, ",")

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

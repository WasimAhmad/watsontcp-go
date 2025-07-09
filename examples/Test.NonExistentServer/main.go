package main

import (
	"log"
	"net"

	"github.com/WasimAhmad/watsontcp-go/client"
)

func main() {
	addr := "127.0.0.1:65534" // assume no server is listening
	cli := client.New(addr, nil, client.Callbacks{}, nil)
	if err := cli.Connect(); err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			log.Printf("connection to %s timed out: %v", addr, err)
		} else {
			log.Printf("unable to connect to %s: %v", addr, err)
		}
		log.Println("Ensure the address is correct, the server is running, and consider retry logic for transient failures.")
		return
	}
	defer cli.Disconnect()
}

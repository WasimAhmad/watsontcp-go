package main

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/WasimAhmad/watsontcp-go/client"
	"github.com/WasimAhmad/watsontcp-go/message"
	"github.com/WasimAhmad/watsontcp-go/server"
)

func main() {
	// Start a simple server that logs the size and MD5 of each message.
	srvCb := server.Callbacks{
		OnMessage: func(id string, msg *message.Message, data []byte) {
			fmt.Printf("received %d bytes from %s: %x\n", len(data), id, md5.Sum(data))
		},
	}
	srv := server.New("127.0.0.1:9300", nil, srvCb, nil)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	defer srv.Stop()

	// Give the server time to start listening.
	time.Sleep(500 * time.Millisecond)

	// Create a client and connect to the server.
	cli := client.New("127.0.0.1:9300", nil, client.Callbacks{}, nil)
	if err := cli.Connect(); err != nil {
		log.Fatal(err)
	}
	defer cli.Disconnect()

	// Build a payload over several megabytes and send it.
	payload := make([]byte, 8*1024*1024)
	if _, err := rand.Read(payload); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("sending %d bytes: %x\n", len(payload), md5.Sum(payload))

	if err := cli.Send(&message.Message{}, payload); err != nil {
		log.Fatal(err)
	}

	// Wait a moment for the server to process the message.
	time.Sleep(time.Second)
}

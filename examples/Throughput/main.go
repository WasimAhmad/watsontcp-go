package main

import (
	"fmt"
	"log"
	"time"

	"github.com/WasimAhmad/watsontcp-go/client"
	"github.com/WasimAhmad/watsontcp-go/message"
	"github.com/WasimAhmad/watsontcp-go/server"
)

const (
	addr        = "127.0.0.1:9103"
	numMessages = 10000
)

func main() {
	srv := server.New(addr, nil, server.Callbacks{}, nil)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	defer srv.Stop()

	cli := client.New(addr, nil, client.Callbacks{}, nil)
	if err := cli.Connect(); err != nil {
		log.Fatal(err)
	}
	defer cli.Disconnect()

	sCli := cli.Statistics()
	sSrv := srv.Statistics()
	sCli.Reset()
	sSrv.Reset()
	start := time.Now()

	payload := []byte("hi")
	for i := 0; i < numMessages; i++ {
		if err := cli.Send(&message.Message{}, payload); err != nil {
			log.Fatal(err)
		}
	}

	// allow server to process remaining messages
	time.Sleep(500 * time.Millisecond)
	elapsed := time.Since(start)

	fmt.Printf("Client sent %d messages (%d bytes) in %v\n", sCli.SentMessages(), sCli.SentBytes(), elapsed)
	fmt.Printf("  Messages/sec: %.2f\n", float64(sCli.SentMessages())/elapsed.Seconds())
	fmt.Printf("  Bytes/sec:    %.2f\n", float64(sCli.SentBytes())/elapsed.Seconds())

	fmt.Printf("Server received %d messages (%d bytes)\n", sSrv.ReceivedMessages(), sSrv.ReceivedBytes())
}

package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/WasimAhmad/watsontcp-go/client"
	"github.com/WasimAhmad/watsontcp-go/message"
	"github.com/WasimAhmad/watsontcp-go/server"
)

const (
	serverAddr    = "127.0.0.1:9300"
	numClients    = 5
	msgsPerClient = 100
)

func startServer() *server.Server {
	cb := server.Callbacks{
		OnConnect: func(id string, _ net.Conn) {
			log.Printf("server: %s connected", id)
		},
		OnDisconnect: func(id string) {
			log.Printf("server: %s disconnected", id)
		},
		OnMessage: func(id string, _ *message.Message, data []byte) {
			log.Printf("server got from %s: %s", id, string(data))
		},
	}
	srv := server.New(serverAddr, nil, cb, nil)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	return srv
}

func runClient(idx int, wg *sync.WaitGroup) {
	defer wg.Done()
	cb := client.Callbacks{
		OnConnect:    func() { log.Printf("client %d connected", idx) },
		OnDisconnect: func() { log.Printf("client %d disconnected", idx) },
	}
	c := client.New(serverAddr, nil, cb, nil)
	if err := c.Connect(); err != nil {
		log.Printf("client %d connect: %v", idx, err)
		return
	}
	defer c.Disconnect()

	for i := 0; i < msgsPerClient; i++ {
		txt := fmt.Sprintf("client %d message %d", idx, i)
		if err := c.Send(&message.Message{}, []byte(txt)); err != nil {
			log.Printf("send error: %v", err)
			return
		}
	}
}

func main() {
	srv := startServer()
	defer srv.Stop()

	time.Sleep(500 * time.Millisecond)

	var wg sync.WaitGroup
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go runClient(i, &wg)
	}
	wg.Wait()

	fmt.Println("--- server statistics ---")
	fmt.Println(srv.Statistics())
}

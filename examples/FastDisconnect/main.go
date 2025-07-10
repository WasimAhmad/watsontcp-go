package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/WasimAhmad/watsontcp-go/client"
	"github.com/WasimAhmad/watsontcp-go/server"
)

const addr = "127.0.0.1:9103"

func main() {
	var (
		mu     sync.Mutex
		active int
	)

	cb := server.Callbacks{
		OnConnect: func(id string, c net.Conn) {
			mu.Lock()
			active++
			fmt.Printf("connect %s (active %d)\n", id, active)
			mu.Unlock()
		},
		OnDisconnect: func(id string) {
			mu.Lock()
			active--
			fmt.Printf("disconnect %s (active %d)\n", id, active)
			mu.Unlock()
		},
	}

	srv := server.New(addr, nil, cb, nil)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	defer srv.Stop()

	for i := 0; i < 100; i++ {
		cli := client.New(addr, nil, client.Callbacks{}, nil)
		if err := cli.Connect(); err != nil {
			log.Println("connect:", err)
			time.Sleep(50 * time.Millisecond)
			continue
		}
		cli.Disconnect()
	}

	time.Sleep(500 * time.Millisecond)
	mu.Lock()
	fmt.Printf("active connections after loop: %d\n", active)
	mu.Unlock()
}

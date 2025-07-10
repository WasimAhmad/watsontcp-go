package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/WasimAhmad/watsontcp-go/client"
	"github.com/WasimAhmad/watsontcp-go/message"
	"github.com/WasimAhmad/watsontcp-go/server"
)

const addr = "127.0.0.1:9300"

func main() {
	srvCb := server.Callbacks{
		OnMessage: func(id string, msg *message.Message, data []byte) {
			fmt.Printf("[%s] %s\n", id, string(data))
		},
	}
	srv := server.New(addr, nil, srvCb, nil)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	defer srv.Stop()

	cli := client.New(addr, nil, client.Callbacks{}, nil)
	if err := cli.Connect(); err != nil {
		log.Fatal(err)
	}
	defer cli.Disconnect()

	// Client.Send is safe for concurrent use by multiple goroutines.
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				txt := fmt.Sprintf("goroutine %d msg %d", n, j)
				if err := cli.Send(&message.Message{}, []byte(txt)); err != nil {
					log.Println("send:", err)
					return
				}
				time.Sleep(100 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
}

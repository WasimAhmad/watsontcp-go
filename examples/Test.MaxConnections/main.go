package main

import (
	"log"
	"sync"
	"time"

	"github.com/WasimAhmad/watsontcp-go/client"
	"github.com/WasimAhmad/watsontcp-go/message"
	"github.com/WasimAhmad/watsontcp-go/server"
)

func main() {
	opts := server.DefaultOptions()
	opts.MaxConnections = 2
	srv := server.New("127.0.0.1:9300", nil, server.Callbacks{}, &opts)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	defer srv.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cli := client.New("127.0.0.1:9300", nil, client.Callbacks{}, nil)
			if err := cli.Connect(); err != nil {
				log.Printf("client %d rejected: %v", id, err)
				return
			}
			log.Printf("client %d connected", id)
			defer cli.Disconnect()
			cli.Send(&message.Message{}, []byte("hi"))
			time.Sleep(time.Second)
		}(i)
		time.Sleep(200 * time.Millisecond)
	}
	wg.Wait()
}

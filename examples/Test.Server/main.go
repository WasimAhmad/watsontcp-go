package main

import (
	"fmt"
	"github.com/WasimAhmad/watsontcp-go/message"
	"github.com/WasimAhmad/watsontcp-go/server"
	"log"
)

func main() {
	cb := server.Callbacks{}
	cb.OnMessage = func(id string, msg *message.Message, data []byte) {
		fmt.Printf("%s: %s\n", id, string(data))
	}
	srv := server.New("127.0.0.1:9000", nil, cb, nil)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	defer srv.Stop()
	select {}
}

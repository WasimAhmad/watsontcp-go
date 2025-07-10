package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/WasimAhmad/watsontcp-go/client"
	"github.com/WasimAhmad/watsontcp-go/message"
	"github.com/WasimAhmad/watsontcp-go/server"
)

const addr = "127.0.0.1:9103"

func startServer() *server.Server {
	var conn net.Conn
	first := true
	cb := server.Callbacks{
		OnConnect: func(id string, c net.Conn) {
			conn = c
			fmt.Println("[server] client connected")
		},
		OnMessage: func(id string, msg *message.Message, data []byte) {
			fmt.Printf("[server] received: %s\n", string(data))
			if first {
				first = false
				// drop the connection without responding
				conn.Close()
				return
			}
			resp := &message.Message{SyncResponse: true, ConversationGUID: msg.ConversationGUID}
			resp.ContentLength = int64(len("ack"))
			hdr, _ := message.BuildHeader(resp)
			conn.Write(hdr)
			conn.Write([]byte("ack"))
		},
		OnDisconnect: func(id string) { fmt.Println("[server] client disconnected") },
	}
	srv := server.New(addr, nil, cb, nil)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	return srv
}

func runClient() {
	cli := client.New(addr, nil, client.Callbacks{OnDisconnect: func() { fmt.Println("[client] disconnected") }}, nil)

	for {
		if err := cli.Connect(); err != nil {
			log.Println("connect:", err)
			time.Sleep(time.Second)
			continue
		}
		fmt.Println("[client] connected")

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		resp, data, err := cli.SendSync(ctx, &message.Message{}, []byte("ping"))
		cancel()
		if err != nil {
			log.Println("send failed:", err)
			cli.Disconnect()
			time.Sleep(time.Second)
			continue
		}
		fmt.Printf("[client] got response %q (%v)\n", string(data), resp.SyncResponse)
		cli.Disconnect()
		break
	}
}

func main() {
	srv := startServer()
	defer srv.Stop()

	// give the server a moment to start
	time.Sleep(500 * time.Millisecond)

	runClient()
}

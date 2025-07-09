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

// This example intentionally causes a send/receive deadlock. The client sends a
// synchronous request and waits for the server to reply. The server waits for
// another message from the client before sending any response, so both sides
// block. After the client's context times out, SendSync returns an error.
// Deadlocks like this can be avoided by clearly defining which side sends next
// or by using asynchronous handlers with timeouts.
func main() {
	var srvConn net.Conn

	srvCb := server.Callbacks{
		OnConnect: func(id string, c net.Conn) { srvConn = c },
		OnMessage: func(id string, msg *message.Message, data []byte) {
			fmt.Printf("server received: %s\n", string(data))
			fmt.Println("server waiting for another message before responding...")
			// Wait indefinitely for a second message that will never arrive.
			_, err := message.ParseHeader(srvConn)
			if err != nil {
				fmt.Println("server read:", err)
			}
		},
	}

	srv := server.New("127.0.0.1:9300", nil, srvCb, nil)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	defer srv.Stop()

	time.Sleep(200 * time.Millisecond)

	cli := client.New("127.0.0.1:9300", nil, client.Callbacks{}, nil)
	if err := cli.Connect(); err != nil {
		log.Fatal(err)
	}
	defer cli.Disconnect()

	fmt.Println("client sending ping and waiting for reply (will deadlock)...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, _, err := cli.SendSync(ctx, &message.Message{}, []byte("ping"))
	fmt.Println("SendSync returned:", err)
}

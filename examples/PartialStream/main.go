package main

import (
	"bytes"
	"io"
	"log"
	"net"
	"time"

	"github.com/WasimAhmad/watsontcp-go/client"
	"github.com/WasimAhmad/watsontcp-go/message"
	"github.com/WasimAhmad/watsontcp-go/server"
)

// This example demonstrates sending a stream in multiple segments
// using io.LimitedReader. The server writes data to the client in
// small chunks with a pause between each chunk.
func main() {
	data := []byte("Streaming data from server in partial chunks using LimitedReader.")

	var srv *server.Server

	srvCb := server.Callbacks{
		OnConnect: func(id string, c net.Conn) {
			go func() {
				r := bytes.NewReader(data)
				// send 16 bytes at a time
				for r.Len() > 0 {
					lr := &io.LimitedReader{R: r, N: 16}
					if err := srv.SendStream(id, &message.Message{}, lr, lr.N); err != nil {
						log.Println("send stream:", err)
						return
					}
					time.Sleep(500 * time.Millisecond)
				}
			}()
		},
	}

	srv = server.New("127.0.0.1:9300", nil, srvCb, nil)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	defer srv.Stop()

	cliCb := client.Callbacks{
		OnStream: func(msg *message.Message, r io.Reader) {
			buf, _ := io.ReadAll(r)
			log.Printf("client received chunk: %q", string(buf))
		},
	}

	cli := client.New("127.0.0.1:9300", nil, cliCb, nil)
	if err := cli.Connect(); err != nil {
		log.Fatal(err)
	}
	defer cli.Disconnect()

	select {}
}

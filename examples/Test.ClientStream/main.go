package main

import (
	"io"
	"log"
	"os"

	"github.com/WasimAhmad/watsontcp-go/client"
	"github.com/WasimAhmad/watsontcp-go/message"
	"github.com/WasimAhmad/watsontcp-go/server"
)

func main() {
	// Prepare some data to stream from the client
	data := []byte("streaming data from client")
	os.WriteFile("client_stream.txt", data, 0644)

	// Server that reads a stream from the client
	srvCb := server.Callbacks{
		OnStream: func(id string, msg *message.Message, r io.Reader) {
			b, err := io.ReadAll(r)
			if err != nil {
				log.Println("read stream:", err)
				return
			}
			log.Printf("[server] received %d bytes: %s", len(b), string(b))
		},
	}

	srv := server.New("127.0.0.1:9300", nil, srvCb, nil)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	defer srv.Stop()

	// Client connects and sends the file using SendStream
	cli := client.New("127.0.0.1:9300", nil, client.Callbacks{}, nil)
	if err := cli.Connect(); err != nil {
		log.Fatal(err)
	}
	defer cli.Disconnect()

	f, err := os.Open("client_stream.txt")
	if err != nil {
		log.Fatal(err)
	}
	fi, _ := f.Stat()
	cli.SendStream(&message.Message{}, f, fi.Size())
	f.Close()

	select {}
}

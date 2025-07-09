package watsontcpgo_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/WasimAhmad/watsontcp-go/client"
	"github.com/WasimAhmad/watsontcp-go/message"
	"github.com/WasimAhmad/watsontcp-go/server"
)

func newTLSConfig() (*tls.Config, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	tmpl := x509.Certificate{SerialNumber: big.NewInt(1), NotBefore: time.Now(), NotAfter: time.Now().Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	key := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	tlsCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}, InsecureSkipVerify: true}, nil
}

func startServer(t *testing.T, addr string, tlsConf *tls.Config, cb server.Callbacks) *server.Server {
	srv := server.New(addr, tlsConf, cb, nil)
	if err := srv.Start(); err != nil {
		t.Fatalf("server start: %v", err)
	}
	return srv
}

func TestClientServerCommunication(t *testing.T) {
	done := make(chan struct{})
	cb := server.Callbacks{
		OnMessage: func(id string, msg *message.Message, data []byte) { close(done) },
	}
	srv := startServer(t, "127.0.0.1:30000", nil, cb)
	defer srv.Stop()

	cli := client.New("127.0.0.1:30000", nil, client.Callbacks{}, nil)
	if err := cli.Connect(); err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer cli.Disconnect()

	if err := cli.Send(&message.Message{}, []byte("hi")); err != nil {
		time.Sleep(200 * time.Millisecond)
		t.Fatalf("send: %v", err)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("message not received")
	}
}

func TestTLSConnection(t *testing.T) {
	tlsConf, err := newTLSConfig()
	if err != nil {
		t.Fatalf("tls: %v", err)
	}
	srv := startServer(t, "127.0.0.1:30001", tlsConf, server.Callbacks{})
	defer srv.Stop()

	cli := client.New("127.0.0.1:30001", &tls.Config{InsecureSkipVerify: true}, client.Callbacks{}, nil)
	if err := cli.Connect(); err != nil {
		t.Fatalf("connect: %v", err)
	}
	cli.Disconnect()
}

func TestSendSync(t *testing.T) {
	var conn net.Conn
	ready := make(chan struct{})
	cb := server.Callbacks{
		OnConnect: func(id string, c net.Conn) {
			conn = c
			close(ready)
		},
		OnMessage: func(id string, msg *message.Message, data []byte) {
			if msg.SyncRequest {
				resp := &message.Message{SyncResponse: true, ConversationGUID: msg.ConversationGUID, ContentLength: int64(len("pong"))}
				hdr, _ := message.BuildHeader(resp)
				conn.Write(hdr)
				conn.Write([]byte("pong"))
			}
		},
	}
	srv := startServer(t, "127.0.0.1:30100", nil, cb)
	defer srv.Stop()

	cli := client.New("127.0.0.1:30100", nil, client.Callbacks{}, nil)
	if err := cli.Connect(); err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer cli.Disconnect()
	<-ready

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, data, err := cli.SendSync(ctx, &message.Message{}, []byte("ping"))
	if err != nil {
		t.Fatalf("SendSync: %v", err)
	}
	if string(data) != "pong" || !resp.SyncResponse {
		t.Fatalf("bad response")
	}
}

func TestClientIdleDisconnect(t *testing.T) {
	disc := make(chan struct{})
	srv := startServer(t, "127.0.0.1:30101", nil, server.Callbacks{})
	defer srv.Stop()

	opts := client.DefaultOptions()
	opts.IdleTimeout = 200 * time.Millisecond
	opts.EvaluationInterval = 50 * time.Millisecond
	cli := client.New("127.0.0.1:30101", nil, client.Callbacks{OnDisconnect: func() { close(disc) }}, &opts)
	if err := cli.Connect(); err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer cli.Disconnect()

	select {
	case <-disc:
	case <-time.After(2 * time.Second):
		t.Fatalf("client did not disconnect after idle timeout")
	}
}

func TestServerIdleDisconnect(t *testing.T) {
	disc := make(chan struct{})
	opts := server.DefaultOptions()
	opts.IdleTimeout = 200 * time.Millisecond
	opts.CheckInterval = 50 * time.Millisecond
	cb := server.Callbacks{OnDisconnect: func(id string) { close(disc) }}
	srv := server.New("127.0.0.1:30102", nil, cb, &opts)
	if err := srv.Start(); err != nil {
		t.Fatalf("server start: %v", err)
	}
	defer srv.Stop()

	cli := client.New("127.0.0.1:30102", nil, client.Callbacks{}, nil)
	if err := cli.Connect(); err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer cli.Disconnect()

	select {
	case <-disc:
	case <-time.After(2 * time.Second):
		t.Fatalf("server did not disconnect idle client")
	}
}

func TestPresharedKeySuccess(t *testing.T) {
	done := make(chan struct{})
	srvOpts := server.DefaultOptions()
	srvOpts.PresharedKey = "secret"
	cb := server.Callbacks{OnMessage: func(id string, msg *message.Message, data []byte) { close(done) }}
	srv := server.New("127.0.0.1:30110", nil, cb, &srvOpts)
	if err := srv.Start(); err != nil {
		t.Fatalf("server start: %v", err)
	}
	defer srv.Stop()

	cliOpts := client.DefaultOptions()
	cliOpts.PresharedKey = "secret"
	cli := client.New("127.0.0.1:30110", nil, client.Callbacks{}, &cliOpts)
	if err := cli.Connect(); err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer cli.Disconnect()
	if err := cli.Send(&message.Message{}, []byte("hi")); err != nil {
		t.Fatalf("send: %v", err)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("message not received")
	}
}

func TestPresharedKeyFailure(t *testing.T) {
	srvOpts := server.DefaultOptions()
	srvOpts.PresharedKey = "secret"
	srv := server.New("127.0.0.1:30111", nil, server.Callbacks{}, &srvOpts)
	if err := srv.Start(); err != nil {
		t.Fatalf("server start: %v", err)
	}
	defer srv.Stop()

	cliOpts := client.DefaultOptions()
	cliOpts.PresharedKey = "wrong"
	cli := client.New("127.0.0.1:30111", nil, client.Callbacks{}, &cliOpts)
	if err := cli.Connect(); err == nil {
		cli.Disconnect()
		t.Fatalf("expected auth failure")
	}
}

func TestMaxConnectionsLimit(t *testing.T) {
	opts := server.DefaultOptions()
	opts.MaxConnections = 1
	srv := server.New("127.0.0.1:30120", nil, server.Callbacks{}, &opts)
	if err := srv.Start(); err != nil {
		t.Fatalf("server start: %v", err)
	}
	defer srv.Stop()

	cli1 := client.New("127.0.0.1:30120", nil, client.Callbacks{}, nil)
	if err := cli1.Connect(); err != nil {
		t.Fatalf("cli1 connect: %v", err)
	}
	defer cli1.Disconnect()

	cli2 := client.New("127.0.0.1:30120", nil, client.Callbacks{}, nil)
	if err := cli2.Connect(); err == nil {
		// attempt a send to verify connection
		err = cli2.Send(&message.Message{}, []byte("test"))
		if err == nil {
			cli2.Disconnect()
			t.Fatalf("expected connection limit to reject second client")
		}
	}
}

func TestIPBlocking(t *testing.T) {
	opts := server.DefaultOptions()
	opts.BlockedIPs = []string{"127.0.0.1"}
	srv := server.New("127.0.0.1:30121", nil, server.Callbacks{}, &opts)
	if err := srv.Start(); err != nil {
		t.Fatalf("server start: %v", err)
	}
	defer srv.Stop()

	cli := client.New("127.0.0.1:30121", nil, client.Callbacks{}, nil)
	if err := cli.Connect(); err == nil {
		err = cli.Send(&message.Message{}, []byte("test"))
		if err == nil {
			cli.Disconnect()
			t.Fatalf("expected connection to be blocked")
		}
	}
}

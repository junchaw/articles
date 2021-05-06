package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spongeprojects/magicconch"
	"net"
)

type ProbeHandlerFunc func(string) string

type ProbeServer struct {
	handlerFunc  ProbeHandlerFunc
	clients      map[*ProbeClient]bool
	registerCh   chan *ProbeClient
	unregisterCh chan *ProbeClient
}

type ProbeClient struct {
	socket net.Conn
	data   chan []byte
}

func (server *ProbeServer) start() {
	for {
		select {
		case client := <-server.registerCh:
			server.clients[client] = true
		case client := <-server.unregisterCh:
			if _, ok := server.clients[client]; ok {
				client.socket.Close()
				delete(server.clients, client)
			}
		}
	}
}

func (server *ProbeServer) receive(client *ProbeClient) {
	for {
		message := make([]byte, 4096)
		l, err := client.socket.Read(message)
		if err != nil {
			server.unregisterCh <- client
			break
		}
		_, err = client.socket.Write([]byte(server.handlerFunc(string(message[:l]))))
		if err != nil {
			server.unregisterCh <- client
			break
		}
	}
}

func startProbeServer(handlerFunc ProbeHandlerFunc) {
	fmt.Println("Starting probe server...")

	listener, err := net.Listen("tcp", ":12345")
	magicconch.Must(err)

	manager := ProbeServer{
		handlerFunc:  handlerFunc,
		clients:      make(map[*ProbeClient]bool),
		registerCh:   make(chan *ProbeClient),
		unregisterCh: make(chan *ProbeClient),
	}

	go manager.start()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(errors.Wrap(err, "accept connection error"))
		}
		client := &ProbeClient{socket: conn}
		manager.registerCh <- client
		go manager.receive(client)
	}
}

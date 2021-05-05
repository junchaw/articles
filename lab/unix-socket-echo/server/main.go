package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spongeprojects/magicconch"
	"net"
)

type ClientManager struct {
	clients      map[*Client]bool
	registerCh   chan *Client
	unregisterCh chan *Client
}

type Client struct {
	socket net.Conn
	data   chan []byte
}

func (manager *ClientManager) start() {
	for {
		select {
		case client := <-manager.registerCh:
			manager.clients[client] = true
			fmt.Println("[REGISTERED]: Client registered!")
		case client := <-manager.unregisterCh:
			if _, ok := manager.clients[client]; ok {
				close(client.data)
				delete(manager.clients, client)
				fmt.Println("[UNREGISTERED]: Client unregistered!")
			}
		}
	}
}

func (manager *ClientManager) receive(client *Client) {
	for {
		message := make([]byte, 4096)
		_, err := client.socket.Read(message)
		if err != nil {
			manager.unregisterCh <- client
			client.socket.Close()
			break
		}
		select {
		case client.data <- message:
		default:
			close(client.data)
			delete(manager.clients, client)
		}
	}
}

func (manager *ClientManager) send(client *Client) {
	defer client.socket.Close()
	for {
		select {
		case message, ok := <-client.data:
			if !ok {
				return
			}
			client.socket.Write(message)
		}
	}
}

func main() {
	fmt.Println("Starting server...")

	listener, err := net.Listen("tcp", ":12345")
	magicconch.Must(err)

	manager := ClientManager{
		clients:      make(map[*Client]bool),
		registerCh:   make(chan *Client),
		unregisterCh: make(chan *Client),
	}

	go manager.start()

	fmt.Println("[WAITING]")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(errors.Wrap(err, "accept connection error"))
		}
		client := &Client{socket: conn, data: make(chan []byte)}
		manager.registerCh <- client
		go manager.receive(client)
		go manager.send(client)
	}
}

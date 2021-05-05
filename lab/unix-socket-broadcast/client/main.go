package main

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spongeprojects/magicconch"
	"net"
	"os"
	"strings"
)

type Client struct {
	socket net.Conn
}

func (client *Client) receive() {
	defer func(socket net.Conn) {
		err := socket.Close()
		if err != nil {
			fmt.Println(errors.Wrap(err, "close socket error"))
		}
	}(client.socket)

	for {
		message := make([]byte, 4096)
		length, err := client.socket.Read(message)
		if err != nil {
			break
		}
		if length > 0 {
			fmt.Println("[RECEIVED]: " + string(message))
		}
	}
}

func main() {
	fmt.Println("Starting client...")

	conn, err := net.Dial("tcp", "localhost:12345")
	magicconch.Must(err)

	client := &Client{socket: conn}

	go client.receive()

	fmt.Println("[WAITING]")
	for {
		reader := bufio.NewReader(os.Stdin)
		message, _ := reader.ReadString('\n')
		message = strings.Trim(message, "\n")
		fmt.Println("[SENDING]: " + message)
		_, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Println(errors.Wrap(err, "send message error"))
		}
	}
}

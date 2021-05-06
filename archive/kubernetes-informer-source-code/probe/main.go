package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spongeprojects/magicconch"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("missing argument: message")
		return
	}

	message := os.Args[1]

	conn, err := net.Dial("tcp", "localhost:12345")
	magicconch.Must(err)

	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println(errors.Wrap(err, "send message error"))
	}

	response := make([]byte, 4096)
	_, err = conn.Read(response)
	magicconch.Must(err)

	fmt.Println("Current state:\n")
	fmt.Println(string(response))
}

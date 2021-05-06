package main

import (
	"fmt"
	"strings"
)

func main() {
	a := make([]byte, 16)

	r := strings.NewReader("test")
	_, _ = r.Read(a)

	b := "test"

	fmt.Println(string(a), b)
	fmt.Println(string(a) == b)
	fmt.Println(len(string(a)), len(b))
	fmt.Println(a, []byte(b))
}

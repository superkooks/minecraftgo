package main

import (
	"fmt"
	"net"
)

func main() {
	addr, err := net.ResolveTCPAddr("tcp4", "yumyumserver.ddnsfree.com:25565")
	if err != nil {
		panic(err)
	}

	fmt.Println("Connecting")
	b, err := Ping(addr)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

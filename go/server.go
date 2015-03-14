package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("HELLO I AM STARTED");

	if (len(os.Args) != 2) {
		fmt.Println("RTFM")
	}
	javaPort := os.Args[1]
	fmt.Println("Trying to connect to java on localhost:" + javaPort)
	addr, err := net.ResolveUDPAddr("udp", ":0")
	checkError(err)

	conn, err := net.ListenUDP("udp", addr)
	checkError(err)

	buf := make([]byte, 1024)
	for {
		message := "my message"
		buf = []byte(message)
		raddr, _ := net.ResolveUDPAddr("udp", ":" + javaPort)
		_, _ = conn.WriteToUDP(buf[:len(message)], raddr)
		fmt.Println("DONE")
		break
	}
}

// If error is non-nil, print it out and halt.
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		os.Exit(1)
	}
}

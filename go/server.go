package main

import "net"
import "fmt"
import "log"
import "time"

func main() {

	addr := net.UDPAddr{
		Port: 1200,
		IP: net.ParseIP("127.0.0.1"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println("here 0")
		log.Fatal(err)
	}
	defer conn.Close()

	var buf []byte = make([]byte, 1500)
	for {
		time.Sleep(100 * time.Millisecond)
		n, addr, err := conn.ReadFromUDP(buf)
		fmt.Println("msg: ", string(buf[0:n]))

		if err != nil {
			log.Fatal(err)
		}

		n, err = conn.WriteToUDP(buf, addr)
		if err != nil {
			log.Fatal(err)
		}
	}
}

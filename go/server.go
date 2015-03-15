package main

import (
	"encoding/json"
	"fmt"
//	"net"
	"os"
//	"time"
)


type GameState struct {
	Round int
}
var gameState GameState

type MyMove struct {
	X int					`json:"x"`
	Y int					`json:"y"`
	Direction string		`json:"direction"`
	Pid int					`json:"pid"`
}

type RoundStart struct {
	Round int				`json:"round"`
	Pid int					`json:"pid"`
}

type RoundStartMessage struct {
	EventName string		`json:"eventName"`
	RoundStart				`json:"roundStart"`
}

type MyMoveMessage struct {
	EventName string		`json:"eventName"`
	MyMove					`json:"myMove"`
}



func newRoundMessage() []byte {
	gameState.Round += 1
	message := RoundStartMessage{EventName:"roundStart", RoundStart{Round:gameState.Round}}
	return json.Marshal(message)
}


func parseMessage(buf []byte) interface{} {
	var dat map[string]interface{}
	err := json.Unmarshal(buf, &dat)
	checkError(err)

	// TODO: error handling, default actions
	eventName := dat["eventName"].(string)

	switch eventName {
	case "myMove":
		res := &MyMoveMessage{}
	default:
		res := &MyMoveMessage{}
	}

	json.Unmarshal(buf, &res)

	return res
}


func encodeMessage(message interface{}) []byte {
	return json.Marshal(message)
}

func main() {
	fmt.Println("HELLO I AM STARTED");

	if (len(os.Args) != 3) {
		fmt.Println("RTFM")
	}
	javaPort := os.Args[1]
	leaderAddr := os.Args[2]

	gameState.Round = 1

	timeToReply := true // TODO: set by ticker or all clients replied


	fmt.Println("Trying to connect to java on localhost:" + javaPort)


	raddr, err := net.ResolveUDPAddr("udp", "localhost:" + javaPort)

	// handle communication between other go clients
	go func() {
		addr, err := net.ResolveUDPAddr("udp", ":0")
		checkError(err)
		conn, err := net.ListenUDP("udp", addr)
		checkError(err)
		buf := make([]byte, 1024)
		for {
			_, raddr, err := goConn.ReadFromUDP(buf)
			checkError(err)


			// TODO: only write back after receiving multiple replies, or after ticker timeout

		}
	}()

	// handle internal communication to java game
	go func() {
		addr, err := net.ResolveUDPAddr("udp", ":0")
		checkError(err)
		conn, err := net.ListenUDP("udp", addr)
		checkError(err)
		for {
			if isNewRound {
				// create message that new round is starting
				byt := newRoundMessage(gameState)

				// send message to everyone that new round has started
				conn.WriteToUDP(byt, raddr)
				isNewRound = false
			}

			// read some reply from the java game (update of move, or death)
			_, raddr, err := conn.ReadFromUDP(buf)
			checkError(err)

			// parse message, figure out if it is myMove message or something else 
			message := parseMessage(buf)

			// TODO: only write back after receiving multiple replies, or after ticker timeout
			if timeToReply {
				// send out response of moves to take (and TODO: death updates if they exist)
				byt, err := encodeMessage(message)
				checkError(err)

				_, err = conn.WriteToUDP(byt, raddr)
				checkError(err)

				isNewRound = true
			}

			time.Sleep(1 * time.Second)
		}
	}()

	
	fmt.Println("GOODBYE")



	/*
	mapD := map[string]int{"apple": 5, "lettuce": 7}
    mapB, _ := json.Marshal(mapD)
    fmt.Println(string(mapB))


	msgD := MyMoveMessage{
		EventName: "myMove",
		MyMove: MyMove{X: 50, Y:10, Direction:"left", Pid:1}}

	msgB, err := json.Marshal(msgD)
	checkError(err)

	fmt.Println(string(msgB))

	str := `{"eventName":"myMove","myMove":{"x":50,"y":10,"direction":"left","pid":1}}`
	res := MyMoveMessage{}

	err = json.Unmarshal([]byte(str), &res)
	checkError(err)

	fmt.Println(res.EventName)
	fmt.Println(res.MyMove.X)
	*/

}

// If error is non-nil, print it out and halt.
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		os.Exit(1)
	}
}

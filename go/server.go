package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
	"bytes"
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

type MovesMessage struct {
	EventName string        `json:"eventName"`
	Moves []MyMove                `json:"moves"`
}


func newRoundMessage() []byte {
	gameState.Round += 1
	message := RoundStartMessage{EventName:"roundStart", RoundStart: RoundStart{Round:gameState.Round, Pid: 1}}
	val, e := json.Marshal(message)
	checkError(e)
	return val
}


func parseMessage(buf []byte) interface{} {
	fmt.Println("Parsing the following message " + string(buf))
	var dat map[string]interface{}
	err := json.Unmarshal(buf, &dat)
	checkError(err)

	fmt.Println("dat: ", dat)

	// TODO: error handling, default actions
	eventName := dat["eventName"].(string)
	var res interface{}
	switch eventName {
	case "myMove":
		x, e := dat["x"].(float64)
		fmt.Println(e)
		y, e := dat["y"].(float64)
		d, e := dat["direction"].(string)
		p, e := dat["pid"].(float64)
		res = MovesMessage{EventName:"moves", Moves: []MyMove{MyMove{X:int(x), Y:int(y), Direction:d, Pid:int(p)}}}

	default:
		x, e := dat["x"].(float64)
		fmt.Println(e)
		y, e := dat["y"].(float64)
		d, e := dat["direction"].(string)
		p, e := dat["pid"].(float64)
		res = MyMoveMessage{EventName:"myMove", MyMove: MyMove{X:int(x), Y:int(y), Direction:d, Pid:int(p)}}
	}

	//json.Unmarshal(buf, &res)
	fmt.Println("parsed message: ", res)
	return res
}


func encodeMessage(message interface{}) []byte {
	val, e := json.Marshal(message)
	checkError(e)
	return val
}

func main() {
	fmt.Println("HELLO I AM STARTED");

	if (len(os.Args) != 3) {
		fmt.Println("RTFM")
	}
	javaPort := os.Args[1]
	leaderAddr := os.Args[2]
	fmt.Println("Leadder address is " + leaderAddr)

	gameState.Round = 1

	timeToReply := true // TODO: set by ticker or all clients replied


	fmt.Println("Trying to connect to java on localhost:" + javaPort)


	raddr, err := net.ResolveUDPAddr("udp", "localhost:" + javaPort)
	checkError(err)

	// handle communication between other go clients
	go func() {
		addr, err := net.ResolveUDPAddr("udp", ":0")
		checkError(err)
		conn, err := net.ListenUDP("udp", addr)
		checkError(err)
		buf := make([]byte, 4096)
		for {
			_, raddr, err := conn.ReadFromUDP(buf)
			fmt.Println(raddr)
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
		var isNewRound = true
		var buf = make([]byte, 4096)
		for {
			if isNewRound {
				// create message that new round is starting
				byt := newRoundMessage()

				// send message to everyone that new round has started
				fmt.Println("Sending the following round start message " + string(byt))
				conn.WriteToUDP(byt, raddr)
				isNewRound = false
			}

			// read some reply from the java game (update of move, or death)
			_, raddr, err := conn.ReadFromUDP(buf)
			buf = bytes.Trim(buf, "\x00")
			checkError(err)

			// parse message, figure out if it is myMove message or something else 
			message := parseMessage(buf)

			// TODO: only write back after receiving multiple replies, or after ticker timeout
			if timeToReply {
				// send out response of moves to take (and TODO: death updates if they exist)
				byt := encodeMessage(message)
				checkError(err)

				_, err = conn.WriteToUDP(byt, raddr)
				checkError(err)

				isNewRound = true
			}

			time.Sleep(200 * time.Millisecond)
		}
	}()
	for {
		time.Sleep(100)
	}
	
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

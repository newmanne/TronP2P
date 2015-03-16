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
	Round int				'json:"round"'
	RoundStart				`json:"roundStart"`
}

type MyMoveMessage struct {
	EventName string		`json:"eventName"`
	Round int				'json:"round"'
	MyMove					`json:"myMove"`
}

type MovesMessage struct {
	EventName string		`json:"eventName"`
	Round int				'json:"round"'
	Moves []MyMove			`json:"moves"`
}


func newRoundMessage() []byte {
	gameState.Round += 1
	message := RoundStartMessage{EventName:"roundStart", Round:gameState.Round, RoundStart: RoundStart{Round:gameState.Round, Pid: 1}}
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
		res = MovesMessage{EventName:"moves", Round:gameState.Round, Moves: []MyMove{MyMove{X:int(x), Y:int(y), Direction:d, Pid:int(p)}}}

	default:
		x, e := dat["x"].(float64)
		fmt.Println(e)
		y, e := dat["y"].(float64)
		d, e := dat["direction"].(string)
		p, e := dat["pid"].(float64)
		res = MovesMessage{EventName:"moves", Round:gameState.Round, Moves: []MyMove{MyMove{X:int(x), Y:int(y), Direction:d, Pid:int(p)}}}
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


// 
func leaderListener(leaderAddr string) {

	playerCount := 2	// somehow know how many people I am waiting for
	var responses map[int]bool // map pid -> response received

	addr, err := net.ResolveUDPAddr("udp", leaderAddr)
	checkError(err)
	conn, err := net.ListenUDP("udp", "localhost:" + addr.Port)
	checkError(err)

	var reply MovesMessage
	raddrs := make([]string, 2)

	// LOBBY PHASE
	// before the general main loop, wait for playerCount messages,
	// this will tell me who I need to send roundStarts to.
	for {
		var buf = make([]byte, 4096)
		_, raddr, err := conn.ReadFromUDP(buf)
		buf = bytes.Trim(buf, "\x00")
		checkError(err)

		in := false
		for _, item := range raddrs {
			if item == raddr {
				in = true
			}
		}
		if !in {
			raddrs = append(raddrs, raddr)
		}
		if len(raddrs) == playerCount {
			break // we can start, and send out a new round message
		}
	}

	// MAIN LOOP SECTION
	isNewRound := true
	for {
		// read a message from some follower
		var buf = make([]byte, 4096)
		_, raddr, err := conn.ReadFromUDP(buf)
		buf = bytes.Trim(buf, "\x00")
		checkError(err)

		// if a new round is starting, let everyone connected to me know
		if isNewRound {
			// create message that new round is starting
			byt := newRoundMessage()

			// send message to everyone that new round has started
			fmt.Println("Sending the following round start message " + string(byt))

			for i := range raddrs {
				_, err := conn.WriteToUDP(byt, raddrs[i])
				checkError(err)		
			}

			isNewRound = false
			responses = make(map[int]bool)
			raddrs = make([]string, 2)

			reply = MovesMessage{EventName:"moves", Round:gameState.Round, Moves: make([]MyMove)}
		}

		// parse the new message into a MovesMessage struct (usually)
		commands := parseMessage(buf)

		// append moves received to list of all moves recieved for current round
		if commands.EventName == "moves" && commands.Round == gameState.Round {
			for index, move := range commands.Moves {
				reply.Moves = append(reply.Moves, move)
				responses[move.Pid] = true
			}
			// keep track of who to respond to
			raddrs = append(raddrs, raddr)
		}

		// end condition; reply to my followers if I have been messaged by all of them
		if len(responses) == playerCount {
			byt := encodeMessage(reply)

			// send message to all followers
			for i := range raddrs {
				_, err = conn.WriteToUDP(byt, raddrs[i])
				checkError(err)
			}
			// start a new round of communication
			isNewRound = true
		}
	}
}

func main() {
	fmt.Println("HELLO I AM STARTED");

	if (len(os.Args) != 4) {
		fmt.Println("RTFM")
	}
	javaPort := os.Args[1]
	leaderAddr := os.Args[2]
	isLeader := os.Args[3]

	fmt.Println("Leadder address is " + leaderAddr)
	gameState.Round = 1

	timeToReply := true // TODO: set by ticker or all clients replied

	sendChan := make(chan string, 1)
	recvChan := make(chan string, 1)

	fmt.Println("Trying to connect to java on localhost:" + javaPort)

	raddr, err := net.ResolveUDPAddr("udp", "localhost:" + javaPort)
	checkError(err)

	// if I am the leader, listen for rounds to confirm them
	if isLeader {
		go leaderListener(leaderAddr)
	}

	// handle communication through leader channel (follower code to leader)
	go func(sendChan <-chan []byte, recvChan chan<- []byte) {
		addr, err := net.ResolveUDPAddr("udp", ":0")
		checkError(err)
		conn, err := net.ListenUDP("udp", addr)
		checkError(err)

		// write message to leader to let them know I exist (LOBBY PHASE)
		_, err = conn.WriteToUDP([]byte("hello leader"), leaderAddr)
		checkError(err)

		// read response from leader
		response := make([]byte, 4096)
		_, _, err := conn.ReadFromUDP(response)
		checkError(err)

		// write back to channel with byte response (let java know to start)
		recvChan <- response

		for {
			// wait for message on leader channel
			message := <- sendChan // TODO make this

			// write message to leader address
			_, err = conn.WriteToUDP(message, leaderAddr)
			checkError(err)

			// read response from leader
			response = make([]byte, 4096)
			_, _, err := conn.ReadFromUDP(response)
			checkError(err)

			// write back to channel with byte response
			recvChan <- response
		}
	}(sendChan, recvChan)

	// handle internal communication to java game
	go func(sendChan chan<- []byte, recvChan <-chan []byte) {
		addr, err := net.ResolveUDPAddr("udp", ":0")
		checkError(err)
		conn, err := net.ListenUDP("udp", addr)
		checkError(err)

		// LOBBY PHASE; need to recv response from leader,
		// pass message onto java side (first round start)

		reply := <- recvChan
		_, err = conn.WriteToUDP(reply, raddr)
		checkError(err)

		for {
			// read some reply from the java game (update of move, or death)
			var buf = make([]byte, 4096)
			_, raddr, err := conn.ReadFromUDP(buf)
			buf = bytes.Trim(buf, "\x00")
			checkError(err)

			// send buf to leader channel
			sendChan <- buf

			// read reply (timeToReply?) from leader (TODO: use a select w/ timeout?)
			reply := <- recvChan

			// TODO: only write back after receiving multiple replies, or after ticker timeout
			if timeToReply {
				_, err = conn.WriteToUDP(reply, raddr)
				checkError(err)
			}

			time.Sleep(10 * time.Millisecond)
		}
	}(sendChan, recvChan)

	// busy function forever
	for {
		time.Sleep(100)
	}
	
	fmt.Println("GOODBYE")

}

// If error is non-nil, print it out and halt.
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		os.Exit(1)
	}
}

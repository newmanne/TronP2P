package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
	"sync"
)



type MyMove struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Direction string `json:"direction"`
	Pid       int    `json:"pid"`
}

type GameState struct {
	Round int
	Positions map[int]MyMove
	AddrToPid map[*net.UDPAddr]int
}
var gameState GameState


type RoundStart struct {
	Round int `json:"round"`
	Pid   int `json:"pid"`
}

type RoundStartMessage struct {
	EventName  string `json:"eventName"`
	Round      int    `json:"round"`
	RoundStart `json:"roundStart"`
}

type MyMoveMessage struct {
	EventName string `json:"eventName"`
	Round     int    `json:"round"`
	MyMove    `json:"myMove"`
}

type MovesMessage struct {
	EventName string   `json:"eventName"`
	Round     int      `json:"round"`
	Moves     []MyMove `json:"moves"`
}
	

func newRoundMessage() []byte {
	gameState.Round += 1
	message := RoundStartMessage{EventName: "roundStart", Round: gameState.Round, RoundStart: RoundStart{Round: gameState.Round, Pid: 1}}
	val, e := json.Marshal(message)
	checkError(e)
	return val
}

func parseMessage(buf []byte) MovesMessage {
	fmt.Println("Parsing the following message " + string(buf))
	var dat map[string]interface{}
	err := json.Unmarshal(buf, &dat)
	checkError(err)

	fmt.Println("dat: ", dat)

	// TODO: error handling, default actions
	eventName := dat["eventName"].(string)
	var res MovesMessage
	switch eventName {
	case "myMove":
		x, e := dat["x"].(float64)
		fmt.Println(e)
		y, e := dat["y"].(float64)
		d, e := dat["direction"].(string)
		p, e := dat["pid"].(float64)
		res = MovesMessage{EventName: "moves", Round: gameState.Round, Moves: []MyMove{MyMove{X: int(x), Y: int(y), Direction: d, Pid: int(p)}}}

	default:
		fmt.Println("ERROR! This should not happen.");
		/*x, e := dat["x"].(float64)
		fmt.Println(e)
		y, e := dat["y"].(float64)
		d, e := dat["direction"].(string)
		p, e := dat["pid"].(float64)
		res = MovesMessage{EventName: "moves", Round: gameState.Round, Moves: []MyMove{MyMove{X: int(x), Y: int(y), Direction: d, Pid: int(p)}}}*/
	}

	fmt.Println("parsed message: ", res)
	return res
}

func encodeMessage(message interface{}) []byte {
	val, e := json.Marshal(message)
	checkError(e)
	return val
}



// TODO: randomize X, Y, Direction based on map size, and other player positions
func CreateInitPlayerPosition(pid int) MyMove {
	return MyMove{X: 0, Y: 0, Direction:"UP", Pid: pid}
}


// TODO: parse message, check if eventName == join
func isJoinMessage(buf []byte) bool {
	return true	
}

// TODO: parse message, check if eventName == start, and pid/addr == monarch?
func isStartMessage(buf []byte) bool {
	return true
}


// TODO: given pid, return message event == startGame (or something), that includes
// all player start positions, this players pid, and their leader they refer to from now on
func startGameMessage(pid int) []byte {
	return nil
}

/*
 At the end of the lobby session, the following things must be true:
 1) all players know the IP, pid, and start positions of all other players
 2) all players know their immediate leader
 3) all players acknowledge game begins?

 General structure:
 - monarch sarts lobby session
 - monarch is pid=1
 - as players join, they are assigned pid in order of arrival
 - monarch closes lobby with start game command
 - upon end of lobby session, monarch sends req info to all players (TCP?)
 - once this is done, lobby session ends and monarch starts game.
*/
func initLobby(conn *net.UDPConn, playerCount int, raddrs []*net.UDPAddr) {
	
	// start of new game
	gameState.Positions = make(map[int]MyMove)
	gameState.AddrToPid = make(map[*net.UDPAddr]int)

	for {
		// wait for message from some client
		var buf = make([]byte, 4096)
		fmt.Println("Waiting for a hello message")
		_, raddr, err := conn.ReadFromUDP(buf)
		buf = bytes.Trim(buf, "\x00")
		checkError(err)
		fmt.Println("Leader has received a hello message")

		// what type of message is it? join or start game?
		// TODO: decide if message is hello or start game.

		if isJoinMessage(buf) {
			// if message is hello
			if _, knownPlayer := gameState.AddrToPid[raddr]; !knownPlayer {
				pid := len(gameState.Positions) + 1
				gameState.Positions[pid] = CreateInitPlayerPosition(pid)
				gameState.AddrToPid[raddr] = pid
			}
		}
		else if isStartMessage(buf) {
			// if message is start game

			// send message to all players to start game
			for player, pid := range gameState.AddrToPid {
				// TODO: use TCP to confirm message is received?
				newGameMsg := startGameMessage(pid)
				_, err = conn.WriteToUDP(newGameMsg, player)
				checkError(err)
			}
			// end lobby phase, prepare to send round messages next
			break
		}
	}
}

func newRound(conn *net.UDPConn, raddrs []*net.UDPAddr) (reply MovesMessage, responses map[int]bool) {
	// create message that new round is starting
	byt := newRoundMessage()
	
	// send message to everyone that new round has started
	fmt.Println("Sending the following round start message " + string(byt))
	
	for i := range raddrs {
		_, err := conn.WriteToUDP(byt, raddrs[i])
		checkError(err)
	}
	
	responses = make(map[int]bool)
	
	reply = MovesMessage{EventName: "moves", Round: gameState.Round, Moves: make([]MyMove, 0)}
	return reply, responses
}

func leaderListener(leaderAddr string) {

	playerCount := 1           // somehow know how many people I am waiting for
	var responses map[int]bool // map pid -> response received

	addr, err := net.ResolveUDPAddr("udp", leaderAddr)
	checkError(err)
	addr2, err := net.ResolveUDPAddr("udp", "localhost:"+strconv.Itoa(addr.Port))
	checkError(err)
	conn, err := net.ListenUDP("udp", addr2)
	checkError(err)
	fmt.Println("Leader has started")

	var reply MovesMessage
	raddrs := make([]*net.UDPAddr, 0)

	// LOBBY PHASE
	// before the general main loop, wait for playerCount messages,
	// this will tell me who I need to send roundStarts to.
	initLobby(conn, playerCount, raddrs)

	// MAIN LOOP SECTION
	isNewRound := true
	for {
		// if a new round is starting, let everyone connected to me know
		if isNewRound {
			reply, responses = newRound(conn, raddrs)
			isNewRound = false
			raddrs = make([]*net.UDPAddr, 0)
		}

		// read a message from some follower
		var buf = make([]byte, 4096)
		_, raddr, err := conn.ReadFromUDP(buf)
		buf = bytes.Trim(buf, "\x00")
		fmt.Println("DEBUG SERVER?", string(buf), conn.LocalAddr());
		checkError(err)

		// parse the new message into a MovesMessage struct (usually)
		commands := parseMessage(buf)

		// append moves received to list of all moves recieved for current round
		fmt.Println("DEBUG BUG?", string(encodeMessage(reply)));
		if commands.EventName == "moves" && commands.Round == gameState.Round {
			for _, move := range commands.Moves {
				reply.Moves = append(reply.Moves, move)
				responses[move.Pid] = true
			}
			// keep track of who to respond to
			raddrs = append(raddrs, raddr)
		}
		fmt.Println("DEBUG BUG?", string(encodeMessage(reply)));

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

func functionOne(sendChan, recvChan chan []byte, leaderAddr string, wg sync.WaitGroup) {
	defer wg.Done()
	addr, err := net.ResolveUDPAddr("udp", ":0")
	
	checkError(err)
	conn, err := net.ListenUDP("udp", addr)
	fmt.Println(conn.LocalAddr());
	checkError(err)

	// write message to leader to let them know I exist (LOBBY PHASE)
	leaderAddrAsAddr, err := net.ResolveUDPAddr("udp", leaderAddr)
	checkError(err)

	fmt.Println("Sending a hello message to the leader")
	_, err = conn.WriteToUDP([]byte("hello leader"), leaderAddrAsAddr)
	checkError(err)

	// read response from leader
	response := make([]byte, 4096)
	_, _, err = conn.ReadFromUDP(response)
	checkError(err)
	fmt.Println("Received a response from the leader:" + string(response))

	// write back to channel with byte response (let java know to start)
	recvChan <- response

	for {
		// wait for message on leader channel
		message := <-sendChan // TODO make this

		// write message to leader address
		//SECOND//
		fmt.Println("DEBUG SENDING MESSAGE?", string(message))
		_, err = conn.WriteToUDP(message, leaderAddrAsAddr)
		checkError(err)

		// read response from leader
		response = make([]byte, 4096)
		_, _, err := conn.ReadFromUDP(response)
		checkError(err)
		response = bytes.Trim(response, "\x00")
		//THIRD//
		fmt.Println("DEBUG RECEIVE RESPONSE?", string(response), " from ", conn.LocalAddr())
		
		// write back to channel with byte response
		recvChan <- response
	}
}


func functionTwo(sendChan, recvChan chan []byte, raddr *net.UDPAddr, timeToReply bool, wg sync.WaitGroup) {
	defer wg.Done()
	addr, err := net.ResolveUDPAddr("udp", ":0")
	checkError(err)
	
	conn, err := net.ListenUDP("udp", addr)
	fmt.Println(conn.LocalAddr());
	checkError(err)

	// LOBBY PHASE; need to recv response from leader,
	// pass message onto java side (first round start)

	reply := <-recvChan
	_, err = conn.WriteToUDP(reply, raddr)
	checkError(err)

	for {
		// read some reply from the java game (update of move, or death)
		time.Sleep(100 * time.Millisecond)
		var buf = make([]byte, 4096)
		_, raddr, err := conn.ReadFromUDP(buf)
		buf = bytes.Trim(buf, "\x00")
		//FIRST//
		fmt.Println("DEBUG RECEIVE", string(buf))
		checkError(err)
		
		// send buf to leader channel
		sendChan <- buf

		// read reply (timeToReply?) from leader (TODO: use a select w/ timeout?)
		reply := <-recvChan

		// TODO: only write back after receiving multiple replies, or after ticker timeout
		//FOURTH//
		fmt.Println("DEBUG SENDING", string(reply))
		if timeToReply {
			_, err = conn.WriteToUDP(reply, raddr)
			checkError(err)
		}
	}
}

func main() {
	fmt.Println("HELLO I AM STARTED")

	if len(os.Args) != 4 {
		fmt.Println("RTFM")
		panic("DYING")
	}
	javaPort := os.Args[1]
	leaderAddr := os.Args[2]
	isLeader, err := strconv.ParseBool(os.Args[3])

	checkError(err)

	fmt.Println("Leadder address is " + leaderAddr)
	gameState.Round = 1

	timeToReply := true // TODO: set by ticker or all clients replied

	sendChan, recvChan := make(chan []byte, 1), make(chan []byte, 1)

	fmt.Println("Trying to connect to java on localhost:" + javaPort)

	javaRAddr, err := net.ResolveUDPAddr("udp", "localhost:"+javaPort)
	checkError(err)

	// if I am the leader, listen for rounds to confirm them
	if isLeader {
		go leaderListener(leaderAddr)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	
	// handle communication through leader channel (follower code to leader)
	go functionOne(sendChan, recvChan, leaderAddr, wg)

	// handle internal communication to java game
	go functionTwo(sendChan, recvChan, javaRAddr, timeToReply, wg)

	wg.Wait();

	fmt.Println("GOODBYE")

}

// If error is non-nil, print it out and halt.
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		os.Exit(1)
	}
}

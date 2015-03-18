package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type MyMove struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Direction string `json:"direction"`
	Pid       string `json:"pid"`
}

type Move struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Direction string `json:"direction"`
}

type GameState struct {
	Round      int
	MyPid      int
	GridWidth  int
	GridHeight int
	Positions  map[string]Move
	AddrToPid  map[*net.UDPAddr]string
}

var gameState GameState

type RoundStart struct {
	Round int    `json:"round"`
	Pid   string `json:"pid"`
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

type GameStart struct {
	Pid               string          `json:"pid"`
	StartingPositions map[string]Move `json:"startingPositions"`
}

type GameStartMessage struct {
	EventName string `json:"eventName"`
	GameStart `json:"gameStart"`
}

func random(min, max int) int {
	return rand.Intn(max-min) + min
}

func newRoundMessage() []byte {
	gameState.Round += 1
	// TODO: the pid...
	message := RoundStartMessage{EventName: "roundStart", Round: gameState.Round, RoundStart: RoundStart{Round: gameState.Round, Pid: "1"}}
	return encodeMessage(message)
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
		p, e := dat["pid"].(string)
		res = MovesMessage{EventName: "moves", Round: gameState.Round, Moves: []MyMove{MyMove{X: int(x), Y: int(y), Direction: d, Pid: p}}}

	default:
		panic("Did not understand event " + eventName)
	}

	fmt.Println("parsed message: ", res)
	return res
}

func encodeMessage(message interface{}) []byte {
	val, e := json.Marshal(message)
	checkError(e)
	return val
}

func CreateInitPlayerPosition() Move {
	return Move{X: random(0, gameState.GridWidth), Y: random(0, gameState.GridHeight), Direction: "UP"}
}

func isJoinMessage(buf []byte) bool {
	return string(buf) == "JOIN"
}

func isStartMessage(buf []byte) bool {
	return string(buf) == "START\n"
}

// TODO: given pid, return message event == startGame (or something), that includes
// all player start positions, this players pid, and their leader they refer to from now on
func startGameMessage(pid string, startingPositions map[string]Move) GameStartMessage {
	return GameStartMessage{EventName: "gameStart", GameStart: GameStart{Pid: pid, StartingPositions: startingPositions}}
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
func initLobby(conn *net.UDPConn, raddrs []*net.UDPAddr) {
	// start of new game
	for {
		// wait for message from some client
		var buf = make([]byte, 4096)
		fmt.Println("Waiting for a client to join or send a start game message")
		_, raddr, err := conn.ReadFromUDP(buf)
		buf = bytes.Trim(buf, "\x00")
		checkError(err)
		// what type of message is it? join or start game?
		if isJoinMessage(buf) {
			if _, knownPlayer := gameState.AddrToPid[raddr]; !knownPlayer {
				pid := strconv.Itoa(len(gameState.Positions) + 1)
				gameState.Positions[pid] = CreateInitPlayerPosition()
				gameState.AddrToPid[raddr] = pid
				fmt.Println("New player has joined from address " + raddr.String())
				fmt.Println("Assigning pid " + pid + " and starting position " + strconv.Itoa(gameState.Positions[pid].X) + "," + strconv.Itoa(gameState.Positions[pid].Y))
			}
		} else if isStartMessage(buf) {
			fmt.Println("The game start message has been sent! Notifying all players")
			// send message to all players to start game
			for player, pid := range gameState.AddrToPid {
				// TODO: we probably care about whether or not this one is received
				newGameMsg := encodeMessage(startGameMessage(pid, gameState.Positions))
				fmt.Println("Sending a game start message to " + player.String() + ". " + string(newGameMsg))
				_, err = conn.WriteToUDP(newGameMsg, player)
				checkError(err)
			}
			// end lobby phase, prepare to send round messages next
			break
		} else {
			panic("WTF KIND OF MESSAGE IS THIS" + string(buf))
		}
	}
}

func newRound(conn *net.UDPConn, raddrs []*net.UDPAddr) (reply MovesMessage) {
	// create message that new round is starting
	byt := newRoundMessage()

	// send message to everyone that new round has started
	fmt.Println("Sending the following round start message " + string(byt))

	for i := range raddrs {
		_, err := conn.WriteToUDP(byt, raddrs[i])
		checkError(err)
	}

	reply = MovesMessage{EventName: "moves", Round: gameState.Round, Moves: make([]MyMove, 0)}
	return
}

func leaderListener(leaderAddrString string) {
	// Listen
	leaderAddr, err := net.ResolveUDPAddr("udp", leaderAddrString)
	checkError(err)
	conn, err := net.ListenUDP("udp", leaderAddr)
	checkError(err)
	// connBuf := bufio.NewReader(conn)
	fmt.Println("Leader has started")

	raddrs := make([]*net.UDPAddr, 0)

	// LOBBY PHASE
	// before the general main loop, wait for playerCount messages,
	// this will tell me who I need to send roundStarts to.
	initLobby(conn, raddrs)

	// MAIN LOOP SECTION
	isNewRound := true
	var responses map[string]bool
	var reply MovesMessage
	for {
		// if a new round is starting, let everyone connected to me know
		if isNewRound {
			reply = newRound(conn, raddrs)
			responses = make(map[string]bool)
			isNewRound = false
			raddrs = make([]*net.UDPAddr, 0)
		}

		// read a message from some follower
		var buf = make([]byte, 4096)
		_, raddr, err := conn.ReadFromUDP(buf)
		buf = bytes.Trim(buf, "\x00")
		fmt.Println("DEBUG SERVER?", string(buf), conn.LocalAddr())
		checkError(err)

		// parse the new message into a MovesMessage struct (usually)
		commands := parseMessage(buf)

		// append moves received to list of all moves recieved for current round
		if commands.EventName == "moves" && commands.Round == gameState.Round {
			for _, move := range commands.Moves {
				reply.Moves = append(reply.Moves, move)
				responses[move.Pid] = true
			}
			// keep track of who to respond to
			raddrs = append(raddrs, raddr)
		}

		// end condition; reply to my followers if I have been messaged by all of them
		if len(responses) == len(gameState.AddrToPid) {
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

func functionOne(sendChan chan string, recvChan chan string, leaderAddrString string, wg sync.WaitGroup, isLeader bool) {
	defer wg.Done()

	// Get a port for the go client to use
	addr, err := net.ResolveUDPAddr("udp", "localhost:0")
	checkError(err)
	conn, err := net.ListenUDP("udp", addr)
	checkError(err)
	// Resolve the leader address
	leaderAddr, err := net.ResolveUDPAddr("udp", leaderAddrString)
	checkError(err)

	// Write a message to the leader letting it know that I have started
	fmt.Println("Sending a hello message to the leader")
	_, err = conn.WriteToUDP([]byte("JOIN"), leaderAddr)
	checkError(err)

	if isLeader {
		message := <-sendChan
		fmt.Println("Go client got START message. Sending to leader")
		_, err = conn.WriteToUDP([]byte(message), leaderAddr)
		checkError(err)
	}

	// Read response from leader
	response := make([]byte, 4096)
	fmt.Println("Waiting for leader to respond with game start details")
	_, _, err = conn.ReadFromUDP(response)
	checkError(err)
	response = bytes.Trim(response, "\x00")
	fmt.Println("Received a game start response from the leader:" + string(response))

	// write back to channel with byte response (let java know to start)
	recvChan <- string(response)

	for {
		// wait for message on leader channel
		message := <-sendChan // TODO make this

		// write message to leader address
		//SECOND//
		_, err = conn.WriteToUDP([]byte(message), leaderAddr)
		checkError(err)

		// read response from leader
		response = make([]byte, 4096)
		_, _, err := conn.ReadFromUDP(response)
		checkError(err)
		response = bytes.Trim(response, "\x00")
		//THIRD//

		// write back to channel with byte response
		recvChan <- string(response)
	}
}

func javaGoConnection(sendChan chan string, recvChan chan string, javaAddrString string, wg sync.WaitGroup, isLeader bool) {
	defer wg.Done()

	fmt.Println("Trying to connect to java on " + javaAddrString)
	conn, err := net.Dial("tcp", javaAddrString)
	defer conn.Close()
	checkError(err)
	connBuf := bufio.NewReader(conn)

	if isLeader {
		// wait for a start game message from java
		str, err := connBuf.ReadString('\n')
		checkError(err)
		fmt.Println("Received a start game message from java: " + str)
		// now tell the leader
		sendChan <- str
	}

	// LOBBY PHASE; need to recv response from leader,
	// pass message onto java side (first round start)
	fmt.Println("Waiting for the go message to send to java")
	reply := <-recvChan
	fmt.Println("reply" + reply)
	conn.Write([]byte(reply + "\n"))
	checkError(err)
	fmt.Println("Wrote game start message to java")

	for {
		// read some reply from the java game (update of move, or death)
		time.Sleep(100 * time.Millisecond)
		status, err := bufio.NewReader(conn).ReadString('\n')
		//FIRST//
		fmt.Println("DEBUG RECEIVE", status)
		checkError(err)

		// send buf to leader channel
		sendChan <- status

		// read reply (timeToReply?) from leader (TODO: use a select w/ timeout?)
		reply := <-recvChan

		// TODO: only write back after receiving multiple replies, or after ticker timeout
		//FOURTH//
		_, err = fmt.Fprintf(conn, string(reply))
		checkError(err)
	}

}

func main() {
	fmt.Println("Go process started")

	// argument parsing
	if len(os.Args) != 6 {
		panic("RTFM")
	}
	javaPort := os.Args[1]
	leaderAddr := os.Args[2]
	isLeaderString := os.Args[3]
	isLeader, err := strconv.ParseBool(isLeaderString)
	checkError(err)
	gridWidth, err := strconv.Atoi(os.Args[4])
	checkError(err)
	gridHeight, err := strconv.Atoi(os.Args[5])
	checkError(err)

	// init vars
	rand.Seed(time.Now().Unix())
	gameState.Round = 1
	gameState.GridWidth = gridWidth
	gameState.GridHeight = gridHeight
	gameState.Positions = make(map[string]Move)
	gameState.AddrToPid = make(map[*net.UDPAddr]string)
	sendChan, recvChan := make(chan string, 1), make(chan string, 1)

	// if I am the leader, listen for rounds to confirm them
	if isLeader {
		go leaderListener(leaderAddr)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// handle communication through leader channel (follower code to leader)
	go functionOne(sendChan, recvChan, leaderAddr, wg, isLeader)

	// handle internal communication to java game
	go javaGoConnection(sendChan, recvChan, "localhost:"+javaPort, wg, isLeader)

	wg.Wait()

	fmt.Println("GOODBYE")
}

// If error is non-nil, print it out and halt.
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		os.Exit(1)
	}
}

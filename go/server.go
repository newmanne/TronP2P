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
	"strings"
	"sync"
	"time"
)

// STRUCTURES

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
	Alive      map[string]bool
	Finish     []string
	AddrToPid  map[*net.UDPAddr]string
}

type RoundStart struct {
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

type MyDeathMessage struct {
	EventName string `json:"eventName"`
	Round     int    `json:"round"`
}

type MovesMessage struct {
	EventName string `json:"eventName"`
	Round     int    `json:"round"`
	Moves     `json:"moves"`
}

type Moves struct {
	Moves map[string]Move `json:"moves"`
}

type GameStart struct {
	Pid               string          `json:"pid"`
	StartingPositions map[string]Move `json:"startingPositions"`
}

type GameStartMessage struct {
	EventName string `json:"eventName"`
	GameStart `json:"gameStart"`
}

// GLOBAL VARS

var gameState GameState
var DISABLE_GAME_OVER = true // to allow single player game for debugging
var REFRESH_RATE = 50 * time.Millisecond // number of milliseconds per refresh/round

// UTILITY FUNCTIONS

func readFromUDP(conn *net.UDPConn) ([]byte, *net.UDPAddr) {
	buf := make([]byte, 4096)
	_, raddr, err := conn.ReadFromUDP(buf)
	buf = bytes.Trim(buf, "\x00")
	checkError(err)
	return buf, raddr
}

func randomInt(min, max int) int {
	return rand.Intn(max-min) + min
}

func log(message string) {
	fmt.Println(message)
}

func logJava(message string) {
	log("GO2Java: " + message)
}

func logLeader(message string) {
	log("GOLEADER: " + message)
}

func logClient(message string) {
	log("GOCLIENT: " + message)
}

func encodeMessage(message interface{}) []byte {
	val, e := json.Marshal(message)
	checkError(e)
	return val
}

func newRoundMessage() []byte {
	gameState.Round += 1
	message := RoundStartMessage{EventName: "roundStart", Round: gameState.Round, RoundStart: RoundStart{}}
	return encodeMessage(message)
}

func parseMessage(buf []byte) (Move, string) {
	fmt.Println("Parsing the following message " + string(buf))
	var dat map[string]interface{}
	err := json.Unmarshal(buf, &dat)
	checkError(err)

	fmt.Println("dat: ", dat)

	// TODO: error handling, default actions
	eventName := dat["eventName"].(string)
	var res Move
	var pid string
	switch eventName {
	case "myMove":
		x, _ := dat["x"].(float64)
		y, _ := dat["y"].(float64)
		d, _ := dat["direction"].(string)
		pid, _ = dat["pid"].(string)
		res = Move{X: int(x), Y: int(y), Direction: d}
	case "myDeath":
		if gameState.Alive[pid] {
			gameState.Alive[pid] = false
			gameState.Finish = append(gameState.Finish, pid)
		}
		res = gameState.Positions[pid]
	default:
		panic("Did not understand event " + eventName)
	}

	fmt.Println("parsed message: ", res)
	return res, pid
}

func CreateInitPlayerPosition() Move {
	return Move{X: randomInt(0, gameState.GridWidth), Y: randomInt(0, gameState.GridHeight), Direction: "UP"}
}

func isJoinMessage(buf []byte) bool {
	return strings.TrimSpace(string(buf)) == "JOIN"
}

func isStartMessage(buf []byte) bool {
	return strings.TrimSpace(string(buf)) == "START"
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
func initLobby(conn *net.UDPConn) {
	// start of new game
	for {
		// wait for message from some client
		logLeader("Waiting for a client to join or send a start game message")
		buf, raddr := readFromUDP(conn)
		// what type of message is it? join or start game?
		if isJoinMessage(buf) || isStartMessage(buf) {
			if _, knownPlayer := gameState.AddrToPid[raddr]; !knownPlayer {
				pid := strconv.Itoa(len(gameState.Positions) + 1)
				gameState.Positions[pid] = CreateInitPlayerPosition()
				gameState.AddrToPid[raddr] = pid
				logLeader("New player has joined from address " + raddr.String())
				logLeader("Assigning pid " + pid + " and starting position " + strconv.Itoa(gameState.Positions[pid].X) + "," + strconv.Itoa(gameState.Positions[pid].Y))
			}
		}

		if isStartMessage(buf) {
			logLeader("The game start message has been sent! Notifying all players")
			// send message to all players to start game
			for player, pid := range gameState.AddrToPid {
				// TODO: we probably care about whether or not this one is received
				newGameMsg := encodeMessage(startGameMessage(pid, gameState.Positions))
				logLeader("Sending a game start message to " + player.String() + ". " + string(newGameMsg))
				_, err := conn.WriteToUDP(newGameMsg, player)
				checkError(err)
			}
			// end lobby phase, prepare to send round messages next
			break
		} else if !isJoinMessage(buf) {
			panic("WTF KIND OF MESSAGE IS THIS " + string(buf))
		}
	}
}

// determines if game is over (all players are dead except one)
func gameOver() bool {

	if DISABLE_GAME_OVER {
		return false
	}

	count := 0
	for _, alive := range gameState.Alive {
		if alive {
			if count == 1 {
				return false
			}
			count = 1
		}
	}
	return true
}

// main leader function, approves moves of followers, TODO add leader hierarchy
func leaderListener(leaderAddrString string) {
	// Listen
	leaderAddr, err := net.ResolveUDPAddr("udp", leaderAddrString)
	checkError(err)
	conn, err := net.ListenUDP("udp", leaderAddr)
	checkError(err)
	// connBuf := bufio.NewReader(conn)
	logLeader("Leader has started")

	// LOBBY PHASE
	// before the general main loop, wait for playerCount messages,
	// this will tell me who I need to send roundStarts to.
	initLobby(conn)

	// MAIN GAME LOOP
	isNewRound := true
	var roundMoves MovesMessage
	for !gameOver() {
		// if a new round is starting, let everyone connected to me know
		if isNewRound {
			newRoundMessage := newRoundMessage()
			logLeader("Sending the following round start message: " + string(newRoundMessage))
			for addr, pid := range gameState.AddrToPid {
				_, err := conn.WriteToUDP(newRoundMessage, addr)
				checkError(err)
				logLeader("Sent round start to player " + pid)
			}
			roundMoves = MovesMessage{EventName: "moves", Round: gameState.Round, Moves: Moves{Moves: make(map[string]Move)}}
			isNewRound = false
			logLeader("done sending round start messages.")
		}

		// read a message from some follower
		logLeader("Waiting to receive message from follower...")
		buf, _ := readFromUDP(conn)
		// for now, assume its a move message. TODO other stuff
		// TODO: check relevant round
		move, pid := parseMessage(buf)
		roundMoves.Moves.Moves[pid] = move
		logLeader("Received move message " + string(encodeMessage(move)) + " from player " + pid)

		// end condition; reply to my followers if I have been messaged by all of them
		if len(roundMoves.Moves.Moves) == len(gameState.AddrToPid) {
			byt := encodeMessage(roundMoves)

			// send message to all followers
			for addr, pid := range gameState.AddrToPid {
				_, err = conn.WriteToUDP(byt, addr)
				checkError(err)
				logLeader("Replied with round moves to player " + pid)
			}
			// start a new round of communication
			isNewRound = true
		}
	}

	// TODO END GAME SCREEN (RESULTS)
}

func goClient(sendChan chan string, recvChan chan string, leaderAddrString string, wg sync.WaitGroup, isLeader bool) {
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
	logClient("Sending a hello message to the leader")
	_, err = conn.WriteToUDP([]byte("JOIN"), leaderAddr)
	checkError(err)

	if isLeader {
		message := <-sendChan
		logClient("Go client got START message. Sending to leader")
		_, err = conn.WriteToUDP([]byte(message), leaderAddr)
		checkError(err)
	}

	// Read response from leader
	logClient("Waiting for leader to respond with game start details")
	buf, _ := readFromUDP(conn)
	logClient("Received a game start response from the leader:" + string(buf))

	// write back to channel with byte response (let java know to start)
	recvChan <- string(buf)

	logClient("LOBBY PHASE IS OVER. ENTERING MAIN LOOP")
	// MAIN GAME LOOP
	for !gameOver() {
		// read round start from leader
		buf, _ = readFromUDP(conn)
		logClient("Got round start message from leader, passing it to java")
		recvChan <- string(buf)

		// wait for message from java
		message := <-sendChan // TODO make this

		// write message to leader address
		//SECOND//
		_, err = conn.WriteToUDP([]byte(message), leaderAddr)
		checkError(err)

		// read response from leader
		buf, _ = readFromUDP(conn)
		//THIRD//

		// write back to channel with byte response
		recvChan <- string(buf)
	}

	// TODO END GAME SCREEN (RESULTS)
}

func javaGoConnection(sendChan chan string, recvChan chan string, javaAddrString string, wg sync.WaitGroup, isLeader bool) {
	defer wg.Done()

	logJava("Trying to connect to java on " + javaAddrString)
	conn, err := net.Dial("tcp", javaAddrString)
	defer conn.Close()
	checkError(err)
	connBuf := bufio.NewReader(conn)

	if isLeader {
		// wait for a start game message from java
		str, err := connBuf.ReadString('\n')
		checkError(err)
		logJava("Received a start game message from java: " + str)
		// now tell the leader
		sendChan <- str
	}

	// LOBBY PHASE; need to recv response from leader,
	// pass message onto java side (first round start)
	logJava("Waiting for the go message to send to java")
	reply := <-recvChan
	logJava("reply " + reply)
	conn.Write([]byte(reply + "\n"))
	checkError(err)
	logJava("Wrote game start message to java. Lobby phase over, entering main loop")

	// MAIN LOOP
	for !gameOver() {
		// read round start message from channel and send it to java
		logJava("Waiting for round start message from go client")
		roundStartMessage := <-recvChan
		logJava("Sending the following round start message to java:" + roundStartMessage)
		conn.Write([]byte(roundStartMessage + "\n"))
		checkError(err)
		logJava("Round start message sent to java")

		// read some reply from the java game (update of move, or death)
		time.Sleep(REFRESH_RATE)
		status, err := connBuf.ReadString('\n')
		logJava("Received: " + status)
		checkError(err)

		// send buf to leader channel
		sendChan <- status

		// read reply (timeToReply?) from leader (TODO: use a select w/ timeout?)
		reply := <-recvChan

		// TODO: only write back after receiving multiple replies, or after ticker timeout
		conn.Write([]byte(reply + "\n"))
		checkError(err)
	}

	// TODO END GAME SCREEN (RESULTS)
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
	gameState.Alive = make(map[string]bool)
	gameState.AddrToPid = make(map[*net.UDPAddr]string)
	sendChan, recvChan := make(chan string, 1), make(chan string, 1)

	// if I am the leader, listen for rounds to confirm them
	if isLeader {
		go leaderListener(leaderAddr)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// handle communication through leader channel (follower code to leader)
	go goClient(sendChan, recvChan, leaderAddr, wg, isLeader)

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

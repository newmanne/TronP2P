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
	Grace      map[string]int
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
var COLLISION_IS_DEATH = false
var JAVA_REPLY_TIME = 50 * time.Millisecond // time between every new java move
var FOLLOWER_RESPONSE_TIME = 25 * time.Millisecond // time for followers to respond
var MAX_GRACE_PERIOD = 100 // max number of consecutive missed messages
var FOLLOWER_RESPONSE_FAIL_RATE = 0 // out of 1000, fail rate for responses not to be received

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
		if COLLISION_IS_DEATH {
			killPlayer(pid)
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

func killPlayer(pid string) {
	if gameState.Alive[pid] {
		logLeader("killing player " + pid)
		gameState.Alive[pid] = false
		gameState.Finish = append(gameState.Finish, pid)
	}
}

/*
 At the end of the lobby session, the following things must be true:
 1) all players know the IP, pid, and start positions of all other players
 2) all players know their immediate leader
 3) all players acknowledge game begins?

 General structure:
 - monarch starts lobby session
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
				gameState.Alive[pid] = true
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

	// TODO: what happens if there is a tie?
	return len(gameState.Finish) >= len(gameState.AddrToPid) - 1
}

func resetGracePeriod(pid string) {
	gameState.Grace[pid] = 0
	logLeader("player " + pid + "grace period = " + strconv.Itoa(gameState.Grace[pid]))
}

func countGracePeriod(pid string) {
	gameState.Grace[pid] += 1
	logLeader("player " + pid + "grace period = " + strconv.Itoa(gameState.Grace[pid]))

	if gameState.Grace[pid] >= MAX_GRACE_PERIOD {
		killPlayer(pid)
	}
}

// if pid did not respond, their next move is to continue in next direction one move in advance
func addContinuedMove(moves MovesMessage, pid string) {
	nextMove := gameState.Positions[pid]
	
	switch nextMove.Direction {
	case "UP":
		nextMove.Y = nextMove.Y - 1
	case "DOWN":
		nextMove.Y = nextMove.Y + 1
	case "LEFT":
		nextMove.X = nextMove.X - 1
	case "RIGHT":
		nextMove.X = nextMove.X + 1
	default:
		panic("Next move direction unknown")
	}

	moves.Moves.Moves[pid] = nextMove
}

// to simulate missed responses; some of the time we will miss follower responses
// and resort to timeout; this is one of those times
func surviveFollowerResponseInjectedFailure() bool {
	p := randomInt(0, 1000)
	return p >= FOLLOWER_RESPONSE_FAIL_RATE
}

// determines if leader can respond to followers yet
func timeToRespond(moves MovesMessage, timedout bool) bool {
	recvCount := len(moves.Moves.Moves)
	totalNeeded := len(gameState.AddrToPid)
	logLeader("received " + strconv.Itoa(recvCount) + "/" + strconv.Itoa(totalNeeded) + " messages")

	if recvCount == totalNeeded || timedout {
		// count missed messages for those who did not respond, or reset
		for pid, alive := range gameState.Alive {
			logLeader("#### in for loop #####")
			_, responded := moves.Moves.Moves[pid]
			if responded {
				resetGracePeriod(pid)
			} else {
				countGracePeriod(pid)
				// don't move them if they are dead
				if alive {
					addContinuedMove(moves, pid)
				}		
			}
		}
		return true
	} else {
		return false
	}
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
	recvChan := make(chan bool, 1)

	timedout := false
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
			timedout = false
			isNewRound = false
			logLeader("done sending round start messages.")
		}


		// read messages from followers and forward them
		logLeader("Waiting to receive message from follower...")
		buf, _ := readFromUDP(conn)

		// TODO: check relevant round
		move, pid := parseMessage(buf)

		// artifical missed response for testing
		received := surviveFollowerResponseInjectedFailure()
		if received {
			roundMoves.Moves.Moves[pid] = move
			recvChan <- true
		}

		// block until timeout or response received
		logLeader("Waiting for timeout or message...")
		select {
		case <- recvChan:
			logLeader("Received move message " + string(encodeMessage(move)) + " from player " + pid)

		case <- time.After(FOLLOWER_RESPONSE_TIME):
			logLeader("received timeout")
			timedout = true
		}

		// end condition; reply to my followers if I have been messaged by all of them
		if timeToRespond(roundMoves, timedout) {
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
		time.Sleep(JAVA_REPLY_TIME)
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
	gameState.Grace = make(map[string]int)
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

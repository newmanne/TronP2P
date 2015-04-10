package main

import (
//	"reflect"
	"bufio"
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

var gameState GameState
var addressState AddressState
var leaderState LeaderState
var electionState = NORMAL
var DISABLE_GAME_OVER = true // to allow single player game for debugging
var COLLISION_IS_DEATH = true
var MIN_GAME_SPEED = (1000 / 2) * time.Millisecond      // time between every new java move
var FOLLOWER_RESPONSE_TIME = 500*2 * time.Millisecond      // time for followers to respond
var MAX_ALLOWABLE_MISSED_MESSAGES = 5                    // max number of consecutive missed messages
var FOLLOWER_RESPONSE_FAIL_RATE = map[string]int{"1": 0} // out of 1000, fail rate for responses not to be received

// STRUCTURES
type MyMove struct {
	Direction string `json:"direction"`
	Pid       string `json:"pid"`
}

type Move struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Direction string `json:"direction"`
}

type GameState struct {
	//refactor this
	LeaderID       int
	Round          int
	MyPid          int
	MyPriority     int
	GridWidth      int
	GridHeight     int
	Grid           [][]int
	Nickname       string
	Positions      []map[string]Move
	Alive          map[string]bool
	Grace          map[string]int
	Finish         []string
	AddrToPid      map[string]string
	AddrToAddr     map[string]*net.UDPAddr
	PidToNickname  map[string]string
	DroppedForever map[string]bool
}

type LeaderState struct {
	Positions []map[string]Move
}

type AddressState struct {
	javaAddr       string
	leaderAddr     string
	leaderUDPAddr  *net.UDPAddr
	isLeader       bool
	goConnection   *net.UDPConn
	javaConnection net.Conn
	connBuf        *bufio.Reader
	sendChan       chan []byte
	recvChan       chan []byte
}

type RoundStart struct {
	Round int `json:"round"`
}

type RoundStartMessage struct {
	MessageType string `json:"messageType"`
	EventName   string `json:"eventName"`
	Round       int    `json:"round"`
	RoundStart  `json:"roundStart"`
}

type MyMoveMessage struct {
	MessageType string `json:"messageType"`
	EventName   string `json:"eventName"`
	Round       int    `json:"round"`
	MyMove      `json:"myMove"`
}

type MyDeathMessage struct {
	MessageType string `json:"messageType"`
	EventName   string `json:"eventName"`
	Round       int    `json:"round"`
}

type MovesMessage struct {
	MessageType string `json:"messageType"`
	EventName   string `json:"eventName"`
	Round       int    `json:"round"`
	Moves       `json:"moves"`
}

type Moves struct {
	Moves []map[string]Move `json:"moves"`
	Round int               `json:"round"`
}

type GameStart struct {
	Pid               string            `json:"pid"`
	StartingPositions map[string]Move   `json:"startingPositions"`
	Nicknames         map[string]string `json:"nicknames"`
	Addresses         map[string]string `json:"addresses"`
}

type GameStartMessage struct {
	MessageType string `json:"messageType"`
	EventName   string `json:"eventName"`
	Round       int    `json:"round"`
	GameStart   `json:"gameStart"`
}

type GameOver struct {
	PidsInOrderOfDeath []string `json:"pidsInOrderOfDeath"`
}

type GameOverMessage struct {
	MessageType string `json:"messageType"`
	EventName   string `json:"eventName"`
	GameOver    `json:"gameOver"`
}

type LeaderElectionMessage struct {
	MessageType string `json:"messageType"`
	EventName   string `json:"eventName"`
	Round       int    `json:"round"`
	LeaderID    int    `json:"leaderid"`
}

type ElectionState int

const (
	NORMAL ElectionState = 1 + iota
	QUORUM
	NEWLEADER
)

/*
* READ FROM UDP
 */

func readFromUDPWithTimeout(conn *net.UDPConn, timeoutTime time.Time) ([]byte, *net.UDPAddr, bool) {
	buf := make([]byte, 4096)
	conn.SetReadDeadline(timeoutTime)
	n, raddr, err := conn.ReadFromUDP(buf)
	if err != nil {
		if e, ok := err.(net.Error); !ok || !e.Timeout() {
			checkError(err)
			return nil, nil, false
		} else {
			// timeout
			return nil, raddr, true
		}
	} else {
		return buf[0:n], raddr, false
	}
}

func readFromUDP(conn *net.UDPConn) ([]byte, *net.UDPAddr) {
	buf := make([]byte, 4096)
	n, raddr, err := conn.ReadFromUDP(buf)
	checkError(err)
	return buf[0:n], raddr
}

/*
* MATH UTILITIES
 */

func randomInt(min, max int) int {
	return rand.Intn(max-min) + min
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

/*
* LOG UTILITIES
 */
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

/*
* MESSAGE UTILITIES
 */

func encodeMessage(message interface{}) []byte {
	val, e := json.Marshal(message)
	checkError(e)
	return val
}

func decodeMessage(message []byte) map[string]interface{} {
	var dat map[string]interface{}
	err := json.Unmarshal(message, &dat)
	checkError(err)
	return dat
}

func parseMessage(buf []byte) (string, string, int) {
	fmt.Println("Parsing the following message " + string(buf))

	dat := decodeMessage(buf)

	fmt.Println("dat: ", dat)
	roundString, _ := dat["round"].(float64)
	round := int(roundString)
	eventName := dat["eventName"].(string)
	pid, _ := dat["pid"].(string)
	var direction string
	switch eventName {
	case "myMove":
		direction, _ = dat["direction"].(string)
	default:
		panic("Did not understand event " + eventName)
	}
	logLeader("parsed message: player " + pid + " is going in direction " + direction + " on round " + strconv.Itoa(round))
	return direction, pid, round
}

func getMessageType(buf []byte) (messageType string) {
	dat := decodeMessage(buf)
	messageType = dat["messageType"].(string)
	return
}

func getRoundNumber(buf []byte) (round int) {
	dat := decodeMessage(buf)
	roundString, _ := dat["round"].(float64)
	round = int(roundString)
	return
}

func getMoves(buf []byte) (moves []interface{}) {
	dat := decodeMessage(buf)
	moves = dat["moves"].(map[string]interface{})["moves"].([]interface{})
	return
}

func broadcastMessage(conn *net.UDPConn, message []byte) {
	for _, addr := range gameState.AddrToAddr {
		_, err := conn.WriteToUDP(message, addr)
		checkError(err)
		logLeader("Sent message " + string(message) + " to player " + string(addr.IP))
	}
}

/*
* MESSAGE CONSTRUCTORS
 */
func newRoundMessage() []byte {
	gameState.Round++
	message := RoundStartMessage{
		MessageType: "roundstart",
		EventName:   "roundStart",
		Round:       gameState.Round,
		RoundStart: RoundStart{
			Round: gameState.Round,
		},
	}
	return encodeMessage(message)
}

func startGameMessage(pid string, startingPositions map[string]Move) GameStartMessage {
	return GameStartMessage{
		MessageType: "startgame",
		EventName:   "gameStart",
		Round:       gameState.Round,
		GameStart: GameStart{
			Pid:               pid,
			StartingPositions: startingPositions,
			Nicknames:         gameState.PidToNickname,
			Addresses:         gameState.AddrToPid,
		},
	}
}

func endGameMessage() GameOverMessage {
	return GameOverMessage{
		MessageType: "gameover",
		EventName:   "gameOver",
		GameOver: GameOver{
			PidsInOrderOfDeath: gameState.Finish,
		},
	}
}

func newRound(conn *net.UDPConn) (roundMoves MovesMessage) {
	newRoundMessage := newRoundMessage()
	broadcastMessage(conn, newRoundMessage)
	logLeader("done sending round start messages.")
	slideWindow()
	fmt.Println(leaderState.Positions)
	roundMoves = MovesMessage{
		MessageType: "moves",
		EventName:   "moves",
		Round:       gameState.Round,
		Moves: Moves{
			Moves: leaderState.Positions,
			Round: gameState.Round,
		},
	}
	fmt.Println(roundMoves)
	fmt.Println(decodeMessage(encodeMessage(roundMoves)))
	return
}

/*
* INIT FUNCTIONS
 */
func CreateInitPlayerPosition() Move {
	var direction string
	directionId := randomInt(0, 4)
	if directionId == 0 {
		direction = "UP"
	} else if directionId == 1 {
		direction = "DOWN"
	} else if directionId == 2 {
		direction = "LEFT"
	} else {
		direction = "RIGHT"
	}
	// Give a small buffer so they don't crash into a wall immediately!
	if (gameState.GridWidth < 30) || (gameState.GridHeight < 30) {
		panic("game grid dimensions are too small!")
	}
	return Move{
		X:         randomInt(15, gameState.GridWidth-15),
		Y:         randomInt(15, gameState.GridHeight-15),
		Direction: direction,
	}
}

func registerNewPlayer(raddr *net.UDPAddr, buf []byte) {
	address := raddr.String()
	pid := strconv.Itoa(len(getLeaderMoveMap()) + 1)
	getLeaderMoveMap()[pid] = CreateInitPlayerPosition()
	gameState.Alive[pid] = true
	gameState.AddrToPid[address] = pid
	gameState.AddrToAddr[address] = raddr
	gameState.Alive[pid] = true
	nickname := strings.Split(string(buf), ":")[1]
	gameState.PidToNickname[pid] = nickname
	logLeader("New player named " + nickname + " has joined from address " + address)
	logLeader("Assigning pid " + pid + " and starting position " + strconv.Itoa(getLeaderMoveMap()[pid].X) +
		"," + strconv.Itoa(getLeaderMoveMap()[pid].Y))
}

func initLobby(conn *net.UDPConn) {
	for {
		logLeader("Waiting for a client to join or send a start game message")
		buf, raddr := readFromUDP(conn)
		if isJoinMessage(buf) {
			address := raddr.String()
			if _, knownPlayer := gameState.AddrToPid[address]; !knownPlayer {
				registerNewPlayer(raddr, buf)
			}
		} else if isStartMessage(buf) {
			logLeader("Start of the game, sending broadcast")
			for addr, pid := range gameState.AddrToPid {
				newGameMsg := encodeMessage(startGameMessage(pid, getLeaderMoveMap()))
				logLeader("Sending a game start message to " + addr + ". " + string(newGameMsg))
				_, err := conn.WriteToUDP(newGameMsg, gameState.AddrToAddr[addr])
				checkError(err)
			}
			break
		} else {
			panic("Message not regonized:  " + string(buf))
		}
	}
}

func initializeLeader(leaderAddrString string) (conn *net.UDPConn) {
	leaderState.Positions = make([]map[string]Move, MAX_ALLOWABLE_MISSED_MESSAGES)
	for i := 0; i < len(leaderState.Positions); i++ {
		leaderState.Positions[i] = make(map[string]Move)
	}
	leaderAddr, err := net.ResolveUDPAddr("udp", leaderAddrString)
	checkError(err)
	conn, err = net.ListenUDP("udp", leaderAddr)
	checkError(err)
	return
}

func initializeLeaderConnection() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:0")
	checkError(err)
	conn, err := net.ListenUDP("udp", addr)
	checkError(err)
	leaderUDPAddr, err := net.ResolveUDPAddr("udp", addressState.leaderAddr)
	checkError(err)
	addressState.goConnection = conn
	addressState.leaderUDPAddr = leaderUDPAddr
}

func contactLeader(){
	initializeLeaderConnection()
	logClient("Sending a hello message to the leader")
	_, err := addressState.goConnection.WriteToUDP([]byte("JOIN:"+ gameState.Nickname),
		addressState.leaderUDPAddr)
	checkError(err)
	if addressState.isLeader {
		message := <-addressState.sendChan
		logClient("Go client got START message. Sending to leader")
		_, err = addressState.goConnection.WriteToUDP([]byte(message), addressState.leaderUDPAddr)
		checkError(err)
	}

	buf, _ := readFromUDP(addressState.goConnection)
	logClient("Received a game start response from the leader:" + string(buf))
	dat := decodeMessage(buf)["gameStart"].(map[string]interface{})
	pid, _ := strconv.Atoi(dat["pid"].(string))
	gameState.MyPid = pid
	addresses := dat["addresses"].(map[string]interface{})
	for addr, pid := range addresses {
		raddr, err := net.ResolveUDPAddr("udp", addr)
		checkError(err)
		gameState.AddrToPid[addr] = pid.(string)
		gameState.AddrToAddr[addr] = raddr
	}
	addressState.recvChan <- buf
}

func initializeJavaConnection() {
	logJava("Trying to connect to java on " + addressState.javaAddr)
	conn, err := net.Dial("tcp", addressState.javaAddr)
	addressState.connBuf = bufio.NewReader(conn)
	checkError(err)
	if addressState.isLeader {
		str, err := addressState.connBuf.ReadString('\n')
		checkError(err)
		logJava("Received a start game message from java: " + str)
		addressState.sendChan <- []byte(str)
	}
	logJava("Waiting for the go message to send to java")
	reply := <-addressState.recvChan
	logJava("reply " + string(reply))
	conn.Write(append(reply, '\n'))
	logJava("Wrote game start message to java. Lobby phase over, entering main loop")
	addressState.javaConnection = conn
}

func initializeGameState() {
	var err error
	rand.Seed(time.Now().Unix())
	gameState.LeaderID = 0
	gameState.Round = 1
	gameState.GridWidth, err = strconv.Atoi(os.Args[4])
	checkError(err)
	gameState.GridHeight, err = strconv.Atoi(os.Args[5])
	checkError(err)
	gameState.Nickname = os.Args[6]
	gameState.Positions = make([]map[string]Move, MAX_ALLOWABLE_MISSED_MESSAGES)
	for i := 0; i < len(gameState.Positions); i++ {
		gameState.Positions[i] = make(map[string]Move)
	}
	gameState.Grid = make([][]int, gameState.GridWidth)
	for i := 0; i < gameState.GridWidth; i++ {
		gameState.Grid[i] = make([]int, gameState.GridHeight)
	}
	gameState.Alive = make(map[string]bool)
	gameState.Grace = make(map[string]int)
	gameState.AddrToPid = make(map[string]string)
	gameState.AddrToAddr = make(map[string]*net.UDPAddr)
	gameState.PidToNickname = make(map[string]string)
	gameState.DroppedForever = make(map[string]bool)

	isLeader, err := strconv.ParseBool(os.Args[3])
	checkError(err)

	addressState.javaAddr = "localhost:" + os.Args[1]
	addressState.leaderAddr = os.Args[2]
	addressState.isLeader = isLeader
	addressState.sendChan = make(chan []byte, 1)
	addressState.recvChan = make(chan []byte, 1)
}

/*
* CHECK FUNCTIONS
 */

func gameOver() bool {
	if DISABLE_GAME_OVER {
		return false
	}
	//TODO the following two lines don't make sense.
	numDeadToEnd := len(gameState.AddrToPid) - 1
	logLeader("There are " + strconv.Itoa(len(gameState.Finish)) + " dead players and we require at least " +
		strconv.Itoa(numDeadToEnd) + " dead players to call it a game")
	return len(gameState.Finish) >= numDeadToEnd
}

func resetGracePeriod(pid string) {
	gameState.Grace[pid] = 0
}

func countGracePeriod(pid string) {
	if !gameState.DroppedForever[pid] {
		gameState.Grace[pid] += 1
		logLeader("Player " + pid + " grace period = " + strconv.Itoa(gameState.Grace[pid]))
		if gameState.Grace[pid] >= MAX_ALLOWABLE_MISSED_MESSAGES {
			logLeader("Grace period for player " + pid + " exceeded. Force dropping them")
			dropPlayer(pid)
		}
	}
}

//TODO Those functions are pointless and completely stupid. We have a function to get the message type
func isJoinMessage(buf []byte) bool {
	//return getMessageType(message) == "join"
	return strings.Contains(strings.TrimSpace(string(buf)), "JOIN")
}

func isStartMessage(buf []byte) bool {
	//return getMessageType(message) == "start"
	return strings.TrimSpace(string(buf)) == "START"
}

func isGameOverMessage(buf []byte) bool {
	return getMessageType(buf) == "gameover"
	return strings.Contains(string(buf), "GameOver")
}

/*
* UTILITY FUNCTIONS
*/

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		os.Exit(1)
	}
}

// Slight difference from killing player; we owe no obligation to respond to dropped players,
// but we should respond to killed players that are still connected.
func dropPlayer(pid string) {
	logLeader("dropping player " + pid)
	gameState.DroppedForever[pid] = true
	killPlayer(pid)
}

func killPlayer(pid string) {
	if gameState.Alive[pid] {
		logLeader("killing player " + pid)
		gameState.Alive[pid] = false
		gameState.Finish = append(gameState.Finish, pid)
	}
	if gameOver() {
		for pid, alive := range gameState.Alive {
			if alive {
				gameState.Finish = append(gameState.Finish, pid)
				logLeader("Player " + pid + " is the winner! Congrats!")
			}
		}
	}
}

func getCurrentMoveMap() map[string]Move {
	return gameState.Positions[len(gameState.Positions)-1]
}

func getLeaderMoveMap() map[string]Move {
	return leaderState.Positions[len(leaderState.Positions)-1]
}

func slideWindow() {
	fmt.Println("Sliding window")
	fmt.Println(leaderState.Positions)
	// push everything back
	for i := 0; i < len(leaderState.Positions)-1; i++ {
		leaderState.Positions[i] = leaderState.Positions[i+1]
	}
	// clear final move
	leaderState.Positions[len(leaderState.Positions)-1] = make(map[string]Move)
	fmt.Println(leaderState.Positions)
	fmt.Println("Sliding done")
}

//TODO Change this name
func addContinuedMove(pid string) {
	fmt.Println("adding continued move")
	prevMove := leaderState.Positions[len(leaderState.Positions)-2][pid]
	makeMove(prevMove.Direction, pid)
}

// Attempt to move the player pid one space in given direction. If movement
// results in collision, the player dies.
func makeMove(direction string, pid string) Move {
	//TODO what about a linked list?
	fmt.Println("Making move " + direction + " " + pid)
	prevMove := leaderState.Positions[len(leaderState.Positions)-2][pid]
	fmt.Println(leaderState.Positions)
	fmt.Println(prevMove)
	var nextMove Move
	if gameState.Alive[pid] {
		nextMove = Move{
			Direction: direction,
			X:         prevMove.X,
			Y:         prevMove.Y,
		}
		switch nextMove.Direction {
		case "DOWN":
			nextMove.Y = max(1, nextMove.Y-1)
		case "UP":
			nextMove.Y = min(gameState.GridHeight-2, nextMove.Y+1)
		case "LEFT":
			nextMove.X = max(1, nextMove.X-1)
		case "RIGHT":
			nextMove.X = min(gameState.GridWidth-2, nextMove.X+1)
		default:
			panic("Next move direction unknown")
		}

		if isCollision(nextMove.X, nextMove.Y) {
			killPlayer(pid)
			nextMove = prevMove
		} else {
			gameState.Grid[nextMove.X][nextMove.Y], _ = strconv.Atoi(pid)
		}
	} else { // Player is dead, keep old move.
		nextMove = prevMove
	}
	getLeaderMoveMap()[pid] = nextMove
	fmt.Println("move made")
	fmt.Println(leaderState.Positions)
	return nextMove
}

func surviveFollowerResponseInjectedFailure(pid string) bool {
	//TODO the hell is this ok
	if val, ok := FOLLOWER_RESPONSE_FAIL_RATE[pid]; ok {
		p := randomInt(0, 1000)
		return p >= val
	}
	return true
}

func timeToRespond() bool {
	recvCount := len(getLeaderMoveMap())
	totalNeeded := len(gameState.AddrToPid) - len(gameState.DroppedForever)
	logLeader("received " + strconv.Itoa(recvCount) + "/" + strconv.Itoa(totalNeeded) + " messages")
	if recvCount == totalNeeded {
		// count missed messages for those who did not respond, or reset
		for pid, alive := range gameState.Alive {
			_, responded := getLeaderMoveMap()[pid]
			if responded {
				resetGracePeriod(pid)
			} else {
				countGracePeriod(pid)
				if alive {
					addContinuedMove(pid)
				}
			}
		}
		return true
	}
	return false
}

//TODO No comment.
func isCollision(x, y int) bool {
	if !COLLISION_IS_DEATH {
		return false
	} else if (0 <= x && x < gameState.GridWidth) && (0 <= y && y < gameState.GridHeight) {
		return gameState.Grid[x][y] != 0
	} else {
		return true
	}
}

/*
* MAIN FUNCTIONS
 */
func main() {
	fmt.Println("Go process started")
	if FOLLOWER_RESPONSE_TIME < MIN_GAME_SPEED {
		panic("Can't set response time to be less than min game speed")
	}
	if len(os.Args) != 7 {
		panic("RTFM")
	}
	initializeGameState()
	if addressState.isLeader {
		go func() {
			conn := initializeLeader(addressState.leaderAddr)
			logLeader("Leader has started")
			initLobby(conn)
			go leaderListener(conn)
		}()
		//TODO i'm pretty sure there's a better way than that
		time.Sleep(100 * time.Millisecond) // stupid hack to make sure the leader is up before the client
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go goClient(wg)
	go javaGoConnection(wg)
	wg.Wait()
	fmt.Println("GOODBYE")
}

/*
* Routines
*/
func leaderListener(conn *net.UDPConn) {
	defer conn.Close()
	var roundMoves MovesMessage
	var timeoutTimeForRound time.Time
	for {
		roundMoves = newRound(conn)
		fmt.Println("here")
		fmt.Println(roundMoves)
		timeoutTimeForRound = time.Now().Add(FOLLOWER_RESPONSE_TIME)
		for {
			logLeader("Waiting to receive message from follower...")
			buf, _, timedout := readFromUDPWithTimeout(conn, timeoutTimeForRound)
			if timedout {
				break
			}
			direction, pid, round := parseMessage(buf)
			if round == gameState.Round && !gameState.DroppedForever[pid] {
				if surviveFollowerResponseInjectedFailure(pid) {
					logLeader("Received move message " + " from player " + pid)
					move := makeMove(direction, pid)
					fmt.Println("Registered move ", move)
					if timeToRespond() {
						break
					}
				}
			} else {
				logLeader("Recieved a move message from " + pid + " from an old round " +
					strconv.Itoa(round) + " but current round is " +
					strconv.Itoa(gameState.Round) + ". Ignoring message")
			}
		}
		if gameOver() {
			break
		}
		fmt.Println("message to be sent")
		fmt.Println(roundMoves)
		byt := encodeMessage(roundMoves)
		fmt.Println(decodeMessage(byt))
		broadcastMessage(conn, byt)
	}
}

func goClient(wg sync.WaitGroup) {
	var bufChan chan []byte
	gameOver := false
	defer wg.Done()
	contactLeader()
	defer addressState.goConnection.Close()
	logClient("Waiting for leader to respond with game start details")
	for !gameOver {
		timeoutTimeForRound := time.Now().Add(FOLLOWER_RESPONSE_TIME)
		buf, raddr, timedout := readFromUDPWithTimeout(addressState.goConnection, timeoutTimeForRound)
		if timedout {
			fmt.Println("timedout")
			if electionState == NORMAL {
				fmt.Println("new election")
				electionState = QUORUM
				bufChan = make(chan []byte, 1)
				go func() {
					time.Sleep(FOLLOWER_RESPONSE_TIME)
					electionState = NORMAL
					close(bufChan)
				}()
				go startElection(bufChan)
			}
			continue
		}
		//might be usefull to move it to a separate routine? if it gets slow, that is
		messageType := getMessageType(buf)
		switch messageType {
		case "roundstart":
			fmt.Println("Message " + string(buf))
			addressState.recvChan <- buf
			gameState.Round = getRoundNumber(buf)
			message := <-addressState.sendChan
			_, err := addressState.goConnection.WriteToUDP([]byte(message), addressState.leaderUDPAddr)
			checkError(err)
			break
		case "moves":
			fmt.Println("Message " + string(buf))
			fmt.Println("moves")
			fmt.Println(getMoves(buf))
			fmt.Println(gameState.Positions)
			for index, moves := range getMoves(buf) {
				positions := gameState.Positions[index]
				for pid, move := range moves.(map[string]interface{}) {
					castedMove := move.(map[string]interface{})
					fmt.Println(index, pid)
					positions[pid] = Move{
						Direction: castedMove["direction"].(string),
						X: int(castedMove["x"].(float64)),
						Y: int(castedMove["y"].(float64)),
						
					}
				}
			}
			fmt.Println(gameState.Positions)
			addressState.recvChan <- buf
			break
		case "gameover":
			//TODO test if its working
			gameOver = true
			logClient("Cloosing Client")
			break
		case "newleader":
			fmt.Println("Notified about new leader")
			electionState = NORMAL
			break
		case "checkleader":
			var message LeaderElectionMessage
			//TODO add a condition on leaderid
			if gameState.Round > getRoundNumber(buf) {
				message = LeaderElectionMessage{
					MessageType: "leaderalive",
					Round: gameState.Round,
					LeaderID: gameState.LeaderID,
				}
			} else {
				pid, _ := strconv.Atoi(gameState.AddrToPid[raddr.String()])
				if electionState == QUORUM && pid < gameState.MyPid {
					electionState = NORMAL
					close(bufChan)
				}
				message = LeaderElectionMessage{
					MessageType: "leaderdead",
					Round: gameState.Round,
					LeaderID: gameState.LeaderID,
				}
			}
			byt := encodeMessage(message)
			_, err := addressState.goConnection.WriteToUDP(byt, raddr)
			checkError(err)
			break
		case "leaderalive", "leaderdead":
			bufChan <- buf
			break
		default:
			panic("Cannot understand message type: " + messageType)
		}
	}
	logClient("Cloosing Client")
}

func javaGoConnection(wg sync.WaitGroup) {
	defer wg.Done()
	initializeJavaConnection()
	defer addressState.javaConnection.Close()
	for {
		message := <- addressState.recvChan
		messageType := getMessageType(message)
		switch messageType {
		case "roundstart":
			logJava("Sending the following message to java:" + string(message))
			addressState.javaConnection.Write(append(message, '\n'))
			logJava("Message has been sent to java")
			if isGameOverMessage(message) {
				logJava("A Game Over was sent to java. My work here is done. Goodbye")
				break
			}
			// read some reply from the java game (update of move, or death)
			time.Sleep(MIN_GAME_SPEED)
			status, err := addressState.connBuf.ReadString('\n')
			checkError(err)
			logJava("Received from java " + status)
			addressState.sendChan <- []byte(status)
			fmt.Println("sent")
			break
		case "moves":
			logJava("Reply from java " + string(message))
			addressState.javaConnection.Write(append(message, '\n'))
			break
		default:
			panic("Message from java not recognized: " + messageType)
		}
	}
}

/*
* Leader election
*/

func startElection(bufChan chan []byte) {
	electionState = QUORUM
	message := LeaderElectionMessage{
		MessageType: "checkleader",
		Round: gameState.Round,
		LeaderID: gameState.LeaderID,
	}
	byt := encodeMessage(message)
	broadcastMessage(addressState.goConnection, byt)
	received := 1
	positive := 1
	for {
		buf, timedout := <- bufChan
		if timedout || received == len(gameState.AddrToPid) - 2 {
			break
		}
		messageType := getMessageType(buf)
		received++
		if messageType == "leaderdead" {
			positive++
		}
	}
	if positive > received / 2 {
		electNewLeader()
	}
}

func electNewLeader() {
	newLeaderConn := initializeLeader(":0")
	leaderState.Positions = gameState.Positions
	gameState.LeaderID++
	message := LeaderElectionMessage{
		MessageType: "newleader",
		Round: gameState.Round,
		LeaderID: gameState.LeaderID,
	}
	byt := encodeMessage(message)
	broadcastMessage(newLeaderConn, byt)
	time.Sleep(1000*time.Millisecond)
	go leaderListener(newLeaderConn)
	//TODO sleep might be needed
	//TODO not sure if working
	splitAddress := strings.Split(newLeaderConn.LocalAddr().String(), ":")
	addressState.leaderAddr = "localhost:" + splitAddress[len(splitAddress)-1]
	initializeLeaderConnection()
	//electionState = NORMAL
	fmt.Println("New leader elected")
}


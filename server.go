package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math/rand"
	"net"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
//	"sync"
	"time"

	"bitbucket.org/bestchai/dinv/govec"
)

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

//type GameState struct {
//	//refactor this
var gameState_LeaderID       int
var gameState_Round          int
var gameState_MyPid          int
var gameState_GridWidth      int
var gameState_GridHeight     int
var gameState_Grid           [][]int
var gameState_Nickname       string
var gameState_Positions      []map[string]Move
var gameState_Alive          map[string]bool
var gameState_Grace          map[string]int
var gameState_Finish         []string
var gameState_AddrToPid      map[string]string
var gameState_AddrToAddr     map[string]*net.UDPAddr
var gameState_PidToNickname  map[string]string
var gameState_DroppedForever map[string]bool
//}

type LeaderState struct {
	Positions        []map[string]Move
	leaderConnection *net.UDPConn
}

//type AddressState struct {
var addressState_LocalIP        string
var addressState_javaAddr       string
var addressState_leaderAddr     string
var addressState_leaderUDPAddr  *net.UDPAddr
var addressState_isLeader       bool
var addressState_goConnection   *net.UDPConn
var addressState_javaConnection net.Conn
var addressState_connBuf        *bufio.Reader
var addressState_sendChan       chan []byte
var addressState_recvChan       chan []byte
//}

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

type KillPlayerMessage struct {
	MessageType string `json:"messageType"`
	EventName   string `json:"eventName"`
	PlayerPID   string `json:"playerpid"`
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
	Round       int    `json:"round"`
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

//var gameState GameState
//var addressState AddressState
var leaderState LeaderState
var electionState = NORMAL
var lastTime time.Time

var DISABLE_GAME_OVER = false // to allow single player game for debugging
var COLLISION_IS_DEATH = true
var MIN_GAME_SPEED = 1000 / 40 * time.Millisecond        // time between every new java move
var FOLLOWER_RESPONSE_TIME = 500 * 4 * time.Millisecond  // time for followers to respond
var MAX_ALLOWABLE_MISSED_MESSAGES = 5                    // max number of consecutive missed messages
var FOLLOWER_RESPONSE_FAIL_RATE = map[string]int{"1": 0} // out of 1000, fail rate for responses not to be received
var METRICS = false                                       // disable metrics

var ROUND_LATENCY_FILENAME = ""
var READ_THROUGHPUT_FILENAME = ""
var WRITE_THROUGHPUT_FILENAME = ""


var DIRECTIONS = [...]string{"DOWN", "LEFT", "UP", "RIGHT"}

var Logger *govec.GoLog

/*
* READ FROM UDP
 */

func readFromUDPWithTimeout(conn *net.UDPConn, timeoutTime time.Time) ([]byte, *net.UDPAddr, bool) {
	//@dump
	buf := make([]byte, 4096)
	conn.SetReadDeadline(timeoutTime)
	n, raddr, err := conn.ReadFromUDP(buf[0:])

	if err != nil {
		if e, ok := err.(net.Error); !ok || !e.Timeout() {
			checkError(err)
			return nil, nil, false
		} else {
			// timeout
			// TODO: how to deal with udp timeout and govec unpack recv? 
			fmt.Println("timed out")
			return nil, raddr, true
		}
	} else {
		buf2 := Logger.UnpackReceive("Received", buf[0:])
		//@dump
		recordReadThroughput(n)

		return buf2, raddr, false
	}
}

func readFromUDP(conn *net.UDPConn) ([]byte, *net.UDPAddr) {
	buf := make([]byte, 4096)
	n, raddr, err := conn.ReadFromUDP(buf[0:])
	buf2 := Logger.UnpackReceive("Received", buf[0:])
	checkError(err)
	recordReadThroughput(n)

	return buf2, raddr
}

/*
* MATH UTILITIES
 */
func randomDir() string {
	return DIRECTIONS[randomInt(0, 4)]
}

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
	Logger.LogLocalEvent(message)
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
* PERFORMANCE METRIC UTILITIES
 */
// TODO: measure at different failure rates, write to csv file

// what is the smoothness/time delta between consecutive rounds?

func recordRoundLatency() {
	if !METRICS {
		return
	}
	diff := time.Since(lastTime)

	csvfile, err := os.OpenFile(ROUND_LATENCY_FILENAME, os.O_APPEND|os.O_WRONLY, 0600)
	checkError(err)
	defer csvfile.Close()
	writer := csv.NewWriter(csvfile)

	err = writer.Write([]string{strconv.Itoa(gameState_Round),
		strconv.Itoa(gameState_LeaderID), strconv.FormatInt(diff.Nanoseconds(), 10)})
	checkError(err)
	writer.Flush()

	lastTime = time.Now()
}

// length of messages received by this player by round
func recordReadThroughput(n int) {
	if !METRICS {
		return
	}
	csvfile, err := os.OpenFile(READ_THROUGHPUT_FILENAME, os.O_APPEND|os.O_WRONLY, 0600)
	checkError(err)
	defer csvfile.Close()
	writer := csv.NewWriter(csvfile)

	err = writer.Write([]string{strconv.Itoa(gameState_Round),
		strconv.Itoa(gameState_MyPid), strconv.Itoa(n)})
	checkError(err)
	writer.Flush()
}

func recordWriteThroughput(n int) {
	if !METRICS {
		return
	}
	csvfile, err := os.OpenFile(WRITE_THROUGHPUT_FILENAME, os.O_APPEND|os.O_WRONLY, 0600)
	defer csvfile.Close()
	writer := csv.NewWriter(csvfile)

	err = writer.Write([]string{strconv.Itoa(gameState_Round),
		strconv.Itoa(gameState_MyPid), strconv.Itoa(n)})
	checkError(err)
	writer.Flush()
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

func getPID(buf []byte) (pid string) {
	dat := decodeMessage(buf)
	pid = dat["playerpid"].(string)
	return
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

func getLeaderID(buf []byte) (ID int) {
	dat := decodeMessage(buf)
	IDString, _ := dat["leaderid"].(float64)
	ID = int(IDString)
	return
}

func getMoves(buf []byte) (moves []interface{}) {
	dat := decodeMessage(buf)
	moves = dat["moves"].(map[string]interface{})["moves"].([]interface{})
	return
}

func broadcastMessage(conn *net.UDPConn, message []byte) {
	for _, addr := range gameState_AddrToAddr {
		_, err := conn.WriteToUDP(Logger.PrepareSend("", message), addr)
		//@dump
		recordWriteThroughput(len(message))
		checkError(err)
		logLeader("Sent message " + string(message) + " to player " + addr.String())
	}
}

/*
* MESSAGE CONSTRUCTORS
 */
func newRoundMessage() []byte {
	gameState_Round++
	message := RoundStartMessage{
		MessageType: "roundstart",
		EventName:   "roundStart",
		Round:       gameState_Round,
		RoundStart: RoundStart{
			Round: gameState_Round,
		},
	}
	return encodeMessage(message)
}

func startGameMessage(pid string, startingPositions map[string]Move) GameStartMessage {
	return GameStartMessage{
		MessageType: "startgame",
		EventName:   "gameStart",
		Round:       gameState_Round,
		GameStart: GameStart{
			Pid:               pid,
			StartingPositions: startingPositions,
			Nicknames:         gameState_PidToNickname,
			Addresses:         gameState_AddrToPid,
		},
	}
}

func endGameMessage() GameOverMessage {
	return GameOverMessage{
		MessageType: "gameOver",
		EventName:   "gameOver",
		Round:       gameState_Round,
		GameOver: GameOver{
			PidsInOrderOfDeath: gameState_Finish,
		},
	}
}

func newRound(conn *net.UDPConn) (roundMoves MovesMessage) {
	newRoundMessage := newRoundMessage()
	broadcastMessage(conn, newRoundMessage)
	logLeader("Done sending round start messages.")
	slideWindow()
	roundMoves = MovesMessage{
		MessageType: "moves",
		EventName:   "moves",
		Round:       gameState_Round,
		Moves: Moves{
			Moves: leaderState.Positions,
			Round: gameState_Round,
		},
	}
	return
}

/*
* INIT FUNCTIONS
 */
func CreateInitPlayerPosition() Move {
	direction := randomDir()
	// Give a small buffer so they don't crash into a wall immediately!
	if (gameState_GridWidth < 30) || (gameState_GridHeight < 30) {
		panic("game grid dimensions are too small!")
	}
	return Move{
		X:         randomInt(15, gameState_GridWidth-15),
		Y:         randomInt(15, gameState_GridHeight-15),
		Direction: direction,
	}
}

func registerNewPlayer(raddr *net.UDPAddr, buf []byte) {
	address := raddr.String()
	pid := strconv.Itoa(len(getLeaderMoveMap()) + 1)
	getLeaderMoveMap()[pid] = CreateInitPlayerPosition()
	gameState_Alive[pid] = true
	gameState_AddrToPid[address] = pid
	gameState_AddrToAddr[address] = raddr
	gameState_Alive[pid] = true
	nickname := strings.Split(string(buf), ":")[1]
	gameState_PidToNickname[pid] = nickname
	logLeader("New player named " + nickname + " has joined from address " + address)
	logLeader("Assigning pid " + pid + " and starting position " + strconv.Itoa(getLeaderMoveMap()[pid].X) +
		"," + strconv.Itoa(getLeaderMoveMap()[pid].Y))
	
}

func initLobby() {
	for {
		logLeader("Waiting for a client to join or send a start game message")
		buf, raddr, timedout := readFromUDPWithTimeout(leaderState.leaderConnection, time.Now().Add(time.Second * 15))
		if timedout || isStartMessage(buf) {
			logLeader("Start of the game, sending broadcast")
			for addr, pid := range gameState_AddrToPid {
				newGameMsg := encodeMessage(startGameMessage(pid, getLeaderMoveMap()))
				logLeader("Sending a game start message to " + addr + ". " + string(newGameMsg))
				_, err := leaderState.leaderConnection.WriteToUDP(Logger.PrepareSend("", newGameMsg), gameState_AddrToAddr[addr])
				//@dump
				checkError(err)
			}
			break
		} else if isJoinMessage(buf) {
			address := raddr.String()
			if _, knownPlayer := gameState_AddrToPid[address]; !knownPlayer {
				registerNewPlayer(raddr, buf)
			}
		} else {
			panic("Message not recognized:  " + string(buf))
		}
	}
}

func initializeLeader(leaderAddrString string) {
	leaderState.Positions = make([]map[string]Move, MAX_ALLOWABLE_MISSED_MESSAGES)
	for i := 0; i < len(leaderState.Positions); i++ {
		leaderState.Positions[i] = make(map[string]Move)
	}
	leaderAddr, err := net.ResolveUDPAddr("udp", leaderAddrString)
	checkError(err)
	conn, err := net.ListenUDP("udp", leaderAddr)
	fmt.Println(conn.LocalAddr(), conn.RemoteAddr())
	checkError(err)
	leaderState.leaderConnection = conn
}

func initializeConnection() {
	addr, err := net.ResolveUDPAddr("udp", addressState_LocalIP+":0")
	checkError(err)
	conn, err := net.ListenUDP("udp", addr)
	checkError(err)
	addressState_goConnection = conn
}

func initializeLeaderConnection() {
	leaderUDPAddr, err := net.ResolveUDPAddr("udp", addressState_leaderAddr)
	checkError(err)
	addressState_leaderUDPAddr = leaderUDPAddr
}

func contactLeader() {
	initializeConnection()
	initializeLeaderConnection()
	fmt.Print("GOCLIENT: ")
	fmt.Println("Sending a hello message to the leader", addressState_leaderAddr, addressState_goConnection.LocalAddr(), addressState_goConnection.RemoteAddr())
	_, err := addressState_goConnection.WriteToUDP(Logger.PrepareSend("", []byte("JOIN:"+gameState_Nickname)),
		addressState_leaderUDPAddr)
	logClient("hello")
	checkError(err)
	if addressState_isLeader {
		message := <-addressState_sendChan
		logClient("Go client got START message. Sending to leader")
		_, err = addressState_goConnection.WriteToUDP(Logger.PrepareSend("", []byte(message)), addressState_leaderUDPAddr)
		checkError(err)

	}

	buf, _ := readFromUDP(addressState_goConnection)
	logClient("Received a game start response from the leader:" + string(buf))
	fmt.Println("not this one")
	dat := decodeMessage(buf)["gameStart"].(map[string]interface{})
	pid, _ := strconv.Atoi(dat["pid"].(string))
	gameState_MyPid = pid
	addresses := dat["addresses"].(map[string]interface{})
	for addr, pid := range addresses {
		raddr, err := net.ResolveUDPAddr("udp", addr)
		checkError(err)
		gameState_AddrToPid[addr] = pid.(string)
		gameState_AddrToAddr[addr] = raddr
		gameState_Alive[pid.(string)] = true
	}
	addressState_recvChan <- buf
}

func initializeJavaConnection() {
	logJava("Trying to connect to java on " + addressState_javaAddr)
	conn, err := net.Dial("tcp", addressState_javaAddr)
	addressState_connBuf = bufio.NewReader(conn)
	checkError(err)
	if addressState_isLeader {
		str, err := addressState_connBuf.ReadString('\n')
		checkError(err)
		logJava("Received a start game message from java: " + str)
		addressState_sendChan <- []byte(str)
	}
	logJava("Waiting for the go message to send to java")
	reply := <-addressState_recvChan
	logJava("reply " + string(reply))
	conn.Write(append(reply, '\n'))
	logJava("Wrote game start message to java. Lobby phase over, entering main loop")
	addressState_javaConnection = conn
}

func initializeGameState() {
	var err error
	gameState_LeaderID = 0
	gameState_Round = 1
	gameState_GridWidth, err = strconv.Atoi(os.Args[4])
	checkError(err)
	gameState_GridHeight, err = strconv.Atoi(os.Args[5])
	checkError(err)
	gameState_Nickname = os.Args[6]
	hash := fnv.New64a()
	hash.Write([]byte(gameState_Nickname))
	rand.Seed(time.Now().Unix() + int64(hash.Sum64()))
	gameState_Positions = make([]map[string]Move, MAX_ALLOWABLE_MISSED_MESSAGES)
	for i := 0; i < len(gameState_Positions); i++ {
		gameState_Positions[i] = make(map[string]Move)
	}
	gameState_Grid = make([][]int, gameState_GridWidth)
	for i := 0; i < gameState_GridWidth; i++ {
		gameState_Grid[i] = make([]int, gameState_GridHeight)
	}
	// walls
	for i := 0; i < gameState_GridHeight; i++ {
		gameState_Grid[i][0] = -1
		gameState_Grid[i][gameState_GridHeight-1] = -1

	}
	for j := 0; j < gameState_GridWidth; j++ {
		gameState_Grid[0][j] = -1
		gameState_Grid[gameState_GridWidth-1][j] = -1
	}
	gameState_Alive = make(map[string]bool)
	gameState_Grace = make(map[string]int)
	gameState_AddrToPid = make(map[string]string)
	gameState_AddrToAddr = make(map[string]*net.UDPAddr)
	gameState_PidToNickname = make(map[string]string)
	gameState_DroppedForever = make(map[string]bool)

	isLeader, err := strconv.ParseBool(os.Args[3])
	checkError(err)

	addressState_javaAddr = "localhost:" + os.Args[1]
	addressState_leaderAddr = os.Args[2]
	addressState_isLeader = isLeader
	addressState_sendChan = make(chan []byte, 1)
	addressState_recvChan = make(chan []byte, 1)

	//@dump
}

func initializePerformanceMetrics() {
	if !METRICS {
		return
	}
	latencyFile, err := os.Create(ROUND_LATENCY_FILENAME)
	checkError(err)
	defer latencyFile.Close()

	readFile, err := os.Create(READ_THROUGHPUT_FILENAME)
	checkError(err)
	defer readFile.Close()

	writeFile, err := os.Create(WRITE_THROUGHPUT_FILENAME)
	defer writeFile.Close()

	lastTime = time.Now()
}

/*
* CHECK FUNCTIONS
 */
func isAi() bool {
	return addressState_javaAddr == addressState_LocalIP + ":" || len(os.Args) == 8
}

func gameOver() bool {
	if DISABLE_GAME_OVER {
		return false
	}

	numDeadToEnd := len(gameState_Alive) - 1
	//numDeadToEnd := 1 // for metric debugging
	logLeader("There are " + strconv.Itoa(len(gameState_Finish)) + " dead players and we require at least " +
		strconv.Itoa(numDeadToEnd) + " dead players to call it a game")
	return len(gameState_Finish) >= numDeadToEnd
}

func resetGracePeriod(pid string) {
	gameState_Grace[pid] = 0
}

func countGracePeriod(pid string) {
	if !gameState_DroppedForever[pid] {
		gameState_Grace[pid] += 1
		logLeader("Player " + pid + " grace period = " + strconv.Itoa(gameState_Grace[pid]))
		if gameState_Grace[pid] >= MAX_ALLOWABLE_MISSED_MESSAGES {
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

/*
* UTILITY FUNCTIONS
 */

func checkError(err error) {
	if err != nil {
		debug.PrintStack()
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		os.Exit(1)
	}
}

// Slight difference from killing player; we owe no obligation to respond to dropped players,
// but we should respond to killed players that are still connected.
func dropPlayer(pid string) {
	logLeader("dropping player " + pid)
	gameState_DroppedForever[pid] = true
	killPlayer(pid)
}

func killPlayer(pid string) {
	if gameState_Alive[pid] {
		logLeader("killing player " + pid)
		gameState_Alive[pid] = false
		gameState_Finish = append(gameState_Finish, pid)
	}
	if gameOver() {
		for pid, alive := range gameState_Alive {
			if alive {
				gameState_Finish = append(gameState_Finish, pid)
				logLeader("Player " + pid + " is the winner! Congrats!")
			}
		}
	}
}

func getCurrentMoveMap() map[string]Move {
	return gameState_Positions[len(gameState_Positions)-1]
}

func getLeaderMoveMap() map[string]Move {
	return leaderState.Positions[len(leaderState.Positions)-1]
}

func slideWindow() {
	// push everything back
	for i := 0; i < len(leaderState.Positions)-1; i++ {
		leaderState.Positions[i] = leaderState.Positions[i+1]
	}
	// clear final move
	leaderState.Positions[len(leaderState.Positions)-1] = make(map[string]Move)
}

//TODO Change this name
func addContinuedMove(pid string) {
	fmt.Println("adding continued move")
	prevMove := leaderState.Positions[len(leaderState.Positions)-2][pid]
	makeMove(prevMove.Direction, pid)
}

func createContinuedMove(direction string, prevMove Move) Move {
	nextMove := Move{
		Direction: direction,
		X:         prevMove.X,
		Y:         prevMove.Y,
	}
	switch nextMove.Direction {
	case "DOWN":
		nextMove.Y = max(0, nextMove.Y-1)
	case "UP":
		nextMove.Y = min(gameState_GridHeight-1, nextMove.Y+1)
	case "LEFT":
		nextMove.X = max(0, nextMove.X-1)
	case "RIGHT":
		nextMove.X = min(gameState_GridWidth-1, nextMove.X+1)
	default:
		panic("Next move direction unknown")
	}
	return nextMove
}

// Attempt to move the player pid one space in given direction. If movement
// results in collision, the player dies.
func makeMove(direction string, pid string) Move {
	prevMove := leaderState.Positions[len(leaderState.Positions)-2][pid]
	fmt.Println("Make move", direction, pid, prevMove)
	var nextMove Move
	if gameState_Alive[pid] {
		nextMove = createContinuedMove(direction, prevMove)
		if isCollision(nextMove.X, nextMove.Y) {

			message := KillPlayerMessage{
				MessageType: "killplayer",
				EventName:   "killplayer",
				PlayerPID:   pid,
				Round:       gameState_Round,
			}
			broadcastMessage(leaderState.leaderConnection, encodeMessage(message))
			nextMove = prevMove
		} else {
			gameState_Grid[nextMove.X][nextMove.Y], _ = strconv.Atoi(pid)
		}
	} else { // Player is dead, keep old move.
		nextMove = prevMove
	}
	fmt.Println("done")
	fmt.Println(nextMove)
	fmt.Println(getLeaderMoveMap())
	getLeaderMoveMap()[pid] = nextMove
	fmt.Println(getLeaderMoveMap())
	return nextMove
}

func surviveFollowerResponseInjectedFailure(pid string) bool {
	if val, ok := FOLLOWER_RESPONSE_FAIL_RATE[pid]; ok {
		p := randomInt(0, 1000)
		return p >= val
	}
	return true
}

func updateGracePeriod() {
	// count missed messages for those who did not respond, or reset
	for pid, alive := range gameState_Alive {
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
}

func timeToRespond() bool {
	recvCount := len(getLeaderMoveMap())
	totalNeeded := len(gameState_AddrToPid) - len(gameState_DroppedForever)
	logLeader("received " + strconv.Itoa(recvCount) + "/" + strconv.Itoa(totalNeeded) + " messages")
	return recvCount == totalNeeded
}

func isCollision(x, y int) bool {
	if !COLLISION_IS_DEATH {
		return false
	} else if (0 <= x && x < gameState_GridWidth) && (0 <= y && y < gameState_GridHeight) {
		if gameState_Grid[x][y] != 0 {
			fmt.Println("                      collision at", x, y, gameState_Grid[x][y])
		}
		return gameState_Grid[x][y] != 0
	} else {
		fmt.Println("                      collision out of bound")
		return true
	}
}

/*
* MAIN FUNCTIONS
 */
func main() {

	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)

	gameState_Nickname = os.Args[6]
	name := "server-" + gameState_Nickname
	Logger = govec.Initialize(name, name + ".log")
	//@dump	

	initializePerformanceMetrics()
	initializeGameState()

	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			addressState_LocalIP = ipv4.String()
			break
		}
	}
	if addressState_LocalIP == "" {
		panic("No open ipv4 interface")
	}
	fmt.Println("Go process started")
	if FOLLOWER_RESPONSE_TIME < MIN_GAME_SPEED {
		panic("Can't set response time to be less than min game speed")
	}
	if len(os.Args) < 7 {
		panic("RTFM")
	}



	if addressState_isLeader {
		go func() {
			initializeLeader(addressState_leaderAddr)
			logLeader("Leader has started")
			initLobby()
			go leaderListener()
		}()
		//TODO i'm pretty sure there's a better way than that
		time.Sleep(100 * time.Millisecond) // stupid hack to make sure the leader is up before the client
	}

	goClient()
	
	fmt.Println("GOODBYE")
}

/*
* Routines
 */
func leaderListener() {
	defer leaderState.leaderConnection.Close()
	var roundMoves MovesMessage
	var timeoutTimeForRound time.Time
	for {
		roundMoves = newRound(leaderState.leaderConnection)
		fmt.Println("newRoundMoves", roundMoves)
		timeoutTimeForRound = time.Now().Add(FOLLOWER_RESPONSE_TIME)
		for {
			logLeader("Waiting to receive message from follower...")
			buf, _, timedout := readFromUDPWithTimeout(leaderState.leaderConnection, timeoutTimeForRound)
			if timedout {
				break
			}
			direction, pid, round := parseMessage(buf)
			if round == gameState_Round && !gameState_DroppedForever[pid] {
				if surviveFollowerResponseInjectedFailure(pid) {
					logLeader("Received move message " + " from player " + pid)
					fmt.Println(leaderState.Positions)
					move := makeMove(direction, pid)
					fmt.Println("Registered move ", move)
					if timeToRespond() {
						break
					}
				}
			} else {
				logLeader("Received a move message from " + pid + " from an old round " +
					strconv.Itoa(round) + " but current round is " +
					strconv.Itoa(gameState_Round) + ". Ignoring message")
			}
		}
		updateGracePeriod()
		if gameOver() {
			logLeader("Broadcasting end of game!")
			broadcastMessage(leaderState.leaderConnection, encodeMessage(endGameMessage()))
			return
		}
		byt := encodeMessage(roundMoves)
		fmt.Println("roundmvoes", roundMoves)
		broadcastMessage(leaderState.leaderConnection, byt)
	}
}

func goClient() {
	var bufChan chan []byte
	gameOver := false

	logClient("Starting go client")
	
	if isAi() {
		logClient("I'm an AI player named " + gameState_Nickname)
		go aiGoConnection()
	} else {
		logClient("I'm a normal player named " + gameState_Nickname)
		go javaGoConnection()
	}

	contactLeader()
	defer addressState_goConnection.Close()
	logClient("Waiting for leader to respond with game start details")

	for !gameOver {
		fmt.Println("setitng up timeout", time.Now(), FOLLOWER_RESPONSE_TIME)
		timeoutTimeForRound := time.Now().Add(FOLLOWER_RESPONSE_TIME)
		fmt.Println(timeoutTimeForRound)
		leaderID := gameState_LeaderID
		buf, raddr, timedout := readFromUDPWithTimeout(addressState_goConnection, timeoutTimeForRound)
		if timedout {
			fmt.Println(time.Now())
			if leaderID != gameState_LeaderID {
				fmt.Println("$$$$$$$$$$$$$$$$$$")
				continue
			}
			fmt.Println("timedout", gameState_LeaderID, leaderID)
			if electionState == NORMAL {
				electionState = QUORUM
				fmt.Println("new election")
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
		case "killplayer":
			//check round rumber?
			killPlayer(getPID(buf))
		case "roundstart":
			logClient("Round start message: " + string(buf))
			recordRoundLatency()

			gameState_Round = getRoundNumber(buf)
			addressState_recvChan <- buf
			message := <-addressState_sendChan

			_, err := addressState_goConnection.WriteToUDP(Logger.PrepareSend("", []byte(message)), addressState_leaderUDPAddr)
			recordWriteThroughput(len(message))
			checkError(err)
			break
		case "moves":
			logClient("Moves message: " + string(buf))
			for index, moves := range getMoves(buf) {
				positions := gameState_Positions[index]
				for pid, move := range moves.(map[string]interface{}) {
					castedMove := move.(map[string]interface{})
					move := Move{
						Direction: castedMove["direction"].(string),
						X:         int(castedMove["x"].(float64)),
						Y:         int(castedMove["y"].(float64)),
					}
					positions[pid] = move
					// If you are the leader, you should probably not do this board update (shared state between the leader and client always gets us in to trouble...)
					if !addressState_isLeader {
						gameState_Grid[move.X][move.Y], _ = strconv.Atoi(pid)
					}
				}
			}
			addressState_recvChan <- buf
			break
		case "gameOver":
			gameOver = true
			addressState_recvChan <- buf
			logClient("Closing Client")
			return
		case "newleader":
			fmt.Println("Notified about new leader", raddr.String())
			addressState_leaderAddr = raddr.String()
			initializeLeaderConnection()
			newLeaderId := getLeaderID(buf)
			addressState_isLeader = newLeaderId == gameState_LeaderID
			gameState_LeaderID = newLeaderId
			electionState = NORMAL
			break
		case "checkleader":
			var message LeaderElectionMessage
			if gameState_Round > getRoundNumber(buf) || gameState_LeaderID > getLeaderID(buf) {
				message = LeaderElectionMessage{
					MessageType: "leaderalive",
					Round:       gameState_Round,
					LeaderID:    gameState_LeaderID,
				}
			} else {
				pid, _ := strconv.Atoi(gameState_AddrToPid[raddr.String()])
				if electionState == QUORUM && pid < gameState_MyPid {
					electionState = NORMAL
					close(bufChan)
				}
				message = LeaderElectionMessage{
					MessageType: "leaderdead",
					Round:       gameState_Round,
					LeaderID:    gameState_LeaderID,
				}
			}
			byt := encodeMessage(message)
			logLeader("Sent message " + string(byt))
			_, err := addressState_goConnection.WriteToUDP(Logger.PrepareSend("", byt), raddr)
			recordWriteThroughput(len(byt))
			checkError(err)
			break
		case "leaderalive", "leaderdead":
			if gameState_LeaderID >= getLeaderID(buf) {
				bufChan <- buf
			}
			break
		default:
			panic("Cannot understand message type: " + messageType)
		}
	}
	logClient("Closing Client")
}

func javaGoConnection() {
	initializeJavaConnection()
	defer addressState_javaConnection.Close()
	for {
		message := <-addressState_recvChan
		messageType := getMessageType(message)
		logJava("Sending a " + messageType + " message to java:" + string(message))
		switch messageType {
		case "roundstart":
			addressState_javaConnection.Write(append(message, '\n'))
			// read some reply from the java game (update of move, or death)
			time.Sleep(MIN_GAME_SPEED)
			status, err := addressState_connBuf.ReadString('\n')
			checkError(err)
			logJava("Received from java " + status)
			addressState_sendChan <- []byte(status)
			break
		case "moves":
			addressState_javaConnection.Write(append(message, '\n'))
			break
		case "gameOver":
			addressState_javaConnection.Write(append(message, '\n'))
			logJava("A Game Over was sent to java. My work here is done. Goodbye")
			return
		default:
			panic("Message to send to java not recognized: " + messageType)
		}
	}
	fmt.Println("Java connection closed")
}

func aiGoConnection() {
	_ = <-addressState_recvChan // drain the first message with the starting positions. We aren't sophisticated enough right now
	directionShuffleOrder := rand.Perm(len(DIRECTIONS))
	for {
		message := <-addressState_recvChan
		messageType := getMessageType(message)
		switch messageType {
		case "roundstart":
			time.Sleep(MIN_GAME_SPEED)

			// all AI goes here
			// Pick a random permutation of the directions ever X moves. Try each direction in order, and pick the first that doesn't give you a collision.
			direction := randomDir()
			prevMove := getCurrentMoveMap()[strconv.Itoa(gameState_MyPid)]
			if gameState_Round%30 == 0 {
				// reshuffle every x moves
				directionShuffleOrder = rand.Perm(len(DIRECTIONS))
			}
			for i := 0; i < len(DIRECTIONS); i++ {
				dir := DIRECTIONS[directionShuffleOrder[i]]
				move := createContinuedMove(dir, prevMove)
				if !isCollision(move.X, move.Y) {
					direction = dir
					break
				}
			}

			log("AI DECIDED TO MOVE: " + direction)
			move := map[string]interface{}{"eventName": "myMove", "direction": direction, "pid": strconv.Itoa(gameState_MyPid), "round": gameState_Round}
			addressState_sendChan <- []byte(encodeMessage(move))
			break
		case "moves":
			// do nothing, since we aren't adapting our strategy to the state of things
			break
		case "gameOver":
			log("A Game Over was sent to ai player. My work here is done. Goodbye")
			return
		default:
			panic("Message to AI not recognized: " + messageType)
		}
	}
	fmt.Println("AI OVER")
}

/*
* Leader election
 */

func startElection(bufChan chan []byte) {

	leaderID := gameState_LeaderID

	message := LeaderElectionMessage{
		MessageType: "checkleader",
		Round:       gameState_Round,
		LeaderID:    leaderID,
	}
	byt := encodeMessage(message)
	broadcastMessage(addressState_goConnection, byt)
	received := 0
	positive := 0
	for {
		buf, timedout := <-bufChan

		if !timedout {
			break
		}
		fmt.Println(timedout)
		fmt.Println(buf)
		fmt.Println(len(buf))

		if getLeaderID(buf) > leaderID {
			return
		}

		received++
		if getMessageType(buf) == "leaderdead" {
			positive++
		}
	}

	if positive > received/2 {
		if gameState_LeaderID == leaderID {
			electNewLeader()
		}
	}
}

func electNewLeader() {
	initializeLeader(":0")
	//leaderState.Positions = gameState_Positions
	for index, positions := range gameState_Positions {
		position := leaderState.Positions[index]
		for key, value := range positions {
			position[key] = value
		}
	}
	gameState_LeaderID++
	message := LeaderElectionMessage{
		MessageType: "newleader",
		Round:       gameState_Round,
		LeaderID:    gameState_LeaderID,
	}
	addressState_isLeader = true
	byt := encodeMessage(message)
	broadcastMessage(leaderState.leaderConnection, byt)
	go leaderListener()
	splitAddress := strings.Split(leaderState.leaderConnection.LocalAddr().String(), ":")
	addressState_leaderAddr = addressState_LocalIP + ":" + splitAddress[len(splitAddress)-1]
	initializeLeaderConnection()
	fmt.Println("New leader elected")
}

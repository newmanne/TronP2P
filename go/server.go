package main
//BUG we are not setting this
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

// GLOBAL VARS

var gameState GameState
var electionState = NORMAL
var DISABLE_GAME_OVER = true // to allow single player game for debugging
var COLLISION_IS_DEATH = true
var MIN_GAME_SPEED = 50 * time.Millisecond          // time between every new java move
var FOLLOWER_RESPONSE_TIME = 500 * time.Millisecond // time for followers to respond
var MAX_ALLOWABLE_MISSED_MESSAGES = 5               // max number of consecutive missed messages
var FOLLOWER_RESPONSE_FAIL_RATE = map[string]int{
	"1": 0,
} // out of 1000, fail rate for responses not to be received

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

type Client struct {
	addr *net.UDPAddr
	pid string
}

type GameState struct {
	Round          int
	MyPid          int
	MyPriority     int
	GridWidth      int
	GridHeight     int
	Positions      []map[string]Move
	Alive          map[string]bool
	Grace          map[string]int
	Finish         []string
	AddrToPid      map[string]Client
	PidToNickname  map[string]string
	DroppedForever map[string]bool
}

type RoundStart struct {
	Round int `json:"round"`
}

type RoundStartMessage struct {
	MessageType string `json:"messageType"`
	EventName  string `json:"eventName"`
	Round      int    `json:"round"`
	RoundStart `json:"roundStart"`
}

type MyMoveMessage struct {
	MessageType string `json:"messageType"`
	EventName string `json:"eventName"`
	Round     int    `json:"round"`
	MyMove    `json:"myMove"`
}

type MyDeathMessage struct {
	MessageType string `json:"messageType"`
	EventName string `json:"eventName"`
	Round     int    `json:"round"`
}

type MovesMessage struct {
	MessageType string `json:"messageType"`
	EventName string `json:"eventName"`
	Round     int    `json:"round"`
	Moves     `json:"moves"`
}

type Moves struct {
	Moves []map[string]Move `json:"moves"`
	Round int               `json:"round"`
}

type GameStart struct {
	Pid               string            `json:"pid"`
	StartingPositions map[string]Move   `json:"startingPositions"`
	Nicknames         map[string]string `json:"nicknames"`
}

type GameStartMessage struct {
	MessageType string `json:"messageType"`
	EventName string `json:"eventName"`
	Round     int    `json:"round"`
	GameStart `json:"gameStart"`
}

type GameOver struct {
	PidsInOrderOfDeath []string `json:"pidsInOrderOfDeath"`
}

type GameOverMessage struct {
	MessageType string `json:"messageType"`
	EventName string `json:"eventName"`
	GameOver  `json:"gameOver"`
}

type LeaderElectionMessage struct {
	MessageType string `json:"messageType"`
	EventName string `json:"eventName"`
	Round int `json:"round"`
}

type ElectionState int

const (
	NORMAL ElectionState = 1 + iota
	QUORUM
	NEWLEADER
)
/*
// GLOBAL VARS

var gameState GameState
var electionState = NORMAL
var DISABLE_GAME_OVER = true // to allow single player game for debugging
var COLLISION_IS_DEATH = true
var MIN_GAME_SPEED = 50 * time.Millisecond          // time between every new java move
var FOLLOWER_RESPONSE_TIME = 500 * time.Millisecond // time for followers to respond
var MAX_GRACE_PERIOD = 3                            // max number of consecutive missed messages
var FOLLOWER_RESPONSE_FAIL_RATE = 0                 // out of 1000, fail rate for responses not to be received
*/

// UTILITY FUNCTIONS

func readFromUDPWithTimeout(conn *net.UDPConn, timeoutTime time.Time) ([]byte, *net.UDPAddr, bool) {
	buf := make([]byte, 4096)
	conn.SetReadDeadline(timeoutTime)
	_, raddr, err := conn.ReadFromUDP(buf)
	if err != nil {
		if e, ok := err.(net.Error); !ok || !e.Timeout() {
			checkError(err)
			return nil, nil, false
		} else {
			// timeout
			return nil, raddr, true
		}
	} else {
		buf = bytes.Trim(buf, "\x00")
		return buf, raddr, false
	}
}

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

func newRoundMessage() []byte {
	gameState.Round++
	message := RoundStartMessage{MessageType: "roundstart", EventName: "roundStart", Round: gameState.Round, RoundStart: RoundStart{Round: gameState.Round}}
	return encodeMessage(message)
}

func getMessageType(buf []byte) (messageType string) {
	fmt.Println(buf)
	dat := decodeMessage(buf)
	fmt.Println(dat)
	messageType = dat["messageType"].(string)
	fmt.Println(messageType)
	return
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
	return Move{X: randomInt(1, gameState.GridWidth-2), Y: randomInt(1, gameState.GridHeight-2), Direction: direction}
}

func isJoinMessage(buf []byte) bool {
	return strings.Contains(strings.TrimSpace(string(buf)), "JOIN")
}

func isStartMessage(buf []byte) bool {
	return strings.TrimSpace(string(buf)) == "START"
}

func isGameOverMessage(message string) bool {
	return strings.Contains(message, "GameOver")
}

func startGameMessage(pid string, startingPositions map[string]Move) GameStartMessage {
	return GameStartMessage{MessageType: "startgame", EventName: "gameStart", Round: gameState.Round, GameStart: GameStart{Pid: pid, StartingPositions: startingPositions, Nicknames: gameState.PidToNickname}}
}

func endGameMessage() GameOverMessage {
	return GameOverMessage{MessageType: "gameover", EventName: "gameOver", GameOver: GameOver{PidsInOrderOfDeath: gameState.Finish}}
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

func slideWindow() {
	// push everything back
	for i := 0; i < len(gameState.Positions)-1; i++ {
		gameState.Positions[i] = gameState.Positions[i+1]
	}
	// clear final move
	gameState.Positions[len(gameState.Positions)-1] = make(map[string]Move)
}

// TODO: when the board state is moved over
func isCollision(x, y int) bool {
	return false
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
	// LOBBY PHASE
	// before the general main loop, wait for playerCount messages,
	// this will tell me who I need to send roundStarts to.
	
	// start of new game
	for {
		// wait for message from some client
		logLeader("Waiting for a client to join or send a start game message")
		buf, raddr := readFromUDP(conn)
		// what type of message is it? join or start game?
		if isJoinMessage(buf) {
			fmt.Println(raddr.String())
			if _, knownPlayer := gameState.AddrToPid[raddr.String()]; !knownPlayer {
				pid := strconv.Itoa(len(getCurrentMoveMap()) + 1)
				getCurrentMoveMap()[pid] = CreateInitPlayerPosition()
				gameState.Alive[pid] = true
				gameState.AddrToPid[raddr.String()] = Client{addr:raddr, pid: pid}
				gameState.Alive[pid] = true
				nickname := strings.Split(string(buf), ":")[1]
				gameState.PidToNickname[pid] = nickname
				logLeader("New player named " + nickname + " has joined from address " + raddr.String())
				logLeader("Assigning pid " + pid + " and starting position " + strconv.Itoa(getCurrentMoveMap()[pid].X) + "," + strconv.Itoa(getCurrentMoveMap()[pid].Y))
			}
		} else if isStartMessage(buf) {
			logLeader("The game start message has been sent! Notifying all players")
			// send message to all players to start game
			for addr, client := range gameState.AddrToPid {
				// TODO: we probably care about whether or not this one is received
				newGameMsg := encodeMessage(startGameMessage(client.pid, getCurrentMoveMap()))
				logLeader("Sending a game start message to " + addr + ". " + string(newGameMsg))
				_, err := conn.WriteToUDP(newGameMsg, client.addr)
				checkError(err)
			}
			// end lobby phase, prepare to send round messages next
			break
		} else {
			panic("WTF KIND OF MESSAGE IS THIS " + string(buf))
		}
	}
}

// determines if game is over (all players are dead except one)
func gameOver() bool {
	if DISABLE_GAME_OVER {
		return false
	}
	numDeadToEnd := len(gameState.AddrToPid) - 1
	logLeader("There are " + strconv.Itoa(len(gameState.Finish)) + " dead players and we require at least " + strconv.Itoa(numDeadToEnd) + " dead players to call it a game")
	return len(gameState.Finish) >= numDeadToEnd
}

func resetGracePeriod(pid string) {
	gameState.Grace[pid] = 0
	logLeader("player " + pid + " grace period = " + strconv.Itoa(gameState.Grace[pid]))
}

func countGracePeriod(pid string) {
	if !gameState.DroppedForever[pid] {
		gameState.Grace[pid] += 1
		logLeader("player " + pid + " grace period = " + strconv.Itoa(gameState.Grace[pid]))

		if gameState.Grace[pid] >= MAX_ALLOWABLE_MISSED_MESSAGES {
			logLeader("Grace period for player " + pid + " exceeded. Force dropping them")
			killPlayer(pid)
			gameState.DroppedForever[pid] = true
		}
	}
}

// if pid did not respond, their next move is to continue in next direction one move in advance
func addContinuedMove(pid string) {
	prevMove := gameState.Positions[len(gameState.Positions)-2][pid]
	makeMove(prevMove.Direction, pid)
}

func makeMove(direction string, pid string) Move {
	prevMove := gameState.Positions[len(gameState.Positions)-2][pid]
	nextMove := Move{Direction: direction, X: prevMove.X, Y: prevMove.Y}
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
	getCurrentMoveMap()[pid] = nextMove
	return nextMove
}

// to simulate missed responses; some of the time we will miss follower responses
// and resort to timeout; this is one of those times
func surviveFollowerResponseInjectedFailure(pid string) bool {
	if val, ok := FOLLOWER_RESPONSE_FAIL_RATE[pid]; ok {
		p := randomInt(0, 1000)
		return p >= val
	} else {
		return true
	}
}

// determines if leader can respond to followers yet
func timeToRespond(timedout bool) bool {
	recvCount := len(getCurrentMoveMap())
	totalNeeded := len(gameState.AddrToPid) - len(gameState.DroppedForever)
	logLeader("received " + strconv.Itoa(recvCount) + "/" + strconv.Itoa(totalNeeded) + " messages")

	if recvCount == totalNeeded || timedout {
		// count missed messages for those who did not respond, or reset
		for pid, alive := range gameState.Alive {
			_, responded := getCurrentMoveMap()[pid]
			if responded {
				resetGracePeriod(pid)
			} else {
				countGracePeriod(pid)
				// don't move them if they are dead
				if alive {
					addContinuedMove(pid)
				}
			}
		}
		return true
	} else {
		return false
	}
}

func broadcastMessage(conn *net.UDPConn, message []byte) {
	for _, client := range gameState.AddrToPid {
		_, err := conn.WriteToUDP(message, client.addr)
		checkError(err)
		logLeader("Sent message " + string(message) + " to player " + client.pid)
	}
}


func initializeLeader(leaderAddrString string) (conn *net.UDPConn) {
	// Listen
	leaderAddr, err := net.ResolveUDPAddr("udp", leaderAddrString)
	checkError(err)
	conn, err = net.ListenUDP("udp", leaderAddr)
	checkError(err)
	return
}


func newRound(conn *net.UDPConn) (roundMoves MovesMessage){
	newRoundMessage := newRoundMessage()
	broadcastMessage(conn, newRoundMessage)
	logLeader("done sending round start messages.")
	slideWindow()
	roundMoves = MovesMessage{MessageType: "moves", EventName: "moves", Round: gameState.Round, Moves: Moves{Moves: gameState.Positions, Round: gameState.Round}}
	return
}

// main leader function, approves moves of followers
func leaderListener(conn *net.UDPConn) {
	//conn := initializeLeader(leaderAddr)
	//logLeader("Leader has started")
	//initLobby(conn)
	defer conn.Close()
	// MAIN GAME LOOP
	isNewRound := true
	var roundMoves MovesMessage
	var timeoutTimeForRound time.Time
	for {
		// if a new round is starting, let everyone connected to me know
		if isNewRound {
			roundMoves = newRound(conn)
			isNewRound = false
			timeoutTimeForRound = time.Now().Add(FOLLOWER_RESPONSE_TIME)
		}
		// read messages from followers and forward them
		logLeader("Waiting to receive message from follower...")
		buf, _, timedout := readFromUDPWithTimeout(conn, timeoutTimeForRound)

		if !timedout {
			direction, pid, round := parseMessage(buf)
			if round == gameState.Round && !gameState.DroppedForever[pid] {
				// artifical missed response for testing
				received := surviveFollowerResponseInjectedFailure(pid)
				if received {
					move := makeMove(direction, pid)
					if COLLISION_IS_DEATH && isCollision(move.X, move.Y) {
						logLeader("Player " + pid + " is dead")
						killPlayer(pid)
					}
					getCurrentMoveMap()[pid] = move
					logLeader("Received move message " + string(encodeMessage(move)) + " from player " + pid)
				} else {
					logLeader("Fault injection removed message")
				}
			} else {
				logLeader("Recieved a move message from " + pid + " from an old round " + strconv.Itoa(round) + " but current round is " + strconv.Itoa(gameState.Round) + ". Ignoring message")
			}
		} else {
			logLeader("Timed out")
		}
		if gameOver() {
			break
		}

		// end condition; reply to my followers if I have been messaged by all of them
		if timeToRespond(timedout) {
			byt := encodeMessage(roundMoves)
			// send message to all followers
			broadcastMessage(conn, byt)
			// start a new round of communication
			isNewRound = true
		}
	}

	// END GAME SCREEN (RESULTS)
	broadcastMessage(conn, encodeMessage(endGameMessage()))
	logLeader("My work here as leader is done. Goodbye.")
}


func initializeConnection(sendChan chan string, leaderAddrString, nickname string, isLeader bool) (conn *net.UDPConn, leaderAddr *net.UDPAddr) {
	// Get a port for the go client to use
	addr, err := net.ResolveUDPAddr("udp", "localhost:0")
	checkError(err)
	conn, err = net.ListenUDP("udp", addr)
	checkError(err)
	// Resolve the leader address
	leaderAddr, err = net.ResolveUDPAddr("udp", leaderAddrString)
	checkError(err)
	// Write a message to the leader letting it know that I have started
	logClient("Sending a hello message to the leader")
	_, err = conn.WriteToUDP([]byte("JOIN:"+nickname), leaderAddr)
	checkError(err)
	if isLeader {
		message := <-sendChan
		logClient("Go client got START message. Sending to leader")
		_, err = conn.WriteToUDP([]byte(message), leaderAddr)
		checkError(err)
	}
	return
}

//NOTE maybe we don't need to define timeout outside. message needs to be defined outside to receive it though
//NOTE argh, so many parameters. solvable someway? maybe a struct since we're always passing the same parameters to everything?
func dealWithGameMessages(killChan chan bool, sendChan, recvChan chan string, messageChan, messageChan2 chan []byte, leaderAddr *net.UDPAddr, conn *net.UDPConn ) {	
	//NOTE need to interrupt it at some point	
//	var buf []byte
//	var timedout bool
	var timedout bool
	var buf []byte
	for {
		//NOTE current timeout keeps running. im afraid that if a new leader is not elected in time it will trigger another election. mayne is solvable with an ID? not sure though
		timeout := make(chan bool, 1)
		go func () {
			time.Sleep(FOLLOWER_RESPONSE_TIME)
			fmt.Println("TIMEOUT");
			timeout <- true
		}()
		select {
		case buf = <- messageChan:
			timedout = false;
			break
		case <-timeout:
			timedout = true;
			break
		}
		
		if timedout {
			if electionState == NORMAL {
				electionState = QUORUM
				go initElection(conn, killChan, messageChan2)
			}
			continue
		} else if electionState == QUORUM {
			electionState = NORMAL
			//BUG this is currently ineffective since killchan is not used inside
			killChan <- true
		}
		
		logClient("Got message " + string(buf) + " from leader, passing it to java")
		recvChan <- string(buf)

		if getMessageType(buf) == "roundstart" {
			dat := decodeMessage(buf)
			roundString, _ := dat["round"].(float64)
			gameState.Round = int(roundString)
			// wait for message from java
			message := <-sendChan
			// write message to leader address
			_, err := conn.WriteToUDP([]byte(message), leaderAddr)
			checkError(err)
		}
	}
}

func goClient(sendChan chan string, recvChan chan string, leaderAddrString string, wg sync.WaitGroup, isLeader bool, nickname string) {
	defer wg.Done()
	conn, leaderAddr := initializeConnection(sendChan, leaderAddrString, nickname, isLeader)
	defer conn.Close()
	// Read response from leader
	logClient("Waiting for leader to respond with game start details")

	buf, _ := readFromUDP(conn)
	logClient("Received a game start response from the leader:" + string(buf))

	// write back to channel with byte response (let java know to start)
	recvChan <- string(buf)

	logClient("LOBBY PHASE IS OVER. ENTERING MAIN LOOP")
	// MAIN GAME LOOP
	messageChan := make(chan []byte, 1)
	//TODO find a better name
	messageChan2 := make(chan []byte, 1)
	killChan := make(chan bool, 1)
	go dealWithGameMessages(killChan, sendChan, recvChan, messageChan, messageChan2, leaderAddr, conn)
	for {
		// read round start from leader
		buf, raddr := readFromUDP(conn)

		//NOTE might be worth to decode the message it and pass it decoded? not sure with java after though
		//NOTE might be worth to check the roundmessage here, we have a bunch of code duplication
		//NOTE I'm considering extracting checkleader code to another method
		switch getMessageType(buf) {
		case "roundstart":
			messageChan <- buf
			break;
		case "moves":
			messageChan <- buf
			break;
		case "gameover":
			logClient("Delivered a game over message. My work here is done. Goodbye")
			//BUG currently not functioning, need to break out of the loop in someway.
			//a bool tag should work, but its ugly
			break;
		case "newLeader":
			dat := decodeMessage(buf)
			roundString, _ := dat["round"].(float64)
			round := int(roundString)
			if gameState.Round  > round {
				break;
			} else {
				leaderAddr = raddr
			}
		case "checkleader":
			dat := decodeMessage(buf)
			roundString, _ := dat["round"].(float64)
			round := int(roundString)

			if gameState.Round > round {
				_, err := conn.WriteToUDP(encodeMessage(LeaderElectionMessage{MessageType: "leaderalive", Round: gameState.Round}), raddr)
				checkError(err)
				break
			}
			pid, _ := strconv.Atoi(gameState.AddrToPid[raddr.String()].pid)
			if electionState == QUORUM && pid > gameState.MyPid {
				electionState = NORMAL
				//BUG this is currently ineffective since killchan is not used inside
				killChan <- true
			} else {
				_, err := conn.WriteToUDP(encodeMessage(LeaderElectionMessage{MessageType: "leaderdead", Round: gameState.Round}), raddr)
				checkError(err)
			}
			break;
		case "leaderdead":
			messageChan2 <- buf
		case "leaderalive":
			messageChan2 <- buf
		default:
			panic("Cannot understdand message type")
		}
	}
	logClient("My work here as a client is done. Goodbye")
}

func initializeJavaConnection(sendChan chan string, javaAddrString string) (conn net.Conn) {
	logJava("Trying to connect to java on " + javaAddrString)
	conn, err := net.Dial("tcp", javaAddrString)
	checkError(err)
	return
}

func javaGoConnection(sendChan chan string, recvChan chan string, javaAddrString string, wg sync.WaitGroup, isLeader bool) {
	defer wg.Done()

	conn := initializeJavaConnection(sendChan, javaAddrString)
	defer conn.Close()
	
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
//	checkError(err)
	logJava("Wrote game start message to java. Lobby phase over, entering main loop")

	// MAIN LOOP
	for {
		// read round start message from channel and send it to java
		logJava("Waiting for message from go client")
		message := <-recvChan

		logJava("Sending the following message to java:" + message)
		conn.Write([]byte(message + "\n"))
		logJava("Message has been sent to java")
		if isGameOverMessage(message) {
			logJava("A Game Over was sent to java. My work here is done. Goodbye")
			break
		}

		// read some reply from the java game (update of move, or death)
		time.Sleep(MIN_GAME_SPEED)
		status, err := connBuf.ReadString('\n')

		
		logJava("Received: " + status)
		checkError(err)
		// send buf to leader channel
		sendChan <- status

		reply := <-recvChan

		conn.Write([]byte(reply + "\n"))
		checkError(err)
	}
}

func initializeGameState() {
	var err error
	// init vars
	rand.Seed(time.Now().Unix())
	gameState.Round = 1
	gameState.GridWidth, err = strconv.Atoi(os.Args[4])
	checkError(err)
	gameState.GridHeight, err = strconv.Atoi(os.Args[5])
	checkError(err)
	gameState.Positions = make([]map[string]Move, MAX_ALLOWABLE_MISSED_MESSAGES)
	for i := 0; i < len(gameState.Positions); i++ {
		gameState.Positions[i] = make(map[string]Move)
	}
	gameState.Alive = make(map[string]bool)
	gameState.Grace = make(map[string]int)
	gameState.AddrToPid = make(map[string]Client)
	gameState.PidToNickname = make(map[string]string)
	gameState.DroppedForever = make(map[string]bool)
}

func main() {
	fmt.Println("Go process started")
	if FOLLOWER_RESPONSE_TIME < MIN_GAME_SPEED {
		panic("Can't set response time to be less than min game speed")
	}

	// argument parsing
	if len(os.Args) != 7 {
		panic("RTFM")
	}
	initializeGameState()
	javaPort, leaderAddr, nickname := os.Args[1], os.Args[2], os.Args[6]
	isLeader, err := strconv.ParseBool(os.Args[3])
	checkError(err)
	sendChan, recvChan := make(chan string, 1), make(chan string, 1)

	// if I am the leader, listen for rounds to confirm them
	if isLeader {
		go func() {
			conn := initializeLeader(leaderAddr)
			logLeader("Leader has started")
			initLobby(conn)
			go leaderListener(conn)
		}()
		time.Sleep(100 * time.Millisecond) // stupid hack to make sure the leader is up before the client
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// handle communication through leader channel (follower code to leader)
	go goClient(sendChan, recvChan, leaderAddr, wg, isLeader, nickname)

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

func initElection(conn *net.UDPConn, killChan chan bool, messageChan chan []byte) {
	var buf []byte
	var count = 0
	var positive = 0

	message := LeaderElectionMessage{MessageType: "checkleader", Round: gameState.Round}
	byt := encodeMessage(message)
	broadcastMessage(conn, byt)

	timeout := make(chan bool, 1)
	timedout := false

	for !timedout {
		go func() {
			//NOTE this should be decreased maybe?
			time.Sleep(FOLLOWER_RESPONSE_TIME)
			timeout <- true
		}()
		select {
		case buf = <- messageChan:
			break
		case timedout = <- timeout:
			break
		}
		if timedout {
			break
		}
		dat := decodeMessage(buf)
		messageType, _ := dat["messagetype"].(string)
		count ++;
		if messageType == "leaderdead" {
			roundString, _ := dat["round"].(float64)
			round := int(roundString)
			if round >= gameState.Round {
				positive++
			}
		}		
	}
	if positive <  count/2 {
		electionState = NORMAL
		return
	}

	electionState = NEWLEADER

	/*elect leader */
	newLeaderConn := initializeLeader(":0")
	message = LeaderElectionMessage{MessageType: "newleader", Round: gameState.Round}
	byt = encodeMessage(message)
	broadcastMessage(newLeaderConn, byt)
	go leaderListener(newLeaderConn)
	electionState = NORMAL
}

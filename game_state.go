package main

import (
	"github.com/gorilla/websocket"
	"log"
	"strings"
	"sync"
	"time"
)

/*
	Game creation, player join, and start code
	i.e. how to websocket
*/
type newPlayer struct {
	socket    *websocket.Conn
	username  string
	maxMoveId *int32
	moves     chan PlayerMovement
}

var waiting_games map[string](chan newPlayer)
var active_games_lock = &sync.RWMutex{}

func initGlobals() {
	waiting_games = make(map[string](chan newPlayer))
}

/*
	Type defs for Game-specific code
	i.e. how do we store game data in memory

	Important!: if you want to write these as JSON,
	then you have to write them as capitals (to export)

	We should pack this using msgpack or protobufs
*/
type unitType int32

const (
	Empty = iota
	Mountain
	Town
	Capital
)

type cellValue struct {
	Owner    int32
	Troops   int32
	CellType unitType
}

func cellValueToProto(cell cellValue) SquareValue {
	return SquareValue{Owner: cell.Owner, Troops: cell.Troops, Type: SquareType(cell.CellType)}
}

type gameState struct {
	Cells   [][]cellValue
	Players []string
	Round   int
}

func gameExists(gameid string) bool {
	active_games_lock.RLock()
	if _, ok := waiting_games[gameid]; ok {
		active_games_lock.RUnlock()
		return true
	} else {
		active_games_lock.RUnlock()
		return false
	}
}

// we create a new game state here and pass it along to all other functions
func setupNewGame(gameid string) {
	active_games_lock.Lock()
	waiting_games[gameid] = make(chan newPlayer)
	new_game := gameState{}

	// start the gameLoop for that game
	go gameLoop(waiting_games[gameid], &new_game, gameid)
	active_games_lock.Unlock()
}

func endGame(gameid string) {
	active_games_lock.Lock()
	close(waiting_games[gameid])
	delete(waiting_games, gameid)
	active_games_lock.Unlock()
}

func addNewPlayer(username string, c *websocket.Conn, gameid string) {
	// we need to lock because apparently concurrent accesses to maps
	// are not thread-safe
	active_games_lock.Lock()
	// we need buffered channel so we initialize each user with 50 queueable moves
	waiting_games[gameid] <- newPlayer{
		username:  username,
		socket:    c,
		maxMoveId: new(int32),
		moves:     make(chan PlayerMovement, 50),
	}
	active_games_lock.Unlock()
}

func countTrue(arr map[string]bool) int {
	sum := 0
	for _, v := range arr {
		if v {
			sum += 1
		}
	}
	return sum
}

// we need this intermediate representation because we don't want to pack just yet
type socketMsg struct {
	msgType ServerMessageType
	data    interface{}
}

func broadcast(msgType ServerMessageType, data interface{}, chans []chan socketMsg) {
	for _, c := range chans {
		// so we don't block
		go func(c chan socketMsg) {
			c <- socketMsg{msgType: msgType, data: data}
		}(c)
	}
}

func cleanup(username string, ready *map[string]bool, players *[]newPlayer, websocketChans *[]chan socketMsg) {
	log.Printf("closing channels for player %s", username)
	// clean up! want to avoid memory leaks
	for i := 0; i < len(*players); i++ {
		if (*players)[i].username == username {
			*players = append((*players)[:i], (*players)[i+1:]...)
			close((*websocketChans)[i])
			(*websocketChans) = append((*websocketChans)[:i], (*websocketChans)[i+1:]...)
			break
		}
	}
	delete(*ready, username)
}

func gameLoop(playerChan chan newPlayer, game *gameState, gameid string) {
	log.Printf("entering game loop... %s", gameid)

	ready := make(map[string]bool)
	players := make([]newPlayer, 0)
	websocketChans := make([]chan socketMsg, 0)

	// websocket listeners can push to this when the player changes readiness
	playerConnect := make(chan string)

	for countTrue(ready) < len(players) || (len(players) < 2 && !cmd_flags.debug) || (len(players) < 1 && cmd_flags.debug) {
		select {
		case player := <-playerChan:
			players = append(players, player)
			newSocketChan := make(chan socketMsg)
			websocketChans = append(websocketChans, newSocketChan)
			ready[player.username] = false
			go websocketListener(&player, playerConnect, game, newSocketChan)
		case msg := <-playerConnect:
			log.Printf("received message: %s", msg)
			pieces := strings.Split(msg, ":")
			if pieces[0] == "Ready" {
				ready[pieces[1]] = true
			} else if pieces[0] == "Unready" {
				ready[pieces[1]] = false
			} else if pieces[0] == "Disconnect" {
				cleanup(pieces[1], &ready, &players, &websocketChans)
			}
		}
	}

	log.Printf("starting game: %s", gameid)

	// make a map of username to int
	usernameToID := make(map[string]int32)
	for i := 0; i < len(players); i++ {
		usernameToID[players[i].username] = int32(i + 1)
	}

	initializeTerrain(game, players, usernameToID)
	randomizeTerrain(game)
	broadcast(ServerMessageType_FullBoard, map[string]int32{}, websocketChans)

	for {
		time.Sleep(1000 * time.Millisecond)
		// unfortunately, when someone disconnects it means the update won't happen that frame
		select {
		case msg := <-playerConnect:
			pieces := strings.Split(msg, ":")
			if pieces[0] == "disconnect" {
				cleanup(pieces[1], &ready, &players, &websocketChans)
			}
		default:
			movesExecuted := updateGameState(players, game)
			broadcast(ServerMessageType_FullBoard, *movesExecuted, websocketChans)
		}
	}
}

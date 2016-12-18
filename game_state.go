package main

import (
	"github.com/gorilla/websocket"
	"log"
	"strings"
	"sync"
)

/*
	Game creation, player join, and start code
	i.e. how to websocket
*/
type newPlayer struct {
	socket   *websocket.Conn
	username string
}

var waiting_games map[string](chan newPlayer)
var active_games_lock = &sync.RWMutex{}

func initGlobals() {
	waiting_games = make(map[string](chan newPlayer))
}

/*
	Type defs for Game-specific code
	i.e. how do we store game data in memory
*/
type unitType int

const (
	empty = iota
	mountain
	town
	capital
)

type cellValue struct {
	owner    int
	troops   int
	cellType unitType
}

type gameState struct {
	cells         [][]cellValue
	players       []string
	ready_players []string
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
	waiting_games[gameid] <- newPlayer{username: username, socket: c}
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

func gameLoop(playerChan chan newPlayer, game *gameState, gameid string) {
	ready := make(map[string]bool)
	players := make([]newPlayer, 0)

	// websocket listeners can push to this when the player changes readiness
	playerConnect := make(chan string)

	// we'll need to define a new channel (or channels) for communicating
	// during the game itself.

	for countTrue(ready) < len(players) || (len(players) < 2 && !cmd_flags.debug) {
		select {
		case player := <-playerChan:
			players = append(players, player)
			ready[player.username] = false
			go websocketListener(&player, playerConnect, game)
		case msg := <-playerConnect:
			log.Printf("received message: %s", msg)
			pieces := strings.Split(msg, ":")
			if pieces[0] == "ready" {
				ready[pieces[1]] = true
			} else if pieces[0] == "unready" {
				ready[pieces[1]] = false
			} else if pieces[0] == "disconnect" {
				for i := 0; i < len(players); i++ {
					if players[i].username == pieces[1] {
						players = append(players[:i], players[i+1:]...)
						break
					}
				}
				delete(ready, pieces[1])
			}
		}
	}

	log.Printf("Starting game: %s", gameid)
	// playGame()
}

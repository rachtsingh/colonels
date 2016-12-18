package main

import (
	"log"
	"math/rand"
	"strings"
)

// used as a goroutine
func websocketListener(player *newPlayer, readyChan chan string, game *gameState, gameChan chan socketMsg) {
	defer player.socket.Close()
	open := true
	log.Printf("opened websocket connection to user: %s", player.username)

	player.socket.SetCloseHandler(func(code int, text string) error {
		log.Printf("read closing message from user: %s", player.username)
		pieces := []string{"disconnect", player.username}
		readyChan <- strings.Join(pieces, ":")
		open = false
		return nil
	})

	// start the writer (listening on game logic) as a new goroutine
	go func() {
		for {
			gameMsg, openc := <-gameChan
			if openc {
				switch gameMsg.msgType {
				case 0:
					// game update, we can just read from the gameState ptr
					err := player.socket.WriteJSON(game)
					if err != nil {
						log.Println("write:", err)
						break
					}
				}
			} else {
				log.Printf("closing websocket writer thread for player %s", player.username)
				break
			}
		}
	}()

	// start the reader (this goroutine)
	var response map[string]interface{}

	for open {
		err := player.socket.ReadJSON(&response)
		if err != nil {
			// ok if the websocket thread dies now
			log.Println("read:", err)
			break
		}
		msg_type := response["m"].(string)
		log.Printf("read message: %s from user: %s", msg_type, player.username)
		if msg_type == "ready" || msg_type == "unready" {
			pieces := []string{msg_type, player.username}
			readyChan <- strings.Join(pieces, ":")
		} else if msg_type == "move" {
			// do something else here
		}
	}

}

func initializeTerrain(game *gameState, players []newPlayer) {
	game.Cells = make([][]cellValue, 20)
	for i := 0; i < 20; i++ {
		game.Cells[i] = make([]cellValue, 20)
	}
	for _, player := range players {
		game.Players = append(game.Players, player.username)
	}
}

func randomizeTerrain(game *gameState) {
	for i := 0; i < len(game.Cells); i++ {
		for j := 0; j < len(game.Cells[i]); j++ {
			if rand.Float64() < 0.2 {
				game.Cells[i][j].CellType = Mountain
			} else {
				game.Cells[i][j].CellType = Empty
			}
		}
	}
}

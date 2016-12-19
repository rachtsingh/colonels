package main

import (
	"github.com/golang/protobuf/proto"
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
	for open {
		mtype, bytes, err := player.socket.ReadMessage()
		if err != nil {
			// ok if the websocket thread dies now
			log.Println("read:", err)
			break
		}
		if mtype != 2 {
			log.Printf("improper websocket client message - should be binary.")
		}
		msg := &PlayerStatus{}
		err = proto.Unmarshal(bytes, msg)
		if err != nil {
			log.Println("error unmarshaling: ", err)
		} else {
			pieces := []string{
				msg.GetStatus().String(),
				player.username,
			}
			readyChan <- strings.Join(pieces, ":")
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

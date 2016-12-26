package main

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"strings"
	"sync/atomic"
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
				msg := new(ServerToClient)
				msg.Which = gameMsg.msgType
				switch gameMsg.msgType {
				case ServerMessageType_FullBoard:
					board := new(FullBoard)
					for i := 0; i < len(game.Cells); i++ {
						board.Rows = append(board.Rows, new(FullBoardInnerRow))
						for j := 0; j < len(game.Cells[i]); j++ {
							value := cellValueToProto(game.Cells[i][j])
							board.Rows[i].Column = append(board.Rows[i].Column, &value)
						}
					}
					msg.Board = board
				case ServerMessageType_SingleCellUpdate:
					update := gameMsg.data.(SingleCellUpdate)
					msg.Update = &update
				}
				data, err := proto.Marshal(msg)
				if err != nil {
					log.Println("error marshaling:", err)
					continue
				}
				err = player.socket.WriteMessage(websocket.BinaryMessage, data)
				if err != nil {
					log.Println("write:", err)
					break
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
			log.Println("read:", err)
			// ok if the websocket thread dies now
			break
		}
		if mtype != 2 {
			log.Printf("improper websocket client message - should be binary.")
			continue
		}
		msg := new(ClientToServer)
		err = proto.Unmarshal(bytes, msg)
		if err != nil {
			log.Println("error unmarshaling: ", err)
			continue
		}
		switch msg.GetWhich() {
		case ClientMessageType_ClientStatus:
			pieces := []string{
				msg.GetStatus().String(),
				player.username,
			}
			readyChan <- strings.Join(pieces, ":")
		case ClientMessageType_PlayerMovement:
			player.moves <- *msg.GetMovement()
		case ClientMessageType_CancelQueue:
			atomic.StoreInt32(player.maxMoveId, msg.GetCancel().GetId())
		}
	}

}

func initializeTerrain(game *gameState, players []newPlayer, playerToID map[string]int32) {
	game.Cells = make([][]cellValue, 20)
	for i := 0; i < 20; i++ {
		game.Cells[i] = make([]cellValue, 20)
	}

	// give each player a capital
	var x int
	var y int
	for i := 0; i < len(players); i++ {
		x = rand.Intn(len(game.Cells))
		y = rand.Intn(len(game.Cells[x]))
		game.Cells[x][y].CellType = Capital
		game.Cells[x][y].Troops = 50
		game.Cells[x][y].Owner = int32(i + 1)
	}

	for _, player := range players {
		game.Players = append(game.Players, player.username)
	}
}

func randomizeTerrain(game *gameState) {
	for i := 0; i < len(game.Cells); i++ {
		for j := 0; j < len(game.Cells[i]); j++ {
			if game.Cells[i][j].CellType == Capital {
				continue
			}
			if rand.Float64() < 0.2 {
				game.Cells[i][j].CellType = Mountain
			} else {
				game.Cells[i][j].CellType = Empty
			}
		}
	}
}

func incrementallyChangeTerrain(game *gameState) (int32, int32, int32) {
	x := rand.Intn(len(game.Cells))
	y := rand.Intn(len(game.Cells[x]))
	if game.Cells[x][y].CellType == Mountain {
		game.Cells[x][y].CellType = Empty
		return int32(x), int32(y), Empty
	} else if game.Cells[x][y].CellType != Capital {
		game.Cells[x][y].CellType = Mountain
		return int32(x), int32(y), Mountain
	} else {
		return incrementallyChangeTerrain(game) // recurse!
	}
}

func readMoveNonblocking(moves chan PlayerMovement, maxMoveId int32) (PlayerMovement, error) {
	// let's do a nonblocking select until either we read something that is good, or we exhaust the channel
	for {
		select {
		case move := <-moves:
			log.Println("read move!", move)
			log.Println("maxMoveId:", maxMoveId)
			if move.GetId() > maxMoveId {
				return move, nil
			} else {
				continue
			}
		default:
			return PlayerMovement{}, errors.New("No valid moves for this player")
		}
	}
}

type IdPlayerMovement struct {
	move   *PlayerMovement
	player *newPlayer
}

func updateGameState(players []newPlayer, game *gameState) {
	// increment round
	game.Round += 1

	// increment troops
	for i := 0; i < len(game.Cells); i++ {
		for j := 0; j < len(game.Cells[i]); j++ {
			cell := &game.Cells[i][j]
			if cell.Owner != 0 {
				if game.Round%25 == 0 || cell.CellType == Town || cell.CellType == Capital {
					cell.Troops += 1
				}
			}
		}
	}
	validMoves := make([]IdPlayerMovement, 0)
	for _, player := range players {
		move, err := readMoveNonblocking(player.moves, atomic.LoadInt32(player.maxMoveId))
		if err == nil {
			log.Println("read move:", move)
			validMoves = append(validMoves, IdPlayerMovement{move: &move, player: &player})
		}
	}
	perm := rand.Perm(len(validMoves))
	for i := 0; i < len(validMoves); i++ {
		// execute the move
		move := validMoves[perm[i]].move
		oldCell := &game.Cells[move.GetOldx()][move.GetOldy()]
		newCell := &game.Cells[move.GetNewx()][move.GetNewy()]
		if oldCell.Owner > 0 && (game.Players[oldCell.Owner-1] == validMoves[perm[i]].player.username) {
			if newCell.CellType == Empty && oldCell.Troops > 1 {
				newCell.Owner = oldCell.Owner
				newCell.Troops = oldCell.Troops - 1
				oldCell.Troops = 1
			}
		}
	}
}

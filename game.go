package main

import (
	"log"
	"strings"
)

// used as a goroutine
func websocketListener(player *newPlayer, readyChan chan string) {
	defer player.socket.Close()

	// set the close handler
	player.socket.SetCloseHandler(func(code int, text string) error {
		pieces := []string{"disconnect", player.username}
		readyChan <- strings.Join(pieces, ":")
		return nil
	})

	var response map[string]interface{}

	for {
		err := player.socket.ReadJSON(&response)
		if err != nil {
			log.Println("read:", err)
			break
		}
		msg_type := response["m"].(string)
		if msg_type == "ready" || msg_type == "unready" {
			pieces := []string{msg_type, player.username}
			readyChan <- strings.Join(pieces, ":")
		} else if msg_type == "move" {
			// do something else here
		}
	}
}

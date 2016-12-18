package main

import (
	"log"
	"strings"
)

// used as a goroutine
func websocketListener(player *newPlayer, readyChan chan string) {
	defer player.socket.Close()
	log.Printf("Opened websocket connection to user: %s", player.username)

	// set the close handler
	player.socket.SetCloseHandler(func(code int, text string) error {
		log.Printf("read closing message from user: %s", player.username)
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
		log.Printf("Read message: %s from user: %s", msg_type, player.username)
		if msg_type == "ready" || msg_type == "unready" {
			pieces := []string{msg_type, player.username}
			readyChan <- strings.Join(pieces, ":")
		} else if msg_type == "move" {
			// do something else here
		}
	}
}

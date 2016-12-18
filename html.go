package main

import (
	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"strings"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SECRET_KEY")))
var upgrader = websocket.Upgrader{} // use default options

// hook from main.go
func setupUserMethods(r *mux.Router) {
	r.HandleFunc("/user/setusername", SetUsername).Methods("POST")
}

func setupGameMethods(r *mux.Router) {
	// initial page load
	r.HandleFunc("/game/{gameid}", GamePageHandler).Methods("GET")

	// start the websocket connection
	r.HandleFunc("/game/{gameid}/ws", GameSetupHandler)
}

var homePage = pongo2.Must(pongo2.FromFile("templates/index.html"))
var gamePage = pongo2.Must(pongo2.FromFile("templates/game.html"))

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "colonels")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// check whether we're logged in
	var logged_in = false
	var username = session.Values["username"]
	if username != nil {
		logged_in = true
	} else {
		username = "anonymous"
	}

	err = homePage.ExecuteWriter(pongo2.Context{"username": username, "logged_in": logged_in, "random_game": RandString(15)}, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func SetUsername(w http.ResponseWriter, r *http.Request) {
	log.Printf("got request to change username to: %s", r.FormValue("username"))

	if r.FormValue("username") == "" {
		http.Error(w, "Invalid username", http.StatusBadRequest)
		return
	}

	session, err := store.Get(r, "colonels")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// let's not worry about databases for now
	session.Values["username"] = r.FormValue("username")
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func GamePageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	session, err := store.Get(r, "colonels")
	gameid := vars["gameid"]

	// set an anonymous username if there isn't a username right now
	if session.Values["username"] == "" {
		s := []string{"Anonymous", RandString(2)}
		session.Values["username"] = strings.Join(s, "")
		session.Save(r, w)
	}

	log.Printf("%s tried to access game: %s", session.Values["username"], gameid)

	err = gamePage.ExecuteWriter(pongo2.Context{"matchid": gameid}, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func GameSetupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	session, err := store.Get(r, "colonels")
	gameid := vars["gameid"]
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("user: %s joined game: %s", session.Values["username"], gameid)

	// upgrade the connection to a websocket connection
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("user: %s failed to upgrade: %s", session.Values["username"], err)
		return
	}

	// has the game been seen before?
	if !gameExists(gameid) {
		log.Printf("creating new game: %s", gameid)
		setupNewGame(gameid)
	}

	// note! we can't use c again, since we've given ownership of it away
	// in particular we shouldn't defer c.Close(), the goroutine will take care of it
	addNewPlayer(session.Values["username"].(string), c, gameid)
}

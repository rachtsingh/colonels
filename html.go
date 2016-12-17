package main

import (
	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"os"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SECRET_KEY")))

// hook from main.go
func setupUserMethods(r *mux.Router) {
	r.HandleFunc("/user/setusername", SetUsername).Methods("POST")
}

var homePage = pongo2.Must(pongo2.FromFile("templates/index.html"))

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

	err = homePage.ExecuteWriter(pongo2.Context{"username": username, "logged_in": logged_in}, w)
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

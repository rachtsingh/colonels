package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type commandLineFlags struct {
	port int
}

func parseCmdLine() (*commandLineFlags, error) {
	portPtr := flag.Int("port", 8000, "port to serve on")
	flag.Parse()

	// because Unix!
	if *portPtr < 1024 {
		return nil, errors.New("Invalid port number: must be root to access ports below 1024")
	}

	return &commandLineFlags{port: *portPtr}, nil
}

func main() {
	cmd_flags, err := parseCmdLine()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting colonels server on port: %d", cmd_flags.port)

	r := mux.NewRouter()

	// main pages to serve up HTML
	r.HandleFunc("/", HomeHandler).Methods("GET")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// allow the user to set various settings
	setupUserMethods(r)
	setupGameMethods(r)
	initGlobals()

	// start up the server
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cmd_flags.port), r))
}

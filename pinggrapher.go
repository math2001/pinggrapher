package main

import (
	"flag"
	"log"
)

func startserver(port int) {
	log.Printf("Starting server on %d", port)
	log.Fatal("Not implemented")
}

func main() {
	var server bool
	var port int
	flag.BoolVar(&server, "server", false, "start the server")
	flag.IntVar(&port, "port", 9998, "port to run the server on")
	flag.Parse()
	if server {
		startserver(port)
	}
}

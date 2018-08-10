package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

func startserver(port int) {
	log.Printf("Starting server on %d", port)
	log.Fatal("Not implemented")
}

func read(delay int) {
	var cache []string
	ticker := time.NewTicker(time.Duration(delay) * 1000 * 1000 * time.Nanosecond)
	defer ticker.Stop()
	go func() {
		var line string
		var err error
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err = reader.ReadString('\n')
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Fatal(err)
			}
			cache = append(cache, strings.TrimSpace(line))
		}
	}()
	for {
		<-ticker.C
		fmt.Printf("Sending %#v\n", cache)
		cache = []string{}
	}
}

func main() {
	var server bool
	var port int
	var delay int
	flag.BoolVar(&server, "server", false, "start the server")
	flag.IntVar(&port, "port", 9998, "port to use")
	flag.IntVar(&delay, "delay", 10000, "mms to wait before sending the data")
	flag.Parse()
	if server {
		startserver(port)
	}
	read(delay)
}

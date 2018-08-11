package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

func read(delay int, pings chan int) {
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
			float, err := strconv.ParseFloat(strings.TrimSpace(line), 64)
			if err != nil {
				log.Printf("Couldn't convert '%s': %s", line, err)
			}
			pings <- int(math.Round(float))

		}
	}()
	// for {
	// 	<-ticker.C
	// 	for _, e := range cache {
	// 		float, err := strconv.ParseFloat(e, 64)
	// 		if err != nil {
	// 			log.Printf("Couldn't convert '%s': %s", e, err)
	// 		}
	// 		pings <- int(math.Round(float))
	// 	}
	// 	cache = []string{}
	// }
}

func main() {
	var port int
	var delay int
	flag.IntVar(&port, "port", 9998, "port to use")
	flag.IntVar(&delay, "delay", 10000, "mms to wait before sending the data")
	flag.Parse()
	var pings = make(chan int)
	go read(delay, pings)
	startserver(port, pings)
}

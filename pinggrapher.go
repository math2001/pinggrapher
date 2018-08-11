package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// use a map instead of slice cause otherwise it's a pain to delete them when
// they decide to leave
type Clients struct {
	m   map[int]Client
	mux sync.Mutex
}

func (c *Clients) Set(n int, e Client) int {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.m[n] = e
	return n
}

func (c *Clients) Get(n int) (Client, bool) {
	c.mux.Lock()
	defer c.mux.Unlock()
	e, ok := c.m[n]
	return e, ok
}

func (c *Clients) Length() int {
	c.mux.Lock()
	defer c.mux.Unlock()
	return len(c.m)
}

func (c *Clients) Delete(id int) {
	c.mux.Lock()
	defer c.mux.Unlock()
	delete(c.m, id)
}

func (c *Clients) ForEach(fn func(int, Client)) {
	c.mux.Lock()
	defer c.mux.Unlock()
	// lock for the whole duration of every for loop, just in case
	// not too sure if I could just lock during the .Get
	for id, client := range c.m {
		fn(id, client)
	}
}

var clients = Clients{m: make(map[int]Client)}
var clientsmutex sync.Mutex

type Client struct {
	Writer  *wsutil.Writer
	Encoder *json.Encoder
	Conn    net.Conn
}

type Stats struct {
	Timestamp int64   `json:"timestamp"`
	Average   float64 `json:"average"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
}

func NewStats(times []float64, timestamp int64) Stats {
	return Stats{
		Timestamp: timestamp,
		Max:       Max(times),
		Min:       Min(times),
		Average:   Average(times),
	}
}

func read(pings chan float64) {
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
			ping, err := strconv.ParseFloat(strings.TrimSpace(line), 64)
			if err != nil {
				log.Printf("Couldn't convert '%s': %s", line, err)
			}
			pings <- ping
		}
	}()
}

func sendpast(file *os.File, delay int) {
	defer file.Close()
	decoder := json.NewDecoder(file)
	var statsarr []Stats
	// read the file, compute the stats, and send
	// every line represents on Lap (which will be converted to one Stats)
	for {
		var s Stats
		if err := decoder.Decode(&s); err == io.EOF {
			break
		} else if err != nil {
			log.Printf("Couldn't decode: %s", err)
		}
		statsarr = append(statsarr, s)
	}
	if len(statsarr) != 0 {
		log.Printf("Sending past (%d elements)...", len(statsarr))
		send(statsarr)
	} else {
		log.Printf("No past to send.")
	}
}

func startserver(port int, delay int, path string, pings chan float64) {
	log.Printf("listening on :%d\n", port)
	var clientidcount = 0
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("New client #%d.", clientidcount)
		conn, _, _, err := ws.UpgradeHTTP(r, w, nil)
		if err != nil {
			log.Print(err)
			return
		}
		writer := wsutil.NewWriter(conn, ws.StateServerSide, ws.OpText)
		id := clients.Set(clientidcount, Client{
			Writer:  writer,
			Encoder: json.NewEncoder(writer),
			Conn:    conn,
		})
		file, err := os.Open(path)
		if err != nil {
			log.Fatalf("Couldn't open cache file '%s': %s", path, err)
		}
		clientidcount += 1
		go func() {
			defer conn.Close()
			sendpast(file, delay)
			// this will stop blocking as soon as the client does something
			// That is, send a message (which they shouldn't do) or close the
			// connection
			header, err := wsutil.NewReader(conn, ws.StateServerSide).NextFrame()
			if err != nil {
				log.Printf("Error while reading header: %s", err)
			}
			if header.OpCode == ws.OpClose {
				log.Printf("Client %d left", id)
				clients.Delete(id)
			}
		}()
	})
	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func send(statsarr []Stats) {
	clients.ForEach(func(id int, client Client) {
		if err := client.Encoder.Encode(statsarr); err != nil {
			log.Print("Couldn't encode/write:", err)
			clients.Delete(id)
		}
		if err := client.Writer.Flush(); err != nil {
			log.Print("Couldn't flush:", err)
			clients.Delete(id)
		}
	})
}

func write(delay int, path string, pings chan float64) {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		log.Fatal("Target file already exists.")
	}
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Couldn't create '%s': %s", path, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	// we write in json and not gob for reusability (and so we don't have to
	// re-write everything everytime)
	encoder := json.NewEncoder(w)

	ticker := time.NewTicker(time.Duration(delay) * time.Millisecond)
	defer ticker.Stop()

	var times []float64

	for {
		select {
		case ping := <-pings:
			times = append(times, ping)
		case <-ticker.C:
			var stats = NewStats(times, time.Now().Unix())
			log.Printf("Save lap to file and send to %d client(s)", clients.Length())
			if err := encoder.Encode(stats); err != nil {
				log.Fatalf("Couldn't write Lap to file: %s", err)
			}
			if err := w.Flush(); err != nil {
				log.Fatalf("Couldn't flush file: %s", err)
			}
			go send([]Stats{stats})
			times = []float64{}
		}
	}
}

func main() {
	var port, delay int
	var path string
	flag.IntVar(&port, "port", 9998, "port to use")
	flag.IntVar(&delay, "delay", 60*1000, "ms to wait before sending the data")
	flag.StringVar(&path, "path", "./.pings", "path to the filename to store information")
	flag.Parse()
	var pings = make(chan float64)
	go read(pings)
	go write(delay, path, pings)
	startserver(port, delay, path, pings)
}

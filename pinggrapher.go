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
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/sparrc/go-ping"
)

// Clients is a map with a mutex. People shouldn't touch .m themself, but use
// the API. Note that it's a map and not a slice, just to make easier to delete
// clients when they decide to leave
type Clients struct {
	m   map[int]Client
	mux sync.Mutex
}

// Set client for id
func (c *Clients) Set(n int, e Client) int {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.m[n] = e
	return n
}

// Get client from id
func (c *Clients) Get(n int) (Client, bool) {
	c.mux.Lock()
	defer c.mux.Unlock()
	e, ok := c.m[n]
	return e, ok
}

// Length returns the number of active clients
func (c *Clients) Length() int {
	c.mux.Lock()
	defer c.mux.Unlock()
	return len(c.m)
}

// Delete a client from its id
func (c *Clients) Delete(id int) {
	c.mux.Lock()
	defer c.mux.Unlock()
	delete(c.m, id)
}

// ForEach loops around each client. Note that the mutex for the entire
// duration of all the callbacks *combined*
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

// Client doesn't have a writer because we don't care what the browser says
type Client struct {
	Writer  *wsutil.Writer
	Encoder *json.Encoder
	Conn    net.Conn
}

// Stats are the intersting stuff. They are computed from a list of times
type Stats struct {
	Average float64 `json:"average"`
	// the percentage of pings that took longer than x ms
	Above     float64 `json:"above"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	Timestamp int64   `json:"timestamp"`
}

// NewStats compute the stats from times and returns a Stats object
func NewStats(times []float64, timestamp int64) Stats {
	return Stats{
		Timestamp: timestamp,
		Max:       Max(times),
		Min:       Min(times),
		Average:   Average(times),
		Above:     AboveBy(10, times),
	}
}

func read(pings chan float64, host string) {
	go func() {
		log.Printf("Pinging %v", host)
		pinger, err := ping.NewPinger(host)
		if err != nil {
			panic(err)
		}
		pinger.Count = 0
		pinger.Interval = time.Second * 1
		pinger.Timeout = time.Second * 5
		pinger.OnRecv = func(pkt *ping.Packet) {
			ping := float64(pkt.Rtt / time.Millisecond)
			pings <- ping
		}
		pinger.Run() // blocks until finished
	}()
}

func sendpast(file *os.File) {
	defer file.Close()
	decoder := json.NewDecoder(file)
	var timesarr [][]float64
	var statsarr []Stats
	// read the file, compute the stats, and send every line (that are a list)
	// represents one set of times (which will be converted to one Stats)
	for {
		var times []float64
		if err := decoder.Decode(&times); err == io.EOF {
			break
		} else if err != nil {
			log.Printf("Couldn't decode: %s", err)
		}
		timesarr = append(timesarr, times)
		// compute the stats. The first element is the timestamp
		statsarr = append(statsarr, NewStats(times[1:], int64(times[0])))
	}
	if len(timesarr) != 0 {
		log.Printf("Sending past (%d elements)...", len(statsarr))
		send(statsarr)
	} else {
		log.Printf("No past to send.")
	}
}

func startserver(port int, path string, pings chan float64) {
	log.Printf("listening on :%d\n", port)
	var clientidcount = 0
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("New client #%d.", clientidcount)
		conn, _, _, err := ws.UpgradeHTTP(r, w)
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
		clientidcount++
		go func() {
			defer conn.Close()
			sendpast(file)
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
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Couldn't create '%s': %s", path, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	// we write in json and not gob for reusability (and so we don't have to
	// re-write everything everytime)
	encoder := json.NewEncoder(w)

	ticker := time.NewTicker(time.Duration(delay) * time.Second)
	defer ticker.Stop()

	var times = []float64{float64(time.Now().Unix())}

	for {
		select {
		case ping := <-pings:
			times = append(times, ping)
		case <-ticker.C:
			log.Printf("Save times to file and send to %d client(s)", clients.Length())
			// we save the raw times so that if we want add some metric (like
			// average, etc), it's easy to do
			if err := encoder.Encode(times); err != nil {
				log.Fatalf("Couldn't write times to file: %s", err)
			}
			if err := w.Flush(); err != nil {
				log.Fatalf("Couldn't flush file: %s", err)
			}
			go send([]Stats{NewStats(times[1:], int64(times[0]))})
			times = []float64{float64(time.Now().Unix())}
		}
	}
}

func main() {
	var port, delay int
	var path string
	flag.IntVar(&port, "port", 9998, "port to use")
	flag.IntVar(&delay, "delay", 60, "seconds to wait before sending the data")
	flag.StringVar(&path, "path", "./.pings", "path to the filename to store information")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("Usage: pinggrapher <hostname/IP>")
	}
	var pings = make(chan float64)
	go read(pings, flag.Args()[0])
	go write(delay, path, pings)
	startserver(port, path, pings)
}

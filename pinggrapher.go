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

func (c *Clients) Set(n int, e Client) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.m[n] = e
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

type Lap []float64

type Stats struct {
	Average float64 `json:"average"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
}

func NewStats(lap Lap) Stats {
	return Stats{
		Max:     Max(lap),
		Min:     Min(lap),
		Average: Average(lap),
	}
}

func (l Lap) WriteTo(w io.Writer) (int64, error) {
	var written int64 = 0
	for _, f := range l {
		b, err := fmt.Fprintf(w, "%f", f)
		written += int64(b)
		if err != nil {
			return written, err
		}
	}
	b, err := fmt.Fprint(w, "\n")
	written += int64(b)
	if err != nil {
		return written, err
	}
	return written, nil
}

type Storer struct {
	Start time.Time
	Laps  []Lap
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

func startserver(port int, pings chan float64) {
	log.Printf("listening on :%d\n", port)
	var clientidcount = 0
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("New client.")
		conn, _, _, err := ws.UpgradeHTTP(r, w, nil)
		if err != nil {
			log.Print(err)
			return
		}
		clientidcount += 1
		writer := wsutil.NewWriter(conn, ws.StateServerSide, ws.OpText)
		clients.Set(clientidcount, Client{
			Writer:  writer,
			Encoder: json.NewEncoder(writer),
			Conn:    conn,
		})
	})
	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func send(stats Stats) {
	clients.ForEach(func(id int, client Client) {
		if err := client.Encoder.Encode(stats); err != nil {
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
		log.Fatal("Target file already exists. Not implemented.")
	}
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Couldn't create '%s': %s", path, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	storer := Storer{Start: time.Now()}

	fmt.Fprintf(w, "%d # %s\n", storer.Start.Unix(), storer.Start.Format(time.UnixDate))

	ticker := time.NewTicker(time.Duration(delay) * time.Millisecond)
	defer ticker.Stop()

	var lap Lap

	for {
		select {
		case ping := <-pings:
			lap = append(lap, ping)
		case <-ticker.C:
			log.Printf("Save lap to file and send to %d client(s)", clients.Length())
			if _, err := lap.WriteTo(w); err != nil {
				log.Fatalf("Couldn't write Lap to file: %s", err)
			}
			if err := w.Flush(); err != nil {
				log.Fatalf("Couldn't flush file: %s", err)
			}
			go send(NewStats(lap))
			lap = Lap{}
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
	startserver(port, pings)
}

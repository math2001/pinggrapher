package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func startserver(port int, pings chan int) {
	fmt.Printf("listening on :%d\n", port)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w, nil)
		if err != nil {
			log.Print(err)
			return
		}
		go func() {
			defer conn.Close()
			var (
				w       = wsutil.NewWriter(conn, ws.StateServerSide, ws.OpText)
				encoder = json.NewEncoder(w)
			)
			for {
				ping := <-pings
				encoder.Encode(ping)
				if err := w.Flush(); err != nil {
					log.Print(err)
					return
				}
			}
		}()
	})
	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

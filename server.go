package main

import (
	"fmt"
	"net/http"
)

func startserver(port int) {
	fmt.Printf("listening on :%d\n", port)
	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

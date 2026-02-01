package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func main() {
	var port int
	var sleep int
	flag.IntVar(&port, "port", 8081, "port to listen on")
	flag.IntVar(&sleep, "sleep", 0, "milliseconds to sleep before responding")
	flag.Parse()

	handler := func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if s := q.Get("sleep"); s != "" {
			if ms, err := strconv.Atoi(s); err == nil && ms > 0 {
				time.Sleep(time.Duration(ms) * time.Millisecond)
			}
		} else if sleep > 0 {
			time.Sleep(time.Duration(sleep) * time.Millisecond)
		}
		msg := fmt.Sprintf("Привет, я сервер на порту %d\n", port)
		w.Write([]byte(msg))
	}

	http.HandleFunc("/", handler)
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Test backend listening on %s (sleep=%dms)", addr, sleep)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("backend server failed: %v", err)
	}
}

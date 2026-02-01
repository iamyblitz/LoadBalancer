package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	lb "loadbalancer/pkg/loadbalancer"
)

func main() {
	var port int
	var backends string
	var mode string

	flag.IntVar(&port, "port", 8000, "port for load balancer")
	flag.StringVar(&backends, "backends", "http://localhost:8081,http://localhost:8082,http://localhost:8083", "comma-separated backend URLs")
	flag.StringVar(&mode, "mode", "rr", "load balancing mode: rr or least")
	flag.Parse()

	pool := lb.NewServerPool()

	urls := strings.Split(backends, ",")
	for _, u := range urls {
		parsed, err := url.Parse(strings.TrimSpace(u))
		if err != nil {
			log.Fatalf("invalid backend URL %s: %v", u, err)
		}
		b := lb.NewBackend(parsed)
		// set ErrorHandler so that failed requests mark backend dead
		b.ReverseProxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
			b.SetAlive(false)
			http.Error(rw, "Bad Gateway", http.StatusBadGateway)
		}
		pool.AddBackend(b)
	}

	// start health checking every 20s with 2s timeout
	pool.RunHealthCheck(20*time.Second, 2*time.Second)

	handler := lb.NewLoadBalancer(pool, mode)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Load balancer started on %s, mode=%s", addr, mode)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server failed: %v", err)
		os.Exit(1)
	}
}

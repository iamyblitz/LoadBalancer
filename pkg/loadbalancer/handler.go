package loadbalancer

import (
	"context"
	"net/http"
	"time"
)

// LoadBalancer is an HTTP handler that proxies requests to a ServerPool.
// Mode can be "rr" (round-robin) or "least" (least-connections).
type LoadBalancer struct {
	Pool *ServerPool
	Mode string
	// RequestTimeout controls per-request timeout; zero means no timeout.
	RequestTimeout time.Duration
	// MaxRetries controls how many times to retry another backend on error.
	MaxRetries int
}

// NewLoadBalancer constructs a LoadBalancer with sensible defaults.
func NewLoadBalancer(pool *ServerPool, mode string) *LoadBalancer {
	return &LoadBalancer{
		Pool:           pool,
		Mode:           mode,
		RequestTimeout: 10 * time.Second,
		MaxRetries:     3,
	}
}

type captureResponseWriter struct {
	rw         http.ResponseWriter
	statusCode int
}

func (c *captureResponseWriter) Header() http.Header { return c.rw.Header() }
func (c *captureResponseWriter) Write(b []byte) (int, error) {
	if c.statusCode == 0 {
		c.statusCode = http.StatusOK
	}
	return c.rw.Write(b)
}
func (c *captureResponseWriter) WriteHeader(code int) {
	c.statusCode = code
	c.rw.WriteHeader(code)
}

// ServeHTTP implements http.Handler. It selects a backend according to Mode
// and proxies the request, retrying on backend errors up to MaxRetries.
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for i := 0; i < lb.MaxRetries; i++ {
		var peer *Backend
		if lb.Mode == "least" {
			peer = lb.Pool.GetLeastConnectedPeer()
		} else {
			peer = lb.Pool.GetNextPeer()
		}

		if peer == nil {
			http.Error(w, "Service not available", http.StatusServiceUnavailable)
			return
		}

		// per-request timeout
		ctx := r.Context()
		var cancel context.CancelFunc
		if lb.RequestTimeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, lb.RequestTimeout)
		}
		req := r.Clone(ctx)

		crw := &captureResponseWriter{rw: w}

		peer.IncConn()
		peer.ReverseProxy.ServeHTTP(crw, req)
		peer.DecConn()
		if cancel != nil {
			cancel()
		}

		// retry on server errors or if no status was written
		if crw.statusCode >= 500 || crw.statusCode == 0 {
			peer.SetAlive(false)
			continue
		}
		return
	}
	http.Error(w, "Service not available after retries", http.StatusServiceUnavailable)
}

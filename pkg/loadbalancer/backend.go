package loadbalancer

import (
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

// Backend represents a single upstream server.
// It holds the target URL, a reverse proxy and runtime state
// such as Alive and ActiveConnections.
type Backend struct {
	URL               *url.URL
	Alive             bool
	mux               sync.RWMutex
	ReverseProxy      *httputil.ReverseProxy
	ActiveConnections int32
}

// NewBackend creates a Backend with a reverse proxy for the given URL.
func NewBackend(u *url.URL) *Backend {
	proxy := httputil.NewSingleHostReverseProxy(u)
	return &Backend{
		URL:          u,
		Alive:        true,
		ReverseProxy: proxy,
	}
}

// SetAlive sets the alive state of the backend (thread-safe).
func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

// IsAlive returns whether the backend is marked alive.
func (b *Backend) IsAlive() (alive bool) {
	b.mux.RLock()
	alive = b.Alive
	b.mux.RUnlock()
	return
}

// IncConn increments the active connection counter.
func (b *Backend) IncConn() {
	atomic.AddInt32(&b.ActiveConnections, 1)
}

// DecConn decrements the active connection counter.
func (b *Backend) DecConn() {
	atomic.AddInt32(&b.ActiveConnections, -1)
}

// GetActiveConn returns the current active connections count.
func (b *Backend) GetActiveConn() int32 {
	return atomic.LoadInt32(&b.ActiveConnections)
}

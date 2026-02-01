package loadbalancer

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// ServerPool manages a set of backends and selection algorithms.
type ServerPool struct {
	backends []*Backend
	current  uint64
	mux      sync.RWMutex
}

// NewServerPool creates an empty ServerPool.
func NewServerPool() *ServerPool {
	return &ServerPool{}
}

// AddBackend appends a backend to the pool.
func (s *ServerPool) AddBackend(b *Backend) {
	s.mux.Lock()
	s.backends = append(s.backends, b)
	s.mux.Unlock()
}

// NextIndex returns the next index for round-robin selection.
func (s *ServerPool) NextIndex() int {
	if len(s.backends) == 0 {
		return -1
	}
	return int(atomic.AddUint64(&s.current, 1) % uint64(len(s.backends)))
}

// GetNextPeer returns the next alive backend using round-robin.
func (s *ServerPool) GetNextPeer() *Backend {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.backends) == 0 {
		return nil
	}
	next := s.NextIndex()
	l := len(s.backends) + next
	for i := next; i < l; i++ {
		idx := i % len(s.backends)
		if s.backends[idx].IsAlive() {
			return s.backends[idx]
		}
	}
	return nil
}

// GetLeastConnectedPeer returns the alive backend with minimal active connections.
func (s *ServerPool) GetLeastConnectedPeer() *Backend {
	s.mux.RLock()
	defer s.mux.RUnlock()
	var selected *Backend
	for _, b := range s.backends {
		if !b.IsAlive() {
			continue
		}
		if selected == nil || b.GetActiveConn() < selected.GetActiveConn() {
			selected = b
		}
	}
	return selected
}

// HealthCheck attempts a TCP dial to each backend and marks Alive accordingly.
func (s *ServerPool) HealthCheck(timeout time.Duration) {
	s.mux.RLock()
	backends := make([]*Backend, len(s.backends))
	copy(backends, s.backends)
	s.mux.RUnlock()

	for _, b := range backends {
		host := b.URL.Host
		conn, err := net.DialTimeout("tcp", host, timeout)
		if err != nil {
			b.SetAlive(false)
			continue
		}
		_ = conn.Close()
		b.SetAlive(true)
	}
}

// RunHealthCheck starts a background ticker that runs HealthCheck at interval.
func (s *ServerPool) RunHealthCheck(interval, timeout time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			s.HealthCheck(timeout)
			<-ticker.C
		}
	}()
}

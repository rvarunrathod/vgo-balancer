package algo

import (
	"net/http"
	"sync"
	bc "vgo-balancer/pkg/backend"
)

type RoundRobin struct {
	mu    sync.Mutex
	index int
}

func (r *RoundRobin) NextBackend(pool []*bc.Backend, w http.ResponseWriter, req *http.Request) *bc.Backend {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(pool) == 0 {
		return nil
	}

	r.index = (r.index + 1) % len(pool)

	for i := 0; i < len(pool); i++ {
		r.index = (r.index + i) % len(pool)
		if pool[r.index].IsAlive.Load() {
			return pool[r.index]
		}
	}

	return nil
}

func (r *RoundRobin) Name() string {
	return "round-robin"
}

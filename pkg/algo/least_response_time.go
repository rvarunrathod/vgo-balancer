package algo

import (
	"net/http"
	"sync"
	bc "vgo-balancer/pkg/backend"
)

type LeastResponseTime struct {
	mu sync.Mutex
}

func (lrt *LeastResponseTime) NextBackend(pool []*bc.Backend, w http.ResponseWriter, r *http.Request) *bc.Backend {
	lrt.mu.Lock()
	defer lrt.mu.Unlock()

	var selectedBackend *bc.Backend
	for _, backend := range pool {
		if backend.IsAlive.Load() {
			if selectedBackend == nil || backend.ResponseTime < selectedBackend.ResponseTime {
				selectedBackend = backend
			}
		}
	}
	return selectedBackend
}

func (lrt *LeastResponseTime) Name() string {
	return "least-response-time"
}

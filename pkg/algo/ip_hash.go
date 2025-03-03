package algo

import (
	"hash/fnv"
	"net/http"
	"strings"
	bc "vgo-balancer/pkg/backend"
)

type IPHash struct{}

func (i *IPHash) NextBackend(pool []*bc.Backend, w http.ResponseWriter, r *http.Request) *bc.Backend {
	if len(pool) == 0 {
		return nil
	}

	availablePool := make([]*bc.Backend, 0)
	for _, p := range pool {
		if p.IsAlive.Load() {
			availablePool = append(availablePool, p)
		}
	}

	if len(availablePool) == 0 {
		return nil
	}

	clientIP := getClientIP(r)
	hash := fnv.New32a()
	hash.Write([]byte(clientIP))
	index := hash.Sum32() % uint32(len(availablePool))

	return availablePool[index]
}

func (i *IPHash) Name() string {
	return "ip-hash"
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}
	return ip
}

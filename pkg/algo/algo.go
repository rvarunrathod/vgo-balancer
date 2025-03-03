package algo

import (
	"net/http"
	bc "vgo-balancer/pkg/backend"
)

type Algorithm interface {
	NextBackend(pool []*bc.Backend, w http.ResponseWriter, r *http.Request) *bc.Backend
	Name() string
}

func CreateAlgorithm(name string, pool []*bc.Backend) Algorithm {
	switch name {
	case "round-robin":
		return &RoundRobin{}
	case "weighted-round-robin":
		return NewWeightedRoundRobin(pool)
	case "ip-hash":
		return &IPHash{}
	case "least-response-time":
		return &LeastResponseTime{}
	default:
		return &RoundRobin{} // default algorithm
	}
}

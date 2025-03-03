package algo

import (
	"net/http"
	"sync"
	bc "vgo-balancer/pkg/backend"
)

type WeightedRoundRobin struct {
	mx            sync.Mutex
	currentIndex  int
	maxWeight     int
	gcdWeight     int
	backendCount  int
	currentWeight int
}

func NewWeightedRoundRobin(pool []*bc.Backend) *WeightedRoundRobin {
	maxWeight := getMaxWeight(pool)
	gcdWeight := getGCDWeight(pool)

	return &WeightedRoundRobin{
		mx:           sync.Mutex{},
		currentIndex: -1,
		maxWeight:    maxWeight,
		gcdWeight:    gcdWeight,
		backendCount: len(pool),
	}
}

func (wrr *WeightedRoundRobin) NextBackend(pool []*bc.Backend, w http.ResponseWriter, r *http.Request) *bc.Backend {
	wrr.mx.Lock()
	defer wrr.mx.Unlock()

	if len(pool) == 0 {
		return nil
	}

	for i := 0; i < wrr.backendCount; i++ {
		wrr.currentIndex = (wrr.currentIndex + 1) % wrr.backendCount

		if wrr.currentIndex == 0 {
			wrr.currentWeight -= wrr.gcdWeight
			if wrr.currentWeight <= 0 {
				wrr.currentWeight = wrr.maxWeight
			}
		}

		if pool[wrr.currentIndex].Weight >= wrr.currentWeight && pool[wrr.currentIndex].IsAlive.Load() {
			return pool[wrr.currentIndex]
		}
	}

	return nil
}

func (w *WeightedRoundRobin) Name() string {
	return "weighted-round-robin"
}

func getMaxWeight(pool []*bc.Backend) int {
	max := 0
	for _, p := range pool {
		if p.Weight > max {
			max = p.Weight
		}
	}
	return max
}

func getGCDWeight(pool []*bc.Backend) int {
	if len(pool) == 0 {
		return 1
	}

	gcd := pool[0].Weight
	for _, p := range pool[1:] {
		gcd = getGCD(gcd, p.Weight)
	}
	return gcd
}

func getGCD(a, b int) int {
	for b != 0 {
		a, b = b, b%a
	}
	return a
}

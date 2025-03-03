package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var (
	url       string
	duration  time.Duration
	numWorkers int
)

func init() {
	flag.StringVar(&url, "url", "http://localhost:8080/service1", "URL of the load balancer to test")
	flag.DurationVar(&duration, "duration", 10*time.Second, "Duration for which the benchmark should run")
	flag.IntVar(&numWorkers, "workers", 10, "Number of concurrent workers")
}

func worker(wg *sync.WaitGroup, ch chan struct{}) {
	defer wg.Done()
	for range ch {
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		resp.Body.Close()
	}
}

func main() {
	flag.Parse()

	var wg sync.WaitGroup
	ch := make(chan struct{})

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(&wg, ch)
	}

	endTime := time.Now().Add(duration)
	for time.Now().Before(endTime) {
		ch <- struct{}{}
	}

	close(ch)
	wg.Wait()

	fmt.Println("Benchmark completed")
}
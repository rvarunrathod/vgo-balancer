package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var (
	url           string
	duration      time.Duration
	concurrency   int
	totalRequests int
)

func init() {
	flag.StringVar(&url, "url", "http://localhost:8080/service1", "URL to benchmark")
	flag.IntVar(&concurrency, "c", 10, "Number of concurrent requests")
	flag.IntVar(&totalRequests, "n", 1000, "Total number of requests")
	flag.DurationVar(&duration, "d", 10*time.Second, "Duration of the test")
}

func main() {
	flag.Parse()

	results := make(chan time.Duration, totalRequests)
	errors := make(chan error, totalRequests)
	var wg sync.WaitGroup

	start := time.Now()
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	if duration > 0 {
		timer := time.NewTimer(duration)
		go func() {
			<-timer.C
			fmt.Println("Duration reached, stopping...")
			totalRequests = 0
		}()
	}

	// Start workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker(client, results, errors, &wg)
	}

	// Wait for completion
	wg.Wait()
	close(results)
	close(errors)

	// Process results
	processResults(results, errors, start)
}

func worker(client *http.Client, results chan<- time.Duration, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < totalRequests/concurrency; i++ {
		requestStart := time.Now()
		resp, err := client.Get(url)
		if err != nil {
			errors <- err
			continue
		}
		resp.Body.Close()
		results <- time.Since(requestStart)
	}
}

func processResults(results <-chan time.Duration, errors <-chan error, start time.Time) {
	var total time.Duration
	var count int
	var min, max time.Duration
	errCount := 0

	for d := range results {
		if min == 0 || d < min {
			min = d
		}
		if d > max {
			max = d
		}
		total += d
		count++
	}

	for range errors {
		errCount++
	}

	// Print results
	fmt.Printf("\nBenchmark Results:\n")
	fmt.Printf("URL: %s\n", url)
	fmt.Printf("Concurrency Level: %d\n", concurrency)
	fmt.Printf("Time taken: %v\n", time.Since(start))
	fmt.Printf("Complete requests: %d\n", count)
	fmt.Printf("Failed requests: %d\n", errCount)
	fmt.Printf("Requests per second: %.2f\n", float64(count)/time.Since(start).Seconds())
	fmt.Printf("Mean latency: %v\n", total/time.Duration(count))
	fmt.Printf("Min latency: %v\n", min)
	fmt.Printf("Max latency: %v\n", max)
}
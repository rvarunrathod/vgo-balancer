package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

var requestCount int

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		log.Printf("Method: %s, URL: %s, RemoteAddr: %s, Headers: %#v", r.Method, r.URL, r.RemoteAddr, r.Header)
		
	})

	log.Printf("Server starting on port %s", port)
	go http.ListenAndServe(":"+port, nil)
	for {
		select {
		case <-c:
			log.Printf("Server shutting down. Total requests: %d", requestCount)
			return
		case <- time.After(20*time.Second):
			log.Printf("Server shutting down. Total requests: %d", requestCount)
			return	
		}
	}
}
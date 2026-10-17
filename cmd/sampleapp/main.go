package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ambi/e-go-e/internal/sampleapp"
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "check whether the sample application is ready")
	flag.Parse()
	addr := os.Getenv("SAMPLE_ADDR")
	if addr == "" {
		addr = "127.0.0.1:3000"
	}
	if *healthcheck {
		// The container health check always uses the fixed container port.
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://127.0.0.1:3000/health", http.NoBody)
		if err != nil {
			log.Fatal(err)
		}
		client := &http.Client{Timeout: time.Second}
		response, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		if err := response.Body.Close(); err != nil {
			log.Fatal(err)
		}
		if response.StatusCode != http.StatusOK {
			log.Fatalf("health check returned HTTP %d", response.StatusCode)
		}
		return
	}
	server := &http.Server{
		Addr: addr, Handler: sampleapp.New(), ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("sample application listening on %s", addr)
	log.Fatal(server.ListenAndServe())
}

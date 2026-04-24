// cmd/healthcheck/main.go
//
// Tiny health-check binary used by the Docker HEALTHCHECK directive.
// It performs a single HTTP GET to http://localhost:PORT/health and exits 0
// on a 2xx response or 1 on any error.  The port defaults to 8080 and can be
// overridden with the SERVER_ADDR environment variable (":8080" format) or the
// -addr flag.
//
// Usage (Dockerfile HEALTHCHECK):
//
//	HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
//	  CMD ["/app/healthcheck"]
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	addr := flag.String("addr", envOrDefault("SERVER_ADDR", ":8080"), "server address, e.g. :8080")
	flag.Parse()

	// Strip leading colon if the address is just a port.
	host := *addr
	if strings.HasPrefix(host, ":") {
		host = "localhost" + host
	}

	url := fmt.Sprintf("http://%s/health", host)

	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "healthcheck: unexpected status %d\n", resp.StatusCode)
		os.Exit(1)
	}

	os.Exit(0)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

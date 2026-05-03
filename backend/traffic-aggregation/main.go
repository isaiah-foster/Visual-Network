package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"visual-network/traffic-aggregation/aggregator"
	"visual-network/traffic-aggregation/ebpf"
)

const bearerPrefix = "Bearer "

// corsMiddleware sets CORS headers. ALLOWED_ORIGINS env var controls allowed origins
// (comma-separated). Defaults to "*" when unset so local dev works without config.
func corsMiddleware(next http.Handler) http.Handler {
	allowed := os.Getenv("ALLOWED_ORIGINS")
	if allowed == "" {
		allowed = "*"
	}
	origins := strings.Split(allowed, ",")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		var allow string
		if allowed == "*" {
			allow = "*"
		} else {
			for _, o := range origins {
				if strings.TrimSpace(o) == origin {
					allow = origin
					break
				}
			}
		}
		// Only set header when the origin is permitted; unlisted origins
		// receive no ACAO header so the browser blocks the request.
		if allow != "" {
			w.Header().Set("Access-Control-Allow-Origin", allow)
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		}

		if r.Method == http.MethodOptions {
			if allow == "" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// apiKeyMiddleware enforces bearer-token auth when COLLECTOR_API_KEY is set.
// When the env var is empty the middleware is a no-op so local dev needs no config.
func apiKeyMiddleware(apiKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiKey == "" {
			next.ServeHTTP(w, r)
			return
		}
		auth := r.Header.Get("Authorization")
		if auth != bearerPrefix+apiKey {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <interface>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s eth0\n", os.Args[0])
		os.Exit(1)
	}

	iface := os.Args[1]

	log.Printf("Starting Visual Network Monitor on interface: %s", iface)
	log.Println("Press Ctrl+C to stop")

	// Create eBPF monitor
	monitor, err := ebpf.NewMonitor(iface)
	if err != nil {
		log.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Close()

	// Create aggregator
	agg := aggregator.NewAggregator(monitor)
	defer agg.Close()

	apiAddr := os.Getenv("BACKEND_API_ADDR")
	if apiAddr == "" {
		apiAddr = ":9090"
	}

	apiKey := os.Getenv("COLLECTOR_API_KEY")
	if apiKey != "" {
		log.Println("API key authentication enabled")
	}

	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	apiMux.HandleFunc("/topology", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(agg.GetCurrentTopology())
	})
	apiMux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(agg.GetCurrentMetrics())
	})

	apiServer := &http.Server{
		Addr:         apiAddr,
		Handler:      corsMiddleware(apiKeyMiddleware(apiKey, apiMux)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Backend API listening on %s", apiAddr)
		if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Backend API server failed: %v", err)
		}
	}()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start update printer
	go func() {
		for {
			select {
			case update := <-agg.UpdateChannel():
				if len(update.NewFlows) > 0 {
					log.Printf("New flows: %d", len(update.NewFlows))
					for _, flow := range update.NewFlows {
						log.Printf("  %s:%d -> %s:%d [%s/%s] %s",
							flow.SrcName, flow.Key.SrcPort,
							flow.DstName, flow.Key.DstPort,
							flow.ProtocolName, flow.L7ProtoName,
							flow.Key,
						)
					}
				}

				if len(update.Updated) > 0 {
					log.Printf("Updated flows: %d", len(update.Updated))
				}

				if len(update.Removed) > 0 {
					log.Printf("Removed flows: %d", len(update.Removed))
				}

				log.Printf("Summary: %d flows, %d packets, %d bytes, %d unique IPs, avg latency: %.2fms",
					update.Summary.TotalFlows,
					update.Summary.TotalPackets,
					update.Summary.TotalBytes,
					update.Summary.UniqueIPs,
					update.Summary.AvgLatencyMs,
				)

			case metrics := <-agg.MetricsChannel():
				if len(metrics.TopLatencySrc) > 0 {
					log.Println("Top 5 latency sources:")
					for i, info := range metrics.TopLatencySrc {
						log.Printf("  %d. %s -> %s: %.2fms [%s]",
							i+1, info.SrcName, info.DstName,
							info.AvgLatencyMs, info.Protocol,
						)
					}
				}
			}
		}
	}()

	// Periodic stats dump
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			flowCount, hsCount, err := monitor.GetMapStats()
			if err != nil {
				log.Printf("Error getting map stats: %v", err)
			} else {
				log.Printf("Map stats: %d flows, %d handshakes", flowCount, hsCount)
			}
			
			agg.DumpTopology()
		}
	}()

	// Wait for signal
	<-sigChan
	log.Println("\nShutting down...")
	ticker.Stop()
	if err := apiServer.Close(); err != nil {
		log.Printf("Backend API close error: %v", err)
	}

	log.Println("Done!")
}

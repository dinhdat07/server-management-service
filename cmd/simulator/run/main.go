package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"server-management-service/internal/app/monitoring"
)

type simConfig struct {
	serverCount   int
	rounds        int
	togglePercent int
	simulatorURL  string
}

type roundMetrics struct {
	round       int
	duration    time.Duration
	serversUp   int
	serversDown int
	memAllocMB  float64
	goroutines  int
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("simulation failed: %v", err)
	}
}

func run() error {
	cfg := simConfig{
		serverCount:   envInt("SIMULATOR_IP_COUNT", 10000),
		rounds:        envInt("SIMULATION_ROUNDS", 10),
		togglePercent: envInt("SIMULATION_TOGGLE_PCT", 5),
		simulatorURL:  envStr("SIMULATOR_URL", "http://localhost:8080"),
	}

	log.Printf("=== Simulation Config ===")
	log.Printf("Servers: %d, Rounds: %d, Toggle: %d%%", cfg.serverCount, cfg.rounds, cfg.togglePercent)

	// Warmup: reset all IPs to UP
	log.Println("=== Warmup: resetting all servers UP ===")
	if err := resetAll(cfg.simulatorURL); err != nil {
		return fmt.Errorf("warmup reset: %w", err)
	}

	allIPs := generateIPs(cfg.serverCount)

	// Create monitoring worker app
	app, err := monitoring.NewApp()
	if err != nil {
		return fmt.Errorf("init monitoring app: %w", err)
	}

	// Run rounds
	var metrics []roundMetrics
	for round := 1; round <= cfg.rounds; round++ {
		log.Printf("=== Round %d/%d ===", round, cfg.rounds)

		toggleCount := cfg.serverCount * cfg.togglePercent / 100
		ipsToDown := pickRandom(allIPs, toggleCount)
		if err := toggleIPs(cfg.simulatorURL, "down", ipsToDown); err != nil {
			log.Printf("WARN: toggle down round %d: %v", round, err)
		}

		ipsToUp := pickRandom(allIPs, toggleCount/2)
		if err := toggleIPs(cfg.simulatorURL, "up", ipsToUp); err != nil {
			log.Printf("WARN: toggle up round %d: %v", round, err)
		}

		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		start := time.Now()
		if err := app.Pool.Run(context.Background()); err != nil {
			log.Printf("WARN: worker run round %d: %v", round, err)
		}
		elapsed := time.Since(start)

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		status := getStatus(cfg.simulatorURL)
		rm := roundMetrics{
			round:       round,
			duration:    elapsed,
			serversUp:   status.Total - status.Down,
			serversDown: status.Down,
			memAllocMB:  float64(m2.Alloc-m1.Alloc) / 1024 / 1024,
			goroutines:  runtime.NumGoroutine(),
		}
		metrics = append(metrics, rm)

		log.Printf("Round %d: %v, up=%d, down=%d, mem=%.1fMB, goroutines=%d",
			round, elapsed.Round(time.Millisecond), rm.serversUp, rm.serversDown, rm.memAllocMB, rm.goroutines)
	}

	log.Println("=== Summary ===")
	var totalDur time.Duration
	for _, m := range metrics {
		totalDur += m.duration
	}
	avg := totalDur / time.Duration(len(metrics))
	log.Printf("Rounds: %d", len(metrics))
	log.Printf("Avg round duration: %v", avg.Round(time.Millisecond))
	log.Printf("Servers per second: %.0f", float64(cfg.serverCount)/avg.Seconds())
	log.Printf("Peak mem delta: %.1f MB", maxMemDelta(metrics))

	return nil
}

func generateIPs(count int) []string {
	ips := make([]string, 0, count)
	octet3 := 0
	octet4 := 1
	for i := 0; i < count; i++ {
		ips = append(ips, fmt.Sprintf("10.1.%d.%d", octet3, octet4))
		octet4++
		if octet4 > 254 {
			octet4 = 1
			octet3++
		}
	}
	return ips
}

func pickRandom(ips []string, n int) []string {
	if n > len(ips) {
		n = len(ips)
	}
	perm := rand.Perm(len(ips))
	result := make([]string, n)
	for i := 0; i < n; i++ {
		result[i] = ips[perm[i]]
	}
	return result
}

func resetAll(simulatorURL string) error {
	resp, err := http.Post(simulatorURL+"/reset", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("reset returned %d", resp.StatusCode)
	}
	return nil
}

func toggleIPs(simulatorURL, action string, ips []string) error {
	if len(ips) == 0 {
		return nil
	}
	// Batch 200 IPs per request to avoid huge request bodies
	batchSize := 200
	for i := 0; i < len(ips); i += batchSize {
		end := i + batchSize
		if end > len(ips) {
			end = len(ips)
		}
		body := map[string][]string{"ips": ips[i:end]}
		jsonBody, _ := json.Marshal(body)
		resp, err := http.Post(simulatorURL+"/"+action, "application/json", bytes.NewReader(jsonBody))
		if err != nil {
			return err
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("simulator returned %d", resp.StatusCode)
		}
	}
	return nil
}

func getStatus(simulatorURL string) struct{ Total, Down int } {
	resp, err := http.Get(simulatorURL + "/status")
	if err != nil {
		return struct{ Total, Down int }{}
	}
	defer resp.Body.Close()
	var s struct{ Total, Down int }
	json.NewDecoder(resp.Body).Decode(&s)
	return s
}

func envInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		n, _ := strconv.Atoi(v)
		return n
	}
	return defaultVal
}

func envStr(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func maxMemDelta(metrics []roundMetrics) float64 {
	max := 0.0
	for _, m := range metrics {
		if m.memAllocMB > max {
			max = m.memAllocMB
		}
	}
	return max
}

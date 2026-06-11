package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var totalIPs int

type toggleRequest struct {
	IPs []string `json:"ips"`
}

type statusResponse struct {
	Total int `json:"total"`
	Down  int `json:"down"`
}

func main() {
	totalIPs = 10000
	if v := os.Getenv("SIMULATOR_IP_COUNT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			totalIPs = n
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/up", handleUp)
	mux.HandleFunc("/down", handleDown)
	mux.HandleFunc("/status", handleStatus)
	mux.HandleFunc("/reset", handleReset)

	log.Printf("Simulator API listening on :8080 (%d IPs)", totalIPs)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cmd := exec.Command("nft", "flush", "set", "inet", "sim", "down_ips")
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("nft flush failed: %v, output: %s", err, out)
		http.Error(w, "nft error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleDown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req toggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	for _, ip := range req.IPs {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		cmd := exec.Command("nft", "add", "element", "inet", "sim", "down_ips", "{", ip, "}")
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Printf("nft add failed for %s: %v, output: %s", ip, err, out)
			http.Error(w, "nft error", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req toggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	for _, ip := range req.IPs {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		cmd := exec.Command("nft", "delete", "element", "inet", "sim", "down_ips", "{", ip, "}")
		cmd.Run() // Ignore error if IP was not in set
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("nft", "list", "set", "inet", "sim", "down_ips")
	out, _ := cmd.Output()
	down := 0
	output := string(out)
	if idx := strings.Index(output, "elements = {"); idx != -1 {
		rest := output[idx+len("elements = {"):]
		endIdx := strings.Index(rest, "}")
		if endIdx != -1 {
			elements := strings.TrimSpace(rest[:endIdx])
			if elements != "" {
				down = len(strings.Split(elements, ","))
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(statusResponse{
		Total: totalIPs,
		Down:  down,
	})
}

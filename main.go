package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	const PORT = "8080"
	VERSION := os.Getenv("VERSION")
	if VERSION == "" {
		VERSION = "UNKNOWN"
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{
			"message": "Welcome to the main nerve center!",
		}
		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("GET /config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := os.ReadFile("/config.json")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(data)
	})

	mux.HandleFunc("GET /nameserver", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		data, err := os.ReadFile("/etc/resolv.conf")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		type ResolvConfig struct {
			Nameservers []string `json:"nameservers"`
			Search      []string `json:"search,omitempty"`
			Options     []string `json:"options,omitempty"`
		}

		config := ResolvConfig{
			Nameservers: []string{},
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
				continue // skip empty lines and comments
			}

			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}

			switch fields[0] {
			case "nameserver":
				config.Nameservers = append(config.Nameservers, fields[1])
			case "search":
				config.Search = fields[1:]
			case "options":
				config.Options = fields[1:]
			}
		}

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.Encode(config)
	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{
			"status":  "healthy",
			"version": VERSION,
		}
		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("GET /secrets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{}
		response["POSTGRES_PASSWORD"] = os.Getenv("POSTGRES_PASSWORD")
		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("GET /visit", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		volumePath := "/app/data/visits.txt"

		f, err := os.OpenFile(volumePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			http.Error(w, "Volume write error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		timestamp := time.Now().Format(time.RFC3339)
		if _, err := f.WriteString(timestamp + "\n"); err != nil {
			f.Close()
			http.Error(w, "Failed to write data", http.StatusInternalServerError)
			return
		}
		f.Close()

		data, err := os.ReadFile(volumePath)
		if err != nil {
			http.Error(w, "Volume read error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		response := map[string]interface{}{
			"message":      "Visit logged successfully!",
			"total_visits": len(lines),
			"history":      lines,
		}
		json.NewEncoder(w).Encode(response)
	})

	fmt.Printf("Server blasting off on port: %s\n", PORT)
	log.Fatalln(http.ListenAndServe(":"+PORT, mux))
}

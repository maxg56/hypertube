package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type standardResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func respondJSON(w http.ResponseWriter, status int, body standardResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body) //nolint:errcheck
}

func notImplemented(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, standardResponse{
		Success: false,
		Error:   "comment service not yet implemented",
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8005"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, standardResponse{
			Success: true,
			Data:    map[string]string{"status": "ok", "service": "comment-service"},
		})
	})

	mux.HandleFunc("/api/v1/comments/", notImplemented)

	log.Printf("comment-service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

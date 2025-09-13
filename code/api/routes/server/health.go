package server

import (
	"encoding/json"
	"net/http"
)

// HandleHealth handles health check requests
func HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// TODO: add actual health check here somewhere please lol
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "GhostFS",
	})
}

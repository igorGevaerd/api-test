package handler

import (
	"encoding/json"
	"net/http"
)

// Health handles GET requests to /health and returns API status.
//
// HTTP Response:
//   - Status 200: API is healthy
//   - Content-Type: application/json
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

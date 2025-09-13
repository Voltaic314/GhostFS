package folders

import (
	"encoding/json"
	"net/http"
)

// DeleteRequest represents a request to delete a folder
type DeleteRequest struct {
	TableID  string `json:"table_id"`
	FolderID string `json:"folder_id"`
}

// DeleteResponse represents the response from folder deletion
type DeleteResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// HandleDelete handles requests to delete a folder
func HandleDelete(w http.ResponseWriter, r *http.Request, server interface{}) {
	var req DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual folder deletion logic
	// For now, return a placeholder response
	response := DeleteResponse{
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

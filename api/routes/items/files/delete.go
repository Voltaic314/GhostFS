package files

import (
	"encoding/json"
	"net/http"
)

// DeleteRequest represents a request to delete a file
type DeleteRequest struct {
	TableID string `json:"table_id"`
	FileID  string `json:"file_id"`
}

// DeleteResponse represents the response from file deletion
type DeleteResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	FileID  string `json:"file_id,omitempty"`
	TableID string `json:"table_id,omitempty"`
}

// HandleDelete handles requests to delete a file
func HandleDelete(w http.ResponseWriter, r *http.Request, server interface{}) {
	var req DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual file deletion logic
	// For now, return a placeholder response
	response := DeleteResponse{
		Success: true,
		FileID:  req.FileID,
		TableID: req.TableID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

package files

import (
	"encoding/json"
	"net/http"
)

// ListRequest represents a request to list files
type ListRequest struct {
	TableID   string   `json:"table_id"`
	FolderID  string   `json:"folder_id"`
	FolderIDs []string `json:"folder_ids,omitempty"` // For batch listing
}

// ListResponse represents the response with file contents
type ListResponse struct {
	Success bool     `json:"success"`
	Error   string   `json:"error,omitempty"`
	Items   []FSItem `json:"items,omitempty"`
}

// FSItem represents a filesystem item (file or folder)
type FSItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Path    string `json:"path"`
	Type    string `json:"type"`
	Size    int64  `json:"size"`
	Level   int    `json:"level"`
	Checked bool   `json:"checked"`
}

// HandleList handles requests to list items
func HandleList(w http.ResponseWriter, r *http.Request, server interface{}) {
	var req ListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual item listing logic using server
	// For now, return a placeholder response
	response := ListResponse{
		Success: true,
		Items:   []FSItem{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

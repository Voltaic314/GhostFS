package folders

import (
	"encoding/json"
	"net/http"
)

// NewRequest represents a request to create a new folder
type NewRequest struct {
	TableID  string `json:"table_id"`
	ParentID string `json:"parent_id"`
	Name     string `json:"name"`
}

// NewResponse represents the response from folder creation
type NewResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	FolderID  string `json:"folder_id,omitempty"`
	TableID   string `json:"table_id,omitempty"`
	TableName string `json:"table_name,omitempty"`
	ParentID  string `json:"parent_id,omitempty"`
	Name      string `json:"name,omitempty"`
}

// HandleNew handles requests to create a new folder
func HandleNew(w http.ResponseWriter, r *http.Request, server interface{}) {
	var req NewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual folder creation logic using server
	// For now, return a placeholder response
	response := NewResponse{
		Success:   true,
		FolderID:  "placeholder-folder-id",
		TableID:   req.TableID,
		TableName: "placeholder-table-name",
		ParentID:  req.ParentID,
		Name:      req.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

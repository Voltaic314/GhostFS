package items

import (
	"encoding/json"
	"net/http"
)

// ListRequest represents a request to list all items in a folder
type ListRequest struct {
	TableID  string `json:"table_id"`
	FolderID string `json:"folder_id"`
}

// ListResponse represents the response with folder contents
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
	Type    string `json:"type"` // "file" or "folder"
	Size    int64  `json:"size"`
	Level   int    `json:"level"`
	Checked bool   `json:"checked"`
}

// HandleList handles requests to list all items (files and folders) in a folder
func HandleList(w http.ResponseWriter, r *http.Request, server interface{}) {
	var req ListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual item listing logic using server.GetDB() and server.GetTableManager()
	// Query database for all children of the given folder_id
	// Return both files and folders in a single response

	// For now, return a placeholder response
	response := ListResponse{
		Success: true,
		Items: []FSItem{
			{
				ID:   "folder-1",
				Name: "Documents",
				Path: "/Documents",
				Type: "folder",
				Size: 0,
			},
			{
				ID:   "file-1",
				Name: "readme.txt",
				Path: "/readme.txt",
				Type: "file",
				Size: 1024,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

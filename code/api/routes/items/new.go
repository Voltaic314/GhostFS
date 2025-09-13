package items

import (
	"encoding/json"
	"net/http"
)

// NewItemRequest represents a single item to create
type NewItemRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`           // "file" or "folder"
	Size int64  `json:"size,omitempty"` // Only for files
}

// NewRequest represents a request to create one or more items
type NewRequest struct {
	TableID  string           `json:"table_id"`
	ParentID string           `json:"parent_id"`
	Items    []NewItemRequest `json:"items"`
}

// NewItemResponse represents a single created item
type NewItemResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Type    string `json:"type,omitempty"`
	Size    int64  `json:"size,omitempty"`
}

// NewResponse represents the response from creating items
type NewResponse struct {
	Success  bool              `json:"success"`
	Error    string            `json:"error,omitempty"`
	TableID  string            `json:"table_id,omitempty"`
	ParentID string            `json:"parent_id,omitempty"`
	Items    []NewItemResponse `json:"items,omitempty"`
}

// HandleNew handles requests to create one or more items (files and/or folders)
func HandleNew(w http.ResponseWriter, r *http.Request, server interface{}) {
	var req NewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual item creation logic using server
	// Loop through req.Items and create each one in the database
	// Return success/failure for each item

	// For now, return placeholder responses
	var createdItems []NewItemResponse
	for _, item := range req.Items {
		createdItems = append(createdItems, NewItemResponse{
			Success: true,
			ID:      "placeholder-" + item.Type + "-id",
			Name:    item.Name,
			Type:    item.Type,
			Size:    item.Size,
		})
	}

	response := NewResponse{
		Success:  true,
		TableID:  req.TableID,
		ParentID: req.ParentID,
		Items:    createdItems,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

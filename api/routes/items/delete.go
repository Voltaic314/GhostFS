package items

import (
	"encoding/json"
	"net/http"
)

// DeleteRequest represents a request to delete one or more items
type DeleteRequest struct {
	TableID string   `json:"table_id"`
	ItemIDs []string `json:"item_ids"` // Array of item IDs to delete
}

// DeleteItemResponse represents the result of deleting a single item
type DeleteItemResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	ID      string `json:"id"`
}

// DeleteResponse represents the response from deleting items
type DeleteResponse struct {
	Success bool                 `json:"success"`
	Error   string               `json:"error,omitempty"`
	TableID string               `json:"table_id,omitempty"`
	Items   []DeleteItemResponse `json:"items,omitempty"`
}

// HandleDelete handles requests to delete one or more items (files and/or folders)
func HandleDelete(w http.ResponseWriter, r *http.Request, server interface{}) {
	var req DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual item deletion logic using server
	// Loop through req.ItemIDs and delete each one from the database
	// Return success/failure for each item

	// For now, return placeholder responses
	var deletedItems []DeleteItemResponse
	for _, itemID := range req.ItemIDs {
		deletedItems = append(deletedItems, DeleteItemResponse{
			Success: true,
			ID:      itemID,
		})
	}

	response := DeleteResponse{
		Success: true,
		TableID: req.TableID,
		Items:   deletedItems,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

package items

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Voltaic314/GhostFS/code/types/api"
)

// Request/Response structs for this endpoint
type NewItemRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`           // "file" or "folder"
	Size int64  `json:"size,omitempty"` // Only for files
}

type CreateRequest struct {
	TableID  string           `json:"table_id"`
	ParentID string           `json:"parent_id"`
	Items    []NewItemRequest `json:"items"`
}

type CreatedItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Size int64  `json:"size,omitempty"`
}

type CreateResponseData struct {
	TableID  string        `json:"table_id"`
	ParentID string        `json:"parent_id"`
	Items    []CreatedItem `json:"items"`
}

// HandleNew handles requests to create one or more items (files and/or folders)
func HandleNew(w http.ResponseWriter, r *http.Request, server interface{}) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "Invalid JSON")
		return
	}

	// TODO: Implement actual item creation logic using server
	// Loop through req.Items and create each one in the database
	// Return success/failure for each item

	// For now, return placeholder responses
	var createdItems []CreatedItem
	for _, item := range req.Items {
		createdItems = append(createdItems, CreatedItem{
			ID:   fmt.Sprintf("placeholder-%s-id", item.Type),
			Name: item.Name,
			Type: item.Type,
			Size: item.Size,
		})
	}

	// Return successful response
	responseData := CreateResponseData{
		TableID:  req.TableID,
		ParentID: req.ParentID,
		Items:    createdItems,
	}
	api.Success(w, responseData)
}

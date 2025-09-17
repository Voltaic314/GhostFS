package items

import (
	"encoding/json"
	"net/http"

	"github.com/Voltaic314/GhostFS/code/core/items"
	"github.com/Voltaic314/GhostFS/code/db"
	"github.com/Voltaic314/GhostFS/code/db/tables"
	"github.com/Voltaic314/GhostFS/code/types/api"
	dbTypes "github.com/Voltaic314/GhostFS/code/types/db"
)

// Request/Response structs for this endpoint
type ListRequest struct {
	TableID     string `json:"table_id"`
	FolderID    string `json:"folder_id"`
	FoldersOnly bool   `json:"folders_only,omitempty"` // Optional: only return folders
}

type ListResponseData struct {
	Items []dbTypes.Node `json:"items"`
}

// HandleList handles requests to list all items (files and folders) in a folder
func HandleList(w http.ResponseWriter, r *http.Request, server interface{}) {
	var req ListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "Invalid JSON")
		return
	}

	// Cast server to get access to DB and TableManager
	s := server.(interface {
		GetTableManager() *tables.TableManager
		GetDB() *db.DB
		GetDeterministicGenerator() *tables.DeterministicGenerator
	})

	tableManager := s.GetTableManager()
	database := s.GetDB()
	generator := s.GetDeterministicGenerator()

	// Convert API request to core request
	coreReq := items.ListItemsRequest{
		TableID:     req.TableID,
		FolderID:    req.FolderID,
		FoldersOnly: req.FoldersOnly,
	}

	// Call core logic
	coreResp, err := items.ListItems(tableManager, database, generator, coreReq)
	if err != nil {
		api.InternalError(w, err.Error())
		return
	}

	// Convert core response to API response
	responseData := ListResponseData{Items: coreResp.Items}
	api.Success(w, responseData)
}

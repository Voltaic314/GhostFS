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
type GetRootRequest struct {
	TableID string `json:"table_id"`
}

type GetRootResponseData struct {
	Root dbTypes.Node `json:"root"`
}

// HandleGetRoot handles requests to get the root node for a table
func HandleGetRoot(w http.ResponseWriter, r *http.Request, server interface{}) {
	var req GetRootRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "Invalid JSON")
		return
	}

	// Cast server to get access to DB and TableManager
	s := server.(interface {
		GetTableManager() *tables.TableManager
		GetDB() *db.DB
	})

	tableManager := s.GetTableManager()
	database := s.GetDB()

	// Convert API request to core request
	coreReq := items.GetRootRequest{
		TableID: req.TableID,
	}

	// Call core logic
	coreResp, err := items.GetRoot(tableManager, database, coreReq)
	if err != nil {
		api.InternalError(w, err.Error())
		return
	}

	// Convert core response to API response
	responseData := GetRootResponseData{Root: coreResp.Root}
	api.Success(w, responseData)
}

package items

import (
	"encoding/json"
	"fmt"
	"net/http"

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

	// Get table name from table ID (check cache first)
	tableName, exists := tableManager.GetTableNameByID(req.TableID)
	if !exists {
		// Not in cache, try to load from lookup table
		var err error
		tableName, err = tables.GetTableName(database, req.TableID)
		if err != nil {
			api.BadRequest(w, fmt.Sprintf("Invalid table_id: %s", req.TableID))
			return
		}
	}

	// Build SQL query to get the root node (level = 0)
	query := fmt.Sprintf("SELECT id, name, path, type, size, level, checked FROM %s WHERE level = 0 LIMIT 1", tableName)

	// Execute query
	rows, err := database.Query(tableName, query)
	if err != nil {
		api.InternalError(w, fmt.Sprintf("Database query failed: %v", err))
		return
	}
	defer rows.Close()

	// Parse result - should only be one root node
	var rootNode dbTypes.Node
	if rows.Next() {
		err := rows.Scan(&rootNode.ID, &rootNode.Name, &rootNode.Path, &rootNode.Type, &rootNode.Size, &rootNode.Level, &rootNode.Checked)
		if err != nil {
			api.InternalError(w, fmt.Sprintf("Failed to parse database results: %v", err))
			return
		}
	} else {
		api.NotFound(w, "Root node not found for this table")
		return
	}

	// Return successful response
	responseData := GetRootResponseData{Root: rootNode}
	api.Success(w, responseData)
}

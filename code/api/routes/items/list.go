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

	// Build SQL query to get children of the specified folder
	query := fmt.Sprintf("SELECT id, name, path, type, size, level, checked FROM %s WHERE parent_id = ?", tableName)

	// Add folders_only filter if requested
	if req.FoldersOnly {
		query += " AND type = 'folder'"
	}

	// Order by type (folders first) then by name
	query += " ORDER BY type DESC, name ASC"

	// Execute query
	rows, err := database.Query(tableName, query, req.FolderID)
	if err != nil {
		api.InternalError(w, fmt.Sprintf("Database query failed: %v", err))
		return
	}
	defer rows.Close()

	// Parse results
	var items []dbTypes.Node
	for rows.Next() {
		var item dbTypes.Node
		err := rows.Scan(&item.ID, &item.Name, &item.Path, &item.Type, &item.Size, &item.Level, &item.Checked)
		if err != nil {
			api.InternalError(w, fmt.Sprintf("Failed to parse database results: %v", err))
			return
		}
		items = append(items, item)
	}

	// Queue an update to mark the parent folder as "checked" (accessed)
	// This tracks which folders have been listed/accessed without impacting performance
	updateQuery := fmt.Sprintf("UPDATE %s SET checked = TRUE WHERE id = ?", tableName)
	database.QueueWrite(tableName, updateQuery, req.FolderID)

	// Return successful response
	responseData := ListResponseData{Items: items}
	api.Success(w, responseData)
}

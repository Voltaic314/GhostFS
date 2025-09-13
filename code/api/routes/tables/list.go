package tables

import (
	"encoding/json"
	"net/http"

	"github.com/Voltaic314/GhostFS/code/db"
	"github.com/Voltaic314/GhostFS/code/db/tables"
)

// ListTablesResponse represents the response with table information
type ListTablesResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Tables  []TableInfo `json:"tables,omitempty"`
}

// TableInfo represents information about a table
type TableInfo struct {
	TableID   string `json:"table_id"`
	TableName string `json:"table_name"`
	Type      string `json:"type"` // "primary" or "secondary"
}

// HandleListTables handles requests to list all node tables
func HandleListTables(w http.ResponseWriter, r *http.Request, serverInterface interface{}) {
	// Cast to the actual server type
	server := serverInterface.(interface {
		GetTableManager() *tables.TableManager
		GetDB() *db.DB
	})

	tableManager := server.GetTableManager()

	var tableList []TableInfo

	// Get primary table
	primaryTableName := tableManager.GetPrimaryTableName()
	if primaryTableID, exists := tableManager.GetTableIDByName(primaryTableName); exists {
		tableList = append(tableList, TableInfo{
			TableID:   primaryTableID,
			TableName: primaryTableName,
			Type:      "primary",
		})
	}

	// Get secondary tables
	secondaryConfigs := tableManager.GetSecondaryTableConfigs()
	for _, config := range secondaryConfigs {
		if secondaryTableID, exists := tableManager.GetTableIDByName(config.TableName); exists {
			tableList = append(tableList, TableInfo{
				TableID:   secondaryTableID,
				TableName: config.TableName,
				Type:      "secondary",
			})
		}
	}

	response := ListTablesResponse{
		Success: true,
		Tables:  tableList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

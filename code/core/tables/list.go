package tables

import (
	"fmt"

	"github.com/Voltaic314/GhostFS/code/db"
	"github.com/Voltaic314/GhostFS/code/db/tables"
	dbTypes "github.com/Voltaic314/GhostFS/code/types/db"
)

// ListTablesResponse represents the output for listing tables
type ListTablesResponse struct {
	Tables []dbTypes.TableInfo
}

// ListTables lists all node tables
func ListTables(database *db.DB) (*ListTablesResponse, error) {
	// Get all table mappings with types from the database
	tableMappingsWithTypes, err := tables.GetAllTableMappingsWithTypes(database)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve table mappings from database: %w", err)
	}

	var tableList []dbTypes.TableInfo
	for tableID, info := range tableMappingsWithTypes {
		tableList = append(tableList, dbTypes.TableInfo{
			TableID:   tableID,
			TableName: info["table_name"],
			Type:      info["type"],
		})
	}

	return &ListTablesResponse{Tables: tableList}, nil
}

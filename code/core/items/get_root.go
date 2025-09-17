package items

import (
	"fmt"

	"github.com/Voltaic314/GhostFS/code/db"
	"github.com/Voltaic314/GhostFS/code/db/tables"
	dbTypes "github.com/Voltaic314/GhostFS/code/types/db"
)

// GetRootRequest represents the input for getting root node
type GetRootRequest struct {
	TableID string
}

// GetRootResponse represents the output for getting root node
type GetRootResponse struct {
	Root dbTypes.Node
}

// GetRoot gets the root node for a table
func GetRoot(tableManager *tables.TableManager, database *db.DB, req GetRootRequest) (*GetRootResponse, error) {
	// Get table name from table ID (check cache first)
	tableName, exists := tableManager.GetTableNameByID(req.TableID)
	if !exists {
		// Not in cache, try to load from lookup table
		var err error
		tableName, err = tables.GetTableName(database, req.TableID)
		if err != nil {
			return nil, fmt.Errorf("invalid table_id: %s", req.TableID)
		}
	}

	// Build SQL query to get the root node (level = 0)
	query := fmt.Sprintf("SELECT id, name, path, type, size, level, checked FROM %s WHERE level = 0 LIMIT 1", tableName)

	// Execute query
	rows, err := database.Query(tableName, query)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer rows.Close()

	// Parse result - should only be one root node
	var rootNode dbTypes.Node
	if rows.Next() {
		err := rows.Scan(&rootNode.ID, &rootNode.Name, &rootNode.Path, &rootNode.Type, &rootNode.Size, &rootNode.Level, &rootNode.Checked)
		if err != nil {
			return nil, fmt.Errorf("failed to parse database results: %w", err)
		}
	} else {
		return nil, fmt.Errorf("root node not found for this table")
	}

	return &GetRootResponse{Root: rootNode}, nil
}

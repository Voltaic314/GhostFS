package items

import (
	"fmt"

	"github.com/Voltaic314/GhostFS/code/db"
	"github.com/Voltaic314/GhostFS/code/db/tables"
	dbTypes "github.com/Voltaic314/GhostFS/code/types/db"
)

// ListItemsRequest represents the input for listing items
type ListItemsRequest struct {
	TableID     string
	FolderID    string
	FoldersOnly bool
}

// ListItemsResponse represents the output for listing items
type ListItemsResponse struct {
	Items []dbTypes.Node
}

// ListItems lists all items (files and folders) in a folder
func ListItems(tableManager *tables.TableManager, database *db.DB, generator *tables.DeterministicGenerator, req ListItemsRequest) (*ListItemsResponse, error) {
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

	// Get folder information from database (we need path and level for generation)
	folderInfo, err := generator.GetFolderInfo(req.FolderID, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get folder info: %w", err)
	}

	// Use deterministic generator instead of database query
	items, err := generator.GenerateChildren(req.FolderID, folderInfo.Path, folderInfo.Level, req.FoldersOnly, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate children: %w", err)
	}

	// Mark the parent folder as accessed (async)
	generator.MarkFolderAccessed(req.FolderID, tableName)

	return &ListItemsResponse{Items: items}, nil
}

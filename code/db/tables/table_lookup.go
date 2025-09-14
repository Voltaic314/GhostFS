package tables

import (
	"github.com/Voltaic314/GhostFS/code/db"
	"github.com/google/uuid"
)

// TableLookup represents a lookup table for table IDs to their respective table names
type TableLookup struct{}

func (t *TableLookup) Name() string {
	return "table_id_lookup"
}

func (t *TableLookup) Schema() string {
	return `
		table_id VARCHAR NOT NULL PRIMARY KEY,
		table_name VARCHAR NOT NULL,
		type VARCHAR NOT NULL
	`
}

// Init creates the table_id_lookup table asynchronously.
func (t *TableLookup) Init(db *db.DB) error {
	done := make(chan error)
	go func() {
		done <- db.CreateTable(t.Name(), t.Schema())
	}()
	return <-done
}

// GetTableName returns the table name for a given table ID
func GetTableName(db *db.DB, tableID string) (string, error) {
	var tableName string
	query := "SELECT table_name FROM table_id_lookup WHERE table_id = ?"
	err := db.QueryRow(query, tableID).Scan(&tableName)
	return tableName, err
}

// SetTableName sets the table name and type for a given table ID
func SetTableName(db *db.DB, tableID, tableName, tableType string) error {
	query := "INSERT OR REPLACE INTO table_id_lookup (table_id, table_name, type) VALUES (?, ?, ?)"
	_, err := db.Exec(query, tableID, tableName, tableType)
	return err
}

// GetAllTableMappings returns all table ID to name mappings with their types
func GetAllTableMappings(db *db.DB) (map[string]string, error) {
	query := "SELECT table_id, table_name, type FROM table_id_lookup"
	rows, err := db.Query("", query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mappings := make(map[string]string)
	for rows.Next() {
		var tableID, tableName, tableType string
		if err := rows.Scan(&tableID, &tableName, &tableType); err != nil {
			return nil, err
		}
		mappings[tableID] = tableName
	}
	return mappings, nil
}

// GetAllTableMappingsWithTypes returns all table mappings including type information
func GetAllTableMappingsWithTypes(db *db.DB) (map[string]map[string]string, error) {
	query := "SELECT table_id, table_name, type FROM table_id_lookup"
	rows, err := db.Query("", query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mappings := make(map[string]map[string]string)
	for rows.Next() {
		var tableID, tableName, tableType string
		if err := rows.Scan(&tableID, &tableName, &tableType); err != nil {
			return nil, err
		}
		mappings[tableID] = map[string]string{
			"table_name": tableName,
			"type":       tableType,
		}
	}
	return mappings, nil
}

// GenerateTableID generates a new UUID for a table
func GenerateTableID() string {
	return uuid.New().String()
}

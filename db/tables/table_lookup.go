package tables

import "github.com/Voltaic314/GhostFS/db"

// TableLookup represents a lookup table for folder IDs to their respective tables
type TableLookup struct{}

func (t *TableLookup) Name() string {
	return "table_lookup"
}

func (t *TableLookup) Schema() string {
	return `
		item_id VARCHAR NOT NULL PRIMARY KEY,
		table_name VARCHAR NOT NULL,
	`
}

// Init creates the table_lookup table asynchronously.
func (t *TableLookup) Init(db *db.DB) error {
	done := make(chan error)
	go func() {
		done <- db.CreateTable(t.Name(), t.Schema())
	}()
	return <-done
}

// GetTableName returns the table name for a given item ID
func GetTableName(db *db.DB, itemID string) (string, error) {
	var tableName string
	query := "SELECT table_name FROM table_lookup WHERE item_id = ?"
	err := db.QueryRow(query, itemID).Scan(&tableName)
	return tableName, err
}

// SetTableName sets the table name for a given item ID
func SetTableName(db *db.DB, itemID, tableName string) error {
	query := "INSERT OR REPLACE INTO table_lookup (item_id, table_name) VALUES (?, ?)"
	_, err := db.Exec(query, itemID, tableName)
	return err
}

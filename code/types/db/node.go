package db

import "time"

// Node represents a filesystem node (file or folder) as stored in the database
type Node struct {
	ID                    string    `json:"id" db:"id"`
	ParentID              string    `json:"parent_id" db:"parent_id"`
	Name                  string    `json:"name" db:"name"`
	Path                  string    `json:"path" db:"path"`
	Type                  string    `json:"type" db:"type"` // "file" or "folder"
	Size                  int64     `json:"size" db:"size"`
	Level                 int       `json:"level" db:"level"`
	Checked               bool      `json:"checked" db:"checked"`
	SecondaryExistenceMap string    `json:"secondary_existence_map,omitempty" db:"secondary_existence_map"` // JSON string
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// TableInfo represents information about a database table
type TableInfo struct {
	TableID   string `json:"table_id"`
	TableName string `json:"table_name"`
	Type      string `json:"type"` // "primary" or "secondary"
}

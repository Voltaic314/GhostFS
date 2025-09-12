package tables

import "github.com/Voltaic314/GhostFS/db"

// NodesTable represents a generic nodes table for file system emulation
type NodesTable struct {
	TableName string
}

// NewNodesTable creates a new nodes table with the specified name
func NewNodesTable(tableName string) *NodesTable {
	return &NodesTable{TableName: tableName}
}

func (t *NodesTable) Name() string {
	return t.TableName
}

func (t *NodesTable) Schema() string {
	return `
		id VARCHAR NOT NULL PRIMARY KEY,
		parent_id VARCHAR NOT NULL,
		name VARCHAR NOT NULL,
		path VARCHAR NOT NULL,
		type VARCHAR NOT NULL CHECK(type IN ('file', 'folder')),
		size BIGINT,
		level INTEGER NOT NULL,
		checked BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	`
}

// Init creates the nodes table asynchronously.
func (t *NodesTable) Init(db *db.DB) error {
	done := make(chan error)
	go func() {
		done <- db.CreateTable(t.Name(), t.Schema())
	}()
	return <-done
}

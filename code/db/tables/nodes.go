package tables

import (
	"encoding/json"

	"github.com/Voltaic314/GhostFS/code/db"
)

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
		secondary_existence_map JSON,
		child_seed BIGINT,
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

// SecondaryExistenceMap represents which secondary tables contain a node
type SecondaryExistenceMap map[string]bool

// NewSecondaryExistenceMap creates a new existence map for all secondary tables
func NewSecondaryExistenceMap(secondaryTableNames []string) SecondaryExistenceMap {
	existenceMap := make(SecondaryExistenceMap)
	for _, tableName := range secondaryTableNames {
		existenceMap[tableName] = false
	}
	return existenceMap
}

// ToJSON converts the existence map to JSON string
func (sem SecondaryExistenceMap) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(sem)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// FromJSON creates an existence map from JSON string
func FromJSON(jsonStr string) (SecondaryExistenceMap, error) {
	var existenceMap SecondaryExistenceMap
	if jsonStr == "" {
		return make(SecondaryExistenceMap), nil
	}
	err := json.Unmarshal([]byte(jsonStr), &existenceMap)
	return existenceMap, err
}

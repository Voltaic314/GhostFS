package tables

import (
	"fmt"
	"strings"

	"github.com/Voltaic314/GhostFS/db"
)

// TableManager handles table operations for single/multi table modes
type TableManager struct {
	config       *TestConfig
	tableIDMap   map[string]string // table_id -> table_name cache
	tableNameMap map[string]string // table_name -> table_id cache
}

// NewTableManager creates a new table manager
func NewTableManager(config *TestConfig) *TableManager {
	return &TableManager{
		config:       config,
		tableIDMap:   make(map[string]string),
		tableNameMap: make(map[string]string),
	}
}

// IsMultiTableMode returns true if we have secondary tables
func (tm *TableManager) IsMultiTableMode() bool {
	return len(tm.config.Database.Tables.Secondary) > 0
}

// GetPrimaryTableName returns the primary table name
func (tm *TableManager) GetPrimaryTableName() string {
	return tm.config.Database.Tables.Primary.TableName
}

// GetPrimaryConfig returns the primary table configuration
func (tm *TableManager) GetPrimaryConfig() PrimaryTableConfig {
	return tm.config.Database.Tables.Primary
}

// GetTableNames returns all table names that should be created
func (tm *TableManager) GetTableNames() []string {
	tables := []string{tm.GetPrimaryTableName()}
	tables = append(tables, tm.GetSecondaryTableNames()...)
	return tables
}

// GetTableForNode returns the appropriate table name for a node based on dst_prob
// Uses weighted random selection based on dst_prob values
func (tm *TableManager) GetTableForNode(nodeID string) string {
	if !tm.IsMultiTableMode() {
		return tm.GetPrimaryTableName()
	}

	// Simple hash-based distribution for now
	// In a real implementation, you might want weighted random selection
	hash := 0
	for _, char := range nodeID {
		hash += int(char)
	}

	// Get all secondary tables
	secondaryConfigs := tm.GetSecondaryTableConfigs()
	if len(secondaryConfigs) == 0 {
		return tm.GetPrimaryTableName()
	}

	// Use hash to select from available tables (primary + secondary)
	allTables := []string{tm.GetPrimaryTableName()}
	for _, config := range secondaryConfigs {
		allTables = append(allTables, config.TableName)
	}

	tableIndex := hash % len(allTables)
	return allTables[tableIndex]
}

// GetQueryTables returns the table names to query for listing contents
func (tm *TableManager) GetQueryTables() []string {
	return tm.GetTableNames()
}

// GetSecondaryTableNames returns only the secondary table names
func (tm *TableManager) GetSecondaryTableNames() []string {
	var secondaryNames []string
	for _, config := range tm.config.Database.Tables.Secondary {
		secondaryNames = append(secondaryNames, config.TableName)
	}
	return secondaryNames
}

// GetSecondaryTableConfigs returns the secondary table configurations
func (tm *TableManager) GetSecondaryTableConfigs() map[string]SecondaryTableConfig {
	return tm.config.Database.Tables.Secondary
}

// BuildUnionQuery builds a UNION query for listing contents across multiple tables
// This is now used only by the API layer for querying across all tables
func (tm *TableManager) BuildUnionQuery(baseQuery string) string {
	tables := tm.GetQueryTables()
	if len(tables) == 1 {
		return strings.Replace(baseQuery, "{{TABLE}}", tables[0], -1)
	}

	// Build UNION query
	var unionParts []string
	for _, table := range tables {
		tableQuery := strings.Replace(baseQuery, "{{TABLE}}", table, -1)
		// Clean up whitespace and newlines for proper SQL formatting
		tableQuery = strings.ReplaceAll(tableQuery, "\n", " ")
		tableQuery = strings.ReplaceAll(tableQuery, "\t", " ")
		// Remove multiple spaces
		for strings.Contains(tableQuery, "  ") {
			tableQuery = strings.ReplaceAll(tableQuery, "  ", " ")
		}
		tableQuery = strings.TrimSpace(tableQuery)
		unionParts = append(unionParts, tableQuery)
	}

	return strings.Join(unionParts, " UNION ALL ")
}

// GetTableCreationOrder returns the order in which tables should be created
// Primary table is always created first
func (tm *TableManager) GetTableCreationOrder() []string {
	return tm.GetTableNames()
}

// ValidateConfig validates the table configuration
func (tm *TableManager) ValidateConfig() error {
	// Validate primary table
	if tm.config.Database.Tables.Primary.TableName == "" {
		return fmt.Errorf("primary table name cannot be empty")
	}

	// Validate secondary tables
	for tableID, config := range tm.config.Database.Tables.Secondary {
		if config.TableName == "" {
			return fmt.Errorf("secondary table %s name cannot be empty", tableID)
		}
		if config.TableName == tm.config.Database.Tables.Primary.TableName {
			return fmt.Errorf("secondary table %s name cannot be the same as primary table name", tableID)
		}
		if config.DstProb < 0.0 || config.DstProb > 1.0 {
			return fmt.Errorf("secondary table %s dst_prob must be between 0.0 and 1.0", tableID)
		}
	}

	// Check for duplicate table names
	tableNames := make(map[string]bool)
	tableNames[tm.config.Database.Tables.Primary.TableName] = true

	for _, config := range tm.config.Database.Tables.Secondary {
		if tableNames[config.TableName] {
			return fmt.Errorf("duplicate table name: %s", config.TableName)
		}
		tableNames[config.TableName] = true
	}

	return nil
}

// GetGenerationConfigForTable returns the generation configuration for a specific table
func (tm *TableManager) GetGenerationConfigForTable(tableName string) PrimaryTableConfig {
	// Always return primary config since only primary table has generation config
	return tm.GetPrimaryConfig()
}

// GetSecondaryTableIDs returns the IDs of all secondary tables
func (tm *TableManager) GetSecondaryTableIDs() []string {
	var ids []string
	for id := range tm.config.Database.Tables.Secondary {
		ids = append(ids, id)
	}
	return ids
}

// GetTableConfigByID returns the table configuration for a given table ID
func (tm *TableManager) GetTableConfigByID(tableID string) (interface{}, bool) {
	if tableID == "primary" {
		return tm.GetPrimaryConfig(), true
	}

	config, exists := tm.config.Database.Tables.Secondary[tableID]
	return config, exists
}

// InitializeTableIDs generates and caches table IDs for all tables
func (tm *TableManager) InitializeTableIDs() {
	// Clear existing maps
	tm.tableIDMap = make(map[string]string)
	tm.tableNameMap = make(map[string]string)

	// Generate ID for primary table
	primaryTableName := tm.GetPrimaryTableName()
	primaryTableID := GenerateTableID()
	tm.tableIDMap[primaryTableID] = primaryTableName
	tm.tableNameMap[primaryTableName] = primaryTableID

	// Generate IDs for secondary tables
	for _, config := range tm.config.Database.Tables.Secondary {
		tableID := GenerateTableID()
		tm.tableIDMap[tableID] = config.TableName
		tm.tableNameMap[config.TableName] = tableID
	}
}

// GetTableNameByID returns the table name for a given table ID
func (tm *TableManager) GetTableNameByID(tableID string) (string, bool) {
	tableName, exists := tm.tableIDMap[tableID]
	return tableName, exists
}

// GetTableIDByName returns the table ID for a given table name
func (tm *TableManager) GetTableIDByName(tableName string) (string, bool) {
	tableID, exists := tm.tableNameMap[tableName]
	return tableID, exists
}

// GetTableIDForQuery returns the table ID to use for a query
// If single table mode, returns the primary table ID
// If multi table mode, requires tableID parameter
func (tm *TableManager) GetTableIDForQuery(tableID string) (string, error) {
	if !tm.IsMultiTableMode() {
		// Single table mode - return primary table ID
		primaryTableName := tm.GetPrimaryTableName()
		if tableID, exists := tm.tableNameMap[primaryTableName]; exists {
			return tableID, nil
		}
		return "", fmt.Errorf("primary table ID not found in cache")
	}

	// Multi table mode - validate the provided table ID
	if _, exists := tm.tableIDMap[tableID]; !exists {
		return "", fmt.Errorf("invalid table ID: %s", tableID)
	}

	return tableID, nil
}

// LoadTableMappingsFromDB loads table ID mappings from the database
func (tm *TableManager) LoadTableMappingsFromDB(db *db.DB) error {
	mappings, err := GetAllTableMappings(db)
	if err != nil {
		return fmt.Errorf("load table mappings from DB: %w", err)
	}

	// Update cache with loaded mappings
	tm.tableIDMap = make(map[string]string)
	tm.tableNameMap = make(map[string]string)

	for tableID, tableName := range mappings {
		tm.tableIDMap[tableID] = tableName
		tm.tableNameMap[tableName] = tableID
	}

	return nil
}

// SaveTableMappingsToDB saves current table ID mappings to the database
func (tm *TableManager) SaveTableMappingsToDB(db *db.DB) error {
	for tableID, tableName := range tm.tableIDMap {
		if err := SetTableName(db, tableID, tableName); err != nil {
			return fmt.Errorf("save table mapping %s->%s: %w", tableID, tableName, err)
		}
	}
	return nil
}

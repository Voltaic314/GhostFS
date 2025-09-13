package tables

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/marcboeker/go-duckdb"
)

// TestRunner handles DB operations for testing
type TestRunner struct {
	db           *sql.DB
	config       *TestConfig
	tableManager *TableManager
}

// NewTestRunner creates a new test runner with the given config
func NewTestRunner(config *TestConfig) (*TestRunner, error) {
	db, err := sql.Open("duckdb", config.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// Create table manager
	tableManager := NewTableManager(config)
	if err := tableManager.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid table config: %w", err)
	}

	return &TestRunner{
		db:           db,
		config:       config,
		tableManager: tableManager,
	}, nil
}

// Close closes the database connection
func (r *TestRunner) Close() error {
	return r.db.Close()
}

// InitTables creates all tables based on the configuration
func (r *TestRunner) InitTables(ctx context.Context) error {
	// Create table lookup table first
	lookupTable := &TableLookup{}
	ddl := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", lookupTable.Name(), lookupTable.Schema())
	if _, err := r.db.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("creating table %q: %w", lookupTable.Name(), err)
	}
	fmt.Printf("📜 Created table: %s\n", lookupTable.Name())

	// Create nodes tables based on configuration
	tableNames := r.tableManager.GetTableNames()
	for _, tableName := range tableNames {
		nodesTable := NewNodesTable(tableName)
		ddl := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", nodesTable.Name(), nodesTable.Schema())
		if _, err := r.db.ExecContext(ctx, ddl); err != nil {
			return fmt.Errorf("creating table %q: %w", nodesTable.Name(), err)
		}
		fmt.Printf("📜 Created table: %s\n", nodesTable.Name())
	}

	// Print mode information
	if r.tableManager.IsMultiTableMode() {
		fmt.Printf("✅ Initialized %d tables in multi-table mode\n", len(tableNames)+1)
		fmt.Printf("   Primary: %s\n", r.tableManager.GetPrimaryTableName())
		secondaryNames := r.tableManager.GetSecondaryTableNames()
		for i, name := range secondaryNames {
			fmt.Printf("   Secondary %d: %s\n", i, name)
		}
	} else {
		fmt.Printf("✅ Initialized %d tables in single-table mode\n", len(tableNames)+1)
		fmt.Printf("   Primary: %s\n", r.tableManager.GetPrimaryTableName())
	}

	return nil
}

// GetTableManager returns the table manager for this test runner
func (r *TestRunner) GetTableManager() *TableManager {
	return r.tableManager
}

// GetPrimaryTableName returns the primary table name
func (r *TestRunner) GetPrimaryTableName() string {
	return r.tableManager.GetPrimaryTableName()
}

// GetSecondaryTableNames returns all secondary table names
func (r *TestRunner) GetSecondaryTableNames() []string {
	return r.tableManager.GetSecondaryTableNames()
}

// GetAllTableNames returns all table names
func (r *TestRunner) GetAllTableNames() []string {
	return r.tableManager.GetTableNames()
}

// IsMultiTableMode returns true if running in multi-table mode
func (r *TestRunner) IsMultiTableMode() bool {
	return r.tableManager.IsMultiTableMode()
}

// QueryAllTables runs a query across all tables using UNION
func (r *TestRunner) QueryAllTables(ctx context.Context, baseQuery string, args ...interface{}) (*sql.Rows, error) {
	unionQuery := r.tableManager.BuildUnionQuery(baseQuery)
	return r.db.QueryContext(ctx, unionQuery, args...)
}

// QueryPrimaryTable runs a query on just the primary table
func (r *TestRunner) QueryPrimaryTable(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// Replace {{TABLE}} placeholder with primary table name
	primaryQuery := fmt.Sprintf(query, r.tableManager.GetPrimaryTableName())
	return r.db.QueryContext(ctx, primaryQuery, args...)
}

// QuerySecondaryTable runs a query on a specific secondary table
func (r *TestRunner) QuerySecondaryTable(ctx context.Context, tableID, query string, args ...interface{}) (*sql.Rows, error) {
	config, exists := r.tableManager.GetTableConfigByID(tableID)
	if !exists {
		return nil, fmt.Errorf("secondary table %s not found", tableID)
	}

	// Get table name from config
	var tableName string
	switch c := config.(type) {
	case SecondaryTableConfig:
		tableName = c.TableName
	default:
		return nil, fmt.Errorf("invalid config type for table %s", tableID)
	}

	// Replace {{TABLE}} placeholder with table name
	tableQuery := fmt.Sprintf(query, tableName)
	return r.db.QueryContext(ctx, tableQuery, args...)
}

// GetTableStats returns statistics about all tables
func (r *TestRunner) GetTableStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Get count for each table
	for _, tableName := range r.tableManager.GetTableNames() {
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
		var count int64
		err := r.db.QueryRowContext(ctx, query).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("failed to get count for table %s: %w", tableName, err)
		}
		stats[tableName] = count
	}

	return stats, nil
}

// Cleanup removes the database file
func (r *TestRunner) Cleanup() error {
	if err := r.db.Close(); err != nil {
		return err
	}

	// Remove the database file
	if err := os.Remove(r.config.Database.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove db file: %w", err)
	}

	// Remove WAL file if it exists
	walFile := r.config.Database.Path + ".wal"
	if err := os.Remove(walFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove wal file: %w", err)
	}

	fmt.Println("🧹 Cleaned up database files")
	return nil
}

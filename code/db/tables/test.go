package tables

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestSchema(t *testing.T) {
	// Load config
	configData, err := os.ReadFile("config.json")
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	var config TestConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Create test runner
	runner, err := NewTestRunner(&config)
	if err != nil {
		t.Fatalf("Failed to create test runner: %v", err)
	}
	defer runner.Close()

	// Clean up any existing test files
	runner.Cleanup()

	// Recreate runner after cleanup
	runner, err = NewTestRunner(&config)
	if err != nil {
		t.Fatalf("Failed to recreate test runner: %v", err)
	}
	defer runner.Close()

	// Initialize tables
	ctx := context.Background()
	if err := runner.InitTables(ctx); err != nil {
		t.Fatalf("Failed to initialize tables: %v", err)
	}

	// Test that all tables exist by querying them
	allTableNames := runner.GetAllTableNames()
	for _, tableName := range allTableNames {
		var count int
		err = runner.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+tableName).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query table %s: %v", tableName, err)
		}
		t.Logf("✅ Table %s exists with %d rows", tableName, count)
	}

	// Test table lookup table
	var lookupCount int
	err = runner.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM table_lookup").Scan(&lookupCount)
	if err != nil {
		t.Fatalf("Failed to query table_lookup: %v", err)
	}
	t.Logf("✅ Table lookup table exists with %d rows", lookupCount)

	// Test mode detection
	if runner.IsMultiTableMode() {
		t.Logf("✅ Running in multi-table mode with %d tables", len(allTableNames))
		secondaryNames := runner.GetSecondaryTableNames()
		for i, name := range secondaryNames {
			t.Logf("   Secondary %d: %s", i, name)
		}
	} else {
		t.Logf("✅ Running in single-table mode")
	}

	t.Logf("✅ Schema test passed - all tables created successfully")
}

func TestTableManager(t *testing.T) {
	// Load config
	configData, err := os.ReadFile("config.json")
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	var config TestConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Create test runner
	runner, err := NewTestRunner(&config)
	if err != nil {
		t.Fatalf("Failed to create test runner: %v", err)
	}
	defer runner.Close()

	// Test table manager functionality
	tableManager := runner.GetTableManager()

	// Test primary table name
	primaryName := tableManager.GetPrimaryTableName()
	if primaryName == "" {
		t.Fatalf("Primary table name should not be empty")
	}
	t.Logf("✅ Primary table: %s", primaryName)

	// Test secondary tables
	secondaryNames := tableManager.GetSecondaryTableNames()
	t.Logf("✅ Secondary tables: %v", secondaryNames)

	// Test table assignment
	testNodeIDs := []string{"node1", "node2", "node3", "node4", "node5"}
	for _, nodeID := range testNodeIDs {
		tableName := tableManager.GetTableForNode(nodeID)
		if tableName == "" {
			t.Fatalf("Table name should not be empty for node %s", nodeID)
		}
		t.Logf("✅ Node %s -> Table %s", nodeID, tableName)
	}

	// Test UNION query building
	baseQuery := "SELECT COUNT(*) FROM {{TABLE}}"
	unionQuery := tableManager.BuildUnionQuery(baseQuery)
	t.Logf("✅ UNION query: %s", unionQuery)

	t.Logf("✅ Table manager test passed")
}

func TestTableStats(t *testing.T) {
	// Load config
	configData, err := os.ReadFile("config.json")
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	var config TestConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Create test runner
	runner, err := NewTestRunner(&config)
	if err != nil {
		t.Fatalf("Failed to create test runner: %v", err)
	}
	defer runner.Close()

	// Initialize tables
	ctx := context.Background()
	if err := runner.InitTables(ctx); err != nil {
		t.Fatalf("Failed to initialize tables: %v", err)
	}

	// Get table statistics
	stats, err := runner.GetTableStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get table stats: %v", err)
	}

	// Print statistics
	t.Logf("✅ Table statistics:")
	for tableName, count := range stats {
		t.Logf("   %s: %d rows", tableName, count)
	}

	t.Logf("✅ Table stats test passed")
}

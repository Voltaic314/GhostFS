package sdk

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Voltaic314/GhostFS/code/core/items"
	coreTables "github.com/Voltaic314/GhostFS/code/core/tables"
	"github.com/Voltaic314/GhostFS/code/db"
	"github.com/Voltaic314/GhostFS/code/db/seed"
	"github.com/Voltaic314/GhostFS/code/db/tables"
	dbTypes "github.com/Voltaic314/GhostFS/code/types/db"
)

// GhostFSClient provides a clean SDK interface for ByteWave to use
type GhostFSClient struct {
	tableManager *tables.TableManager
	database     *db.DB
	generator    *tables.DeterministicGenerator
}

// NewGhostFSClient creates a new SDK client with auto-discovery
// It will look for GhostFS.db in the current directory and parent directories
// Options:
//   - generateDB: if true, creates a new database with root folders if none exists
func NewGhostFSClient(options ...bool) (*GhostFSClient, error) {
	generateDB := false
	if len(options) > 0 {
		generateDB = options[0]
	}

	// Try to find existing database file
	dbPath, err := findDatabaseFile()
	if err != nil {
		if !generateDB {
			return nil, fmt.Errorf("failed to find database file: %w", err)
		}

		// Generate a new database with root folders using existing init_db function
		fmt.Println("🗑️  No existing database found, generating new one...")
		seed.InitDB("") // Use default config.json path
		fmt.Println("✅ Database generated successfully!")

		// Now find the database file that was just created
		dbPath, err = findDatabaseFile()
		if err != nil {
			return nil, fmt.Errorf("failed to find generated database file: %w", err)
		}
	}

	return NewGhostFSClientWithDB(dbPath)
}

// NewGhostFSClientWithDB creates a new SDK client with a specific database file
func NewGhostFSClientWithDB(dbPath string) (*GhostFSClient, error) {
	// Initialize database
	database, err := db.NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create table manager
	tableManager := tables.NewTableManager(config)
	if err := tableManager.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid table configuration: %w", err)
	}

	// Initialize table IDs from database
	tableManager.InitializeTableIDs()

	// Get master seed from database
	masterSeed, err := getMasterSeed(database)
	if err != nil {
		return nil, fmt.Errorf("failed to get master seed: %w", err)
	}

	// Create deterministic generator
	generator := tables.NewDeterministicGenerator(
		database,
		tableManager.GetPrimaryConfig(),
		tableManager.GetSecondaryTableConfigs(),
		masterSeed,
		tableManager,
	)

	// Load existing seeds from database
	tableNames := tableManager.GetTableNames()
	for _, tableName := range tableNames {
		if err := generator.LoadSeedsFromDatabase(tableName); err != nil {
			// Log warning but don't fail - this is expected for new databases
			fmt.Printf("⚠️  Warning: Could not load seeds from table %s: %v\n", tableName, err)
		}
	}

	// Set up write queues for tables
	for _, tableName := range tableNames {
		database.InitWriteQueue(tableName, dbTypes.NodeWriteQueue, 1000, 100*time.Millisecond)
	}

	return &GhostFSClient{
		tableManager: tableManager,
		database:     database,
		generator:    generator,
	}, nil
}

// findDatabaseFile searches for GhostFS.db relative to the package location
func findDatabaseFile() (string, error) {
	// Get the directory of the current file (this SDK package)
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}

	// Start from the directory containing this SDK file (code/sdk/)
	packageDir := filepath.Dir(currentFile)

	// The GhostFS.db should be 2 levels up from code/sdk/ (at the project root)
	// code/sdk/ -> code/ -> project root
	projectRoot := filepath.Join(packageDir, "..", "..")
	projectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path to project root: %w", err)
	}

	// Look for GhostFS.db in the project root
	dbPath := filepath.Join(projectRoot, "GhostFS.db")
	if _, err := os.Stat(dbPath); err == nil {
		return dbPath, nil
	}

	// Fallback: try current working directory
	currentDir, err := os.Getwd()
	if err == nil {
		dbPath := filepath.Join(currentDir, "GhostFS.db")
		if _, err := os.Stat(dbPath); err == nil {
			return dbPath, nil
		}
	}

	return "", fmt.Errorf("GhostFS.db not found in project root (%s) or current directory", projectRoot)
}

// loadConfig loads the configuration from config.json
func loadConfig() (*tables.TestConfig, error) {
	// Look for config.json in current directory and parent directories
	configPath, err := findConfigFile()
	if err != nil {
		return nil, err
	}

	return tables.LoadConfig(configPath)
}

// findConfigFile searches for config.json relative to the package location
func findConfigFile() (string, error) {
	// Get the directory of the current file (this SDK package)
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}

	// Start from the directory containing this SDK file (code/sdk/)
	packageDir := filepath.Dir(currentFile)

	// The config.json should be 2 levels up from code/sdk/ (at the project root)
	// code/sdk/ -> code/ -> project root
	projectRoot := filepath.Join(packageDir, "..", "..")
	projectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path to project root: %w", err)
	}

	// Look for config.json in the project root
	configPath := filepath.Join(projectRoot, "config.json")
	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	// Fallback: search up the directory tree from the package location
	dir := packageDir
	for {
		configPath := filepath.Join(dir, "config.json")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		// Move up one directory
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			// Reached root directory
			break
		}
		dir = parentDir
	}

	// Final fallback: try current working directory
	currentDir, err := os.Getwd()
	if err == nil {
		configPath := filepath.Join(currentDir, "config.json")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	return "", fmt.Errorf("config.json not found in project root (%s) or relative to package location", projectRoot)
}

// getMasterSeed retrieves the master seed from the database
func getMasterSeed(database *db.DB) (int64, error) {
	query := "SELECT seed FROM seed_info LIMIT 1"
	var seed int64
	err := database.QueryRow(query).Scan(&seed)
	if err != nil {
		return 0, fmt.Errorf("failed to get master seed from database: %w", err)
	}
	return seed, nil
}

// Close closes the database connection
func (c *GhostFSClient) Close() error {
	if c.database != nil {
		c.database.Close()
	}
	return nil
}

// ListItems lists all items (files and folders) in a folder
func (c *GhostFSClient) ListItems(tableID, folderID string, foldersOnly bool) ([]dbTypes.Node, error) {
	req := items.ListItemsRequest{
		TableID:     tableID,
		FolderID:    folderID,
		FoldersOnly: foldersOnly,
	}

	resp, err := items.ListItems(c.tableManager, c.database, c.generator, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}

	return resp.Items, nil
}

// GetRoot gets the root node for a table
func (c *GhostFSClient) GetRoot(tableID string) (dbTypes.Node, error) {
	req := items.GetRootRequest{
		TableID: tableID,
	}

	resp, err := items.GetRoot(c.tableManager, c.database, req)
	if err != nil {
		return dbTypes.Node{}, fmt.Errorf("failed to get root: %w", err)
	}

	return resp.Root, nil
}

// ListTables lists all available tables
func (c *GhostFSClient) ListTables() ([]dbTypes.TableInfo, error) {
	resp, err := coreTables.ListTables(c.database)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}

	return resp.Tables, nil
}

// GetCacheStats returns cache statistics
func (c *GhostFSClient) GetCacheStats() map[string]int {
	return c.generator.GetCacheStats()
}

// ClearCache clears the in-memory cache
func (c *GhostFSClient) ClearCache() {
	c.generator.ClearCache()
}

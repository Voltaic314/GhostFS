package sdk

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Voltaic314/GhostFS/code/core/items"
	coreTables "github.com/Voltaic314/GhostFS/code/core/tables"
	"github.com/Voltaic314/GhostFS/code/db"
	"github.com/Voltaic314/GhostFS/code/db/seed"
	"github.com/Voltaic314/GhostFS/code/db/tables"
	dbTypes "github.com/Voltaic314/GhostFS/code/types/db"
)

// SDKConfig represents the configuration for the SDK
type SDKConfig struct {
	Database SDKDatabaseConfig `json:"database"`
}

// SDKDatabaseConfig represents the database configuration for the SDK
type SDKDatabaseConfig struct {
	Path                string          `json:"path,omitempty"`         // Optional: path to database file
	GenerateIfNotExists bool            `json:"generate_if_not_exists"` // Whether to generate database if it doesn't exist
	Tables              SDKTablesConfig `json:"tables"`
}

// SDKTablesConfig represents the tables configuration for the SDK
type SDKTablesConfig struct {
	Primary   tables.PrimaryTableConfig              `json:"primary"`
	Secondary map[string]tables.SecondaryTableConfig `json:"secondary,omitempty"`
}

// GhostFSClient provides a clean SDK interface for ByteWave to use
type GhostFSClient struct {
	tableManager *tables.TableManager
	database     *db.DB
	generator    *tables.DeterministicGenerator
}

// NewGhostFSClient creates a new SDK client with config file
// It will look for config.json in the current directory and parent directories
func NewGhostFSClient(configPath string) (*GhostFSClient, error) {
	// Load SDK config
	config, err := loadSDKConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Determine database path
	dbPath := config.Database.Path
	if dbPath == "" {
		// No path specified, use current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
		dbPath = filepath.Join(cwd, "GhostFS.db")
	}

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		if !config.Database.GenerateIfNotExists {
			return nil, fmt.Errorf("database file does not exist at %s and generate_if_not_exists is false", dbPath)
		}

		// Generate a new database with root folders using existing init_db function
		fmt.Println("🗑️  No existing database found, generating new one...")
		seed.InitDB(configPath) // Use the existing init_db function
		fmt.Println("✅ Database generated successfully!")
	}

	return NewGhostFSClientWithDB(dbPath, config.Database.Tables)
}

// NewGhostFSClientWithDB creates a new SDK client with a specific database file
func NewGhostFSClientWithDB(dbPath string, config SDKTablesConfig) (*GhostFSClient, error) {
	// Initialize database
	database, err := db.NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Convert SDK config to TestConfig format
	testConfig := &tables.TestConfig{}
	testConfig.Database.Path = dbPath
	testConfig.Database.Tables.Primary = config.Primary
	testConfig.Database.Tables.Secondary = config.Secondary

	// Create table manager
	tableManager := tables.NewTableManager(testConfig)
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

// loadSDKConfig loads the SDK configuration from a config file
func loadSDKConfig(configPath string) (*SDKConfig, error) {
	// If no config path provided, look for config.json in current directory
	if configPath == "" {
		configPath = "config.json"
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Read and parse config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config SDKConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate required fields
	if config.Database.Tables.Primary.TableName == "" {
		return nil, fmt.Errorf("config must have database.tables.primary.table_name")
	}

	return &config, nil
}

// getMasterSeed retrieves the master seed from the database
func getMasterSeed(database *db.DB) (int64, error) {
	query := "SELECT seed_value FROM seed_info LIMIT 1"
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

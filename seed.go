package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/Voltaic314/GhostFS/db"
	"github.com/Voltaic314/GhostFS/db/tables"
	typesdb "github.com/Voltaic314/GhostFS/types/db"
	"github.com/google/uuid"
	_ "github.com/marcboeker/go-duckdb"
)

// ParentNode represents a parent node from the DB for child generation
type ParentNode struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

// Node struct removed - using direct DB inserts instead of in-memory accumulation
func main() {
	cfgPath := "config.json"
	if env := os.Getenv("SEED_CONFIG"); env != "" {
		cfgPath = env
	}
	cfg, err := loadConfig(cfgPath)
	if err != nil {
		fatalf("load config: %v", err)
	}

	// Validate config
	if err := validateConfig(cfg); err != nil {
		fatalf("invalid config: %v", err)
	}

	// Create table manager
	tableManager := tables.NewTableManager(cfg)
	if err := tableManager.ValidateConfig(); err != nil {
		fatalf("invalid table config: %v", err)
	}

	// Set up RNG using primary table seed
	primaryConfig := tableManager.GetPrimaryConfig()
	seed := primaryConfig.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	rng := rand.New(rand.NewSource(seed))
	fmt.Printf("🎲 Seed: %d\n", seed)

	// Clean up existing DB
	path, _ := filepath.Abs(cfg.Database.Path)
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		fatalf("remove existing db: %v", err)
	}
	if err := os.RemoveAll(path + ".wal"); err != nil && !os.IsNotExist(err) {
		fatalf("remove existing wal: %v", err)
	}

	// Initialize DB with write queues
	DB, err := db.NewDB(path)
	if err != nil {
		fatalf("create db: %v", err)
	}
	defer DB.Close()

	// Set up write queues for tables
	tableNames := tableManager.GetTableNames()
	for _, tableName := range tableNames {
		DB.InitWriteQueue(tableName, typesdb.NodeWriteQueue, 1000, 100*time.Millisecond)
	}

	// Create tables
	if err := createTables(DB, tableManager); err != nil {
		fatalf("create tables: %v", err)
	}

	// Generate tree structure using sliding window approach
	fmt.Println("🌳 Generating tree structure...")
	totalNodes, err := generateTreeLevelByLevel(cfg, rng, DB, tableManager)
	if err != nil {
		fatalf("generate tree: %v", err)
	}

	// Force flush all queues
	for _, tableName := range tableNames {
		DB.ForceFlushTable(tableName)
	}

	if tableManager.IsMultiTableMode() {
		fmt.Printf("✅ Generated %d nodes across %d tables successfully!\n", totalNodes, len(tableNames))
	} else {
		fmt.Printf("✅ Generated %d nodes in single table mode successfully!\n", totalNodes)
	}
}

func loadConfig(path string) (*tables.TestConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg tables.TestConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func validateConfig(cfg *tables.TestConfig) error {
	// Validate primary table config
	primary := cfg.Database.Tables.Primary
	if primary.MinChildFolders < 0 || primary.MaxChildFolders < primary.MinChildFolders {
		return fmt.Errorf("invalid primary child folder range: min=%d, max=%d", primary.MinChildFolders, primary.MaxChildFolders)
	}
	if primary.MinChildFiles < 0 || primary.MaxChildFiles < primary.MinChildFiles {
		return fmt.Errorf("invalid primary child file range: min=%d, max=%d", primary.MinChildFiles, primary.MaxChildFiles)
	}
	if primary.MinDepth < 1 || primary.MaxDepth < primary.MinDepth {
		return fmt.Errorf("invalid primary depth range: min=%d, max=%d", primary.MinDepth, primary.MaxDepth)
	}

	// Validate secondary table configs
	for tableID, config := range cfg.Database.Tables.Secondary {
		if config.TableName == "" {
			return fmt.Errorf("secondary table %s name cannot be empty", tableID)
		}
		if config.DstProb < 0.0 || config.DstProb > 1.0 {
			return fmt.Errorf("invalid secondary table %s dst_prob: %f (must be 0.0-1.0)", tableID, config.DstProb)
		}
	}

	return nil
}

func createTables(db *db.DB, tableManager *tables.TableManager) error {
	// Create table lookup table first
	lookupTable := &tables.TableLookup{}
	ddl := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", lookupTable.Name(), lookupTable.Schema())
	if err := db.Write(ddl); err != nil {
		return fmt.Errorf("creating table %q: %w", lookupTable.Name(), err)
	}
	fmt.Printf("📜 Created table: %s\n", lookupTable.Name())

	// Create nodes tables
	tableNames := tableManager.GetTableNames()
	for _, tableName := range tableNames {
		nodesTable := tables.NewNodesTable(tableName)
		ddl := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", nodesTable.Name(), nodesTable.Schema())
		if err := db.Write(ddl); err != nil {
			return fmt.Errorf("creating table %q: %w", nodesTable.Name(), err)
		}
		fmt.Printf("📜 Created table: %s\n", nodesTable.Name())
	}

	return nil
}

func generateTreeLevelByLevel(cfg *tables.TestConfig, rng *rand.Rand, db *db.DB, tableManager *tables.TableManager) (int64, error) {
	// Use primary table config for generation parameters
	primaryConfig := tableManager.GetPrimaryConfig()

	// Generate random depth within range
	depth := primaryConfig.MinDepth + rng.Intn(primaryConfig.MaxDepth-primaryConfig.MinDepth+1)
	fmt.Printf("🎯 Target depth: %d\n", depth)

	var totalNodes int64

	// Generate root node
	rootID := generateUUID()

	// Insert root node
	if err := insertRootNode(db, tableManager, rootID); err != nil {
		return 0, fmt.Errorf("insert root node: %w", err)
	}
	totalNodes++

	// Process each level by querying the database
	currentLevel := 1
	for currentLevel <= depth {
		fmt.Printf("📁 Processing level %d...\n", currentLevel)

		// Query database for nodes that need children at this level
		nodeCount, err := generateChildrenForLevelFromDB(cfg, rng, db, tableManager, currentLevel)
		if err != nil {
			return 0, fmt.Errorf("generate children for level %d: %w", currentLevel, err)
		}

		if nodeCount == 0 {
			fmt.Printf("📁 No more nodes to process at level %d, stopping\n", currentLevel)
			break
		}

		totalNodes += nodeCount
		currentLevel++
	}

	return totalNodes, nil
}

func insertRootNode(db *db.DB, tableManager *tables.TableManager, rootID string) error {
	// Root path is always "/"
	rootPath := "/"

	// Determine which table to use for the root node
	tableName := tableManager.GetTableForNode(rootID)

	// Insert root node
	query := `INSERT INTO {{TABLE}} (id, parent_id, name, path, type, size, level, checked) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	query = tableManager.BuildUnionQuery(query)

	// For root node, we need to insert into the specific table
	rootQuery := fmt.Sprintf("INSERT INTO %s (id, parent_id, name, path, type, size, level, checked) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", tableName)
	db.QueueWrite(tableName, rootQuery, rootID, "", "root", rootPath, "folder", 0, 0, false)

	// Set table lookup for root
	if err := tables.SetTableName(db, rootID, tableName); err != nil {
		return fmt.Errorf("set table lookup for root: %w", err)
	}

	return nil
}

func generateChildrenForLevelFromDB(cfg *tables.TestConfig, rng *rand.Rand, db *db.DB, tableManager *tables.TableManager, level int) (int64, error) {
	var totalNodeCount int64
	const batchSize = 1000 // Process 1000 parents at a time
	var lastSeenRowID int64 = 0

	// Force flush before querying to ensure we have the latest data
	tableNames := tableManager.GetTableNames()
	for _, tableName := range tableNames {
		db.ForceFlushTable(tableName)
	}

	for {
		// Query for folder nodes at the current level that need children
		// Use rowid-based pagination for O(1) performance (rowid is monotonic)
		query := `SELECT s.rowid, s.id, s.path FROM {{TABLE}} s
			WHERE s.level = ? AND s.type = 'folder' AND s.rowid > ? 
			ORDER BY s.rowid LIMIT ?`

		// Build UNION query for all tables
		unionQuery := tableManager.BuildUnionQuery(query)

		rows, err := db.Query(tableManager.GetPrimaryTableName(), unionQuery, level-1, lastSeenRowID, batchSize)
		if err != nil {
			return 0, fmt.Errorf("query parents for level %d: %w", level, err)
		}

		var parents []ParentNode
		var maxRowID int64 = lastSeenRowID
		for rows.Next() {
			var rowID int64
			var ID, parentPath string
			if err := rows.Scan(&rowID, &ID, &parentPath); err != nil {
				rows.Close()
				return 0, fmt.Errorf("scan parent row: %w", err)
			}
			parents = append(parents, ParentNode{ID: ID, Path: parentPath})
			if rowID > maxRowID {
				maxRowID = rowID
			}
		}
		rows.Close()

		if len(parents) == 0 {
			// No more parents to process
			break
		}

		// Update lastSeenRowID for pagination
		lastSeenRowID = maxRowID

		// Generate children for this batch of parents
		nodeCount, err := generateChildrenForBatch(cfg, rng, db, tableManager, parents, level)
		if err != nil {
			return 0, fmt.Errorf("generate children for batch: %w", err)
		}

		totalNodeCount += nodeCount

		fmt.Printf("📁 Processed %d parents at level %d, generated %d children\n",
			len(parents), level, nodeCount)

		// If we got fewer results than batch size, we've reached the end
		if len(parents) < batchSize {
			break
		}
	}

	return totalNodeCount, nil
}

func generateChildrenForBatch(cfg *tables.TestConfig, rng *rand.Rand, db *db.DB, tableManager *tables.TableManager, parents []ParentNode, level int) (int64, error) {
	var nodeCount int64
	primaryConfig := tableManager.GetPrimaryConfig()

	// Process each parent in this batch
	for _, parent := range parents {
		// Generate random number of folders for this parent
		numFolders := primaryConfig.MinChildFolders + rng.Intn(primaryConfig.MaxChildFolders-primaryConfig.MinChildFolders+1)
		for i := 0; i < numFolders; i++ {
			folderID := generateUUID()
			folderName := fmt.Sprintf("folder_%d", i)
			folderPath := buildPath(parent.Path, folderName)

			// Determine which table to use for this folder
			tableName := tableManager.GetTableForNode(folderID)

			// Insert folder
			query := fmt.Sprintf("INSERT INTO %s (id, parent_id, name, path, type, size, level, checked) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", tableName)
			db.QueueWrite(tableName, query, folderID, parent.ID, folderName, folderPath, "folder", 0, level, false)
			nodeCount++

			// Set table lookup for this folder
			if err := tables.SetTableName(db, folderID, tableName); err != nil {
				return 0, fmt.Errorf("set table lookup for folder %s: %w", folderID, err)
			}
		}

		// Generate random number of files for this parent
		numFiles := primaryConfig.MinChildFiles + rng.Intn(primaryConfig.MaxChildFiles-primaryConfig.MinChildFiles+1)
		for i := 0; i < numFiles; i++ {
			fileID := generateUUID()
			fileName := fmt.Sprintf("file_%d.txt", i)
			filePath := buildPath(parent.Path, fileName)

			// Determine which table to use for this file
			tableName := tableManager.GetTableForNode(fileID)

			// Insert file
			query := fmt.Sprintf("INSERT INTO %s (id, parent_id, name, path, type, size, level, checked) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", tableName)
			fileSize := int64(100 + rng.Intn(900)) // Random size 100-999 bytes
			db.QueueWrite(tableName, query, fileID, parent.ID, fileName, filePath, "file", fileSize, level, false)
			nodeCount++

			// Set table lookup for this file
			if err := tables.SetTableName(db, fileID, tableName); err != nil {
				return 0, fmt.Errorf("set table lookup for file %s: %w", fileID, err)
			}
		}
	}

	return nodeCount, nil
}

func generateUUID() string {
	return uuid.New().String()
}

// buildPath constructs the full path for a node based on its parent's path and name
func buildPath(parentPath, name string) string {
	if parentPath == "" {
		// Root node
		return "/"
	}
	return parentPath + "/" + name
}

func fatalf(f string, a ...any) {
	fmt.Printf("❌ "+f+"\n", a...)
	os.Exit(1)
}

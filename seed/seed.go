package seed

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

// ParentNodeWithExistence represents a parent node with its secondary table existence map
type ParentNodeWithExistence struct {
	ID           string
	Path         string
	ExistenceMap tables.SecondaryExistenceMap
}

// Node struct removed - using direct DB inserts instead of in-memory accumulation
func Seed() {
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

	// Initialize table IDs
	tableManager.InitializeTableIDs()

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

	// Save table mappings to database
	if err := tableManager.SaveTableMappingsToDB(DB); err != nil {
		fatalf("save table mappings: %v", err)
	}

	// Generate tree structure using sliding window approach
	fmt.Println("🌳 Generating tree structure...")
	totalNodes, primaryNodes, err := generateTreeLevelByLevel(rng, DB, tableManager, seed)
	if err != nil {
		fatalf("generate tree: %v", err)
	}

	// Force flush all queues
	for _, tableName := range tableNames {
		DB.ForceFlushTable(tableName)
	}

	if tableManager.IsMultiTableMode() {
		// Calculate actual counts from database
		var totalDBEntries int64
		var secondaryEntries int64
		for _, tableName := range tableNames {
			tableCountQuery := `SELECT COUNT(*) FROM ` + tableName
			var tableCount int64
			if err := DB.QueryRow(tableCountQuery).Scan(&tableCount); err == nil {
				totalDBEntries += tableCount
				if tableName != tableManager.GetPrimaryTableName() {
					secondaryEntries += tableCount
				}
			}
		}

		fmt.Printf("📊 Primary table: %d nodes, Secondary tables: %d nodes\n", primaryNodes, secondaryEntries)
		fmt.Printf("💾 Total database entries: %d\n", totalDBEntries)
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
	// Create table ID lookup table first
	lookupTable := &tables.TableLookup{}
	ddl := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", lookupTable.Name(), lookupTable.Schema())
	if err := db.Write(ddl); err != nil {
		return fmt.Errorf("creating table %q: %w", lookupTable.Name(), err)
	}
	fmt.Printf("📜 Created table: %s\n", lookupTable.Name())

	// Create seed info table
	seedInfoTable := &tables.SeedInfoTable{}
	ddl = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", seedInfoTable.Name(), seedInfoTable.Schema())
	if err := db.Write(ddl); err != nil {
		return fmt.Errorf("creating table %q: %w", seedInfoTable.Name(), err)
	}
	fmt.Printf("📜 Created table: %s\n", seedInfoTable.Name())

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

func generateTreeLevelByLevel(rng *rand.Rand, db *db.DB, tableManager *tables.TableManager, seed int64) (int64, int64, error) {
	// Use primary table config for generation parameters
	primaryConfig := tableManager.GetPrimaryConfig()

	// Generate random depth within range
	depth := primaryConfig.MinDepth + rng.Intn(primaryConfig.MaxDepth-primaryConfig.MinDepth+1)
	fmt.Printf("🎯 Target depth: %d\n", depth)

	// Save seed info to database
	if err := tables.SaveSeedInfo(db, seed, depth); err != nil {
		return 0, 0, fmt.Errorf("save seed info: %w", err)
	}

	var totalNodes int64

	// Generate root node
	rootID := generateUUID()

	// Insert root node
	if err := insertRootNode(db, tableManager, rootID); err != nil {
		return 0, 0, fmt.Errorf("insert root node: %w", err)
	}
	totalNodes++

	// Force flush root node before starting generation
	db.ForceFlushTable(tableManager.GetPrimaryTableName())

	// Process each level by querying the database
	// Root is at level 0, so we start generating children for level 1
	currentLevel := 1
	for currentLevel <= depth {
		fmt.Printf("📁 Processing level %d...\n", currentLevel)

		// Query database for parent nodes at level (currentLevel-1) to generate children at currentLevel
		nodeCount, err := generateChildrenForLevelFromDB(rng, db, tableManager, currentLevel)
		if err != nil {
			return 0, 0, fmt.Errorf("generate children for level %d: %w", currentLevel, err)
		}

		if nodeCount == 0 {
			fmt.Printf("📁 No more nodes to process at level %d, stopping\n", currentLevel)
			break
		}

		totalNodes += nodeCount
		currentLevel++
	}

	// Mark generation as completed
	if err := tables.MarkGenerationCompleted(db); err != nil {
		return 0, 0, fmt.Errorf("mark generation completed: %w", err)
	}

	// Get primary table count for more detailed reporting
	primaryTableName := tableManager.GetPrimaryTableName()
	countQuery := `SELECT COUNT(*) FROM ` + primaryTableName
	var primaryNodeCount int64
	if err := db.QueryRow(countQuery).Scan(&primaryNodeCount); err != nil {
		fmt.Printf("⚠️  Could not get primary table count: %v\n", err)
		primaryNodeCount = totalNodes // Fallback to total
	}

	return totalNodes, primaryNodeCount, nil
}

func insertRootNode(db *db.DB, tableManager *tables.TableManager, rootID string) error {
	// Root path is always "/"
	rootPath := "/"

	// Create existence map for root node - root exists in all secondary tables
	secondaryTableNames := tableManager.GetSecondaryTableNames()
	existenceMap := tables.NewSecondaryExistenceMap(secondaryTableNames)

	// Root exists in all secondary tables (set all to true)
	for tableName := range existenceMap {
		existenceMap[tableName] = true
	}

	// Convert existence map to JSON
	existenceMapJSON, err := existenceMap.ToJSON()
	if err != nil {
		return fmt.Errorf("convert root existence map to JSON: %w", err)
	}

	// Insert root node into primary table
	primaryQuery := fmt.Sprintf("INSERT INTO %s (id, parent_id, name, path, type, size, level, checked, secondary_existence_map) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", tableManager.GetPrimaryTableName())
	db.QueueWrite(tableManager.GetPrimaryTableName(), primaryQuery, rootID, "", "root", rootPath, "folder", 0, 0, false, existenceMapJSON)

	// Insert root node into all secondary tables
	for _, tableName := range secondaryTableNames {
		secondaryQuery := fmt.Sprintf("INSERT INTO %s (id, parent_id, name, path, type, size, level, checked) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", tableName)
		db.QueueWrite(tableName, secondaryQuery, rootID, "", "root", rootPath, "folder", 0, 0, false)
	}

	return nil
}

func generateChildrenForLevelFromDB(rng *rand.Rand, db *db.DB, tableManager *tables.TableManager, targetLevel int) (int64, error) {
	var totalNodeCount int64
	const batchSize = 1000       // Process 1000 parents at a time
	var lastSeenRowID int64 = -1 // Start with -1 so we include rowid 0

	// Force flush before querying to ensure we have the latest data
	db.ForceFlushTable(tableManager.GetPrimaryTableName())

	// We need to find parent nodes at level (targetLevel-1) to generate children at targetLevel
	parentLevel := targetLevel - 1

	for {
		// Query ONLY the primary table for folder nodes at the parent level
		// Use rowid-based pagination for O(1) performance (rowid is monotonic)
		query := `SELECT rowid, id, path, secondary_existence_map FROM ` + tableManager.GetPrimaryTableName() + `
			WHERE level = ? AND type = 'folder' AND rowid > ? 
			ORDER BY rowid LIMIT ?`

		fmt.Printf("🔍 Querying for parents at level %d to generate children at level %d\n", parentLevel, targetLevel)
		rows, err := db.Query(tableManager.GetPrimaryTableName(), query, parentLevel, lastSeenRowID, batchSize)
		if err != nil {
			return 0, fmt.Errorf("query parents for level %d: %w", targetLevel, err)
		}

		var parents []ParentNodeWithExistence
		var maxRowID int64 = lastSeenRowID
		for rows.Next() {
			var rowID int64
			var ID, parentPath, existenceMapJSON string
			if err := rows.Scan(&rowID, &ID, &parentPath, &existenceMapJSON); err != nil {
				rows.Close()
				return 0, fmt.Errorf("scan parent row: %w", err)
			}

			// Parse existence map
			existenceMap, err := tables.FromJSON(existenceMapJSON)
			if err != nil {
				rows.Close()
				return 0, fmt.Errorf("parse existence map for %s: %w", ID, err)
			}

			parents = append(parents, ParentNodeWithExistence{
				ID:           ID,
				Path:         parentPath,
				ExistenceMap: existenceMap,
			})
			if rowID > maxRowID {
				maxRowID = rowID
			}
		}
		rows.Close()

		if len(parents) == 0 {
			// No more parents to process
			fmt.Printf("📁 No parents found at level %d, stopping generation\n", parentLevel)
			break
		}

		// Update lastSeenRowID for pagination
		lastSeenRowID = maxRowID

		// Generate children for this batch of parents
		nodeCount, err := generateChildrenForBatch(rng, db, tableManager, parents, targetLevel)
		if err != nil {
			return 0, fmt.Errorf("generate children for batch: %w", err)
		}

		totalNodeCount += nodeCount

		fmt.Printf("📁 Processed %d parents at level %d, generated %d children at level %d\n",
			len(parents), parentLevel, nodeCount, targetLevel)

		// If we got fewer results than batch size, we've reached the end
		if len(parents) < batchSize {
			break
		}
	}

	return totalNodeCount, nil
}

func generateChildrenForBatch(rng *rand.Rand, db *db.DB, tableManager *tables.TableManager, parents []ParentNodeWithExistence, targetLevel int) (int64, error) {
	var nodeCount int64
	primaryConfig := tableManager.GetPrimaryConfig()
	secondaryConfigs := tableManager.GetSecondaryTableConfigs()
	secondaryTableNames := tableManager.GetSecondaryTableNames()

	// Process each parent in this batch
	for _, parent := range parents {
		// Generate random number of folders for this parent
		numFolders := primaryConfig.MinChildFolders + rng.Intn(primaryConfig.MaxChildFolders-primaryConfig.MinChildFolders+1)
		for i := 0; i < numFolders; i++ {
			folderID := generateUUID()
			folderName := fmt.Sprintf("folder_%d", i)
			folderPath := buildPath(parent.Path, folderName)

			// Determine which secondary tables this folder should exist in
			childExistenceMap := determineSecondaryExistence(secondaryConfigs, rng)

			// Check parent dependencies for secondary tables
			childExistenceMap = checkParentDependencies(parent.ExistenceMap, childExistenceMap, secondaryTableNames)

			// Convert existence map to JSON
			existenceMapJSON, err := childExistenceMap.ToJSON()
			if err != nil {
				return 0, fmt.Errorf("convert existence map to JSON for folder %s: %w", folderID, err)
			}

			// Insert folder into primary table
			primaryQuery := fmt.Sprintf("INSERT INTO %s (id, parent_id, name, path, type, size, level, checked, secondary_existence_map) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", tableManager.GetPrimaryTableName())
			db.QueueWrite(tableManager.GetPrimaryTableName(), primaryQuery, folderID, parent.ID, folderName, folderPath, "folder", 0, targetLevel, false, existenceMapJSON)
			nodeCount++

			// Insert into secondary tables where it should exist
			for tableName, exists := range childExistenceMap {
				if exists {
					secondaryQuery := fmt.Sprintf("INSERT INTO %s (id, parent_id, name, path, type, size, level, checked) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", tableName)
					db.QueueWrite(tableName, secondaryQuery, folderID, parent.ID, folderName, folderPath, "folder", 0, targetLevel, false)
				}
			}
		}

		// Generate random number of files for this parent
		numFiles := primaryConfig.MinChildFiles + rng.Intn(primaryConfig.MaxChildFiles-primaryConfig.MinChildFiles+1)
		for i := 0; i < numFiles; i++ {
			fileID := generateUUID()
			fileName := fmt.Sprintf("file_%d.txt", i)
			filePath := buildPath(parent.Path, fileName)

			// Determine which secondary tables this file should exist in
			childExistenceMap := determineSecondaryExistence(secondaryConfigs, rng)

			// Check parent dependencies for secondary tables
			childExistenceMap = checkParentDependencies(parent.ExistenceMap, childExistenceMap, secondaryTableNames)

			// Convert existence map to JSON
			existenceMapJSON, err := childExistenceMap.ToJSON()
			if err != nil {
				return 0, fmt.Errorf("convert existence map to JSON for file %s: %w", fileID, err)
			}

			// Insert file into primary table
			primaryQuery := fmt.Sprintf("INSERT INTO %s (id, parent_id, name, path, type, size, level, checked, secondary_existence_map) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", tableManager.GetPrimaryTableName())
			fileSize := int64(100 + rng.Intn(900)) // Random size 100-999 bytes
			db.QueueWrite(tableManager.GetPrimaryTableName(), primaryQuery, fileID, parent.ID, fileName, filePath, "file", fileSize, targetLevel, false, existenceMapJSON)
			nodeCount++

			// Insert into secondary tables where it should exist
			for tableName, exists := range childExistenceMap {
				if exists {
					secondaryQuery := fmt.Sprintf("INSERT INTO %s (id, parent_id, name, path, type, size, level, checked) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", tableName)
					db.QueueWrite(tableName, secondaryQuery, fileID, parent.ID, fileName, filePath, "file", fileSize, targetLevel, false)
				}
			}
		}
	}

	return nodeCount, nil
}

func generateUUID() string {
	return uuid.New().String()
}

// determineSecondaryExistence determines which secondary tables a node should exist in based on probability
func determineSecondaryExistence(secondaryConfigs map[string]tables.SecondaryTableConfig, rng *rand.Rand) tables.SecondaryExistenceMap {
	existenceMap := make(tables.SecondaryExistenceMap)

	for _, config := range secondaryConfigs {
		// Roll the dice - if random float is less than dst_prob, include in this table
		roll := rng.Float64()
		existenceMap[config.TableName] = roll < config.DstProb
	}

	return existenceMap
}

// checkParentDependencies ensures that a child can only exist in secondary tables where its parent exists
func checkParentDependencies(parentExistenceMap, childExistenceMap tables.SecondaryExistenceMap, secondaryTableNames []string) tables.SecondaryExistenceMap {
	result := make(tables.SecondaryExistenceMap)

	for _, tableName := range secondaryTableNames {
		// Child can only exist in secondary table if parent exists there
		parentExists := parentExistenceMap[tableName]
		childWantsToExist := childExistenceMap[tableName]

		result[tableName] = parentExists && childWantsToExist
	}

	return result
}

// buildPath constructs the full path for a node based on its parent's path and name
func buildPath(parentPath, name string) string {
	return parentPath + "/" + name
}

func fatalf(f string, a ...any) {
	fmt.Printf("❌ "+f+"\n", a...)
	os.Exit(1)
}

package seed

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Voltaic314/GhostFS/code/db/tables"
	"github.com/Voltaic314/GhostFS/code/db"
	"github.com/google/uuid"
)

func InitDB(cfgPath string) {
	// Load configuration
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := loadConfig(cfgPath)
	if err != nil {
		fatalf("load config: %v", err)
	}

	// Validate config
	if err := validateConfig(cfg); err != nil {
		fatalf("invalid config: %v", err)
	}

	// Clean up existing DB
	dbPath, _ := filepath.Abs(cfg.Database.Path)
	fmt.Printf("🗑️  Removing existing database: %s\n", dbPath)
	if err := os.RemoveAll(dbPath); err != nil && !os.IsNotExist(err) {
		fatalf("remove existing db: %v", err)
	}
	if err := os.RemoveAll(dbPath + ".wal"); err != nil && !os.IsNotExist(err) {
		fatalf("remove existing wal: %v", err)
	}

	// Initialize DB
	fmt.Printf("🔧 Creating new database: %s\n", dbPath)
	database, err := db.NewDB(dbPath)
	if err != nil {
		fatalf("create db: %v", err)
	}
	defer database.Close()

	// Create table manager
	tableManager := tables.NewTableManager(cfg)
	if err := tableManager.ValidateConfig(); err != nil {
		fatalf("invalid table config: %v", err)
	}

	// Initialize table IDs
	tableManager.InitializeTableIDs()

	// Get master seed
	masterSeed := cfg.Database.Tables.Primary.Seed
	if masterSeed == 0 {
		masterSeed = time.Now().UnixNano()
	}
	fmt.Printf("🎲 Master seed: %d\n", masterSeed)

	// Create tables
	fmt.Println("📜 Creating tables...")
	if err := createTables(database, tableManager); err != nil {
		fatalf("create tables: %v", err)
	}

	// Save table mappings to database
	if err := tableManager.SaveTableMappingsToDB(database); err != nil {
		fatalf("save table mappings: %v", err)
	}

	// Save seed info to database
	if err := tables.SaveSeedInfo(database, masterSeed, cfg.Database.Tables.Primary.MaxDepth); err != nil {
		fatalf("save seed info: %v", err)
	}

	// Create root nodes for all tables
	fmt.Println("🌱 Creating root nodes...")
	if err := createRootNodes(database, tableManager, masterSeed); err != nil {
		fatalf("create root nodes: %v", err)
	}

	// Mark generation as completed
	if err := tables.MarkGenerationCompleted(database); err != nil {
		fatalf("mark generation completed: %v", err)
	}

	// Force DuckDB to checkpoint
	if err := database.Write("FORCE CHECKPOINT"); err != nil {
		fmt.Printf("⚠️  Could not checkpoint database: %v\n", err)
	}

	fmt.Println("✅ Database initialization complete!")
	fmt.Printf("📊 Created root nodes for %d tables\n", len(tableManager.GetTableNames()))
	fmt.Println("🚀 Ready for deterministic generation!")
}

func createRootNodes(db *db.DB, tableManager *tables.TableManager, masterSeed int64) error {
	// Generate root node ID
	rootID := uuid.New().String()
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

	// Generate root's child seed
	rootChildSeed := generateDeterministicSeed(masterSeed, rootID)

	// Insert root node into primary table
	primaryTableName := tableManager.GetPrimaryTableName()
	primaryQuery := fmt.Sprintf("INSERT INTO %s (id, parent_id, name, path, type, size, level, checked, secondary_existence_map, child_seed) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", primaryTableName)
	if err := db.Write(primaryQuery, rootID, "", "root", rootPath, "folder", 0, 0, false, existenceMapJSON, rootChildSeed); err != nil {
		return fmt.Errorf("insert root into primary table: %w", err)
	}
	fmt.Printf("🌱 Created root in primary table: %s\n", primaryTableName)

	// Insert root node into all secondary tables
	for _, tableName := range secondaryTableNames {
		secondaryQuery := fmt.Sprintf("INSERT INTO %s (id, parent_id, name, path, type, size, level, checked, child_seed) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", tableName)
		if err := db.Write(secondaryQuery, rootID, "", "root", rootPath, "folder", 0, 0, false, rootChildSeed); err != nil {
			return fmt.Errorf("insert root into secondary table %s: %w", tableName, err)
		}
		fmt.Printf("🌱 Created root in secondary table: %s\n", tableName)
	}

	return nil
}

// generateDeterministicSeed generates a deterministic seed based on master seed and folder ID
func generateDeterministicSeed(masterSeed int64, folderID string) int64 {
	// Create a hash of masterSeed + folderID to get deterministic but unique seed
	hasher := sha256.New()
	binary.Write(hasher, binary.LittleEndian, masterSeed)
	hasher.Write([]byte(folderID))
	hash := hasher.Sum(nil)

	// Convert first 8 bytes of hash to int64
	return int64(binary.LittleEndian.Uint64(hash[:8]))
}

func fatalf(f string, a ...any) {
	fmt.Printf("❌ "+f+"\n", a...)
	os.Exit(1)
}

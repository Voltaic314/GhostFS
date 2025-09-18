package tables

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Voltaic314/GhostFS/code/db"
	dbTypes "github.com/Voltaic314/GhostFS/code/types/db"
	"github.com/google/uuid"
)

// CachedNodeData holds both seed and existence map for a node
type CachedNodeData struct {
	ChildSeed    int64
	ExistenceMap SecondaryExistenceMap
}

// DeterministicGenerator generates filesystem nodes deterministically using seeds
type DeterministicGenerator struct {
	db               *db.DB
	config           PrimaryTableConfig
	secondaryConfigs map[string]SecondaryTableConfig
	nodeCache        map[string]CachedNodeData // folder_id -> (child_seed, existence_map) cache
	cacheMutex       sync.RWMutex
	masterSeed       int64
	tableManager     *TableManager
}

// NewDeterministicGenerator creates a new deterministic generator
func NewDeterministicGenerator(database *db.DB, config PrimaryTableConfig, secondaryConfigs map[string]SecondaryTableConfig, masterSeed int64, tableManager *TableManager) *DeterministicGenerator {
	return &DeterministicGenerator{
		db:               database,
		config:           config,
		secondaryConfigs: secondaryConfigs,
		nodeCache:        make(map[string]CachedNodeData),
		masterSeed:       masterSeed,
		tableManager:     tableManager,
	}
}

// LoadSeedsFromDatabase loads all existing ID/seed mappings and existence maps from the database into memory
func (dg *DeterministicGenerator) LoadSeedsFromDatabase(tableName string) error {
	// Only load existence maps from the primary table
	if tableName == dg.config.TableName {
		query := fmt.Sprintf("SELECT id, child_seed, secondary_existence_map FROM %s WHERE child_seed IS NOT NULL", tableName)
		rows, err := dg.db.Query(tableName, query)
		if err != nil {
			return fmt.Errorf("load seeds from database: %w", err)
		}
		defer rows.Close()

		dg.cacheMutex.Lock()
		defer dg.cacheMutex.Unlock()

		for rows.Next() {
			var id string
			var childSeed int64
			var existenceMapJSON string
			if err := rows.Scan(&id, &childSeed, &existenceMapJSON); err != nil {
				return fmt.Errorf("scan seed row: %w", err)
			}

			// Parse and cache the existence map
			existenceMap, err := FromJSON(existenceMapJSON)
			if err != nil {
				return fmt.Errorf("parse existence map for %s: %w", id, err)
			}

			// Store both seed and existence map in single cache entry
			dg.nodeCache[id] = CachedNodeData{
				ChildSeed:    childSeed,
				ExistenceMap: existenceMap,
			}
		}
	} else {
		// For secondary tables, only load the child_seed (no existence map)
		query := fmt.Sprintf("SELECT id, child_seed FROM %s WHERE child_seed IS NOT NULL", tableName)
		rows, err := dg.db.Query(tableName, query)
		if err != nil {
			return fmt.Errorf("load seeds from database: %w", err)
		}
		defer rows.Close()

		dg.cacheMutex.Lock()
		defer dg.cacheMutex.Unlock()

		for rows.Next() {
			var id string
			var childSeed int64
			if err := rows.Scan(&id, &childSeed); err != nil {
				return fmt.Errorf("scan seed row: %w", err)
			}

			// For secondary tables, we don't have existence maps, so we'll need to
			// get the existence info from the primary table when needed
			dg.nodeCache[id] = CachedNodeData{
				ChildSeed:    childSeed,
				ExistenceMap: make(SecondaryExistenceMap), // Empty for now
			}
		}
	}

	return nil
}

// GenerateChildren generates children for a folder deterministically
func (dg *DeterministicGenerator) GenerateChildren(folderID string, folderPath string, level int, foldersOnly bool, tableName string) ([]dbTypes.Node, error) {
	// Get or create child seed for this folder
	childSeed, err := dg.getOrCreateChildSeed(folderID, tableName)
	if err != nil {
		return nil, fmt.Errorf("get child seed for folder %s: %w", folderID, err)
	}

	// Create RNG with this folder's child seed
	rng := rand.New(rand.NewSource(childSeed))

	// Get parent's existence map from cache
	parentExistenceMap, err := dg.getOrCreateParentExistenceMap(folderID, tableName)
	if err != nil {
		return nil, fmt.Errorf("get parent existence map: %w", err)
	}

	// Generate children deterministically
	children := make([]dbTypes.Node, 0)

	// Generate folders
	numFolders := dg.config.MinChildFolders + rng.Intn(dg.config.MaxChildFolders-dg.config.MinChildFolders+1)
	for i := 0; i < numFolders; i++ {
		folderChild := dbTypes.Node{
			ID:        generateDeterministicUUID(childSeed, fmt.Sprintf("folder_%d", i)),
			ParentID:  folderID,
			Name:      fmt.Sprintf("folder_%d", i),
			Path:      buildPath(folderPath, fmt.Sprintf("folder_%d", i)),
			Type:      "folder",
			Size:      0,
			Level:     level + 1,
			Checked:   false,
			UpdatedAt: time.Now(),
			CreatedAt: time.Now(),
		}
		children = append(children, folderChild)
	}

	// Generate files (unless foldersOnly is true)
	if !foldersOnly {
		numFiles := dg.config.MinChildFiles + rng.Intn(dg.config.MaxChildFiles-dg.config.MinChildFiles+1)
		for i := 0; i < numFiles; i++ {
			fileSize := int64(100 + rng.Intn(900)) // Random size 100-999 bytes
			fileChild := dbTypes.Node{
				ID:        generateDeterministicUUID(childSeed, fmt.Sprintf("file_%d.txt", i)),
				ParentID:  folderID,
				Name:      fmt.Sprintf("file_%d.txt", i),
				Path:      buildPath(folderPath, fmt.Sprintf("file_%d.txt", i)),
				Type:      "file",
				Size:      fileSize,
				Level:     level + 1,
				Checked:   false,
				UpdatedAt: time.Now(),
				CreatedAt: time.Now(),
			}
			children = append(children, fileChild)
		}
	}

	// Store the children in the database with their own seeds and secondary table logic
	err = dg.storeChildrenWithSeeds(children, parentExistenceMap, tableName)
	if err != nil {
		return nil, fmt.Errorf("store children with seeds: %w", err)
	}

	return children, nil
}

// getOrCreateChildSeed gets a child seed from cache or database, or creates a new one
func (dg *DeterministicGenerator) getOrCreateChildSeed(folderID string, tableName string) (int64, error) {
	// Check cache first
	dg.cacheMutex.RLock()
	if nodeData, exists := dg.nodeCache[folderID]; exists {
		dg.cacheMutex.RUnlock()
		return nodeData.ChildSeed, nil
	}
	dg.cacheMutex.RUnlock()

	// Check database
	query := fmt.Sprintf("SELECT child_seed FROM %s WHERE id = ? AND child_seed IS NOT NULL LIMIT 1", tableName)
	var childSeed int64
	err := dg.db.QueryRow(query, folderID).Scan(&childSeed)
	if err == nil {
		// Found in database, need to get existence map too
		existenceMap, err := dg.getExistenceMapFromDB(folderID, tableName)
		if err != nil {
			return 0, fmt.Errorf("get existence map for cached seed: %w", err)
		}

		// Cache both
		dg.cacheMutex.Lock()
		dg.nodeCache[folderID] = CachedNodeData{
			ChildSeed:    childSeed,
			ExistenceMap: existenceMap,
		}
		dg.cacheMutex.Unlock()
		return childSeed, nil
	}

	// Generate new seed deterministically based on master seed + folder ID
	newSeed := generateDeterministicSeed(dg.masterSeed, folderID)

	// Generate existence map for this folder
	existenceMap := dg.determineSecondaryExistence(newSeed)

	// Store in database
	updateQuery := fmt.Sprintf("UPDATE %s SET child_seed = ? WHERE id = ?", tableName)
	dg.db.QueueWrite(tableName, updateQuery, newSeed, folderID)

	// Cache both seed and existence map
	dg.cacheMutex.Lock()
	dg.nodeCache[folderID] = CachedNodeData{
		ChildSeed:    newSeed,
		ExistenceMap: existenceMap,
	}
	dg.cacheMutex.Unlock()

	return newSeed, nil
}

// getExistenceMapFromDB gets the existence map from database
func (dg *DeterministicGenerator) getExistenceMapFromDB(folderID string, tableName string) (SecondaryExistenceMap, error) {
	query := fmt.Sprintf("SELECT secondary_existence_map FROM %s WHERE id = ? LIMIT 1", tableName)
	var existenceMapJSON string
	err := dg.db.QueryRow(query, folderID).Scan(&existenceMapJSON)
	if err != nil {
		return nil, fmt.Errorf("get existence map for %s: %w", folderID, err)
	}

	existenceMap, err := FromJSON(existenceMapJSON)
	if err != nil {
		return nil, fmt.Errorf("parse existence map for %s: %w", folderID, err)
	}

	return existenceMap, nil
}

// getOrCreateParentExistenceMap gets the parent's secondary table existence map from cache or creates it
func (dg *DeterministicGenerator) getOrCreateParentExistenceMap(folderID string, tableName string) (SecondaryExistenceMap, error) {
	// Check cache first
	dg.cacheMutex.RLock()
	if nodeData, exists := dg.nodeCache[folderID]; exists {
		dg.cacheMutex.RUnlock()
		return nodeData.ExistenceMap, nil
	}
	dg.cacheMutex.RUnlock()

	// For secondary tables, we need to get the existence map from the primary table
	if tableName != dg.config.TableName {
		primaryTableName := dg.config.TableName
		existenceMap, err := dg.getExistenceMapFromDB(folderID, primaryTableName)
		if err != nil {
			return nil, err
		}

		// Cache it (we don't have the seed, so we'll create a placeholder)
		dg.cacheMutex.Lock()
		dg.nodeCache[folderID] = CachedNodeData{
			ChildSeed:    0, // Placeholder - will be updated when seed is created
			ExistenceMap: existenceMap,
		}
		dg.cacheMutex.Unlock()

		return existenceMap, nil
	}

	// For primary table, get from database and cache it
	existenceMap, err := dg.getExistenceMapFromDB(folderID, tableName)
	if err != nil {
		return nil, err
	}

	// Cache it (we don't have the seed, so we'll create a placeholder)
	dg.cacheMutex.Lock()
	dg.nodeCache[folderID] = CachedNodeData{
		ChildSeed:    0, // Placeholder - will be updated when seed is created
		ExistenceMap: existenceMap,
	}
	dg.cacheMutex.Unlock()

	return existenceMap, nil
}

// storeChildrenWithSeeds stores children in the database with their seeds and secondary table logic
func (dg *DeterministicGenerator) storeChildrenWithSeeds(children []dbTypes.Node, parentExistenceMap SecondaryExistenceMap, tableName string) error {
	secondaryTableNames := dg.tableManager.GetSecondaryTableNames()

	for _, child := range children {
		// Generate child's own seed
		childSeed := generateDeterministicSeed(dg.masterSeed, child.ID)

		// Determine which secondary tables this child should exist in
		childExistenceMap := dg.determineSecondaryExistence(childSeed)

		// Check parent dependencies for secondary tables
		childExistenceMap = dg.checkParentDependencies(parentExistenceMap, childExistenceMap, secondaryTableNames)

		// Convert existence map to JSON
		existenceMapJSON, err := childExistenceMap.ToJSON()
		if err != nil {
			return fmt.Errorf("convert existence map to JSON for child %s: %w", child.ID, err)
		}

		// Insert child into primary table with seed
		primaryQuery := fmt.Sprintf("INSERT OR IGNORE INTO %s (id, parent_id, name, path, type, size, level, checked, secondary_existence_map, child_seed, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", tableName)
		dg.db.QueueWrite(tableName, primaryQuery, child.ID, child.ParentID, child.Name, child.Path, child.Type, child.Size, child.Level, child.Checked, existenceMapJSON, childSeed, child.CreatedAt, child.UpdatedAt)

		// Cache the child's existence map and seed
		dg.cacheMutex.Lock()
		dg.nodeCache[child.ID] = CachedNodeData{
			ChildSeed:    childSeed,
			ExistenceMap: childExistenceMap,
		}
		dg.cacheMutex.Unlock()

		// Insert into secondary tables where it should exist
		for _, secondaryTableName := range secondaryTableNames {
			if childExistenceMap[secondaryTableName] {
				secondaryQuery := fmt.Sprintf("INSERT OR IGNORE INTO %s (id, parent_id, name, path, type, size, level, checked, child_seed, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", secondaryTableName)
				dg.db.QueueWrite(secondaryTableName, secondaryQuery, child.ID, child.ParentID, child.Name, child.Path, child.Type, child.Size, child.Level, child.Checked, childSeed, child.CreatedAt, child.UpdatedAt)
			}
		}
	}

	return nil
}

// determineSecondaryExistence determines which secondary tables a node should exist in based on probability
func (dg *DeterministicGenerator) determineSecondaryExistence(childSeed int64) SecondaryExistenceMap {
	existenceMap := make(SecondaryExistenceMap)
	rng := rand.New(rand.NewSource(childSeed))

	for _, config := range dg.secondaryConfigs {
		// Roll the dice - if random float is less than dst_prob, include in this table
		roll := rng.Float64()
		existenceMap[config.TableName] = roll < config.DstProb
	}

	return existenceMap
}

// checkParentDependencies ensures that a child can only exist in secondary tables where its parent exists
func (dg *DeterministicGenerator) checkParentDependencies(parentExistenceMap, childExistenceMap SecondaryExistenceMap, secondaryTableNames []string) SecondaryExistenceMap {
	result := make(SecondaryExistenceMap)

	for _, tableName := range secondaryTableNames {
		// Child can only exist in secondary table if parent exists there
		parentExists := parentExistenceMap[tableName]
		childWantsToExist := childExistenceMap[tableName]

		result[tableName] = parentExists && childWantsToExist
	}

	return result
}

// GetFolderInfo gets folder information from database (for path, level, etc.)
func (dg *DeterministicGenerator) GetFolderInfo(folderID string, tableName string) (*dbTypes.Node, error) {
	query := fmt.Sprintf("SELECT id, parent_id, name, path, type, size, level, checked FROM %s WHERE id = ? LIMIT 1", tableName)

	var folder dbTypes.Node
	err := dg.db.QueryRow(query, folderID).Scan(
		&folder.ID, &folder.ParentID, &folder.Name, &folder.Path,
		&folder.Type, &folder.Size, &folder.Level, &folder.Checked)

	if err != nil {
		return nil, fmt.Errorf("get folder info for %s: %w", folderID, err)
	}

	return &folder, nil
}

// MarkFolderAccessed marks a folder as accessed (checked = true)
func (dg *DeterministicGenerator) MarkFolderAccessed(folderID string, tableName string) {
	// Queue async update to mark folder as checked
	updateQuery := fmt.Sprintf("UPDATE %s SET checked = TRUE WHERE id = ?", tableName)
	dg.db.QueueWrite(tableName, updateQuery, folderID)
}

// ClearCache clears the node cache (useful for testing or memory management)
func (dg *DeterministicGenerator) ClearCache() {
	dg.cacheMutex.Lock()
	dg.nodeCache = make(map[string]CachedNodeData)
	dg.cacheMutex.Unlock()
}

// GetCacheSize returns the current cache size (for monitoring)
func (dg *DeterministicGenerator) GetCacheSize() int {
	dg.cacheMutex.RLock()
	size := len(dg.nodeCache)
	dg.cacheMutex.RUnlock()
	return size
}

// GetCacheStats returns detailed cache statistics
func (dg *DeterministicGenerator) GetCacheStats() map[string]int {
	dg.cacheMutex.RLock()
	defer dg.cacheMutex.RUnlock()

	return map[string]int{
		"node_cache_size": len(dg.nodeCache),
	}
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

// generateDeterministicUUID generates a deterministic UUID based on seed and name
func generateDeterministicUUID(seed int64, name string) string {
	// Create a hash of seed + name for deterministic UUID generation
	hasher := sha256.New()
	binary.Write(hasher, binary.LittleEndian, seed)
	hasher.Write([]byte(name))
	hash := hasher.Sum(nil)

	// Create a UUID from the hash
	uuid, _ := uuid.FromBytes(hash[:16])
	return uuid.String()
}

// buildPath constructs the full path for a node based on its parent's path and name
func buildPath(parentPath, name string) string {
	if parentPath == "/" {
		return "/" + name
	}
	return parentPath + "/" + name
}

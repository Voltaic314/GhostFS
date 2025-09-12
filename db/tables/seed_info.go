package tables

import (
	"fmt"

	"github.com/Voltaic314/GhostFS/db"
)

// SeedInfoTable represents a table to store seed generation information
type SeedInfoTable struct{}

func (t *SeedInfoTable) Name() string {
	return "seed_info"
}

func (t *SeedInfoTable) Schema() string {
	return `
		seed_value BIGINT NOT NULL,
		target_depth INTEGER NOT NULL,
		generation_completed BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	`
}

// Init creates the seed_info table asynchronously.
func (t *SeedInfoTable) Init(db *db.DB) error {
	done := make(chan error)
	go func() {
		done <- db.CreateTable(t.Name(), t.Schema())
	}()
	return <-done
}

// SaveSeedInfo saves the seed information to the database
func SaveSeedInfo(db *db.DB, seedValue int64, targetDepth int) error {
	// Clear any existing seed info first (should only be one entry)
	clearQuery := `DELETE FROM seed_info`
	_, err := db.Exec(clearQuery)
	if err != nil {
		return fmt.Errorf("clear existing seed info: %w", err)
	}

	// Insert new seed info
	query := `INSERT INTO seed_info (seed_value, target_depth, generation_completed) VALUES (?, ?, ?)`
	_, err = db.Exec(query, seedValue, targetDepth, false)
	return err
}

// GetSeedInfo retrieves the seed information from the database
func GetSeedInfo(db *db.DB) (seedValue int64, targetDepth int, completed bool, err error) {
	query := `SELECT seed_value, target_depth, generation_completed FROM seed_info LIMIT 1`
	err = db.QueryRow(query).Scan(&seedValue, &targetDepth, &completed)
	return
}

// MarkGenerationCompleted marks the generation as completed
func MarkGenerationCompleted(db *db.DB) error {
	query := `UPDATE seed_info SET generation_completed = TRUE`
	_, err := db.Exec(query)
	return err
}

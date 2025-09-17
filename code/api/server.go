package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Voltaic314/GhostFS/code/api/routes"
	"github.com/Voltaic314/GhostFS/code/db"
	"github.com/Voltaic314/GhostFS/code/db/tables"
	"github.com/go-chi/chi/v5"
)

// GhostFSServer represents the GhostFS HTTP server
type GhostFSServer struct {
	router                 *chi.Mux
	db                     *db.DB
	config                 *tables.TestConfig
	tableManager           *tables.TableManager
	deterministicGenerator *tables.DeterministicGenerator
	server                 *http.Server
}

// NewGhostFSServer creates a new GhostFS server instance
func NewGhostFSServer(configPath string) (*GhostFSServer, error) {
	// Load configuration
	cfg, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	// Initialize DB
	dbPath, _ := filepath.Abs(cfg.Database.Path)
	database, err := db.NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("create db: %w", err)
	}

	// Create table manager
	tableManager := tables.NewTableManager(cfg)
	if err := tableManager.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid table config: %w", err)
	}

	// Initialize table IDs
	tableManager.InitializeTableIDs()

	// Get master seed from config or database
	masterSeed := cfg.Database.Tables.Primary.Seed
	if masterSeed == 0 {
		// Try to get seed from database
		if seedValue, _, _, err := tables.GetSeedInfo(database); err == nil {
			masterSeed = seedValue
		} else {
			// Use a default seed if none found
			masterSeed = 12345
		}
	}

	// Create deterministic generator
	generator := tables.NewDeterministicGenerator(
		database,
		cfg.Database.Tables.Primary,
		cfg.Database.Tables.Secondary,
		masterSeed,
		tableManager,
	)

	// Load existing seeds from all tables into memory
	tableNames := tableManager.GetTableNames()
	for _, tableName := range tableNames {
		if err := generator.LoadSeedsFromDatabase(tableName); err != nil {
			// Log warning but don't fail startup - this is for performance optimization
			fmt.Printf("⚠️  Warning: Could not load seeds from table %s: %v\n", tableName, err)
		}
	}

	// Create router
	router := chi.NewRouter()

	server := &GhostFSServer{
		router:                 router,
		db:                     database,
		config:                 cfg,
		tableManager:           tableManager,
		deterministicGenerator: generator,
	}

	// Setup routes with server instance
	routes.RegisterAllRoutes(router, server)

	return server, nil
}

// Start starts the GhostFS server
func (s *GhostFSServer) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Network.Address, s.config.Network.Port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	log.Printf("🚀 GhostFS server starting on %s", addr)
	return s.server.ListenAndServe()
}

// Stop gracefully stops the GhostFS server
func (s *GhostFSServer) Stop(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// GetTableManager returns the table manager
func (s *GhostFSServer) GetTableManager() *tables.TableManager {
	return s.tableManager
}

// GetDB returns the database instance
func (s *GhostFSServer) GetDB() *db.DB {
	return s.db
}

// GetDeterministicGenerator returns the deterministic generator instance
func (s *GhostFSServer) GetDeterministicGenerator() *tables.DeterministicGenerator {
	return s.deterministicGenerator
}

// loadConfig loads the GhostFS configuration
func loadConfig(path string) (*tables.TestConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg tables.TestConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func StartServer(configPath string) {
	// Create GhostFS server
	server, err := NewGhostFSServer(configPath)
	if err != nil {
		log.Fatalf("Failed to create GhostFS server: %v", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("🛑 Shutdown signal received, stopping server...")
		cancel()
	}()

	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Printf("Server error: %v", err)
			cancel()
		}
	}()

	// Wait for shutdown
	<-ctx.Done()

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Stop(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	} else {
		log.Println("✅ Server stopped gracefully")
	}
}

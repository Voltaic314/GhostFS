package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Voltaic314/GhostFS/api/routes"
	"github.com/Voltaic314/GhostFS/db"
	"github.com/Voltaic314/GhostFS/db/tables"
	"github.com/go-chi/chi/v5"
)

// GhostFSServer represents the GhostFS HTTP server
type GhostFSServer struct {
	router       *chi.Mux
	db           *db.DB
	config       *tables.TestConfig
	tableManager *tables.TableManager
	server       *http.Server
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

	// Create router
	router := chi.NewRouter()

	server := &GhostFSServer{
		router:       router,
		db:           database,
		config:       cfg,
		tableManager: tableManager,
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

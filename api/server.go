package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Voltaic314/GhostFS/db"
	"github.com/Voltaic314/GhostFS/db/tables"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// GhostFSServer represents the GhostFS HTTP server
type GhostFSServer struct {
	router       *chi.Mux
	db           *db.DB
	config       *tables.TestConfig
	tableManager *tables.TableManager
	batcher      *RequestBatcher
	server       *http.Server
}

// RequestBatcher handles batching of listChildren requests
type RequestBatcher struct {
	requests     chan ListChildrenRequest
	responses    map[string]chan ListChildrenResponse
	mu           sync.RWMutex
	db           *db.DB
	tableManager *tables.TableManager
	batchSize    int
	batchDelay   time.Duration
}

// ListChildrenRequest represents a request to list folder contents
type ListChildrenRequest struct {
	ParentPath string
	ResponseCh chan ListChildrenResponse
	RequestID  string
}

// ListChildrenResponse represents the response with folder contents
type ListChildrenResponse struct {
	Success bool     `json:"success"`
	Error   string   `json:"error,omitempty"`
	Items   []FSItem `json:"items,omitempty"`
}

// FSItem represents a filesystem item (file or folder)
type FSItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Path    string `json:"path"`
	Type    string `json:"type"` // "file" or "folder"
	Size    int64  `json:"size"`
	Level   int    `json:"level"`
	Checked bool   `json:"checked"`
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
	db, err := db.NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("create db: %w", err)
	}

	// Create table manager
	tableManager := tables.NewTableManager(cfg)
	if err := tableManager.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid table config: %w", err)
	}

	// Create request batcher
	batcher := NewRequestBatcher(db, tableManager, 10, 5*time.Millisecond)

	// Create router
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	server := &GhostFSServer{
		router:       router,
		db:           db,
		config:       cfg,
		tableManager: tableManager,
		batcher:      batcher,
	}

	// Setup routes
	server.setupRoutes()

	// Start batcher
	go batcher.Start()

	return server, nil
}

// setupRoutes configures all the GhostFS API routes
func (s *GhostFSServer) setupRoutes() {
	// Health check
	s.router.Get("/health", s.handleHealth)

	// Filesystem operations
	s.router.Get("/list/{path:.*}", s.handleListChildren)
	s.router.Post("/create-folder", s.handleCreateFolder)
	s.router.Post("/create-file", s.handleCreateFile)
	s.router.Get("/file/{fileID}/{filename}", s.handleGetFileContents)
	s.router.Get("/download/{fileID}/{filename}", s.handleDownloadFile)

	// Utility endpoints
	s.router.Get("/is-directory/{path:.*}", s.handleIsDirectory)
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

// NewRequestBatcher creates a new request batcher
func NewRequestBatcher(db *db.DB, tableManager *tables.TableManager, batchSize int, batchDelay time.Duration) *RequestBatcher {
	return &RequestBatcher{
		requests:     make(chan ListChildrenRequest, 1000),
		responses:    make(map[string]chan ListChildrenResponse),
		db:           db,
		tableManager: tableManager,
		batchSize:    batchSize,
		batchDelay:   batchDelay,
	}
}

// Start begins the batching process
func (rb *RequestBatcher) Start() {
	ticker := time.NewTicker(rb.batchDelay)
	defer ticker.Stop()

	var batch []ListChildrenRequest

	for {
		select {
		case req := <-rb.requests:
			batch = append(batch, req)
			if len(batch) >= rb.batchSize {
				rb.processBatch(batch)
				batch = batch[:0] // Reset slice
			}

		case <-ticker.C:
			if len(batch) > 0 {
				rb.processBatch(batch)
				batch = batch[:0] // Reset slice
			}
		}
	}
}

// processBatch processes a batch of listChildren requests
func (rb *RequestBatcher) processBatch(batch []ListChildrenRequest) {
	if len(batch) == 0 {
		return
	}

	// Collect parent paths
	parentPaths := make([]string, len(batch))
	for i, req := range batch {
		parentPaths[i] = req.ParentPath
	}

	// Query both tables for children
	children, err := rb.queryChildrenBatch(parentPaths)
	if err != nil {
		// Send error to all requests in batch
		for _, req := range batch {
			req.ResponseCh <- ListChildrenResponse{
				Success: false,
				Error:   err.Error(),
			}
		}
		return
	}

	// Group children by parent path and send responses
	childrenByPath := make(map[string][]FSItem)
	for _, child := range children {
		childrenByPath[child.Path] = append(childrenByPath[child.Path], child)
	}

	for _, req := range batch {
		items := childrenByPath[req.ParentPath]
		req.ResponseCh <- ListChildrenResponse{
			Success: true,
			Items:   items,
		}
	}
}

// queryChildrenBatch queries all tables for children
func (rb *RequestBatcher) queryChildrenBatch(parentPaths []string) ([]FSItem, error) {
	// Create placeholders for IN clause
	placeholders := make([]string, len(parentPaths))
	args := make([]interface{}, len(parentPaths))
	for i, path := range parentPaths {
		placeholders[i] = "?"
		args[i] = path
	}

	// Query all tables with UNION ALL
	query := fmt.Sprintf(`
		SELECT id, name, path, type, size, level, checked 
		FROM {{TABLE}} 
		WHERE parent_id IN (%s)
	`, placeholders)

	// Build UNION query for all tables
	unionQuery := rb.tableManager.BuildUnionQuery(query)

	rows, err := rb.db.Query(rb.tableManager.GetPrimaryTableName(), unionQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query children batch: %w", err)
	}
	defer rows.Close()

	var children []FSItem
	for rows.Next() {
		var item FSItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Path, &item.Type, &item.Size, &item.Level, &item.Checked); err != nil {
			return nil, fmt.Errorf("scan child row: %w", err)
		}
		children = append(children, item)
	}

	return children, nil
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

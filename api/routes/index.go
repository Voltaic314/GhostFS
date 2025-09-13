package routes

import (
	"time"

	"github.com/Voltaic314/GhostFS/api/routes/items/files"
	"github.com/Voltaic314/GhostFS/api/routes/items/folders"
	"github.com/Voltaic314/GhostFS/api/routes/tables"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// RegisterAllRoutes registers all API routes
func RegisterAllRoutes(r chi.Router, server interface{}) {
	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Register route groups with server instance
	r.Route("/tables", func(r chi.Router) {
		tables.RegisterRoutes(r, server)
	})
	r.Route("/folders", func(r chi.Router) {
		folders.RegisterRoutes(r, server)
	})
	r.Route("/files", func(r chi.Router) {
		files.RegisterRoutes(r, server)
	})
}

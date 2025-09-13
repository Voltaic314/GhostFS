package routes

import (
	"time"

	"github.com/Voltaic314/GhostFS/api/routes/items"
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
	r.Route("/items", func(r chi.Router) {
		items.RegisterRoutes(r, server)
	})
}

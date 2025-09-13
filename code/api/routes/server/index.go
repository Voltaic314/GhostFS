package server

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all server-related routes
func RegisterRoutes(r chi.Router) {
	// Health check
	r.Get("/health", HandleHealth)
	r.Post("/register", HandleRegister)
}

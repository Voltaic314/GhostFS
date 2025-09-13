package tables

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all table-related routes
func RegisterRoutes(r chi.Router, server interface{}) {
	// Table management
	r.Post("/list", func(w http.ResponseWriter, r *http.Request) {
		HandleListTables(w, r, server)
	})
}

package folders

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all folder-related routes
func RegisterRoutes(r chi.Router, server interface{}) {
	// Folder operations
	r.Post("/list", func(w http.ResponseWriter, r *http.Request) {
		HandleList(w, r, server)
	})
	r.Post("/new", func(w http.ResponseWriter, r *http.Request) {
		HandleNew(w, r, server)
	})
	r.Post("/delete", func(w http.ResponseWriter, r *http.Request) {
		HandleDelete(w, r, server)
	})
}

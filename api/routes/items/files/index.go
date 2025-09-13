package files

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all file-related routes
func RegisterRoutes(r chi.Router, server interface{}) {
	// File operations
	r.Post("/list", func(w http.ResponseWriter, r *http.Request) {
		HandleList(w, r, server)
	})
	r.Post("/new", func(w http.ResponseWriter, r *http.Request) {
		HandleNew(w, r, server)
	})
	r.Post("/delete", func(w http.ResponseWriter, r *http.Request) {
		HandleDelete(w, r, server)
	})
	r.Post("/download", func(w http.ResponseWriter, r *http.Request) {
		HandleDownload(w, r, server)
	})
}

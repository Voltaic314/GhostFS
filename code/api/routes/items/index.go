package items

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all item-related routes (files and folders)
func RegisterRoutes(r chi.Router, server interface{}) {
	// Item operations (works on files and folders)
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
	r.Get("/get_root", func(w http.ResponseWriter, r *http.Request) {
		HandleGetRoot(w, r, server)
	})
}

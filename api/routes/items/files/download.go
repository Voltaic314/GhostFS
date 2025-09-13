package files

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// DownloadRequest represents a request to download a file
type DownloadRequest struct {
	TableID  string `json:"table_id"`
	FileID   string `json:"file_id"`
	Filename string `json:"filename"`
}

// DownloadResponse represents the response from file download
type DownloadResponse struct {
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
	TableID     string `json:"table_id,omitempty"`
	TableName   string `json:"table_name,omitempty"`
	FileID      string `json:"file_id,omitempty"`
	Filename    string `json:"filename,omitempty"`
	DownloadURL string `json:"download_url,omitempty"`
}

// HandleDownload handles requests to download a file
func HandleDownload(w http.ResponseWriter, r *http.Request, server interface{}) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual file download logic
	// For now, return a placeholder response
	response := DownloadResponse{
		Success:     true,
		TableID:     req.TableID,
		TableName:   "placeholder-table-name",
		FileID:      req.FileID,
		Filename:    req.Filename,
		DownloadURL: "http://localhost:8086/download",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleDownloadFile handles actual file downloads (returns random binary data)
func HandleDownloadFile(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual file download logic
	// For now, generate random binary data
	fileSize := int64(1024) // Default size

	// Set headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", req.Filename))
	w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))

	// Generate random binary data
	randomData := make([]byte, fileSize)
	for i := range randomData {
		randomData[i] = byte(i % 256)
	}

	w.Write(randomData)
}

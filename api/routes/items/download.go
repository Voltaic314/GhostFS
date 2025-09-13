package items

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// DownloadRequest represents a request to download one or more files
type DownloadRequest struct {
	TableID string   `json:"table_id"`
	FileIDs []string `json:"file_ids"` // Array of file IDs to download
}

// DownloadItemResponse represents download info for a single file
type DownloadItemResponse struct {
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
	FileID      string `json:"file_id"`
	Filename    string `json:"filename,omitempty"`
	DownloadURL string `json:"download_url,omitempty"`
	Size        int64  `json:"size,omitempty"`
}

// DownloadResponse represents the response with download URLs
type DownloadResponse struct {
	Success bool                   `json:"success"`
	Error   string                 `json:"error,omitempty"`
	TableID string                 `json:"table_id,omitempty"`
	Files   []DownloadItemResponse `json:"files,omitempty"`
}

// HandleDownload handles requests to get download URLs for one or more files
func HandleDownload(w http.ResponseWriter, r *http.Request, server interface{}) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual download logic using server
	// Loop through req.FileIDs and generate download URLs for each
	// Only works for files, not folders

	// For now, return placeholder responses
	var downloadFiles []DownloadItemResponse
	for i, fileID := range req.FileIDs {
		downloadFiles = append(downloadFiles, DownloadItemResponse{
			Success:     true,
			FileID:      fileID,
			Filename:    fmt.Sprintf("file_%d.txt", i+1),
			DownloadURL: fmt.Sprintf("http://localhost:8086/download/%s", fileID),
			Size:        1024,
		})
	}

	response := DownloadResponse{
		Success: true,
		TableID: req.TableID,
		Files:   downloadFiles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleDownloadFile handles actual file downloads (returns file data)
func HandleDownloadFile(w http.ResponseWriter, r *http.Request) {
	// TODO: Get file ID from URL path and serve actual file data
	// For now, generate random binary data
	fileSize := int64(1024)
	filename := "example.txt"

	// Set headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))

	// Generate random binary data
	randomData := make([]byte, fileSize)
	for i := range randomData {
		randomData[i] = byte(i % 256)
	}

	w.Write(randomData)
}

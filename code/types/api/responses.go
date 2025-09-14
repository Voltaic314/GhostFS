package api

import (
	"encoding/json"
	"net/http"
)

// BaseResponse represents the standard response structure for all API endpoints
type BaseResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// Response represents a complete API response with optional data
type Response struct {
	BaseResponse
	Data interface{} `json:"data,omitempty"`
}

// NewSuccessResponse creates a successful response with optional data
func NewSuccessResponse(data interface{}) Response {
	return Response{
		BaseResponse: BaseResponse{Success: true},
		Data:         data,
	}
}

// NewErrorResponse creates an error response with a message
func NewErrorResponse(errorMsg string) Response {
	return Response{
		BaseResponse: BaseResponse{
			Success: false,
			Error:   errorMsg,
		},
	}
}

// SendJSON writes a JSON response to the HTTP response writer
func (r Response) SendJSON(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(r)
}

// SendSuccess sends a successful response with 200 status
func (r Response) SendSuccess(w http.ResponseWriter) {
	r.SendJSON(w, http.StatusOK)
}

// SendError sends an error response with appropriate status code
func (r Response) SendError(w http.ResponseWriter, statusCode int) {
	r.SendJSON(w, statusCode)
}

// Helper functions for common response patterns

// Success sends a successful response with data
func Success(w http.ResponseWriter, data interface{}) {
	NewSuccessResponse(data).SendSuccess(w)
}

// SuccessEmpty sends a successful response with no data
func SuccessEmpty(w http.ResponseWriter) {
	NewSuccessResponse(nil).SendSuccess(w)
}

// BadRequest sends a 400 error response
func BadRequest(w http.ResponseWriter, message string) {
	NewErrorResponse(message).SendError(w, http.StatusBadRequest)
}

// NotFound sends a 404 error response
func NotFound(w http.ResponseWriter, message string) {
	NewErrorResponse(message).SendError(w, http.StatusNotFound)
}

// InternalError sends a 500 error response
func InternalError(w http.ResponseWriter, message string) {
	NewErrorResponse(message).SendError(w, http.StatusInternalServerError)
}

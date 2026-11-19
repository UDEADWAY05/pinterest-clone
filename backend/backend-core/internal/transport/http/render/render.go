package render

import (
	"encoding/json"
	"log"
	"net/http"
)

// JSONResponse is the standard structure for successful API responses.
type JSONResponse interface{}

// ErrorResponse is the standard structure for API error responses.
type ErrorResponse struct {
	Error ErrorDetails `json:"error"`
}

// ErrorDetails contains detailed information about an API error.
type ErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// JSON sends a successful JSON response with the given data and status code.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("render.JSON: failed to write JSON response: %v", err)
		}
	}
}

// Error sends a standardized JSON error response with the given details and status code.
func Error(w http.ResponseWriter, status int, code, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	errResp := ErrorResponse{
		Error: ErrorDetails{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		log.Printf("render.Error: failed to write error JSON response: %v", err)
	}
}

// render.Error(w, "message", http.StatusXXX).
func Message(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"message": message})
}

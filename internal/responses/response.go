package responses

import (
	"encoding/json"
	"log"
	"net/http"
)

// Response represents the standard JSON response structure
type Response struct {
	Success bool              `json:"success"`
	Message string            `json:"message,omitempty"`
	Data    interface{}       `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// JSONResponse sends a standard JSON response
func JSONResponse(w http.ResponseWriter, statusCode int, response *Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// SuccessResponse sends a successful JSON response
func SuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	response := &Response{
		Success: true,
		Message: message,
		Data:    data,
	}
	JSONResponse(w, statusCode, response)
}

func ServeFileResponse(w http.ResponseWriter, r *http.Request, filePath string) {
	http.ServeFile(w, r, filePath)
}

// ErrorResponse sends an error JSON response
func ErrorResponse(w http.ResponseWriter, statusCode int, message string, err string) {
	log.Println(err)
	response := &Response{
		Success: false,
		Message: message,
		Error:   err,
	}
	JSONResponse(w, statusCode, response)
}

// ValidationResponse sends a validation error JSON response
func ValidationResponse(w http.ResponseWriter, message string, fields map[string]string) {
	response := &Response{
		Success: false,
		Message: message,
		Fields:  fields,
	}
	JSONResponse(w, http.StatusUnprocessableEntity, response)
}

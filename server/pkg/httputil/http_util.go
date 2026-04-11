package httputil

import (
	"encoding/json"
	"net/http"
)

// WriteJSON writes the given data as JSON with the provided status code.
func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// ReadJSON decodes the request body into the provided destination with sane limits.
func ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	const maxBytes = 1_048_576 // 1 MB

	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

// WriteJSONError sends a structured error response.
func WriteJSONError(w http.ResponseWriter, status int, message string) error {
	type envelope struct {
		Error string `json:"error"`
	}
	return WriteJSON(w, status, &envelope{Error: message})
}

type Response struct {
	Success   bool       `json:"success"`
	Message   string     `json:"message,omitempty"`
	Data      any        `json:"data,omitempty"`
	Error     *ErrorBody `json:"error,omitempty"`
	RequestID string     `json:"request_id,omitempty"`
	Meta      any        `json:"meta,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Details any    `json:"details,omitempty"`
}

// JSONResponse wraps the provided data in a "data" envelope for consistency.
func JSONResponse(w http.ResponseWriter, status int, data any, message string) error {
	resp := Response{
		Success: status >= 200 && status < 300,
		Message: message,
		Data:    data,
	}

	return WriteJSON(w, status, resp)
}

package apierror

import (
	"errors"
	"net/http"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/httputil"
)

// Define Some Custom Errors for more readable Error Handling
var (
	ErrNotFound     = errors.New("resource is not found")
	ErrConflict     = errors.New("resource already exists")
	ErrBadRequest   = errors.New("bad request")
	ErrUnauthorized = errors.New("Not Authorized")
)

// Error represents a typed API error response.
type Error struct {
	Message string `json:"error"`
}

// Logger is the minimal logging interface used by apierror helpers.
type Logger interface {
	Errorw(msg string, keysAndValues ...any)
	Warnw(msg string, keysAndValues ...any)
}

// Write sends a JSON error envelope using the shared httputil helpers.
func Write(w http.ResponseWriter, status int, message string) error {
	return httputil.WriteJSON(w, status, Error{Message: message})
}

func InternalServerError(logger Logger, w http.ResponseWriter, r *http.Request, err error) {
	logger.Errorw("Internal Server Error", "method", r.Method, "path", r.URL.Path, "error", err)
	_ = Write(w, http.StatusInternalServerError, "the server encountered a problem")
}

func Forbidden(logger Logger, w http.ResponseWriter, r *http.Request, err error) {
	logger.Warnw("Forbidden Response Error", "method", r.Method, "path", r.URL.Path, "error", err)
	_ = Write(w, http.StatusForbidden, "forbidden")
}

func BadRequest(logger Logger, w http.ResponseWriter, r *http.Request, err error) {
	logger.Warnw("Bad Request Response", "method", r.Method, "path", r.URL.Path, "error", err)
	_ = Write(w, http.StatusBadRequest, err.Error())
}

func Conflict(logger Logger, w http.ResponseWriter, r *http.Request, err error) {
	logger.Errorw("Conflict Response", "method", r.Method, "path", r.URL.Path, "error", err)
	_ = Write(w, http.StatusConflict, err.Error())
}

func NotFound(logger Logger, w http.ResponseWriter, r *http.Request, err error) {
	logger.Errorw("Not Found Response", "method", r.Method, "path", r.URL.Path, "error", err)
	_ = Write(w, http.StatusNotFound, "not found")
}

func Unauthorized(logger Logger, w http.ResponseWriter, r *http.Request, err error) {
	logger.Warnw("Unauthorized Response", "method", r.Method, "path", r.URL.Path, "error", err)
	_ = Write(w, http.StatusUnauthorized, "unauthorized")
}

func RateLimitExceeded(logger Logger, w http.ResponseWriter, r *http.Request, retryAfter string) {
	logger.Warnw("rate limit exceeded", "method", r.Method, "path", r.URL.Path)
	w.Header().Set("Retry-After", retryAfter)
	_ = Write(w, http.StatusTooManyRequests, "rate limit exceeded, retry after: "+retryAfter)
}

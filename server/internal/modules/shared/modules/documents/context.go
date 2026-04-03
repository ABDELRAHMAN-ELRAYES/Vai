package sharedDocs

import (
	"context"
	"net/http"

)

type uploadCtxKeyType struct{}
var UploadedFileCtxKey uploadCtxKeyType

// Set the uploaded file in the request context
func SetUploadedFile(r *http.Request, file *UploadedFile) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), UploadedFileCtxKey, file))
}

// Get the uploaded file from the request context
func GetUploadedFile(r *http.Request) (*UploadedFile, bool) {
	file, ok := r.Context().Value(UploadedFileCtxKey).(*UploadedFile)
	return file, ok
}

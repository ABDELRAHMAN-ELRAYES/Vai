package apierror

import "errors"

var (
	ErrFileTooLarge       = errors.New("file too large")
	ErrFailedToSaveFile   = errors.New("failed to save file")
	ErrFailedToCreateFile = errors.New("failed to create file")
	ErrInvalidFilePart    = errors.New("failed to add the part of the file provided")
	ErrNoFileProvided     = errors.New("No file provided")
)

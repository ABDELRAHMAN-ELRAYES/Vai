package middleware

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	sharedDocs "github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/shared/modules/documents"
	apierror "github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/errors"
	"github.com/google/uuid"
)

func FileUploadMiddleware(app *app.Application) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := app.Logger

			// Limit request size (100MB)
			r.Body = http.MaxBytesReader(w, r.Body, 100<<20)

			reader, err := r.MultipartReader()
			if err != nil {
				apierror.BadRequest(logger, w, r, apierror.ErrNoFileProvided)
				return
			}

			var result *sharedDocs.UploadedFile
			for {
				part, err := reader.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					apierror.InternalServerError(logger, w, r, err)
					return
				}

				// Ignore other fields
				if part.FormName() != "file" {
					continue
				}

				result, err = saveFile(part, app.Config.Upload.Dir)
				if err != nil {
					apierror.InternalServerError(logger, w, r, err)
					return
				}

				break
			}

			if result == nil {
				apierror.BadRequest(logger, w, r, err)
				return
			}

			// Inject file into context and pass to next handler
			next.ServeHTTP(w, sharedDocs.SetUploadedFile(r, result))
		})
	}
}

// Save the file parts to its related file in a specific path
func saveFile(part *multipart.Part, uploadDir string) (*sharedDocs.UploadedFile, error) {
	if part.FileName() == "" {
		return nil, apierror.ErrInvalidFilePart
	}
	// Form the full file path
	uploadedFileName := filepath.Base(part.FileName())
	ext := filepath.Ext(uploadedFileName)
	fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	path := filepath.Join(uploadDir, fileName)

	// Create the file
	dst, err := os.Create(path)
	if err != nil {
		return nil, apierror.ErrFailedToCreateFile
	}
	defer dst.Close()

	// Add the part to its file
	size, err := io.Copy(dst, part)
	if err != nil {
		return nil, apierror.ErrFailedToSaveFile
	}

	return &sharedDocs.UploadedFile{
		FileName: fileName,
		Size:     size,
	}, nil
}

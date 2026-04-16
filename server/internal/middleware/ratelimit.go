package middleware

import (
	"net/http"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	apierror "github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/errors"
)

func RateLimiter(app *app.Application) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if app.Config.RateLimiter.Enabled {
				if allow, retryAfter := app.RateLimiter.Allow(r.RemoteAddr); !allow {
					apierror.RateLimitExceeded(app.Logger, w, r, retryAfter.String())
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}


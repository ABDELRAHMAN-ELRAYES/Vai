package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/auth"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/shared"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/apierror"
)

type UserClaims struct {
	UserID string `json:"UserID"`
	jwt.RegisteredClaims
}

type UserFetcher func(ctx context.Context, id string) (any, error)

// Protect check if the user is authenticated and attach its data to the request context
func Protect(app *app.Application, fetcher UserFetcher) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := auth.GetTokenFromCookie(r)
			if err != nil {
				apierror.Unauthorized(app.Logger, w, r, fmt.Errorf("authorization cookie is missing"))
				return
			}

			claims := &UserClaims{}
			_, err = app.Authenticator.ValidateToken(token, claims)
			if err != nil {
				apierror.Unauthorized(app.Logger, w, r, err)
				return
			}

			ctx := r.Context()

			user, err := fetcher(ctx, claims.UserID)
			if err != nil {
				apierror.Unauthorized(app.Logger, w, r, err)
				return
			}

			ctx = context.WithValue(ctx, shared.UserCtxKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

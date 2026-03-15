package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/auth"
	authModule "github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/auth"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/users"

	"github.com/google/uuid"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/apierror"
)

type userCtxKeyType struct{}

var userCtxKey userCtxKeyType

// Protect check if the user is authenticated and attach its data to the request context
func Protect(app *app.Application) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := auth.GetTokenFromCookie(r)
			if err != nil {
				apierror.Unauthorized(app.Logger, w, r, fmt.Errorf("authorization cookie is missing"))
				return
			}

			claims := &authModule.UserClaims{}
			_, err = app.Authenticator.ValidateToken(token, claims)
			if err != nil {
				apierror.Unauthorized(app.Logger, w, r, err)
				return
			}

			id, err := uuid.Parse(claims.UserID)
			if err != nil {
				apierror.Unauthorized(app.Logger, w, r, err)
				return
			}
			ctx := r.Context()

			// TODO: must be udpated not to create the repo here
			userRepo := users.NewRepository(app.DB)

			user, err := userRepo.GetByID(ctx, id)

			if err != nil {
				apierror.Unauthorized(app.Logger, w, r, err)
				return
			}
			userTokenResp := &authModule.UserWithTokenResponse{
				User:  user.ToResponse(),
				Token: token,
			}

			ctx = context.WithValue(ctx, userCtxKey, userTokenResp)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

package auth

import "net/http"

var (
	AuthTokenCookieKey = "access_token"
	AuthTokenCookieExp = 90 * 24 * 60 * 60
)

// Set a value in a browser cookie
func SetCookie(w http.ResponseWriter, key, value string, expiration int) {
	cookie := &http.Cookie{
		Name:     key,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   expiration,
	}

	http.SetCookie(w, cookie)
}

// Get the access token from the cookie
func GetTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(AuthTokenCookieKey)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// BearerMiddleware checks if api key exists.
func (s *AuthenticateHandler) BearerMiddleware(ctx *gin.Context) {

	// possible states:
	// No Header,
	// Invalid header format,
	// Yes header + valid key,
	// Yes header + invalid key

	authHeader := ctx.GetHeader("Authorization")

	// no header -> continue without auth (guest access)
	if authHeader == "" {
		ctx.Next()
		return
	}

	// parse header
	key := strings.Split(authHeader, " ") // Bearer aabbccddeeff

	// invalid header, abort.
	if len(key) == 1 {
		ctx.AbortWithError(http.StatusUnauthorized, errors.New("invalid Authorization header"))
		return
	}

	// get cache

	val, exists := s.Cache.Get("keys")

	// continue only if keys array exists
	if exists == true {

		for i := range val.([]string) {

			// valid key
			if val.([]string)[i] == key[1] {
				ctx.Set("Authenticated", true)
				ctx.Next()
				return
			}

		}

	}

	// invalid key
	ctx.AbortWithError(http.StatusUnauthorized, errors.New("invalid Authorization header"))
	return
}

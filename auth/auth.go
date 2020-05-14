package auth

import (
	"github.com/patrickmn/go-cache"
)

const otpSecretPath = ".otp"

// AuthenticateHandler is a helper struct for all page handlers.
type AuthenticateHandler struct {
	Issuer string
	Cache  *cache.Cache
}

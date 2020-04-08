package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

// BearerMiddleware -
func BearerMiddleware(ctx *gin.Context) {

	key := strings.Split(ctx.GetHeader("Authorization"), " ") // Bearer aabbccddeeff

	keycache := ctx.MustGet("keycache").(*cache.Cache)

	val, exists := keycache.Get("keys")

	if exists != true {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	var found = false

	for i := range val.([]string) {
		if val.([]string)[i] == key[1] {
			found = true
		}
	}

	if found == true {
		ctx.Set("Authenticated", true)
	}

	ctx.Next()
	return
}

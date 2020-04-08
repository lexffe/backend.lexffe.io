package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CheckAuthentication guards admin routes.
func CheckAuthentication(ctx *gin.Context) {
	if ctx.MustGet("Authentication").(bool) == true {
		ctx.Next()
		return
	}
	ctx.Status(http.StatusUnauthorized)
	return
}
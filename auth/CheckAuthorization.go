package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CheckAuthentication guards admin routes.
func CheckAuthentication(ctx *gin.Context) {
	if ctx.MustGet("Authorized").(bool) == true {
		ctx.Next()
		return
	}
	ctx.AbortWithStatus(http.StatusUnauthorized)
	return
}

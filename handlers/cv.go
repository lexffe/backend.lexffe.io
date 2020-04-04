package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/lexffe/backend.lexffe.io/auth"
)

func RegisterCVRoutes(r *gin.RouterGroup) {

	authRoutes := r.Group("/", auth.BearerMiddleware)
}

package handlers

import (
	"context"
	"log"
	"time"

	"github.com/lexffe/backend.lexffe.io/helpers"
	"github.com/lexffe/backend.lexffe.io/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// RegisterHLRoutes registers the router with all highlights-related subroutes.
func RegisterHLRoutes(r *gin.RouterGroup) {
	
	r.GET("/highlights", getHighlightsHandler)
	r.PUT("/highlights", updateHighlightsHandler) // need middleware
	r.DELETE("/highlights", clearHighlightsHandler) // need middleware

}

// Note: max highlights == 5

func getHighlightsHandler(c *gin.Context) {
	db := c.MustGet("db").(mongo.Database)
	coll := db.Collection("highlights")

	cur, err :=	coll.Find(context.Background(), bson.D{})

	if err != nil {
		
	}

}

func updateHighlightsHandler(c *gin.Context) {

}

func clearHighlightsHandler(c *gin.Context) {
	// Clear highlights, populate highlights with blog posts.

}

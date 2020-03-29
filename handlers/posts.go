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

// TODO: write out all db queries for each action

// RegisterPostRoutes registers the router with all subroutes.
func RegisterPostRoutes(r *gin.RouterGroup) {
	
	r.GET("/post", getPostsHandler)
	r.GET("/post/:id", getPostHandler)

	r.POST("/post", createPostHandler) // need middleware
	r.PUT("/post/:id", updatePostHandler) // need middleware
	r.DELETE("/post/:id", deletePostHandler) // need middleware

}

func getPostsHandler(c *gin.Context) {

	db := c.MustGet("db").(*mongo.Client)
	coll := db.Database("site").Collection("post")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	cursor, err := coll.Find(ctx, bson.D{})

	if err != nil {
		log.Fatal(err)
	}

	var results []models.Post

	// for all the next ones
	for cursor.Next(context.Background()) {
		var result models.Post
		err := cursor.Decode(result)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, result)
	}

	c.JSON(200, gin.H{
		"path": c.FullPath(),
	})
}

func getPostHandler(c *gin.Context) {

	// id := c.Param("id")

	c.JSON(200, gin.H{
		"path": c.FullPath(),
	})
}

func createPostHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"path": c.FullPath(),
	})
}

func updatePostHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"path": c.FullPath(),
	})
}

func deletePostHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"path": c.FullPath(),
	})
}

// ---

func newPost(title string, subtitle string, tags []string, markdown string) *models.Post {

	return &models.Post{
		Title: title,
		Subtitle: subtitle,
		Tags: tags,
		Markdown: markdown,
		HTML: helpers.ParseMD(markdown),
	}

}

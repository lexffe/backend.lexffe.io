package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	// "github.com/lexffe/backend.lexffe.io/auth"
	"github.com/lexffe/backend.lexffe.io/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RegisterProjectRoutes registers the router with all project related subroutes.
func RegisterProjectRoutes(r *gin.RouterGroup) {

	r.GET("/", getProjectsHandler)
	r.GET("/:id", getProjectHandler)

	// authRoutes := r.Group("/", auth.BearerMiddleware)

	r.POST("/", createProjectHandler)       // need auth middleware
	r.PUT("/:id", updateProjectHandler)    // need auth middleware
	r.DELETE("/:id", deleteProjectHandler) // need auth middleware

}

func getProjectsHandler(ctx *gin.Context) {
	
	// user-defined skip, for pagination.
	skipParam := ctx.DefaultQuery("skip", "0")
	skip, err := strconv.Atoi(skipParam)

	// Bad request
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		ctx.Error(errors.New("skip is not a number"))
		ctx.Error(err)
		return
	}
	
	db := ctx.MustGet("db").(*mongo.Database)
	coll := db.Collection(collectionProjects)

	filter := bson.M{
		"published": true,
	}

	opts := options.Find().SetLimit(paginationLimit).SetSkip(int64(skip))

	cur, err := coll.Find(ctx, filter, opts)

	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}

	var projects []models.Project

	for cur.Next(ctx) {
		var project models.Project
		
		if err := cur.Decode(&project) ; err != nil {
			ctx.Status(http.StatusInternalServerError)
			ctx.Error(err)
			return
		}

		projects = append(projects, project)
	}

	ctx.JSON(http.StatusOK, projects)
}

func getProjectHandler(ctx *gin.Context) {
	docID := ctx.Param("id")

	if docID == "" {
		ctx.Status(http.StatusBadRequest)
		ctx.Error(errors.New("no document identifier provided"))
		return
	}

	db := ctx.MustGet("db").(*mongo.Database)
	coll := db.Collection(collectionPosts)

	filter := bson.M{
		"_id": docID,
	}

	opts := options.FindOne()

	res := coll.FindOne(ctx, filter, opts)

	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			ctx.Status(http.StatusNotFound)
			ctx.Error(errors.New("document not found"))
		} else {
			ctx.Status(http.StatusInternalServerError)
		}
		ctx.Error(res.Err())
		return
	}

	var project models.Project

	if err := res.Decode(&project) ; err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, project)
}

func createProjectHandler(ctx *gin.Context) {

	var body models.Project

	if err := ctx.BindJSON(&body); err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}

	db := ctx.MustGet("db").(*mongo.Database)
	coll := db.Collection(collectionProjects)

	_, err := coll.InsertOne(ctx, body)
	
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusCreated)
}

func updateProjectHandler(ctx *gin.Context) {

	docID := ctx.Param("id")

	if docID == "" {
		ctx.Status(http.StatusBadRequest)
		ctx.Error(errors.New("no document identifier provided"))
		return
	}

	var body models.Project

	if err := ctx.BindJSON(&body); err != nil {
		ctx.Error(err)
		return
	}

	if body.ID.String() != docID {
		ctx.Status(http.StatusBadRequest)
		ctx.Error(errors.New("document identifier is different than id in path"))
		return
	}

	filter := bson.M{
		"_id":              body.ID,
	}

	db := ctx.MustGet("db").(*mongo.Database)
	res, err := db.Collection(collectionCustomPages).UpdateOne(ctx, filter, body)

	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(errors.New("cannot update document"))
		ctx.Error(err)
		return
	}

	if res.MatchedCount > 1 {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(errors.New("more than one document updated"))
		return
	}

	ctx.Status(http.StatusNoContent)

}

func deleteProjectHandler(ctx *gin.Context) {
	docID := ctx.Param("id") // ObjectID string

	if docID == "" {
		ctx.Status(http.StatusBadRequest)
		ctx.Error(errors.New("no document identifier provided"))
		return
	}

	filter := bson.M{
		"_id": docID,
	}

	db := ctx.MustGet("db").(*mongo.Database)
	res := db.Collection(collectionCustomPages).FindOneAndDelete(ctx, filter)

	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			ctx.Status(http.StatusNotFound)
		} else {
			ctx.Status(http.StatusInternalServerError)
		}
		ctx.Error(res.Err())
		return
	}

	ctx.Status(http.StatusNoContent)
}

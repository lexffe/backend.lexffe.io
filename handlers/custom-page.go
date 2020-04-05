package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	// "github.com/lexffe/backend.lexffe.io/auth"
	"github.com/lexffe/backend.lexffe.io/helpers"
	"github.com/lexffe/backend.lexffe.io/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RegisterCustomPageRoutes registers the router with all custom page related subroutes.
func RegisterCustomPageRoutes(r *gin.RouterGroup) {
	r.GET("/", getCustomPagesHandler)
	r.GET("/:id", getCustomPageHandler)

	// authRoutes := r.Group("/", auth.BearerMiddleware)

	r.POST("/", createCustomPageHandler)
	r.PUT("/:id", updateCustomPageHandler)
	r.DELETE("/:id", deleteCustomPageHandler)
}

func getCustomPagesHandler(ctx *gin.Context) {
	db := ctx.MustGet("db").(*mongo.Database)
	coll := db.Collection(collectionCustomPages)

	filter := bson.M{
		"published": true,
		"is_cv":     false,
	}

	if ctx.MustGet("Authenticated").(bool) == true {
		delete(filter, "published")
	}

	opts := options.Find().SetProjection(bson.M{
		"_id":   true,
		"title": true,
	})

	cur, err := coll.Find(ctx, filter, opts)
	defer cur.Close(ctx)

	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}

	var results []models.CustomPage

	for cur.Next(ctx) {
		var result models.CustomPage

		if err := cur.Decode(&result); err != nil {
			ctx.Status(http.StatusInternalServerError)
			ctx.Error(err)
			return
		}
		results = append(results, result)
	}

	ctx.JSON(http.StatusOK, results)
}

func getCustomPageHandler(ctx *gin.Context) {

	docID := ctx.Param("id")
	if docID == "" {
		ctx.Status(http.StatusBadRequest)
		ctx.Error(errors.New("no document identifier provided"))
	}

	db := ctx.MustGet("db").(*mongo.Database)
	coll := db.Collection(collectionCustomPages)

	filter := bson.M{
		"published": true,
	}

	projection := bson.M{
		"_id":        true,
		"title":      true,
		"full_title": true,
		"html":       true,
	}

	if ctx.MustGet("Authenticated").(bool) == true {
		delete(filter, "published")
		projection["is_cv"] = true
		projection["markdown"] = true
		projection["published"] = true
	}

	opts := options.FindOne().SetProjection(projection)

	res := coll.FindOne(ctx, filter, opts)

	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			ctx.Status(http.StatusNotFound)
			return
		}
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(res.Err())
		return
	}

	var doc models.CustomPage

	if err := res.Decode(&doc); err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(res.Err())
		return
	}

	ctx.JSON(http.StatusOK, doc)
}

func createCustomPageHandler(ctx *gin.Context) {

	var body models.CustomPage

	if err := ctx.BindJSON(&body); err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}

	stitle, err := helpers.ParseKebab(body.Title)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}
	body.SearchableTitle = stitle

	html, err := helpers.ParseMD(body.Markdown)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}
	body.HTML = html

	db := ctx.MustGet("db").(*mongo.Database)
	_, err = db.Collection(collectionCustomPages).InsertOne(ctx, body)

	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(errors.New("createCustomPageHandler: cannot insert document"))
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusCreated)
}

func updateCustomPageHandler(ctx *gin.Context) {

	docID := ctx.Param("id")

	if docID == "" {
		ctx.Status(http.StatusBadRequest)
		ctx.Error(errors.New("no document identifier provided"))
		return
	}

	var body models.CustomPage

	if err := ctx.BindJSON(&body); err != nil {
		ctx.Error(err)
		return
	}

	if body.ID.String() != docID {
		ctx.Status(http.StatusBadRequest)
		ctx.Error(errors.New("document identifier is different than id in path"))
		return
	}

	// generated fields, in case of new title / edited markdown

	stitle, err := helpers.ParseKebab(body.Title)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}
	body.SearchableTitle = stitle

	html, err := helpers.ParseMD(body.Markdown)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}
	body.HTML = html

	filter := bson.M{
		"_id":              body.ID,
		"searchable_title": body.SearchableTitle,
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

func deleteCustomPageHandler(ctx *gin.Context) {
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

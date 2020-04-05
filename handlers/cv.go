package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lexffe/backend.lexffe.io/auth"
	"github.com/lexffe/backend.lexffe.io/helpers"
	"github.com/lexffe/backend.lexffe.io/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RegisterCVRoutes registers the router with all CV related subroutes.
func RegisterCVRoutes(r *gin.RouterGroup) {
	r.GET("/", getCVHandler)
	r.PUT("/", auth.BearerMiddleware, updateCVHandler)
}

func getCVHandler(ctx *gin.Context) {

	db := ctx.MustGet("db").(*mongo.Database)
	coll := db.Collection(collectionCustomPages)

	filter := bson.M{
		"is_cv": true,
	}

	projection := bson.M{
		"published": true,
		"html":      true,
	}

	opts := options.FindOne().SetProjection(projection)

	res := coll.FindOne(ctx, filter, opts)

	if res.Err() != nil {

		ctx.Error(res.Err())

		if res.Err() == mongo.ErrNoDocuments {
			ctx.Status(http.StatusNotFound)
			return
		}

		ctx.Status(http.StatusInternalServerError)
		return
	}

	var cv models.CV

	if err := res.Decode(&cv); err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, cv)
}

func updateCVHandler(ctx *gin.Context) {
	// upsert

	var body models.CV

	if err := ctx.BindJSON(&body); err != nil {
		ctx.Status(http.StatusBadRequest)
		ctx.Error(err)
		return
	}

	db := ctx.MustGet("db").(*mongo.Database)
	coll := db.Collection(collectionCustomPages)

	html, err := helpers.ParseMD(body.Markdown)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}
	body.HTML = html

	filter := bson.M{
		"is_cv": true,
	}

	opts := options.Update().SetUpsert(true)

	_, err = coll.UpdateOne(ctx, filter, body, opts)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
	return
}

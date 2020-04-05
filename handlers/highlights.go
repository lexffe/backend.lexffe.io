package handlers

import (
	"errors"
	"net/http"

	"github.com/lexffe/backend.lexffe.io/auth"

	"github.com/gin-gonic/gin"
	"github.com/lexffe/backend.lexffe.io/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RegisterHLRoutes registers the router with all highlights related subroutes.
func RegisterHLRoutes(r *gin.RouterGroup) {

	r.GET("/", getHighlightsHandler)
	r.POST("/", auth.BearerMiddleware, setHighlightHandler)

}

// Note: max highlights == 5

func getHighlightsHandler(ctx *gin.Context) {

	db := ctx.MustGet("db").(mongo.Database)
	coll := db.Collection(collectionHighlights)

	cur, err := coll.Find(ctx, bson.M{})

	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}

	var highlights []models.Highlight

	if err = cur.All(ctx, highlights); err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, highlights)
}

func setHighlightHandler(ctx *gin.Context) {

	var HighlightsBody struct {
		Highlights []models.Highlight `json:"highlights,dive"`
	}

	if err := ctx.BindJSON(&HighlightsBody); err != nil {
		ctx.Status(http.StatusBadRequest)
		ctx.Error(err)
		return
	}

	db := ctx.MustGet("db").(mongo.Database)
	coll := db.Collection(collectionHighlights)

	models := []mongo.WriteModel{}

	for _, v := range HighlightsBody.Highlights {
		models = append(models, mongo.NewInsertOneModel().SetDocument(v))
	}

	opts := options.BulkWrite().SetOrdered(false)

	res, err := coll.BulkWrite(ctx, models, opts)

	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}

	if int(res.InsertedCount) != len(HighlightsBody.Highlights) {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(errors.New("mismatched insert count"))
		return
	}

	ctx.Status(http.StatusNoContent)
}

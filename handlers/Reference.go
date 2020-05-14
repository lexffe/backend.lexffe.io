package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lexffe/backend.lexffe.io/auth"
	"github.com/lexffe/backend.lexffe.io/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ReferenceHandler is a helper struct for all reference handlers.
type ReferenceHandler struct {
	Router        *gin.RouterGroup
	DB            *mongo.Database
	ReferenceType models.ObjectType
	Collection    string
}

// RegisterRoutes sets the router routes.
func (s *ReferenceHandler) RegisterRoutes() {
	s.Router.GET("/", s.getReferencesHandler)
	s.Router.GET("/:id", s.getReferenceHandler)

	protected := s.Router.Group("/", auth.CheckAuthentication)

	protected.POST("/", s.createReferenceHandler)
	protected.PUT("/:id", s.updateReferenceHandler)
	protected.DELETE("/:id", s.deleteReferenceHandler)
}

func (s *ReferenceHandler) getReferencesHandler(ctx *gin.Context) {

	// user-defined skip, for pagination.
	skipParam := ctx.DefaultQuery("skip", "0")
	skip, err := strconv.ParseInt(skipParam, 10, 64)
	// Bad request
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("skip is not a number"))
		ctx.Error(err)
		return
	}

	// user-defined limit, for pagination
	limitParam := ctx.DefaultQuery("limit", "0")
	limit, err := strconv.ParseInt(limitParam, 10, 64)

	// Bad request
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("limit is not a number"))
		ctx.Error(err)
		return
	}

	// get length of the collection (for pagination)

	count, err := s.DB.Collection(s.Collection).CountDocuments(ctx.Request.Context(), bson.M{})

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("cannot count number of documents in collection"))
		ctx.Error(err)
		return
	}

	ctx.Header("X-Collection-Length", strconv.FormatInt(count, 10))

	opts := options.Find().
		SetLimit(limit).
		SetSkip(skip).
		SetSort(bson.M{
			"_id": -1,
		})
		// .SetProjection

	cur, err := s.DB.Collection(s.Collection).Find(ctx.Request.Context(), bson.M{}, opts)
	//noinspection ALL
	defer cur.Close(ctx.Request.Context())

	// mongo related error
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var references []models.Reference

	if err := cur.All(ctx.Request.Context(), &references); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, references)
}

func (s *ReferenceHandler) getReferenceHandler(ctx *gin.Context) {
	// get the document identifier
	docID := ctx.Param("id")

	if docID == "" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("no document identifier provided"))
		return
	}

	// parse doc id
	objID, err := primitive.ObjectIDFromHex(docID)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	filter := bson.M{
		"_id": objID,
	}

	// search

	res := s.DB.Collection(s.Collection).FindOne(ctx.Request.Context(), filter)

	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			ctx.Status(http.StatusNotFound)
		} else {
			ctx.Status(http.StatusInternalServerError)
			ctx.Error(res.Err())
		}
		return
	}

	// Decode

	var doc models.Reference

	if err := res.Decode(&doc); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// return
	ctx.JSON(http.StatusOK, doc)
}

func (s *ReferenceHandler) createReferenceHandler(ctx *gin.Context) {

	var body models.Reference

	if err := ctx.BindJSON(&body); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("malformed request body"))
	}

	body.ReferenceType = s.ReferenceType

	_, err := s.DB.Collection(s.Collection).InsertOne(ctx.Request.Context(), body)

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("cannot insert document"))
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusCreated)
}

func (s *ReferenceHandler) updateReferenceHandler(ctx *gin.Context) {
	// get the document identifier
	docID := ctx.Param("id")

	if docID == "" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("no document identifier provided"))
		return
	}

	// parse doc id
	objID, err := primitive.ObjectIDFromHex(docID)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var body models.Reference

	if err := ctx.BindJSON(&body); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("failed to parse body"))
		ctx.Error(err)
		return
	}

	if body.ObjectID.String() != docID {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("document identifier is different than id in path"))
		return
	}

	filter := bson.M{
		"_id": objID,
	}

	res, err := s.DB.Collection(s.Collection).UpdateOne(ctx.Request.Context(), filter, body)

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("cannot update document"))
		ctx.Error(err)
		return
	}

	if res.MatchedCount == 0 {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (s *ReferenceHandler) deleteReferenceHandler(ctx *gin.Context) {
	// get the document identifier
	docID := ctx.Param("id")

	if docID == "" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("no document identifier provided"))
		return
	}

	// parse doc id
	objID, err := primitive.ObjectIDFromHex(docID)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	filter := bson.M{
		"_id": objID,
	}

	res := s.DB.Collection(s.Collection).FindOneAndDelete(ctx.Request.Context(), filter)

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

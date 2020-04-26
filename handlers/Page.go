package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lexffe/backend.lexffe.io/auth"
	"github.com/lexffe/backend.lexffe.io/helpers"
	"github.com/lexffe/backend.lexffe.io/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PageHandler is a helper struct for all page handlers.
type PageHandler struct {
	Router     *gin.RouterGroup
	DB         *mongo.Database
	PageType   models.ObjectType
	Collection string
}

// RegisterRoutes sets the router routes.
func (s *PageHandler) RegisterRoutes() {
	s.Router.GET("/", s.getPagesHandler)
	s.Router.GET("/:id", s.getPageHandler)

	protected := s.Router.Group("/", auth.CheckAuthentication)

	protected.POST("/", s.createPageHandler)
	protected.PUT("/:id", s.updatePageHandler)
	protected.DELETE("/:id", s.deletePageHandler)
}

// directory
func (s *PageHandler) getPagesHandler(ctx *gin.Context) {
	// user-defined skip, for pagination.
	skipParam := ctx.DefaultQuery("skip", "0")
	skip, err := strconv.Atoi(skipParam)

	// Bad request
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("skip is not a number"))
		ctx.Error(err)
		return
	}

	// if user wants a simple view

	simpleParam := ctx.DefaultQuery("simple", "false")
	simple, err := strconv.ParseBool(simpleParam)

	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("simple is not a boolean"))
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

	// only get the published pages
	filter := bson.M{
		"published": true,
	}

	projection := bson.M{
		"title":            true,
		"searchable_title": true,
		"tags":             true,
		"subtitle":         true,
		"page_type":        true,
		"html":             true,
		"published":        true,
		"last_updated":     true,
		"updated":          true,
	}

	if simple == true {
		delete(projection, "html")
	}

	// if user is authenticated, get the drafts as well. i.e. no filter
	if ctx.MustGet("Authorized").(bool) == true {
		delete(filter, "published")
	}

	opts := options.Find().
		SetLimit(paginationLimit).
		SetSkip(int64(skip)).
		SetProjection(projection).
		SetSort(bson.M{
			"_id": -1,
		})

	cur, err := s.DB.Collection(s.Collection).Find(ctx.Request.Context(), filter, opts)
	defer cur.Close(ctx.Request.Context())

	// mongo related error
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var results []models.Page

	// cursor.Next gets 0th document on first iteration.
	// i.e. cursor.Decode before first .Next spits out error.
	for cur.Next(ctx.Request.Context()) {

		var result models.Page

		if err := cur.Decode(&result); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		results = append(results, result)

		// if cur.ID() == 0 {
		// 	break
		// }

	}

	if len(results) == 0 {
		ctx.JSON(http.StatusOK, []int{}) // return empty slice
		return
	}

	ctx.JSON(http.StatusOK, results)
}

func (s *PageHandler) getPageHandler(ctx *gin.Context) {

	// get the document identifier (can be either searchable_title or _id)
	docID := ctx.Param("id")

	if docID == "" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("no document identifier provided"))
		return
	}

	// if the identifier is an object id
	isObjID, err := strconv.ParseBool(ctx.DefaultQuery("obj_id", "false"))

	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("malformed doc_id value, should be boolean"))
		ctx.Error(err)
		return
	}

	// only get the published pages
	filter := bson.M{"published": true}

	if isObjID {
		objID, err := primitive.ObjectIDFromHex(docID)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		filter["_id"] = objID
	} else {
		filter["searchable_title"] = docID
	}

	// reserved for projection
	// opts := options.FindOne()

	// if user is authenticated, get the drafts as well. i.e. no filter
	if ctx.MustGet("Authorized").(bool) == true {
		delete(filter, "published")
		// opts.SetProjection(bson.M{})
	}

	// Query

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

	var doc models.Page

	if err := res.Decode(&doc); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// return

	ctx.JSON(http.StatusOK, doc)
}

func (s *PageHandler) createPageHandler(ctx *gin.Context) {

	// parse body
	// body: { title, tags, subtitle, markdown, published }

	var body models.Page

	if err := ctx.BindJSON(&body); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("malformed request body"))
		ctx.Error(err)
		return
	}

	// generated fields: { searchable_title, page_type, html, last_updated }

	stitle, err := helpers.ParseKebab(body.Title)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("cannot generate searchable title"))
		ctx.Error(err)
		return
	}
	body.SearchableTitle = stitle

	html, err := helpers.ParseMD(body.Markdown)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("cannot generate html from markdown"))
		ctx.Error(err)
		return
	}
	body.HTML = html

	// create only: set page type
	body.PageType = s.PageType
	body.LastUpdated = time.Now()
	body.Updated = false

	// database operation

	_, err = s.DB.Collection(s.Collection).InsertOne(ctx.Request.Context(), body)

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("cannot insert document"))
		ctx.Error(err)
		return
	}

	// ok, return
	ctx.Status(http.StatusCreated)
}

func (s *PageHandler) updatePageHandler(ctx *gin.Context) {

	// get the document identifier (must be _id)
	docID := ctx.Param("id")

	if docID == "" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("no document identifier provided"))
		return
	}

	// parse body

	var body models.Page

	if err := ctx.BindJSON(&body); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("failed to parse body"))
		ctx.Error(err)
		return
	}

	if body.ObjectID.String() != docID {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("document identifier is different than id in path"))
		return
	}

	// generated fields, in case of new title / edited markdown: { searchable_title, html, last_updated }

	stitle, err := helpers.ParseKebab(body.Title)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("cannot generate searchable title"))
		ctx.Error(err)
		return
	}
	body.SearchableTitle = stitle

	html, err := helpers.ParseMD(body.Markdown)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("cannot generate html from markdown"))
		ctx.Error(err)
		return
	}
	body.HTML = html

	body.LastUpdated = time.Now()
	body.Updated = true

	filter := bson.M{
		"_id": body.ObjectID,
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

func (s *PageHandler) deletePageHandler(ctx *gin.Context) {

	// get the document identifier (must be _id)
	docID := ctx.Param("id")

	if docID == "" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("no document identifier provided"))
		return
	}

	objID, err := primitive.ObjectIDFromHex(docID)

	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid document identifier"))
		ctx.Error(err)
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

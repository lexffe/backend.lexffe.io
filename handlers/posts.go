package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lexffe/backend.lexffe.io/helpers"
	"github.com/lexffe/backend.lexffe.io/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collectionName = "post"
const retDocLimit = 10

// TODO: write out all db queries for each action
// TODO: determine if re-indexing is needed for descending order of documents (newest post first)

// RegisterPostRoutes registers the router with all post related subroutes.
func RegisterPostRoutes(r *gin.RouterGroup) {

	r.GET("/post", getPostsHandler)
	r.GET("/post/:id", getPostHandler)

	r.POST("/post", createPostHandler)       // need auth middleware
	r.PUT("/post/:id", updatePostHandler)    // need auth middleware
	r.DELETE("/post/:id", deletePostHandler) // need auth middleware

}

func getPostsHandler(ctx *gin.Context) {

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

	// get db from context, and assert type
	db := ctx.MustGet("db").(*mongo.Database)
	coll := db.Collection(collectionName)

	// only get the published posts
	filter := bson.M{
		"published": true,
	}

	// if user is authenticated, get the drafts as well. i.e. no filter
	if ctx.MustGet("Authenticated").(bool) == true {
		delete(filter, "published")
	}

	// setup projection and pagination
	opts := options.Find().SetLimit(retDocLimit).SetSkip(int64(skip)).SetProjection(bson.M{
		"_id":              true,
		"tags":             true,
		"title":            true,
		"searchable_title": true,
		"subtitle":         true,
	})

	cur, err := coll.Find(ctx, filter, opts)
	defer cur.Close(ctx)

	// mongo related error
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}

	var results []models.Post

	// cursor.Next gets 0th document on first iteration.
	// i.e. cursor.Decode before first .Next spits out error.
	for cur.Next(ctx) {
		var result models.Post
		err := cur.Decode(&result)
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			ctx.Error(err)
			return
		}
		results = append(results, result)

		// if cur.ID() == 0 {
		// 	break
		// }

	}

	ctx.JSON(http.StatusOK, results)
}

func getPostHandler(ctx *gin.Context) {

	// get the document identifier (can be either searchable_title or _id)
	docID := ctx.Param("id")

	if docID == "" {
		ctx.Status(http.StatusBadRequest)
		ctx.Error(errors.New("no document identifier provided"))
		return
	}

	// if the identifier is an object id
	isObjID, err := strconv.ParseBool(ctx.DefaultQuery("obj_id", "false"))

	if err != nil {
		ctx.Status(http.StatusBadRequest)
		ctx.Error(errors.New("malformed doc_id value, should be boolean"))
		ctx.Error(err)
		return
	}

	db := ctx.MustGet("db").(*mongo.Database)
	coll := db.Collection(collectionName)

	// only get the published posts
	filter := bson.M{"published": true}

	// if user is authenticated, get the drafts as well. i.e. no filter
	if ctx.MustGet("Authenticated").(bool) == true {
		delete(filter, "published")
	}

	if isObjID {
		filter["_id"] = docID
	} else {
		filter["searchable_title"] = docID
	}

	opts := options.FindOne().SetProjection(bson.M{
		"_id":              true,
		"tags":             true,
		"title":            true,
		"searchable_title": true, // FE note: replace URL with searchable_title.
		"subtitle":         true,
		"html":             true,
	})

	var doc models.Post

	res := coll.FindOne(ctx, filter, opts)

	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			ctx.Status(http.StatusNotFound)
		} else {
			ctx.Status(http.StatusInternalServerError)
		}

		ctx.Error(res.Err())
		return
	}

	if err := res.Decode(&doc); err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(res.Err())
		return
	}

	ctx.JSON(http.StatusOK, doc)
}

func createPostHandler(ctx *gin.Context) {

	// parse body

	var body models.Post

	if err := ctx.BindJSON(&body); err != nil {
		ctx.Error(err)
		return
	}

	// generated fields

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

	body.LastUpdated = time.Now()

	// database operation

	db := ctx.MustGet("db").(*mongo.Database)
	_, err = db.Collection(collectionName).InsertOne(ctx, body)

	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(errors.New("createPostHandler: cannot insert document"))
		ctx.Error(err)
		return
	}

	// ok, return
	ctx.Status(http.StatusCreated)
}

func updatePostHandler(ctx *gin.Context) {

	var body models.Post

	if err := ctx.BindJSON(&body); err != nil {
		ctx.Error(err)
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

	body.LastUpdated = time.Now()

	filter := bson.M{
		"_id":              body.ID,
		"searchable_title": body.SearchableTitle,
	}

	db := ctx.MustGet("db").(*mongo.Database)
	res, err := db.Collection(collectionName).UpdateOne(ctx, filter, body)

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

func deletePostHandler(ctx *gin.Context) {

	docID := ctx.Param("id")

	if docID == "" {
		ctx.Status(http.StatusBadRequest)
		ctx.Error(errors.New("no document identifier provided"))
		return
	}

	filter := bson.M{
		"_id": docID,
	}

	db := ctx.MustGet("db").(*mongo.Database)
	res := db.Collection(collectionName).FindOneAndDelete(ctx, filter)

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

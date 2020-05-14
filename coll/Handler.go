package coll

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lexffe/backend.lexffe.io/auth"
	"github.com/lexffe/backend.lexffe.io/handlers"
	"github.com/lexffe/backend.lexffe.io/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// RegisterRoutes registers all CRUD functions for collections
func (c *CollectionDelegate) RegisterRoutes() {

	// all routes should be guarded.

	router := c.Engine.Group("/coll", auth.CheckAuthentication)

	router.GET("/", c.getCollsHandler)
	//router.GET("/:name", c.getCollHandler)
	router.POST("/", c.createCollHandler)
	router.DELETE("/:name", c.deleteCollHandler)

}

func (c *CollectionDelegate) getCollsHandler(ctx *gin.Context) {

	cur, err := c.DB.Collection(metaCollection).Find(ctx.Request.Context(), bson.M{})

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("meta: error occured at find command"))
		ctx.Error(err)
		return
	}

	var results []MetaCollectionModel

	for cur.Next(ctx.Request.Context()) {

		var result MetaCollectionModel

		if err := cur.Decode(&result); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, errors.New("meta: cannot decode result"))
			ctx.Error(err)
			return
		}

		results = append(results, result)

	}

	if len(results) == 0 {
		ctx.JSON(http.StatusOK, []int{}) // return empty slice
		return
	}

	ctx.JSON(http.StatusOK, results)
}

//func (c *CollectionDelegate) getCollHandler(ctx *gin.Context) {
//
//	collName := ctx.Param("name")
//
//	if collName == "" {
//		ctx.AbortWithError(http.StatusBadRequest, errors.New("no collection name provided"))
//		return
//	}
//
//	filter := bson.M{
//		"_id": collName,
//	}
//
//	res := c.DB.Collection(metaCollection).FindOne(ctx.Request.Context(), filter)
//
//	if res.Err() != nil {
//		if res.Err() == mongo.ErrNoDocuments {
//			ctx.Status(http.StatusNotFound)
//		} else {
//			ctx.Status(http.StatusInternalServerError)
//			ctx.Error(res.Err())
//		}
//		return
//	}
//
//	var coll MetaCollectionModel
//
//	if err := res.Decode(&coll); err != nil {
//		ctx.AbortWithError(http.StatusInternalServerError, err)
//		return
//	}
//
//	ctx.JSON(http.StatusOK, coll)
//
//}

func (c *CollectionDelegate) createCollHandler(ctx *gin.Context) {
	var body MetaCollectionModel

	if err := ctx.BindJSON(&body); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("malformed request body"))
		ctx.Error(err)
		return
	}

	// prevent route collision with "coll" / "auth"
	if body.Name == "coll" || body.Name == "auth" {
		ctx.AbortWithError(http.StatusConflict, errors.New("collection name is in conflict with the router's internal routes"))
		return
	}

	// prevent existing collection collision
	count, err := c.DB.Collection(metaCollection).CountDocuments(ctx.Request.Context(), bson.M{"name": body.Name})

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("cannot check conflict"))
		ctx.Error(err)
		return
	}

	if count > 0 {
		ctx.AbortWithError(http.StatusConflict, errors.New("collection name is in conflict existing collection(s)"))
		return
	}

	// actually append and insert

	_, err = c.DB.Collection(metaCollection).InsertOne(ctx.Request.Context(), body)

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("cannot insert document"))
		ctx.Error(err)
		return
	}

	// Live register new routes

	switch body.Type {

	case models.TypePage:
		h := handlers.PageHandler{
			Router:     c.Engine.Group(body.Name),
			DB:         c.DB,
			PageType:   body.Type,
			Collection: body.Name,
		}
		h.RegisterRoutes()

	case models.TypeRef:
		h := handlers.ReferenceHandler{
			Router:        c.Engine.Group(body.Name),
			DB:            c.DB,
			ReferenceType: body.Type,
			Collection:    body.Name,
		}
		h.RegisterRoutes()

	default:

		// should not be reachable. BindJSON should check type

		// log.Printf("error on collection %v, type %v: %v", body.Name, body.Type, errors.New("collection type is not implemented"))
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("collection type is not implemented"))
		return

	}

	ctx.Status(http.StatusCreated)
}

func (c *CollectionDelegate) deleteCollHandler(ctx *gin.Context) {

	collName := ctx.Param("name")

	if collName == "" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("no collection name provided"))
		return
	}

	filter := bson.M{
		"_id": collName,
	}

	res := c.DB.Collection(metaCollection).FindOneAndDelete(ctx.Request.Context(), filter)

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

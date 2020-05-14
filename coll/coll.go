package coll

import (
	"github.com/gin-gonic/gin"
	"github.com/lexffe/backend.lexffe.io/models"
	"go.mongodb.org/mongo-driver/mongo"
)

/**
This package contains the handlers for CRUD of collections,
as well as the registration of routes of existing collections.
*/

const metaCollection = "meta"

// CollectionDelegate is a helper struct for all Collection related handlers.
type CollectionDelegate struct {
	Engine *gin.Engine
	DB     *mongo.Database
}

// MetaCollectionModel is a metadata document describing all the collections in the database
type MetaCollectionModel struct {
	// Name is the collection name.
	Name string `json:"_id" bson:"_id"`

	// Type is the collection type.
	Type models.ObjectType `json:"type" bson:"type"`
}

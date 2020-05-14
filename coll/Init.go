package coll

import (
	"context"
	"errors"
	"log"

	"github.com/lexffe/backend.lexffe.io/handlers"
	"github.com/lexffe/backend.lexffe.io/models"
	"go.mongodb.org/mongo-driver/bson"
)

/**

Logic
- database has collection "meta"
- function goes through the meta collection
	- get all documents with the filter { "type": "page" }
		- register to router
	- get all documents with the filter { "type": "reference" }
		- register to router
	- get all documents with the filter { "type": "asset" }
		- register to router
*/

// Bootstrap finds all registered collections (in meta) and registers the routes
func (c *CollectionDelegate) Bootstrap(ctx context.Context) error {

	cur, err := c.DB.Collection(metaCollection).Find(ctx, bson.M{})

	if err != nil {
		return err
	}

	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result MetaCollectionModel

		if err := cur.Decode(&result); err != nil {
			return err
		}

		switch result.Type {

		case models.TypePage:
			h := handlers.PageHandler{
				Router:     c.Engine.Group(result.Name),
				DB:         c.DB,
				PageType:   result.Type,
				Collection: result.Name,
			}
			h.RegisterRoutes()

		case models.TypeRef:
			h := handlers.ReferenceHandler{
				Router:        c.Engine.Group(result.Name),
				DB:            c.DB,
				ReferenceType: result.Type,
				Collection:    result.Name,
			}
			h.RegisterRoutes()

		default:

			log.Printf("error on collection %v, type %v: %v", result.Name, result.Type, errors.New("collection type is not implemented"))

		}

	}

	return nil
}

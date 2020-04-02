package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HighlightType is a type alias to the type of the object in highlight
type HighlightType int

const (
	// ProjectHighlight -
	ProjectHighlight HighlightType = iota

	// PostHighlight -
	PostHighlight
)

// Highlight is a complete model for highlights (for the front page)
// This model serves only as a reference. Further query is required.
type Highlight struct {
	ID primitive.ObjectID `json:"_id" bson:"_id, omitempty"`
	ObjectType HighlightType `json:"object_type" bson:"object_type"`
	ObjectID string `json:"object_id" bson:"object_id"`
}

// query: db.coll.find({PostId}, { Title: true, Subtitle: true, ...})

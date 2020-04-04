package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HighlightType is a type alias to the type of the object in highlight
type HighlightType string

const (
	// HighlightProject asserts that the highlight type is a project
	HighlightProject HighlightType = "project"

	// HighlightPost asserts that the highlight type is a blog post
	HighlightPost HighlightType = "post"
)

// Highlight is a complete model for highlights (for the front page)
// This model serves only as a reference. Further query is required.
type Highlight struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Type        HighlightType      `json:"type" bson:"type"`
	ObjectIDRef string             `json:"object_id" bson:"object_id"`
}

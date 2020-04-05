package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Project is 
type Project struct {
	ID primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Source string `json:"source" bson:"source" binding:"required"` // project source e.g. github / gitlab
	Name string `json:"name" bson:"name" binding:"required"`
	Description string `json:"description" bson:"description" binding:"required"`
	Footnote string `json:"footnote" bson:"footnote"`
	URL string `json:"url" bson:"url" binding:"required"`
}

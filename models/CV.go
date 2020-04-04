package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// CV is a special type of Custom Page.
type CV struct {
	ID primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	IsCV bool `json:"is_cv" bson:"is_cv"`
	Markdown string `json:"markdown" bson:"markdown"`
	HTML string `json:"html" bson:"html"`
	Published bool `json:"published" bson:"published"`
}

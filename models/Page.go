package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CustomPage is a generic markdown page.
// The view should render the custom pages in a nav.
// CV will be put in the same collection as CustomPage. The flag IsCV distinguishes that.
type CustomPage struct {
	ID primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Title string `json:"title" bson:"title" binding:"required"`
	FullTitle string `json:"full_title" bson:"full_title"`
	SearchableTitle string `json:"searchable_title" bson:"searchable_title"`
	IsCV bool `json:"is_cv" bson:"is_cv" binding:"required"`
	Markdown string `json:"markdown" bson:"markdown" binding:"required"`
	HTML string `json:"html" bson:"html"`
	Published bool `json:"published" bson:"published" binding:"required"`
}

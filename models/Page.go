package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Page is a generic markdown page.
type Page struct {
	ObjectID primitive.ObjectID `json:"_id" bson:"_id,omitempty"`

	// Title is the title of the page.
	Title string `json:"title" bson:"title" binding:"required"`

	/*
		SearchableTitle is a generated field that turns the arbitrary title into a kebab-case string

		Note: This field is used for MongoDB indexing.

		When a route like /posts/{searchable_title} is used, the controller should use this as filter instead.
	*/
	SearchableTitle string `json:"searchable_title" bson:"searchable_title"` // Generated Field

	// Tags is an array of keywords.
	Tags []string `json:"tags" bson:"tags" binding:"required"`

	// Subtitle is the subtitle/alternative title of the page.
	Subtitle string `json:"subtitle" bson:"subtitle" binding:"required"`

	// PageType denotes the type of this page
	PageType ObjectType `json:"page_type,omitempty" bson:"page_type"`

	// Markdown is the markdown template of the post.
	Markdown string `json:"markdown,omitempty" bson:"markdown" binding:"required"`

	/*
		HTML is a generated field from the markdown template.
		It is used for rendering the blog post itself. (.innerHTML)
	*/
	HTML string `json:"html,omitempty" bson:"html"` // Generated Field

	// Published is a flag for publisher to withhold the post (drafting).
	Published bool `json:"published" bson:"published" binding:"required"`

	// LastUpdated is a timestamp indicating when the document was last edited.
	LastUpdated time.Time `json:"last_updated" bson:"last_updated"`

	Updated bool `json:"updated" bson:"updated"`
}

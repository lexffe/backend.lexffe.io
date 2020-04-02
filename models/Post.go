package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// /posts/{searchable-title}
// /posts/{_id}?obj_id=true

/*
Post is a complete model for the blog posts.

Note: Projection may vary based on the context (e.g. which page the user is on.)

and the projection must be determined when querying the database.

*/
type Post struct {
	// PostId is MongoDB.ObjectId. Used for internal operation.
	ID primitive.ObjectID `json:"_id" bson:"_id,omitempty"`

	// Tags is an array of keywords.
	Tags []string `json:"tags" bson:"tags" binding:"required"`

	// Title is the title of the blog post.
	Title string `json:"title" bson:"title" binding:"required"`

	/*
		SearchableTitle is a generated field that turns the arbitary title into a kebab-case string

		Note: This field is used for MongoDB indexing.

		When a route like /posts/?title={searchable_title}`
	*/
	SearchableTitle string `json:"searchable_title" bson:"searchable_title"`

	// Subtitle is the subtitle of the blog post.
	Subtitle string `json:"subtitle" bson:"subtitle" binding:"required"`

	// Markdown is the markdown template of the post.
	Markdown string `json:"markdown" bson:"markdown" binding:"required"`

	/*
		HTML is a generated field from the markdown template.
		It is used for rendering the blog post itself. .innerHTML
	*/
	HTML string `json:"html" bson:"html"`

	// Published is a flag for publisher to withhold the post (drafting).
	Published bool `json:"published" bson:"published" binding:"required"`

	// LastUpdated is a timestamp indicating when the document was last edited.
	LastUpdated time.Time `json:"last_updated" bson:"last_updated"`
}

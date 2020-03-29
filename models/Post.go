package models

// Post is a complete model for the blog posts.
// Note: Projection may vary based on the context (e.g. which page the user is on.)
// and the projection must be determined when querying the database.
type Post struct {
	// PostId is MongoDB.ObjectId. Used for internal operation.

	// Tags is an array of keywords.
	Tags []string `json:"tags" bson:"tags"`

	// Title is the title of the blog post.
	Title string `json:"title" bson:"title"`

	// SearchableTitle is a generated field that turns 
	SearchableTitle string `json:"searchable_title" bson:"searchable_title"`
	Subtitle string `json:"subtitle" bson:"subtitle"`
	Markdown string `json:"markdown" bson:"markdown"`
	HTML string `json:"html" bson:"html"`
	Published bool `json:"published" bson:"published"`
}

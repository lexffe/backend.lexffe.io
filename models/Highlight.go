package models

type highlightType int

const (
	project highlightType = iota
	post
)

// type Highlight struct {
// 	PostID string `json:"post_id" bson:"post_id"`
// 	Type highlightType `json:"type" bson:"type"`
// }

type Highlight struct {
	ObjectType highlightType `json:"object_type" bson:"object_type"`
	ObjectID string `json:"object_id" bson:"object_id"`
}

// query: db.coll.find({PostId}, { Title: true, Subtitle: true, })

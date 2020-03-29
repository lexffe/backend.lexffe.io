package models

type Project struct {
	Source string `json:"source" bson:"source"` // project source e.g. github / gitlab
	Name string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
	Footnote string `json:"footnote" bson:"footnote"`
	URL string `json:"url" bson:"url"`
}


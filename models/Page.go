package models

type PageType int

const (
	CV PageType = iota
	Custom
)

type Page struct {
	Title string `json:"title" bson:"title"`
	FullTitle string `json:"full_title" bson:"full_title"`
	PageType PageType `json:"type" bson:"type"`
	Markdown string `json:"markdown" bson:"markdown"`
	HTML string `json:"html" bson:"html"`
	Published bool `json:"published" bson:"published"`
}

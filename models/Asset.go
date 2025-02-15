package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Asset is any non text-based asset
type Asset struct {
	ObjectID  primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	MIMEType  string
	AssetPath string // Path relative to the $(cwd)/assets
}

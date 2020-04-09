package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Reference is an object that either points to a resource, or describe a resource (metadata)
type Reference struct {
	ObjectID primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name string `json:"name" bson:"name" binding:"required"`
	Description string `json:"description" bson:"description"`
	ReferenceSource string `json:"reference_source" bson:"reference_source" binding:"required"`
	ReferenceType ObjectType `json:"reference_type" bson:"reference_type"`
	InternalObjectID primitive.ObjectID `json:"internal_id" bson:"internal_id"`
	ExternalURL string `json:"url" bson:"url"`
}

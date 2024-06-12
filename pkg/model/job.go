package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Job struct {
	ID         primitive.ObjectID `bson:"_id"`
	JobId      string             `bson:"jobId"`
	Status     string             `bson:"status"`
	DocumentId string             `bson:"documentId"`
}

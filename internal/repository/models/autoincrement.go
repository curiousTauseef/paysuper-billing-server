package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Autoincrement struct {
	Id         primitive.ObjectID `bson:"_id"`
	Collection string             `bson:"collection"`
	Counter    int64              `bson:"counter"`
	UpdatedAt  time.Time          `bson:"updated_at"`
}

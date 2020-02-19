package models

import (
	"github.com/bxcodec/faker"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
)

// InitFakeCustomProviders is initialize all custom providers for faker library.
func InitFakeCustomProviders() {
	faker.AddProvider("objectId", FakeProviderObjectId)
	faker.AddProvider("objectIdString", FakeProviderObjectIdString)
}

// FakeProviderObjectId returned value as mongo ObjectID
func FakeProviderObjectId(_ reflect.Value) (interface{}, error) {
	return primitive.NewObjectID(), nil
}

// FakeProviderObjectIdString returned value as string of mongo ObjectID
func FakeProviderObjectIdString(_ reflect.Value) (interface{}, error) {
	return primitive.NewObjectID().Hex(), nil
}

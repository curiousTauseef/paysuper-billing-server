package models

import (
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
	"testing"
)

func Test_FakeProviderObjectId_Ok(t *testing.T) {
	result, err := FakeProviderObjectId(reflect.Value{})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	id := result.(primitive.ObjectID)
	assert.NotEmpty(t, id)
}

func Test_FakeProviderObjectIdString_Ok(t *testing.T) {
	result, err := FakeProviderObjectIdString(reflect.Value{})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	id, err := primitive.ObjectIDFromHex(result.(string))
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
}

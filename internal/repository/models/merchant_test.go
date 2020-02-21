package models

import (
	"github.com/bxcodec/faker"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_MerchantMap_Ok(t *testing.T) {
	InitFakeCustomProviders()

	original := &billingpb.Merchant{}
	err := faker.FakeData(original)
	assert.NoError(t, err)

	m := NewMerchantMapper()
	mgo, err := m.MapObjectToMgo(original)
	assert.NoError(t, err)
	assert.NotNil(t, mgo)

	converted, err := m.MapMgoToObject(mgo)
	assert.NoError(t, err)
	assert.NotNil(t, mgo)

	assert.ObjectsAreEqualValues(original, converted)
}

func Test_MerchantMap_Error(t *testing.T) {
	original := &billingpb.Merchant{
		CreatedAt: &timestamp.Timestamp{Seconds: -1, Nanos: -1},
	}
	m := NewMerchantMapper()
	mgo, err := m.MapObjectToMgo(original)
	assert.Error(t, err)
	assert.Nil(t, mgo)
}
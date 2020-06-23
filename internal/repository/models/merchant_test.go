package models

import (
	"bytes"
	"github.com/bxcodec/faker"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_MerchantMap_Ok(t *testing.T) {
	InitFakeCustomProviders()

	original := &billingpb.Merchant{}
	err := faker.FakeData(original)
	original.CentrifugoToken = ""
	original.HasProjects = false
	assert.NoError(t, err)

	m := NewMerchantMapper()
	mgo, err := m.MapObjectToMgo(original)
	assert.NoError(t, err)
	assert.NotNil(t, mgo)

	obj, err := m.MapMgoToObject(mgo)
	assert.NoError(t, err)
	assert.NotNil(t, mgo)

	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	marshaler := &jsonpb.Marshaler{}

	assert.NoError(t, marshaler.Marshal(buf1, original))
	assert.NoError(t, marshaler.Marshal(buf2, obj.(*billingpb.Merchant)))
	assert.JSONEq(t, string(buf1.Bytes()), string(buf2.Bytes()))
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
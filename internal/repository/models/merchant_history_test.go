package models

import (
	"bytes"
	"github.com/bxcodec/faker"
	"github.com/golang/protobuf/jsonpb"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_MerchantHistoryMap_Ok(t *testing.T) {
	InitFakeCustomProviders()

	original := &billingpb.MerchantPaymentMethodHistory{}
	err := faker.FakeData(original)
	assert.NoError(t, err)

	m := NewMerchantPaymentMethodHistoryMapper()
	mgo, err := m.MapObjectToMgo(original)
	assert.NoError(t, err)
	assert.NotNil(t, mgo)

	obj, err := m.MapMgoToObject(mgo)
	assert.NoError(t, err)
	assert.NotNil(t, obj)

	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	marshaler := &jsonpb.Marshaler{}

	assert.NoError(t, marshaler.Marshal(buf1, original))
	assert.NoError(t, marshaler.Marshal(buf2, obj.(*billingpb.MerchantPaymentMethodHistory)))
	assert.JSONEq(t, string(buf1.Bytes()), string(buf2.Bytes()))
}

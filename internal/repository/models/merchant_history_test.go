package models

import (
	"github.com/bxcodec/faker"
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

	converted, err := m.MapMgoToObject(mgo)
	assert.NoError(t, err)
	assert.NotNil(t, mgo)

	assert.ObjectsAreEqualValues(original, converted)
}

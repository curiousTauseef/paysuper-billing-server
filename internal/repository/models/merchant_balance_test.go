package models

import (
	"github.com/bxcodec/faker"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type MerchantBalanceTestSuite struct {
	suite.Suite
	mapper merchantBalanceMapper
}

func TestMerchantBalanceTestSuite(t *testing.T) {
	suite.Run(t, new(MerchantBalanceTestSuite))
}

func (suite *MerchantBalanceTestSuite) SetupTest() {
	InitFakeCustomProviders()
}

func (suite *MerchantBalanceTestSuite) Test_MerchantBalance_NewMapper() {
	mapper := NewMerchantBalanceMapper()
	assert.IsType(suite.T(), &merchantBalanceMapper{}, mapper)
}

func (suite *MerchantBalanceTestSuite) Test_MapToMgo_Ok() {
	original := &billingpb.MerchantBalance{}
	err := faker.FakeData(original)
	assert.NoError(suite.T(), err)

	mgo, err := suite.mapper.MapObjectToMgo(original)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), mgo)

	obj, err := suite.mapper.MapMgoToObject(mgo)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), obj)

	assert.ObjectsAreEqualValues(original, obj)
}

func (suite *MerchantBalanceTestSuite) Test_Error_CreatedAt() {
	original := &billingpb.MerchantBalance{
		CreatedAt: &timestamp.Timestamp{Seconds: -1, Nanos: -1},
	}
	_, err := suite.mapper.MapObjectToMgo(original)
	assert.Error(suite.T(), err)
}
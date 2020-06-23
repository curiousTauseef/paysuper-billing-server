package models

import (
	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type OrderViewPublicTestSuite struct {
	suite.Suite
	mapper orderViewPublicMapper
}

func TestOrderViewPublicTestSuite(t *testing.T) {
	suite.Run(t, new(OrderViewPublicTestSuite))
}

func (suite *OrderViewPublicTestSuite) SetupTest() {
	InitFakeCustomProviders()
}

func (suite *OrderViewPublicTestSuite) Test_OrderViewPublic_NewMapper() {
	mapper := NewOrderViewPublicMapper()
	assert.IsType(suite.T(), &orderViewPublicMapper{}, mapper)
}

func (suite *OrderViewPublicTestSuite) Test_MapFromMgo_Ok() {
	original := &MgoOrderViewPublic{}
	err := faker.FakeData(original)
	assert.NoError(suite.T(), err)

	mgo, err := suite.mapper.MapMgoToObject(original)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), mgo)
}
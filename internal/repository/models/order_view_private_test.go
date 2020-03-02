package models

import (
	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type OrderViewPrivateTestSuite struct {
	suite.Suite
	mapper orderViewPrivateMapper
}

func TestOrderViewPrivateTestSuite(t *testing.T) {
	suite.Run(t, new(OrderViewPrivateTestSuite))
}

func (suite *OrderViewPrivateTestSuite) SetupTest() {
	InitFakeCustomProviders()
}

func (suite *OrderViewPrivateTestSuite) Test_OrderViewPrivate_NewMapper() {
	mapper := NewOrderViewPrivateMapper()
	assert.IsType(suite.T(), &orderViewPrivateMapper{}, mapper)
}

func (suite *OrderViewPrivateTestSuite) Test_MgoFromMgo_Ok() {
	original := &MgoOrderViewPrivate{}
	err := faker.FakeData(original)
	assert.NoError(suite.T(), err)

	mgo, err := suite.mapper.MapMgoToObject(original)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), mgo)
}
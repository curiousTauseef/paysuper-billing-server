package models

import (
	"github.com/bxcodec/faker"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ProjectestSuite struct {
	suite.Suite
	mapper projectMapper
}

func TestProjectTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectestSuite))
}

func (suite *ProjectestSuite) SetupTest() {
	InitFakeCustomProviders()
}

func (suite *ProjectestSuite) Test_NewMapper() {
	mapper := NewProjectMapper()
	assert.IsType(suite.T(), &projectMapper{}, mapper)
}

func (suite *ProjectestSuite) Test_MapToMgo_Ok() {
	original := &billingpb.Project{}
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

func (suite *ProjectestSuite) Test_Error_CreatedAt() {
	original := &billingpb.Project{}
	err := faker.FakeData(original)
	assert.NoError(suite.T(), err)
	original.CreatedAt = &timestamp.Timestamp{Seconds: -1, Nanos: -1}

	_, err = suite.mapper.MapObjectToMgo(original)
	assert.Error(suite.T(), err)
}

func (suite *ProjectestSuite) Test_Error_UpdatedAt() {
	original := &billingpb.Project{}
	err := faker.FakeData(original)
	assert.NoError(suite.T(), err)
	original.UpdatedAt = &timestamp.Timestamp{Seconds: -1, Nanos: -1}

	_, err = suite.mapper.MapObjectToMgo(original)
	assert.Error(suite.T(), err)
}

func (suite *ProjectestSuite) Test_NoError_EmptyId() {
	original := &billingpb.Project{}
	err := faker.FakeData(original)
	assert.NoError(suite.T(), err)
	original.Id = ""
	obj, err := suite.mapper.MapObjectToMgo(original)
	assert.NoError(suite.T(), err)
	ref := obj.(*MgoRefund)
	assert.False(suite.T(), ref.Id.IsZero())
}

func (suite *ProjectestSuite) Test_Error_WrongId() {
	original := &billingpb.Project{}
	err := faker.FakeData(original)
	assert.NoError(suite.T(), err)

	original.Id = "test"

	_, err = suite.mapper.MapObjectToMgo(original)
	assert.Error(suite.T(), err)
}
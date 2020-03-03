package models

import (
	"bytes"
	"github.com/bxcodec/faker"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type RefundTestSuite struct {
	suite.Suite
	mapper refundMapper
}

func TestRefundTestSuite(t *testing.T) {
	suite.Run(t, new(RefundTestSuite))
}

func (suite *RefundTestSuite) SetupTest() {
	InitFakeCustomProviders()
}

func (suite *RefundTestSuite) Test_NewMapper() {
	mapper := NewRefundMapper()
	assert.IsType(suite.T(), &refundMapper{}, mapper)
}

func (suite *RefundTestSuite) Test_MapToMgo_Ok() {
	original := &billingpb.Refund{}
	err := faker.FakeData(original)
	assert.NoError(suite.T(), err)

	mgo, err := suite.mapper.MapObjectToMgo(original)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), mgo)

	obj, err := suite.mapper.MapMgoToObject(mgo)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), obj)

	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	marshaler := &jsonpb.Marshaler{}

	assert.NoError(suite.T(), marshaler.Marshal(buf1, original))
	assert.NoError(suite.T(), marshaler.Marshal(buf2, obj.(*billingpb.Refund)))
	assert.JSONEq(suite.T(), string(buf1.Bytes()), string(buf2.Bytes()))
}

func (suite *RefundTestSuite) Test_Error_CreatedAt() {
	original := &billingpb.Refund{}
	err := faker.FakeData(original)
	assert.NoError(suite.T(), err)
	original.CreatedAt = &timestamp.Timestamp{Seconds: -1, Nanos: -1}

	_, err = suite.mapper.MapObjectToMgo(original)
	assert.Error(suite.T(), err)
}

func (suite *RefundTestSuite) Test_Error_UpdatedAt() {
	original := &billingpb.Refund{}
	err := faker.FakeData(original)
	assert.NoError(suite.T(), err)
	original.UpdatedAt = &timestamp.Timestamp{Seconds: -1, Nanos: -1}

	_, err = suite.mapper.MapObjectToMgo(original)
	assert.Error(suite.T(), err)
}

func (suite *RefundTestSuite) Test_NoError_EmptyId() {
	original := &billingpb.Refund{}
	err := faker.FakeData(original)
	assert.NoError(suite.T(), err)
	original.Id = ""
	obj, err := suite.mapper.MapObjectToMgo(original)
	assert.NoError(suite.T(), err)
	ref := obj.(*MgoRefund)
	assert.False(suite.T(), ref.Id.IsZero())
}

func (suite *RefundTestSuite) Test_Error_WrongId() {
	original := &billingpb.Refund{}
	err := faker.FakeData(original)
	assert.NoError(suite.T(), err)

	original.Id = "test"

	_, err = suite.mapper.MapObjectToMgo(original)
	assert.Error(suite.T(), err)
}

func (suite *RefundTestSuite) Test_Error_WrongCreatorId() {
	original := &billingpb.Refund{}
	err := faker.FakeData(original)
	assert.NoError(suite.T(), err)

	original.CreatorId = "test"

	_, err = suite.mapper.MapObjectToMgo(original)
	assert.Error(suite.T(), err)
}

func (suite *RefundTestSuite) Test_Error_WrongOrderId() {
	original := &billingpb.Refund{}
	err := faker.FakeData(original)
	assert.NoError(suite.T(), err)

	original.OriginalOrder.Id = "test"

	_, err = suite.mapper.MapObjectToMgo(original)
	assert.Error(suite.T(), err)
}
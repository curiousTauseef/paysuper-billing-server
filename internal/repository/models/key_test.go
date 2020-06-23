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

type KeyTestSuite struct {
	suite.Suite
	mapper keyMapper
}

func TestKeyTestSuite(t *testing.T) {
	suite.Run(t, new(KeyTestSuite))
}

func (suite *KeyTestSuite) SetupTest() {
	InitFakeCustomProviders()
}

func (suite *KeyTestSuite) Test_Key_NewMapper() {
	mapper := NewKeyMapper()
	assert.IsType(suite.T(), &keyMapper{}, mapper)
}

func (suite *KeyTestSuite) Test_MapToMgo_Ok() {
	original := &billingpb.Key{}
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
	assert.NoError(suite.T(), marshaler.Marshal(buf2, obj.(*billingpb.Key)))
	assert.JSONEq(suite.T(), string(buf1.Bytes()), string(buf2.Bytes()))
}

func (suite *KeyTestSuite) Test_Error_CreatedAt() {
	original := &billingpb.Key{
		CreatedAt: &timestamp.Timestamp{Seconds: -1, Nanos: -1},
	}
	_, err := suite.mapper.MapObjectToMgo(original)
	assert.Error(suite.T(), err)
}

func (suite *KeyTestSuite) Test_Error_RedeemedAt() {
	original := &billingpb.Key{
		RedeemedAt: &timestamp.Timestamp{Seconds: -1, Nanos: -1},
	}
	_, err := suite.mapper.MapObjectToMgo(original)
	assert.Error(suite.T(), err)
}

func (suite *KeyTestSuite) Test_Error_ReservedTo() {
	original := &billingpb.Key{
		ReservedTo: &timestamp.Timestamp{Seconds: -1, Nanos: -1},
	}
	_, err := suite.mapper.MapObjectToMgo(original)
	assert.Error(suite.T(), err)
}
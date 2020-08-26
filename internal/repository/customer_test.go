package repository

import (
	"context"
	"errors"
	"github.com/bxcodec/faker"
	"github.com/paysuper/paysuper-billing-server/internal/config"
	"github.com/paysuper/paysuper-billing-server/internal/mocks"
	"github.com/paysuper/paysuper-billing-server/internal/repository/models"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
	mongodbMock "gopkg.in/paysuper/paysuper-database-mongo.v2/mocks"
	"testing"
)

type CustomerTestSuite struct {
	suite.Suite
	db          mongodb.SourceInterface
	repository  *customerRepository
	log         *zap.Logger
	zapRecorder *observer.ObservedLogs
}

func Test_Customer(t *testing.T) {
	suite.Run(t, new(CustomerTestSuite))
}

func (suite *CustomerTestSuite) SetupTest() {
	_, err := config.NewConfig()
	assert.NoError(suite.T(), err, "Config load failed")

	suite.db, err = mongodb.NewDatabase()
	assert.NoError(suite.T(), err, "Database connection failed")

	var core zapcore.Core

	lvl := zap.NewAtomicLevel()
	core, suite.zapRecorder = observer.New(lvl)
	suite.log = zap.New(core)
	zap.ReplaceGlobals(suite.log)

	suite.repository = NewCustomerRepository(suite.db).(*customerRepository)
	assert.NotNil(suite.T(), suite.repository)

	models.InitFakeCustomProviders()
}

func (suite *CustomerTestSuite) TearDownTest() {
	if err := suite.db.Drop(); err != nil {
		suite.FailNow("Database deletion failed", "%v", err)
	}

	if err := suite.db.Close(); err != nil {
		suite.FailNow("Database close failed", "%v", err)
	}
}

func (suite *CustomerTestSuite) TestCustomer_FindAll_Ok() {
	cursorMock := &mongodbMock.CursorInterface{}
	cursorMock.On("All", mock.Anything, mock.MatchedBy(func(in *[]*models.MgoCustomer) bool {
		for i := 0; i < 3; i++ {
			it := new(models.MgoCustomer)
			err := faker.FakeData(it)
			assert.NoError(suite.T(), err)

			*in = append(*in, it)
		}

		return true
	})).Return(nil)
	collectionMock := &mongodbMock.CollectionInterface{}
	collectionMock.On("Find", mock.Anything, mock.Anything, mock.Anything).Return(cursorMock, nil)
	dbMock := &mongodbMock.SourceInterface{}
	dbMock.On("Collection", mock.Anything).Return(collectionMock, nil)
	suite.repository.db = dbMock

	_, err := suite.repository.FindAll(context.Background())
	assert.NoError(suite.T(), err)
}

func (suite *CustomerTestSuite) TestCustomer_FindAll_Find_Error() {
	collectionMock := &mongodbMock.CollectionInterface{}
	collectionMock.On("Find", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("TestCustomer_FindAll_Find_Error"))
	dbMock := &mongodbMock.SourceInterface{}
	dbMock.On("Collection", mock.Anything).Return(collectionMock, nil)
	suite.repository.db = dbMock

	_, err := suite.repository.FindAll(context.Background())
	assert.Error(suite.T(), err)
	assert.EqualError(suite.T(), err, "TestCustomer_FindAll_Find_Error")

	logs := suite.zapRecorder.All()
	assert.Len(suite.T(), logs, 1)
	assert.Equal(suite.T(), zapcore.ErrorLevel, logs[0].Level)
	assert.Equal(suite.T(), pkg.ErrorDatabaseQueryFailed, logs[0].Message)
	assert.EqualError(suite.T(), logs[0].Context[0].Interface.(error), "TestCustomer_FindAll_Find_Error")
}

func (suite *CustomerTestSuite) TestCustomer_FindAll_CursorAll_Error() {
	cursorMock := &mongodbMock.CursorInterface{}
	cursorMock.On("All", mock.Anything, mock.Anything).Return(errors.New("TestCustomer_FindAll_CursorAll_Error"))
	collectionMock := &mongodbMock.CollectionInterface{}
	collectionMock.On("Find", mock.Anything, mock.Anything, mock.Anything).Return(cursorMock, nil)
	dbMock := &mongodbMock.SourceInterface{}
	dbMock.On("Collection", mock.Anything).Return(collectionMock, nil)
	suite.repository.db = dbMock

	_, err := suite.repository.FindAll(context.Background())
	assert.Error(suite.T(), err)
	assert.EqualError(suite.T(), err, "TestCustomer_FindAll_CursorAll_Error")

	logs := suite.zapRecorder.All()
	assert.Len(suite.T(), logs, 1)
	assert.Equal(suite.T(), zapcore.ErrorLevel, logs[0].Level)
	assert.Equal(suite.T(), pkg.ErrorQueryCursorExecutionFailed, logs[0].Message)
	assert.EqualError(suite.T(), logs[0].Context[0].Interface.(error), "TestCustomer_FindAll_CursorAll_Error")
}

func (suite *CustomerTestSuite) TestCustomer_FindAll_MapperMapMgoToObject_Error() {
	cursorMock := &mongodbMock.CursorInterface{}
	cursorMock.On("All", mock.Anything, mock.MatchedBy(func(in *[]*models.MgoCustomer) bool {
		for i := 0; i < 3; i++ {
			it := new(models.MgoCustomer)
			err := faker.FakeData(it)
			assert.NoError(suite.T(), err)

			*in = append(*in, it)
		}

		return true
	})).Return(nil)
	collectionMock := &mongodbMock.CollectionInterface{}
	collectionMock.On("Find", mock.Anything, mock.Anything, mock.Anything).Return(cursorMock, nil)
	dbMock := &mongodbMock.SourceInterface{}
	dbMock.On("Collection", mock.Anything).Return(collectionMock, nil)
	suite.repository.db = dbMock

	mapperMock := &mocks.Mapper{}
	mapperMock.On("MapMgoToObject", mock.Anything).Return(nil, errors.New("TestCustomer_FindAll_MapperMapMgoToObject_Error"))
	suite.repository.mapper = mapperMock

	_, err := suite.repository.FindAll(context.Background())
	assert.Error(suite.T(), err)
	assert.EqualError(suite.T(), err, "TestCustomer_FindAll_MapperMapMgoToObject_Error")

	logs := suite.zapRecorder.All()
	assert.Len(suite.T(), logs, 1)
	assert.Equal(suite.T(), zapcore.ErrorLevel, logs[0].Level)
	assert.Equal(suite.T(), pkg.ErrorDatabaseMapModelFailed, logs[0].Message)
	assert.EqualError(suite.T(), logs[0].Context[0].Interface.(error), "TestCustomer_FindAll_MapperMapMgoToObject_Error")
}

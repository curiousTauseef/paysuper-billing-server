package repository

import (
	"context"
	"errors"
	"github.com/paysuper/paysuper-billing-server/internal/config"
	"github.com/paysuper/paysuper-billing-server/internal/repository/models"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
	mongodbMock "gopkg.in/paysuper/paysuper-database-mongo.v2/mocks"
	"testing"
	"time"
)

type AutoincrementTestSuite struct {
	suite.Suite
	db          mongodb.SourceInterface
	repository  *autoincrementRepository
	log         *zap.Logger
	zapRecorder *observer.ObservedLogs
}

func Test_Autoincrement(t *testing.T) {
	suite.Run(t, new(AutoincrementTestSuite))
}

func (suite *AutoincrementTestSuite) SetupTest() {
	_, err := config.NewConfig()
	assert.NoError(suite.T(), err, "Config load failed")

	suite.db, err = mongodb.NewDatabase()
	assert.NoError(suite.T(), err, "Database connection failed")

	var core zapcore.Core

	lvl := zap.NewAtomicLevel()
	core, suite.zapRecorder = observer.New(lvl)
	suite.log = zap.New(core)
	zap.ReplaceGlobals(suite.log)

	suite.repository = NewAutoincrementRepository(suite.db).(*autoincrementRepository)
	assert.NotNil(suite.T(), suite.repository)

	res, err := suite.db.Collection(collectionAutoincrement).InsertMany(
		context.TODO(),
		[]interface{}{
			&models.Autoincrement{
				Id:         primitive.NewObjectID(),
				Collection: collectionPayoutDocuments,
				Counter:    0,
				UpdatedAt:  time.Now(),
			},
		},
	)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), res.InsertedIDs, 1)
}

func (suite *AutoincrementTestSuite) TearDownTest() {
	if err := suite.db.Drop(); err != nil {
		suite.FailNow("Database deletion failed", "%v", err)
	}

	if err := suite.db.Close(); err != nil {
		suite.FailNow("Database close failed", "%v", err)
	}
}

func (suite *AutoincrementTestSuite) TestAutoincrement_GatPayoutAutoincrementId_Ok() {
	for i := 1; i < 100; i++ {
		counter, err := suite.repository.GatPayoutAutoincrementId(context.TODO())
		assert.NoError(suite.T(), err)
		assert.EqualValues(suite.T(), i, counter)
	}
}

func (suite *AutoincrementTestSuite) TestAutoincrement_GatPayoutAutoincrementId_Error() {
	singleResultMock := &mongodbMock.SingleResultInterface{}
	singleResultMock.On("Decode", mock.Anything).Return(errors.New("AutoincrementRepository_SingleResult_Error"))
	collectionMock := &mongodbMock.CollectionInterface{}
	collectionMock.On("FindOneAndUpdate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(singleResultMock, nil)
	dbMock := &mongodbMock.SourceInterface{}
	dbMock.On("Collection", mock.Anything).Return(collectionMock, nil)
	suite.repository.db = dbMock

	logMessages := suite.zapRecorder.All()
	assert.Empty(suite.T(), logMessages)

	_, err := suite.repository.GatPayoutAutoincrementId(context.TODO())
	assert.Error(suite.T(), err)
	assert.EqualError(suite.T(), err, "AutoincrementRepository_SingleResult_Error")

	logMessages = suite.zapRecorder.All()
	assert.Len(suite.T(), logMessages, 1)
	assert.Equal(suite.T(), zapcore.ErrorLevel, logMessages[0].Level)
	assert.Equal(suite.T(), pkg.ErrorDatabaseQueryFailed, logMessages[0].Message)
	assert.EqualError(suite.T(), logMessages[0].Context[0].Interface.(error), "AutoincrementRepository_SingleResult_Error")
}

package repository

import (
	"encoding/base64"
	"errors"
	"github.com/paysuper/paysuper-billing-server/internal/config"
	"github.com/paysuper/paysuper-billing-server/internal/mocks"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-proto/go/postmarkpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
	"testing"
)

type RoyaltyReportTestSuite struct {
	suite.Suite
	db          mongodb.SourceInterface
	repository  *royaltyReportRepository
	logObserver *zap.Logger
	zapRecorder *observer.ObservedLogs
}

func Test_RoyaltyReport(t *testing.T) {
	suite.Run(t, new(RoyaltyReportTestSuite))
}

func (suite *RoyaltyReportTestSuite) SetupTest() {
	_, err := config.NewConfig()
	assert.NoError(suite.T(), err, "Config load failed")

	suite.db, err = mongodb.NewDatabase()
	assert.NoError(suite.T(), err, "Database connection failed")

	cache := &mocks.CacheInterface{}
	cache.On("Get", mock.Anything, mock.Anything).Return(nil)
	cache.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	cache.On("Delete", mock.Anything).Return(nil)

	repository := NewRoyaltyReportRepository(suite.db, cache)
	suite.repository = repository.(*royaltyReportRepository)

	var core zapcore.Core

	lvl := zap.NewAtomicLevel()
	core, suite.zapRecorder = observer.New(lvl)
	suite.logObserver = zap.New(core)

	zap.ReplaceGlobals(suite.logObserver)
}

func (suite *RoyaltyReportTestSuite) TearDownTest() {
	if err := suite.db.Drop(); err != nil {
		suite.FailNow("Database deletion failed", "%v", err)
	}

	if err := suite.db.Close(); err != nil {
		suite.FailNow("Database close failed", "%v", err)
	}
}

func (suite *RoyaltyReportTestSuite) TestRoyaltyReport_SetRoyaltyReportFinanceItem_Ok() {
	items, err := suite.repository.SetRoyaltyReportFinanceItem(
		"ffffffffffffffffffffffff",
		&postmarkpb.PayloadAttachment{
			Content:     base64.StdEncoding.EncodeToString([]byte(``)),
			Name:        "file_name.txt",
			ContentType: "text/plain",
		},
	)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), items, 1)
}

func (suite *RoyaltyReportTestSuite) TestRoyaltyReport_SetRoyaltyReportFinanceItem_ReplaceItem_Ok() {
	content := base64.StdEncoding.EncodeToString([]byte(``))
	cache := &mocks.CacheInterface{}
	cache.On(
		"Get",
		mock.Anything,
		mock.MatchedBy(func(input *RoyaltyReportFinance) bool {
			input.Items = append(
				input.Items,
				&postmarkpb.PayloadAttachment{
					Content:     content,
					Name:        "file_name.txt",
					ContentType: "text/plain",
				},
			)
			return true
		}),
	).Return(nil)
	cache.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.repository.cache = cache

	items, err := suite.repository.SetRoyaltyReportFinanceItem(
		"ffffffffffffffffffffffff",
		&postmarkpb.PayloadAttachment{
			Content:     content,
			Name:        "file_name.txt",
			ContentType: "text/plain",
		},
	)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), items, 1)
}

func (suite *RoyaltyReportTestSuite) TestRoyaltyReport_SetRoyaltyReportFinanceItem_Set_Error() {
	cache := &mocks.CacheInterface{}
	cache.On("Get", mock.Anything, mock.Anything).Return(nil)
	cache.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("set"))
	suite.repository.cache = cache

	_, err := suite.repository.SetRoyaltyReportFinanceItem(
		"ffffffffffffffffffffffff",
		&postmarkpb.PayloadAttachment{
			Content:     base64.StdEncoding.EncodeToString([]byte(``)),
			Name:        "file_name.txt",
			ContentType: "text/plain",
		},
	)
	assert.Error(suite.T(), err)
	assert.EqualError(suite.T(), err, "set")

	logs := suite.zapRecorder.All()
	assert.Len(suite.T(), logs, 1)
	assert.Equal(suite.T(), zapcore.ErrorLevel, logs[0].Level)
	assert.Equal(suite.T(), pkg.ErrorCacheQueryFailed, logs[0].Message)
}

func (suite *RoyaltyReportTestSuite) TestRoyaltyReport_RemoveRoyaltyReportFinanceItems_Ok() {
	err := suite.repository.RemoveRoyaltyReportFinanceItems("ffffffffffffffffffffffff")
	assert.NoError(suite.T(), err)
}

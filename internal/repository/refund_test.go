package repository

import (
	"context"
	"errors"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/paysuper/paysuper-billing-server/internal/config"
	"github.com/paysuper/paysuper-billing-server/internal/mocks"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
	"testing"
)

type RefundTestSuite struct {
	suite.Suite
	db         mongodb.SourceInterface
	repository *refundRepository
	log        *zap.Logger
}

func Test_Refund(t *testing.T) {
	suite.Run(t, new(RefundTestSuite))
}

func (suite *RefundTestSuite) SetupTest() {
	_, err := config.NewConfig()
	assert.NoError(suite.T(), err, "Config load failed")

	suite.log, err = zap.NewProduction()
	assert.NoError(suite.T(), err, "Logger initialization failed")

	suite.db, err = mongodb.NewDatabase()
	assert.NoError(suite.T(), err, "Database connection failed")

	suite.repository = NewRefundRepository(suite.db).(*refundRepository)
}

func (suite *RefundTestSuite) TearDownTest() {
	if err := suite.db.Drop(); err != nil {
		suite.FailNow("Database deletion failed", "%v", err)
	}

	if err := suite.db.Close(); err != nil {
		suite.FailNow("Database close failed", "%v", err)
	}
}

func (suite *CountryTestSuite) TestCountry_NewRefundRepository_Ok() {
	repository := NewRefundRepository(suite.db)
	assert.IsType(suite.T(), &refundRepository{}, repository)
}

func (suite *RefundTestSuite) TestRefund_Insert_Ok() {
	refund := &billingpb.Refund{
		Id:        primitive.NewObjectID().Hex(),
		CreatorId: primitive.NewObjectID().Hex(),
		OriginalOrder: &billingpb.RefundOrder{
			Id: primitive.NewObjectID().Hex(),
		},
	}
	err := suite.repository.Insert(context.TODO(), refund)
	assert.NoError(suite.T(), err)

	refund2, err := suite.repository.GetById(context.TODO(), refund.Id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), refund.Id, refund2.Id)
}

func (suite *RefundTestSuite) TestRefund_Insert_MapperError() {
	refund := &billingpb.Refund{
		Id:        primitive.NewObjectID().Hex(),
		CreatorId: primitive.NewObjectID().Hex(),
		OriginalOrder: &billingpb.RefundOrder{
			Id: primitive.NewObjectID().Hex(),
		},
	}

	mapper := &mocks.Mapper{}
	mapper.On("MapMgoToObject", mock.Anything).Return(nil, errors.New("error"))
	mapper.On("MapObjectToMgo", mock.Anything).Return(nil, errors.New("error"))

	suite.repository.mapper = mapper

	err := suite.repository.Insert(context.TODO(), refund)
	assert.Error(suite.T(), err)
}

// TODO: Use the DB mock for return error on insert entry
func (suite *RefundTestSuite) TestRefund_Insert_DatabaseError() {
	refund := &billingpb.Refund{
		Id:        primitive.NewObjectID().Hex(),
		CreatorId: primitive.NewObjectID().Hex(),
		OriginalOrder: &billingpb.RefundOrder{
			Id: primitive.NewObjectID().Hex(),
		},
		CreatedAt: &timestamp.Timestamp{Seconds: -100000000000000},
	}

	err := suite.repository.Insert(context.TODO(), refund)
	assert.Error(suite.T(), err)
}

// TODO: Use the DB mock for to skip really inserting the entry to DB
func (suite *RefundTestSuite) TestRefund_Insert_DontHaveDbErrorButDontInserted() {
	refund, err := suite.repository.GetById(context.TODO(), primitive.NewObjectID().Hex())
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), refund)
}

func (suite *RefundTestSuite) TestRefund_Update_Ok() {
	refund := &billingpb.Refund{
		Id:        primitive.NewObjectID().Hex(),
		CreatorId: primitive.NewObjectID().Hex(),
		OriginalOrder: &billingpb.RefundOrder{
			Id: primitive.NewObjectID().Hex(),
		},
		Reason: "test1",
	}
	err := suite.repository.Insert(context.TODO(), refund)
	assert.NoError(suite.T(), err)

	refund1, err := suite.repository.GetById(context.TODO(), refund.Id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), refund.Id, refund1.Id)
	assert.Equal(suite.T(), refund.Reason, refund1.Reason)

	refund.Reason = "test2"
	err = suite.repository.Update(context.TODO(), refund)
	assert.NoError(suite.T(), err)

	refund2, err := suite.repository.GetById(context.TODO(), refund.Id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), refund.Id, refund2.Id)
	assert.Equal(suite.T(), refund.Reason, refund2.Reason)
}

// TODO: Use the DB mock for return error on insert entry
func (suite *RefundTestSuite) TestRefund_Update_Error() {
	refund := &billingpb.Refund{
		Id:        primitive.NewObjectID().Hex(),
		CreatorId: primitive.NewObjectID().Hex(),
		OriginalOrder: &billingpb.RefundOrder{
			Id: primitive.NewObjectID().Hex(),
		},
	}
	err := suite.repository.Insert(context.TODO(), refund)
	assert.NoError(suite.T(), err)

	refund.CreatedAt = &timestamp.Timestamp{Seconds: -100000000000000}
	err = suite.repository.Update(context.TODO(), refund)
	assert.Error(suite.T(), err)
}

// TODO: Use the DB mock for to skip really updating the entry to DB
func (suite *RefundTestSuite) TestRefund_Update_DontHaveDbErrorButDontUpdated() {
	refund := &billingpb.Refund{
		Id:        primitive.NewObjectID().Hex(),
		CreatorId: primitive.NewObjectID().Hex(),
		OriginalOrder: &billingpb.RefundOrder{
			Id: primitive.NewObjectID().Hex(),
		},
		Reason: "test1",
	}
	err := suite.repository.Insert(context.TODO(), refund)
	assert.NoError(suite.T(), err)

	refund1, err := suite.repository.GetById(context.TODO(), refund.Id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), refund.Id, refund1.Id)
	assert.Equal(suite.T(), refund.Reason, refund1.Reason)

	refund.Reason = "test2"
	// TODO: Use the mock of DB
	//err = suite.repository.Update(context.TODO(), refund)
	//assert.NoError(suite.T(), err)

	refund2, err := suite.repository.GetById(context.TODO(), refund.Id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), refund.Id, refund2.Id)
	assert.NotEqual(suite.T(), refund.Reason, refund2.Reason)
}

func (suite *RefundTestSuite) TestRefund_GetById_Ok() {
	refund := &billingpb.Refund{
		Id:        primitive.NewObjectID().Hex(),
		CreatorId: primitive.NewObjectID().Hex(),
		OriginalOrder: &billingpb.RefundOrder{
			Id:   primitive.NewObjectID().Hex(),
			Uuid: "uuid",
		},
		Status:         1,
		Currency:       "CUR",
		Amount:         2,
		Reason:         "reason",
		CreatedOrderId: primitive.NewObjectID().Hex(),
		ExternalId:     primitive.NewObjectID().Hex(),
		IsChargeback:   true,
		SalesTax:       3,
		PayerData: &billingpb.RefundPayerData{
			Country: "CTR",
			State:   "state",
		},
		CreatedAt: &timestamp.Timestamp{Seconds: 100},
		UpdatedAt: &timestamp.Timestamp{Seconds: 100},
	}
	err := suite.repository.Insert(context.TODO(), refund)
	assert.NoError(suite.T(), err)

	refund2, err := suite.repository.GetById(context.TODO(), refund.Id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), refund, refund2)
}

func (suite *RefundTestSuite) TestRefund_GetById_MapperError() {
	refund := &billingpb.Refund{
		Id:        primitive.NewObjectID().Hex(),
		CreatorId: primitive.NewObjectID().Hex(),
		OriginalOrder: &billingpb.RefundOrder{
			Id:   primitive.NewObjectID().Hex(),
			Uuid: "uuid",
		},
		Status:         1,
		Currency:       "CUR",
		Amount:         2,
		Reason:         "reason",
		CreatedOrderId: primitive.NewObjectID().Hex(),
		ExternalId:     primitive.NewObjectID().Hex(),
		IsChargeback:   true,
		SalesTax:       3,
		PayerData: &billingpb.RefundPayerData{
			Country: "CTR",
			State:   "state",
		},
		CreatedAt: &timestamp.Timestamp{Seconds: 100},
		UpdatedAt: &timestamp.Timestamp{Seconds: 100},
	}
	err := suite.repository.Insert(context.TODO(), refund)
	assert.NoError(suite.T(), err)

	mapper := &mocks.Mapper{}
	mapper.On("MapMgoToObject", mock.Anything).Return(nil, errors.New("error"))
	mapper.On("MapObjectToMgo", mock.Anything).Return(nil, errors.New("error"))

	suite.repository.mapper = mapper

	refund2, err := suite.repository.GetById(context.TODO(), refund.Id)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), refund2)
}

func (suite *RefundTestSuite) TestRefund_GetById_ErrorNotFound() {
	refund := &billingpb.Refund{
		Id:        primitive.NewObjectID().Hex(),
		CreatorId: primitive.NewObjectID().Hex(),
		OriginalOrder: &billingpb.RefundOrder{
			Id:   primitive.NewObjectID().Hex(),
			Uuid: "uuid",
		},
		Status:         1,
		Currency:       "CUR",
		Amount:         2,
		Reason:         "reason",
		CreatedOrderId: primitive.NewObjectID().Hex(),
		ExternalId:     primitive.NewObjectID().Hex(),
		IsChargeback:   true,
		SalesTax:       3,
		PayerData: &billingpb.RefundPayerData{
			Country: "CTR",
			State:   "state",
		},
		CreatedAt: &timestamp.Timestamp{Seconds: 100},
		UpdatedAt: &timestamp.Timestamp{Seconds: 100},
	}
	err := suite.repository.Insert(context.TODO(), refund)
	assert.NoError(suite.T(), err)

	refund1, err := suite.repository.GetById(context.TODO(), refund.CreatorId)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), refund1)
}

func (suite *RefundTestSuite) TestRefund_FindByOrderId_Ok() {
	refund := &billingpb.Refund{
		Id:        primitive.NewObjectID().Hex(),
		CreatorId: primitive.NewObjectID().Hex(),
		OriginalOrder: &billingpb.RefundOrder{
			Id:   primitive.NewObjectID().Hex(),
			Uuid: "uuid",
		},
		Status:         1,
		Currency:       "CUR",
		Amount:         2,
		Reason:         "reason",
		CreatedOrderId: primitive.NewObjectID().Hex(),
		ExternalId:     primitive.NewObjectID().Hex(),
		IsChargeback:   true,
		SalesTax:       3,
		PayerData: &billingpb.RefundPayerData{
			Country: "CTR",
			State:   "state",
		},
		CreatedAt: &timestamp.Timestamp{Seconds: 100},
		UpdatedAt: &timestamp.Timestamp{Seconds: 100},
	}
	err := suite.repository.Insert(context.TODO(), refund)
	assert.NoError(suite.T(), err)

	refunds, err := suite.repository.FindByOrderUuid(context.TODO(), refund.OriginalOrder.Uuid, 1, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), refunds, 1)
	assert.Equal(suite.T(), refund, refunds[0])
}

func (suite *RefundTestSuite) TestRefund_FindByOrderId_MapperError() {
	refund := &billingpb.Refund{
		Id:        primitive.NewObjectID().Hex(),
		CreatorId: primitive.NewObjectID().Hex(),
		OriginalOrder: &billingpb.RefundOrder{
			Id:   primitive.NewObjectID().Hex(),
			Uuid: "uuid",
		},
		Status:         1,
		Currency:       "CUR",
		Amount:         2,
		Reason:         "reason",
		CreatedOrderId: primitive.NewObjectID().Hex(),
		ExternalId:     primitive.NewObjectID().Hex(),
		IsChargeback:   true,
		SalesTax:       3,
		PayerData: &billingpb.RefundPayerData{
			Country: "CTR",
			State:   "state",
		},
		CreatedAt: &timestamp.Timestamp{Seconds: 100},
		UpdatedAt: &timestamp.Timestamp{Seconds: 100},
	}
	err := suite.repository.Insert(context.TODO(), refund)
	assert.NoError(suite.T(), err)

	mapper := &mocks.Mapper{}
	mapper.On("MapMgoToObject", mock.Anything).Return(nil, errors.New("error"))
	mapper.On("MapObjectToMgo", mock.Anything).Return(nil, errors.New("error"))

	suite.repository.mapper = mapper

	refunds, err := suite.repository.FindByOrderUuid(context.TODO(), refund.OriginalOrder.Uuid, 1, 0)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), refunds)
}

func (suite *RefundTestSuite) TestRefund_FindByOrderId_Empty() {
	refund := &billingpb.Refund{
		Id:        primitive.NewObjectID().Hex(),
		CreatorId: primitive.NewObjectID().Hex(),
		OriginalOrder: &billingpb.RefundOrder{
			Id:   primitive.NewObjectID().Hex(),
			Uuid: "uuid",
		},
		Status:         1,
		Currency:       "CUR",
		Amount:         2,
		Reason:         "reason",
		CreatedOrderId: primitive.NewObjectID().Hex(),
		ExternalId:     primitive.NewObjectID().Hex(),
		IsChargeback:   true,
		SalesTax:       3,
		PayerData: &billingpb.RefundPayerData{
			Country: "CTR",
			State:   "state",
		},
		CreatedAt: &timestamp.Timestamp{Seconds: 100},
		UpdatedAt: &timestamp.Timestamp{Seconds: 100},
	}
	err := suite.repository.Insert(context.TODO(), refund)
	assert.NoError(suite.T(), err)

	refund1, err := suite.repository.FindByOrderUuid(context.TODO(), refund.Id, 1, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), refund1, 0)
}

func (suite *RefundTestSuite) TestRefund_CountByOrderUuid_Ok() {
	refund := &billingpb.Refund{
		Id:        primitive.NewObjectID().Hex(),
		CreatorId: primitive.NewObjectID().Hex(),
		OriginalOrder: &billingpb.RefundOrder{
			Id: primitive.NewObjectID().Hex(),
		},
	}
	err := suite.repository.Insert(context.TODO(), refund)
	assert.NoError(suite.T(), err)

	cnt, err := suite.repository.CountByOrderUuid(context.TODO(), refund.OriginalOrder.Uuid)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), cnt, int64(1))
}

func (suite *RefundTestSuite) TestRefund_CountByOrderUuid_Empty() {
	refund := &billingpb.Refund{
		Id:        primitive.NewObjectID().Hex(),
		CreatorId: primitive.NewObjectID().Hex(),
		OriginalOrder: &billingpb.RefundOrder{
			Id: primitive.NewObjectID().Hex(),
		},
	}
	err := suite.repository.Insert(context.TODO(), refund)
	assert.NoError(suite.T(), err)

	cnt, err := suite.repository.CountByOrderUuid(context.TODO(), refund.Id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), cnt, int64(0))
}

func (suite *RefundTestSuite) TestRefund_GetAmountByOrderId_Ok() {
	refund := &billingpb.Refund{
		Id:        primitive.NewObjectID().Hex(),
		CreatorId: primitive.NewObjectID().Hex(),
		OriginalOrder: &billingpb.RefundOrder{
			Id: primitive.NewObjectID().Hex(),
		},
		Status: pkg.RefundStatusCompleted,
		Amount: 42,
	}
	err := suite.repository.Insert(context.TODO(), refund)
	assert.NoError(suite.T(), err)

	amount, err := suite.repository.GetAmountByOrderId(context.TODO(), refund.OriginalOrder.Id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(42), amount)
}

func (suite *RefundTestSuite) TestRefund_GetAmountByOrderId_SkipRejectStatus() {
	refund := &billingpb.Refund{
		Id:        primitive.NewObjectID().Hex(),
		CreatorId: primitive.NewObjectID().Hex(),
		OriginalOrder: &billingpb.RefundOrder{
			Id: primitive.NewObjectID().Hex(),
		},
		Status: pkg.RefundStatusRejected,
		Amount: 22,
	}
	err := suite.repository.Insert(context.TODO(), refund)
	assert.NoError(suite.T(), err)

	amount, err := suite.repository.GetAmountByOrderId(context.TODO(), refund.OriginalOrder.Id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(0), amount)
}

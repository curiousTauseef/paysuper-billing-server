package repository

import (
	"context"
	"github.com/paysuper/paysuper-billing-server/internal/repository/models"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
)

const (
	// CollectionRefund is name of table for collection the refund.
	CollectionRefund = "refund"
)

type refundRepository repository

// NewRefundRepository create and return an object for working with the refund repository.
// The returned object implements the RefundRepositoryInterface interface.
func NewRefundRepository(db mongodb.SourceInterface) RefundRepositoryInterface {
	s := &refundRepository{db: db, mapper: models.NewRefundMapper()}
	return s
}

func (h *refundRepository) Insert(ctx context.Context, refund *billingpb.Refund) error {
	mgo, err := h.mapper.MapObjectToMgo(refund)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, refund),
		)
		return err
	}

	_, err = h.db.Collection(CollectionRefund).InsertOne(ctx, mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionRefund),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, refund),
		)
		return err
	}

	return nil
}

func (h *refundRepository) Update(ctx context.Context, refund *billingpb.Refund) error {
	oid, err := primitive.ObjectIDFromHex(refund.Id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionRefund),
			zap.String(pkg.ErrorDatabaseFieldQuery, refund.Id),
		)
		return err
	}

	mgo, err := h.mapper.MapObjectToMgo(refund)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, refund),
		)
		return err
	}

	_, err = h.db.Collection(CollectionRefund).ReplaceOne(ctx, bson.M{"_id": oid}, mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionRefund),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationUpdate),
			zap.Any(pkg.ErrorDatabaseFieldQuery, refund),
		)
		return err
	}

	return nil
}

func (h *refundRepository) GetById(ctx context.Context, id string) (*billingpb.Refund, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionRefund),
			zap.String(pkg.ErrorDatabaseFieldQuery, id),
		)
		return nil, err
	}

	mgo := &models.MgoRefund{}
	query := bson.M{"_id": oid}
	err = h.db.Collection(CollectionRefund).FindOne(ctx, query).Decode(mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionRefund),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	obj, err := h.mapper.MapMgoToObject(mgo)
	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, mgo),
		)
		return nil, err
	}

	return obj.(*billingpb.Refund), nil
}

func (h *refundRepository) FindByOrderUuid(ctx context.Context, id string, limit int64, offset int64) ([]*billingpb.Refund, error) {
	query := bson.M{"original_order.uuid": id}
	opts := options.Find().
		SetLimit(limit).
		SetSkip(offset)
	cursor, err := h.db.Collection(CollectionRefund).Find(ctx, query, opts)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionRefund),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	var mgoRefunds []*models.MgoRefund
	err = cursor.All(ctx, &mgoRefunds)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionRefund),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	refunds := make([]*billingpb.Refund, len(mgoRefunds))
	for i, mgo := range mgoRefunds {
		obj, err := h.mapper.MapMgoToObject(mgo)
		if err != nil {
			zap.L().Error(
				pkg.ErrorDatabaseMapModelFailed,
				zap.Error(err),
				zap.Any(pkg.ErrorDatabaseFieldQuery, mgo),
			)
			return nil, err
		}
		refunds[i] = obj.(*billingpb.Refund)
	}

	return refunds, nil
}

func (h *refundRepository) CountByOrderUuid(ctx context.Context, id string) (int64, error) {
	query := bson.M{"original_order.uuid": id}
	count, err := h.db.Collection(CollectionRefund).CountDocuments(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionRefund),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return 0, err
	}

	return count, nil
}

func (h *refundRepository) GetAmountByOrderId(ctx context.Context, orderId string) (float64, error) {
	var res struct {
		Amount float64 `bson:"amount"`
	}

	oid, _ := primitive.ObjectIDFromHex(orderId)
	query := []bson.M{
		{
			"$match": bson.M{
				"status":            bson.M{"$nin": []int32{pkg.RefundStatusRejected}},
				"original_order.id": oid,
			},
		},
		{"$group": bson.M{"_id": "$original_order.id", "amount": bson.M{"$sum": "$amount"}}},
	}

	cursor, err := h.db.Collection(CollectionRefund).Aggregate(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionRefund),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return 0, err
	}

	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		err = cursor.Decode(&res)

		if err != nil {
			zap.L().Error(
				pkg.ErrorQueryCursorExecutionFailed,
				zap.Error(err),
				zap.String(pkg.ErrorDatabaseFieldCollection, CollectionRefund),
				zap.Any(pkg.ErrorDatabaseFieldQuery, query),
			)
			return 0, err
		}
	}

	return res.Amount, nil
}

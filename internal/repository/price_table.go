package repository

import (
	"context"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
)

type priceTableRepository repository

// NewPriceTableRepository create and return an object for working with the price table repository.
// The returned object implements the PriceTableRepositoryInterface interface.
func NewPriceTableRepository(db mongodb.SourceInterface) PriceTableRepositoryInterface {
	s := &priceTableRepository{db: db}
	return s
}

func (r *priceTableRepository) Insert(ctx context.Context, obj *billingpb.PriceTable) error {
	_, err := r.db.Collection(collectionPriceTable).InsertOne(ctx, obj)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPriceTable),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, obj),
		)
		return err
	}

	return nil
}

func (r priceTableRepository) GetByRegion(ctx context.Context, region string) (*billingpb.PriceTable, error) {
	var obj billingpb.PriceTable

	query := bson.M{"currency": region}
	err := r.db.Collection(collectionPriceTable).FindOne(ctx, query).Decode(&obj)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPriceTable),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	return &obj, nil
}

package repository

import (
	"context"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
)

const (
	collectionNotifyNewRegion = "notify_new_region"
)

type notifyRegionRepository repository

// NewNotifyRegionRepository create and return an object for working with the notify new region repository.
// The returned object implements the NotifyRegionRepositoryInterface interface.
func NewNotifyRegionRepository(db mongodb.SourceInterface) NotifyRegionRepositoryInterface {
	s := &notifyRegionRepository{db: db}
	return s
}

func (r *notifyRegionRepository) Insert(ctx context.Context, obj *billingpb.NotifyUserNewRegion) error {
	_, err := r.db.Collection(collectionNotifyNewRegion).InsertOne(ctx, obj)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionNotifyNewRegion),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, obj),
		)
		return err
	}

	return nil
}

func (r notifyRegionRepository) FindByEmail(ctx context.Context, email string) ([]*billingpb.NotifyUserNewRegion, error) {
	query := bson.M{"email": email}
	cursor, err := r.db.Collection(collectionNotifyNewRegion).Find(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionNotifyNewRegion),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	var list []*billingpb.NotifyUserNewRegion

	err = cursor.All(ctx, &list)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionNotifyNewRegion),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	return list, nil
}

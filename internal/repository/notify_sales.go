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
	collectionNotifySales = "notify_sales"
)

type notifySalesRepository repository

// NewNotifySalesRepository create and return an object for working with the notify sales repository.
// The returned object implements the NotifySalesRepositoryInterface interface.
func NewNotifySalesRepository(db mongodb.SourceInterface) NotifySalesRepositoryInterface {
	s := &notifySalesRepository{db: db}
	return s
}

func (r *notifySalesRepository) Insert(ctx context.Context, obj *billingpb.NotifyUserSales) error {
	_, err := r.db.Collection(collectionNotifySales).InsertOne(ctx, obj)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionNotifySales),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, obj),
		)
		return err
	}

	return nil
}

func (r notifySalesRepository) FindByEmail(ctx context.Context, email string) ([]*billingpb.NotifyUserSales, error) {
	query := bson.M{"email": email}
	cursor, err := r.db.Collection(collectionNotifySales).Find(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionNotifySales),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	var list []*billingpb.NotifyUserSales

	err = cursor.All(ctx, &list)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionNotifySales),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	return list, nil
}

package repository

import (
	"context"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
)

const (
	collectionMerchantPaymentMethodHistory = "payment_method_history"
)

type merchantPaymentMethodHistory repository

// NewMerchantPaymentMethodHistoryRepository create and return an object for working with the merchant tariffs settings repository.
// The returned object implements the MerchantPaymentMethodHistoryRepositoryInterface interface.
func NewMerchantPaymentMethodHistoryRepository(db mongodb.SourceInterface) MerchantPaymentMethodHistoryRepositoryInterface {
	s := &merchantPaymentMethodHistory{db: db}
	return s
}

func (r *merchantPaymentMethodHistory) Insert(ctx context.Context, obj *billingpb.MerchantPaymentMethodHistory) error {
	_, err := r.db.Collection(collectionMerchantPaymentMethodHistory).InsertOne(ctx, obj)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionMerchantPaymentMethodHistory),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, obj),
		)
		return err
	}

	return nil
}

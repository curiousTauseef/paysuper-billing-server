package repository

import (
	"context"
	"fmt"
	"github.com/paysuper/paysuper-billing-server/internal/database"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
)

const (
	collectionMerchantsPaymentTariffs = "merchants_payment_tariffs"

	onboardingTariffRatesCacheKeyGetBy = "onboarding_tariff_rates:%s:%s:%s:%f:%f"
)

type merchantPaymentTariffsRepository repository

// NewMerchantTariffsSettingsRepository create and return an object for working with the merchant payment tariffs repository.
// The returned object implements the MerchantTariffsSettingsInterface interface.
func NewMerchantPaymentTariffsRepository(db mongodb.SourceInterface, cache database.CacheInterface) MerchantPaymentTariffsInterface {
	s := &merchantPaymentTariffsRepository{db: db, cache: cache}
	return s
}

func (r *merchantPaymentTariffsRepository) MultipleInsert(ctx context.Context, objs []*billingpb.MerchantTariffRatesPayment) error {
	m := make([]interface{}, len(objs))
	for i, v := range objs {
		m[i] = v
	}

	_, err := r.db.Collection(collectionMerchantsPaymentTariffs).InsertMany(ctx, m)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionMerchantsPaymentTariffs),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, m),
		)
		return err
	}

	return nil
}

func (r *merchantPaymentTariffsRepository) Find(
	ctx context.Context, homeRegion, payerRegion, mccCode string, minAmount, maxAmount float64,
) ([]*billingpb.MerchantTariffRatesPayment, error) {
	var items []*billingpb.MerchantTariffRatesPayment
	key := fmt.Sprintf(onboardingTariffRatesCacheKeyGetBy, homeRegion, payerRegion, mccCode, minAmount, maxAmount)

	if err := r.cache.Get(key, &items); err == nil {
		return items, nil
	}

	query := bson.M{
		"merchant_home_region": homeRegion,
		"mcc_code":             mccCode,
		"is_active":            true,
	}

	if payerRegion != "" {
		query["payer_region"] = payerRegion
	}

	if minAmount >= 0 && maxAmount > minAmount {
		query["min_amount"] = minAmount
		query["max_amount"] = maxAmount
	}

	opts := options.Find().
		SetSort(bson.M{
			"min_amount":   1,
			"max_amount":   1,
			"payer_region": 1,
			"position":     1,
		})
	cursor, err := r.db.Collection(collectionMerchantsPaymentTariffs).Find(ctx, query, opts)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionMerchantsPaymentTariffs),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	err = cursor.All(ctx, &items)
	_ = cursor.Close(ctx)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionMerchantsPaymentTariffs),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	err = r.cache.Set(key, items, 0)

	if err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorDatabaseFieldQuery, items),
		)

		return nil, err
	}

	return items, nil
}

package repository

import (
	"context"
	"fmt"
	"github.com/paysuper/paysuper-billing-server/internal/database"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
)

const (
	collectionMerchantTariffsSettings = "merchant_tariffs_settings"

	merchantTariffsSettingsKey = "merchant_tariffs_settings:mcc:%s"
)

type merchantTariffsSettingsRepository repository

// NewMerchantTariffsSettingsRepository create and return an object for working with the merchant tariffs settings repository.
// The returned object implements the MerchantTariffsSettingsInterface interface.
func NewMerchantTariffsSettingsRepository(db mongodb.SourceInterface, cache database.CacheInterface) MerchantTariffsSettingsInterface {
	s := &merchantTariffsSettingsRepository{db: db, cache: cache}
	return s
}

func (r *merchantTariffsSettingsRepository) Insert(ctx context.Context, obj *billingpb.MerchantTariffRatesSettings) error {
	_, err := r.db.Collection(collectionMerchantTariffsSettings).InsertOne(ctx, obj)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionMerchantTariffsSettings),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, obj),
		)
		return err
	}

	return nil
}

func (r *merchantTariffsSettingsRepository) GetByMccCode(ctx context.Context, code string) (*billingpb.MerchantTariffRatesSettings, error) {
	item := new(billingpb.MerchantTariffRatesSettings)
	key := fmt.Sprintf(merchantTariffsSettingsKey, code)
	err := r.cache.Get(key, &item)

	if err == nil {
		return item, nil
	}

	query := bson.M{"mcc_code": code}
	err = r.db.Collection(collectionMerchantTariffsSettings).FindOne(ctx, query).Decode(item)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionMerchantTariffsSettings),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	err = r.cache.Set(key, item, 0)

	if err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorDatabaseFieldQuery, item),
		)
	}

	return item, nil
}

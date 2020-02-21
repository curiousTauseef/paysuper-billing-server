package repository

import (
	"context"
	"fmt"
	"github.com/paysuper/paysuper-billing-server/internal/database"
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
	CollectionMerchant = "merchant"

	cacheMerchantId       = "merchant:id:%s"
	cacheMerchantCommonId = "merchant:common:id:%s"
)

type merchantRepository repository

// NewMerchantRepository create and return an object for working with the merchant repository.
// The returned object implements the MerchantRepositoryInterface interface.
func NewMerchantRepository(db mongodb.SourceInterface, cache database.CacheInterface) MerchantRepositoryInterface {
	s := &merchantRepository{db: db, cache: cache, mapper: models.NewMerchantMapper()}
	return s
}

func (h *merchantRepository) Insert(ctx context.Context, merchant *billingpb.Merchant) error {
	mgo, err := h.mapper.MapObjectToMgo(merchant)
	if err != nil {
		zap.L().Error(
			pkg.ErrorMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, merchant),
		)
		return err
	}

	_, err = h.db.Collection(CollectionMerchant).InsertOne(ctx, mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, merchant),
		)
		return err
	}

	key := fmt.Sprintf(cacheMerchantId, merchant.Id)
	err = h.cache.Set(key, merchant, 0)

	if err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorDatabaseFieldQuery, merchant),
		)
		return err
	}

	return nil
}

func (h *merchantRepository) MultipleInsert(ctx context.Context, merchants []*billingpb.Merchant) error {
	m := make([]interface{}, len(merchants))
	for i, v := range merchants {
		mgo, err := h.mapper.MapObjectToMgo(v)
		if err != nil {
			zap.L().Error(
				pkg.ErrorMapModelFailed,
				zap.Error(err),
				zap.Any(pkg.ErrorDatabaseFieldQuery, v),
			)
			return err
		}

		m[i] = mgo
	}

	_, err := h.db.Collection(CollectionMerchant).InsertMany(ctx, m)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, m),
		)
		return err
	}

	return nil
}

func (h *merchantRepository) Update(ctx context.Context, merchant *billingpb.Merchant) error {
	oid, err := primitive.ObjectIDFromHex(merchant.Id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
			zap.String(pkg.ErrorDatabaseFieldQuery, merchant.Id),
		)
		return err
	}

	mgo, err := h.mapper.MapObjectToMgo(merchant)
	if err != nil {
		zap.L().Error(
			pkg.ErrorMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, merchant),
		)
		return err
	}

	filter := bson.M{"_id": oid}
	_, err = h.db.Collection(CollectionMerchant).ReplaceOne(ctx, filter, mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
			zap.Any(pkg.ErrorDatabaseFieldQuery, merchant),
		)
		return err
	}

	key := fmt.Sprintf(cacheMerchantId, merchant.Id)
	err = h.cache.Set(key, merchant, 0)

	if err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorDatabaseFieldQuery, merchant),
		)
		return err
	}

	return nil
}

func (h *merchantRepository) Upsert(ctx context.Context, merchant *billingpb.Merchant) error {
	oid, err := primitive.ObjectIDFromHex(merchant.Id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
			zap.String(pkg.ErrorDatabaseFieldQuery, merchant.Id),
		)
		return err
	}

	mgo, err := h.mapper.MapObjectToMgo(merchant)
	if err != nil {
		zap.L().Error(
			pkg.ErrorMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, merchant),
		)
		return err
	}

	filter := bson.M{"_id": oid}
	opts := options.Replace().SetUpsert(true)
	_, err = h.db.Collection(CollectionMerchant).ReplaceOne(ctx, filter, mgo, opts)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
			zap.Any(pkg.ErrorDatabaseFieldQuery, merchant),
		)
		return err
	}

	key := fmt.Sprintf(cacheMerchantId, merchant.Id)
	err = h.cache.Set(key, merchant, 0)

	if err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorDatabaseFieldQuery, merchant),
		)
		return err
	}

	return nil
}

func (h *merchantRepository) UpdateTariffs(ctx context.Context, merchantId string, tariff *billingpb.PaymentChannelCostMerchant) error {
	set := bson.M{}
	oid, err := primitive.ObjectIDFromHex(merchantId)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
			zap.String(pkg.ErrorDatabaseFieldQuery, merchantId),
		)
		return err
	}

	query := bson.M{
		"_id": oid,
		"tariff.payment": bson.M{
			"$elemMatch": bson.M{
				"method_name":               primitive.Regex{Pattern: tariff.Name, Options: "i"},
				"payer_region":              tariff.Region,
				"mcc_code":                  tariff.MccCode,
				"method_fixed_fee_currency": tariff.MethodFixAmountCurrency,
				"ps_fixed_fee_currency":     tariff.PsFixedFeeCurrency,
			},
		},
	}

	set["$set"] = bson.M{
		"tariff.payment.$.min_amount":         tariff.MinAmount,
		"tariff.payment.$.method_percent_fee": tariff.MethodPercent,
		"tariff.payment.$.method_fixed_fee":   tariff.MethodFixAmount,
		"tariff.payment.$.ps_fixed_fee":       tariff.PsFixedFee,
		"tariff.payment.$.ps_percent_fee":     tariff.PsPercent,
		"tariff.payment.$.is_active":          tariff.IsActive,
	}

	_, err = h.db.Collection(CollectionMerchant).UpdateOne(ctx, query, set)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
			zap.Any(pkg.ErrorDatabaseFieldSet, set),
		)
		return err
	}

	// We need to remove from cache partially updated entity because of multiple parallel update queries
	key := fmt.Sprintf(cacheMerchantId, merchantId)

	if err := h.cache.Delete(key); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "DELETE"),
			zap.String(pkg.ErrorCacheFieldKey, key),
		)
		return err
	}

	return nil
}

func (h *merchantRepository) GetMerchantsWithAutoPayouts(ctx context.Context) (merchants []*billingpb.Merchant, err error) {
	query := bson.M{
		"manual_payouts_enabled": false,
	}
	cursor, err := h.db.Collection(CollectionMerchant).Find(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return
	}

	var mgoMerchants []*models.MgoMerchant
	err = cursor.All(ctx, &mgoMerchants)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return
	}

	merchants = make([]*billingpb.Merchant, len(mgoMerchants))
	for i, merchant := range mgoMerchants {
		obj, err := h.mapper.MapMgoToObject(merchant)
		if err != nil {
			zap.L().Error(
				pkg.ErrorMapModelFailed,
				zap.Error(err),
				zap.Any(pkg.ErrorDatabaseFieldQuery, merchant),
			)
			return nil, err
		}
		merchants[i] = obj.(*billingpb.Merchant)
	}

	return merchants, nil
}

func (h *merchantRepository) GetById(ctx context.Context, id string) (*billingpb.Merchant, error) {
	merchant := &billingpb.Merchant{}
	key := fmt.Sprintf(cacheMerchantId, id)

	if err := h.cache.Get(key, merchant); err == nil {
		return merchant, nil
	}

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
			zap.String(pkg.ErrorDatabaseFieldQuery, id),
		)
		return nil, err
	}

	query := bson.M{"_id": oid}
	mgo := &models.MgoMerchant{}
	err = h.db.Collection(CollectionMerchant).FindOne(ctx, query).Decode(mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	obj, err := h.mapper.MapMgoToObject(mgo)
	if err != nil {
		zap.L().Error(
			pkg.ErrorMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, merchant),
		)
		return nil, err
	}

	merchant = obj.(*billingpb.Merchant)

	if err := h.cache.Set(key, merchant, 0); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorCacheFieldData, merchant),
		)
	}

	return merchant, nil
}

func (h *merchantRepository) GetByUserId(ctx context.Context, userId string) (*billingpb.Merchant, error) {
	merchant := &billingpb.Merchant{}

	mgo := &models.MgoMerchant{}
	query := bson.M{"user.id": userId}
	err := h.db.Collection(CollectionMerchant).FindOne(ctx, query).Decode(mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	obj, err := h.mapper.MapMgoToObject(mgo)
	if err != nil {
		zap.L().Error(
			pkg.ErrorMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, merchant),
		)
		return nil, err
	}

	merchant = obj.(*billingpb.Merchant)

	return merchant, nil
}

func (h *merchantRepository) GetCommonById(ctx context.Context, id string) (*billingpb.MerchantCommon, error) {
	merchant := &billingpb.MerchantCommon{}
	key := fmt.Sprintf(cacheMerchantCommonId, id)

	if err := h.cache.Get(key, merchant); err == nil {
		return merchant, nil
	}

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
			zap.String(pkg.ErrorDatabaseFieldQuery, id),
		)
		return nil, err
	}

	mgo := &models.MgoMerchantCommon{}
	query := bson.M{"_id": oid}
	err = h.db.Collection(CollectionMerchant).FindOne(ctx, query).Decode(mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	m := models.NewCommonMerchantMapper()
	obj, err := m.MapMgoToObject(mgo)
	if err != nil {
		zap.L().Error(
			pkg.ErrorMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, mgo),
		)
		return nil, err
	}

	merchant = obj.(*billingpb.MerchantCommon)

	if err := h.cache.Set(key, merchant, 0); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorCacheFieldData, merchant),
		)
	}

	return merchant, nil
}

func (h *merchantRepository) GetAll(ctx context.Context) ([]*billingpb.Merchant, error) {
	cursor, err := h.db.Collection(CollectionMerchant).Find(ctx, bson.D{})

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
		)
		return nil, err
	}


	var mgoMerchants []*models.MgoMerchant
	err = cursor.All(ctx, &mgoMerchants)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
		)
		return nil, err
	}

	merchants := make([]*billingpb.Merchant, len(mgoMerchants))
	for i, merchant := range mgoMerchants {
		obj, err := h.mapper.MapMgoToObject(merchant)
		if err != nil {
			zap.L().Error(
				pkg.ErrorMapModelFailed,
				zap.Error(err),
				zap.Any(pkg.ErrorDatabaseFieldQuery, merchant),
			)
			return nil, err
		}
		merchants[i] = obj.(*billingpb.Merchant)
	}

	return merchants, nil
}

func (h *merchantRepository) Find(ctx context.Context, query bson.M, sort []string, offset, limit int64) ([]*billingpb.Merchant, error) {
	opts := options.Find().
		SetSort(mongodb.ToSortOption(sort)).
		SetLimit(limit).
		SetSkip(offset)
	cursor, err := h.db.Collection(CollectionMerchant).Find(ctx, query, opts)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	var mgoMerchants []*models.MgoMerchant
	err = cursor.All(ctx, &mgoMerchants)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	merchants := make([]*billingpb.Merchant, len(mgoMerchants))
	for i, merchant := range mgoMerchants {
		obj, err := h.mapper.MapMgoToObject(merchant)
		if err != nil {
			zap.L().Error(
				pkg.ErrorMapModelFailed,
				zap.Error(err),
				zap.Any(pkg.ErrorDatabaseFieldQuery, merchant),
			)
			return nil, err
		}
		merchants[i] = obj.(*billingpb.Merchant)
	}

	return merchants, nil
}

func (h *merchantRepository) FindCount(ctx context.Context, query bson.M) (int64, error) {
	count, err := h.db.Collection(CollectionMerchant).CountDocuments(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, CollectionMerchant),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return int64(0), err
	}

	return count, nil
}

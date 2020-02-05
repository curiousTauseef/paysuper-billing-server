package repository

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-billing-server/internal/database"
	internalPkg "github.com/paysuper/paysuper-billing-server/internal/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	tools "github.com/paysuper/paysuper-tools/number"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
)

const (
	cachePaymentChannelCostSystemKey   = "pccs:n:%s:r:%s:c:%s:mcc:%s:oc:%s"
	cachePaymentChannelCostSystemKeyId = "pccs:id:%s"
	cachePaymentChannelCostSystemAll   = "pccs:all"

	collectionPaymentChannelCostSystem = "payment_channel_cost_system"
)

type paymentChannelCostSystemRepository repository

// NewPaymentChannelCostSystemRepository create and return an object for working with the price group repository.
// The returned object implements the PaymentChannelCostSystemRepositoryInterface interface.
func NewPaymentChannelCostSystemRepository(db mongodb.SourceInterface, cache database.CacheInterface) PaymentChannelCostSystemRepositoryInterface {
	s := &paymentChannelCostSystemRepository{db: db, cache: cache}
	return s
}

func (r *paymentChannelCostSystemRepository) Insert(ctx context.Context, obj *billingpb.PaymentChannelCostSystem) error {
	obj.FixAmount = tools.FormatAmount(obj.FixAmount)
	obj.Percent = tools.ToPrecise(obj.Percent)
	obj.CreatedAt = ptypes.TimestampNow()
	obj.UpdatedAt = ptypes.TimestampNow()
	obj.IsActive = true

	_, err := r.db.Collection(collectionPaymentChannelCostSystem).InsertOne(ctx, obj)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentChannelCostSystem),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, obj),
		)
		return err
	}

	return r.updateCaches(obj)
}

func (r *paymentChannelCostSystemRepository) MultipleInsert(ctx context.Context, obj []*billingpb.PaymentChannelCostSystem) error {
	c := make([]interface{}, len(obj))
	for i, v := range obj {
		if v.Id == "" {
			v.Id = primitive.NewObjectID().Hex()
		}
		v.FixAmount = tools.FormatAmount(v.FixAmount)
		v.Percent = tools.ToPrecise(v.Percent)
		v.CreatedAt = ptypes.TimestampNow()
		v.UpdatedAt = ptypes.TimestampNow()
		v.IsActive = true
		c[i] = v
	}

	_, err := r.db.Collection(collectionPaymentChannelCostSystem).InsertMany(ctx, c)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentChannelCostSystem),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, c),
		)
		return err
	}

	for _, v := range obj {
		if err := r.updateCaches(v); err != nil {
			return err
		}
	}

	return nil
}

func (r *paymentChannelCostSystemRepository) Update(ctx context.Context, obj *billingpb.PaymentChannelCostSystem) error {
	obj.FixAmount = tools.FormatAmount(obj.FixAmount)
	obj.Percent = tools.ToPrecise(obj.Percent)
	obj.UpdatedAt = ptypes.TimestampNow()
	obj.IsActive = true

	oid, err := primitive.ObjectIDFromHex(obj.Id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentChannelCostSystem),
			zap.String(pkg.ErrorDatabaseFieldQuery, obj.Id),
		)
	}

	filter := bson.M{"_id": oid}
	_, err = r.db.Collection(collectionPaymentChannelCostSystem).ReplaceOne(ctx, filter, obj)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentChannelCostSystem),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationUpdate),
			zap.Any(pkg.ErrorDatabaseFieldQuery, obj),
		)
		return err
	}

	return r.updateCaches(obj)
}

func (r *paymentChannelCostSystemRepository) GetById(ctx context.Context, id string) (*billingpb.PaymentChannelCostSystem, error) {
	var c billingpb.PaymentChannelCostSystem
	key := fmt.Sprintf(cachePaymentChannelCostSystemKeyId, id)

	if err := r.cache.Get(key, c); err == nil {
		return &c, nil
	}

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentChannelCostSystem),
			zap.String(pkg.ErrorDatabaseFieldQuery, id),
		)
	}

	filter := bson.M{"_id": oid, "is_active": true}
	err = r.db.Collection(collectionPaymentChannelCostSystem).FindOne(ctx, filter).Decode(&c)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentChannelCostSystem),
			zap.Any(pkg.ErrorDatabaseFieldQuery, filter),
		)
		return nil, err
	}

	_ = r.cache.Set(key, c, 0)

	return &c, nil
}

func (r *paymentChannelCostSystemRepository) Find(
	ctx context.Context, name, region, country, mccCode, operatingCompanyId string,
) (*billingpb.PaymentChannelCostSystem, error) {
	var c *billingpb.PaymentChannelCostSystem
	key := fmt.Sprintf(cachePaymentChannelCostSystemKey, name, region, country, mccCode, operatingCompanyId)

	if err := r.cache.Get(key, c); err == nil {
		return c, nil
	}

	matchQuery := bson.M{
		"name":                 primitive.Regex{Pattern: "^" + name + "$", Options: "i"},
		"mcc_code":             mccCode,
		"operating_company_id": operatingCompanyId,
		"is_active":            true,
		"$or": []bson.M{
			{
				"country": country,
				"region":  region,
			},
			{
				"$or": []bson.M{
					{"country": ""},
					{"country": bson.M{"$exists": false}},
				},
				"region": region,
			},
		},
	}

	query := []bson.M{
		{
			"$match": matchQuery,
		},
		{
			"$group": bson.M{
				"_id": "$country",
				"set": bson.M{"$push": "$$ROOT"},
			},
		},
		{
			"$sort": bson.M{"_id": -1},
		},
		{
			"$limit": 1,
		},
	}

	set := &internalPkg.PaymentChannelCostSystemSet{}
	cursor, err := r.db.Collection(collectionPaymentChannelCostSystem).Aggregate(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentChannelCostSystem),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		err = cursor.Decode(&set)
		if err != nil {
			zap.L().Error(
				pkg.ErrorQueryCursorExecutionFailed,
				zap.Error(err),
				zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentChannelCostSystem),
				zap.Any(pkg.ErrorDatabaseFieldQuery, query),
			)
			return nil, err
		}
	}

	if len(set.Set) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	c = set.Set[0]

	if err = r.cache.Set(key, c, 0); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorCacheFieldData, c),
		)
	}

	return c, nil
}

func (r *paymentChannelCostSystemRepository) Delete(ctx context.Context, obj *billingpb.PaymentChannelCostSystem) error {
	obj.UpdatedAt = ptypes.TimestampNow()
	obj.IsActive = false

	oid, err := primitive.ObjectIDFromHex(obj.Id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentChannelCostSystem),
			zap.String(pkg.ErrorDatabaseFieldQuery, obj.Id),
		)
	}

	filter := bson.M{"_id": oid}
	_, err = r.db.Collection(collectionPaymentChannelCostSystem).ReplaceOne(ctx, filter, obj)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationDelete),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentChannelCostSystem),
			zap.Any(pkg.ErrorDatabaseFieldQuery, filter),
		)
		return err
	}

	return r.updateCaches(obj)
}

func (r *paymentChannelCostSystemRepository) GetAll(ctx context.Context) ([]*billingpb.PaymentChannelCostSystem, error) {
	c := []*billingpb.PaymentChannelCostSystem{}
	key := cachePaymentChannelCostSystemAll
	err := r.cache.Get(key, &c)

	if err == nil {
		return c, nil
	}

	query := bson.M{"is_active": true}
	opts := options.Find().SetSort(bson.M{"name": 1, "region": 1, "country": 1})
	cursor, err := r.db.Collection(collectionPaymentChannelCostSystem).Find(ctx, query, opts)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentChannelCostSystem),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	err = cursor.All(ctx, &c)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentChannelCostSystem),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	err = r.cache.Set(key, c, 0)

	if err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorCacheFieldData, c),
		)
	}

	return c, nil
}

func (r *paymentChannelCostSystemRepository) updateCaches(obj *billingpb.PaymentChannelCostSystem) (err error) {
	groupKeys := []string{
		fmt.Sprintf(cachePaymentChannelCostSystemKey, obj.Name, obj.Region, obj.Country, obj.MccCode, obj.OperatingCompanyId),
		fmt.Sprintf(cachePaymentChannelCostSystemKey, obj.Name, obj.Region, "", obj.MccCode, obj.OperatingCompanyId),
		cachePaymentChannelCostSystemAll,
	}
	for _, key := range groupKeys {
		err = r.cache.Delete(key)
		if err != nil {
			return
		}
	}

	keys := []string{
		fmt.Sprintf(cachePaymentChannelCostSystemKeyId, obj.Id),
	}

	for _, key := range keys {
		err = r.cache.Delete(key)
		if err != nil {
			return
		}
	}

	if obj.IsActive {
		for _, key := range keys {
			err = r.cache.Set(key, obj, 0)
			if err != nil {
				zap.L().Error(
					pkg.ErrorCacheQueryFailed,
					zap.Error(err),
					zap.String(pkg.ErrorCacheFieldCmd, "SET"),
					zap.String(pkg.ErrorCacheFieldKey, key),
					zap.Any(pkg.ErrorCacheFieldData, obj),
				)
				return
			}
		}
		return
	}

	return
}

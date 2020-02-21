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
	collectionPaymentMethod = "payment_method"

	cachePaymentMethodId                     = "payment_method:id:%s"
	cachePaymentMethodGroup                  = "payment_method:group:%s"
	cachePaymentMethodAll                    = "payment_method:all"
	cachePaymentMethodModeCurrencyMccCompany = "payment_method:mode:%s:currency:%s:mcc:%s:oc:%s"

	fieldTestSettings       = "test_settings"
	fieldProductionSettings = "production_settings"
)

type paymentMethodRepository repository

// NewPaymentMethodRepository create and return an object for working with the payment method repository.
// The returned object implements the PaymentMethodRepositoryInterface interface.
func NewPaymentMethodRepository(db mongodb.SourceInterface, cache database.CacheInterface) PaymentMethodRepositoryInterface {
	s := &paymentMethodRepository{db: db, cache: cache, mapper: models.NewPaymentMethodMapper()}
	return s
}

func (r *paymentMethodRepository) Insert(ctx context.Context, pm *billingpb.PaymentMethod) error {
	mgo, err := r.mapper.MapObjectToMgo(pm)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, pm),
		)
		return err
	}

	_, err = r.db.Collection(collectionPaymentMethod).InsertOne(ctx, mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentMethod),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, pm),
		)
		return err
	}

	err = r.resetCaches(pm)

	if err != nil {
		return err
	}

	if err := r.cache.Set(fmt.Sprintf(cachePaymentMethodId, pm.Id), pm, 0); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, fmt.Sprintf(cachePaymentMethodId, pm.Id)),
			zap.Any(pkg.ErrorDatabaseFieldQuery, pm),
		)
		return err
	}

	return nil
}

func (r *paymentMethodRepository) MultipleInsert(ctx context.Context, pm []*billingpb.PaymentMethod) error {
	pms := make([]interface{}, len(pm))

	for i, v := range pm {
		mgo, err := r.mapper.MapObjectToMgo(v)

		if err != nil {
			zap.L().Error(
				pkg.ErrorMapModelFailed,
				zap.Error(err),
				zap.Any(pkg.ErrorDatabaseFieldQuery, v),
			)
			return err
		}

		pms[i] = mgo
	}

	_, err := r.db.Collection(collectionPaymentMethod).InsertMany(ctx, pms)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentMethod),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, pms),
		)
		return err
	}

	if err := r.cache.Delete(cachePaymentMethodAll); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "DELETE"),
			zap.String(pkg.ErrorCacheFieldKey, cachePaymentMethodAll),
		)
		return err
	}

	return nil
}

func (r *paymentMethodRepository) Update(ctx context.Context, pm *billingpb.PaymentMethod) error {
	oid, err := primitive.ObjectIDFromHex(pm.Id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentMethod),
			zap.String(pkg.ErrorDatabaseFieldQuery, pm.Id),
		)
		return err
	}

	mgo, err := r.mapper.MapObjectToMgo(pm)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, pm),
		)
		return err
	}

	filter := bson.M{"_id": oid}
	err = r.db.Collection(collectionPaymentMethod).FindOneAndReplace(ctx, filter, mgo).Err()

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentMethod),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationUpdate),
			zap.Any(pkg.ErrorDatabaseFieldQuery, pm),
		)
		return err
	}

	if err = r.resetCaches(pm); err != nil {
		return err
	}

	if err := r.cache.Set(fmt.Sprintf(cachePaymentMethodId, pm.Id), pm, 0); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, fmt.Sprintf(cachePaymentMethodId, pm.Id)),
			zap.Any(pkg.ErrorDatabaseFieldQuery, pm),
		)
		return err
	}

	return nil
}

func (r *paymentMethodRepository) Delete(ctx context.Context, pm *billingpb.PaymentMethod) error {
	oid, err := primitive.ObjectIDFromHex(pm.Id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentMethod),
			zap.String(pkg.ErrorDatabaseFieldQuery, pm.Id),
		)
		return err
	}

	query := bson.M{"_id": oid}
	err = r.db.Collection(collectionPaymentMethod).FindOneAndDelete(ctx, query).Err()

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentMethod),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationDelete),
			zap.Any(pkg.ErrorDatabaseFieldQuery, pm),
		)
		return err
	}

	if err = r.resetCaches(pm); err != nil {
		return err
	}

	return nil
}

func (r *paymentMethodRepository) GetById(ctx context.Context, id string) (*billingpb.PaymentMethod, error) {
	pm := &billingpb.PaymentMethod{}
	key := fmt.Sprintf(cachePaymentMethodId, id)

	if err := r.cache.Get(key, pm); err == nil {
		return pm, nil
	}

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentMethod),
			zap.String(pkg.ErrorDatabaseFieldQuery, id),
		)
		return nil, err
	}

	mgo := &models.MgoPaymentMethod{}
	filter := bson.M{"_id": oid}
	err = r.db.Collection(collectionPaymentMethod).FindOne(ctx, filter).Decode(mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentMethod),
			zap.Any(pkg.ErrorDatabaseFieldQuery, mgo),
		)
		return nil, err
	}

	obj, err := r.mapper.MapMgoToObject(mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, mgo),
		)
		return nil, err
	}

	pm = obj.(*billingpb.PaymentMethod)

	if err := r.cache.Set(key, pm, 0); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorDatabaseFieldQuery, pm),
		)
	}

	return pm, nil
}

func (r *paymentMethodRepository) GetAll(ctx context.Context) ([]*billingpb.PaymentMethod, error) {
	c := []*billingpb.PaymentMethod{}

	if err := r.cache.Get(cachePaymentMethodAll, &c); err == nil {
		return c, nil
	}

	cursor, err := r.db.Collection(collectionPaymentMethod).Find(ctx, bson.M{})

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentMethod),
		)
		return nil, err
	}

	var mgoPaymentMethods []*models.MgoPaymentMethod
	err = cursor.All(ctx, &mgoPaymentMethods)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentMethod),
		)
		return nil, err
	}

	pms := make([]*billingpb.PaymentMethod, len(mgoPaymentMethods))
	for i, pm := range mgoPaymentMethods {
		obj, err := r.mapper.MapMgoToObject(pm)
		if err != nil {
			zap.L().Error(
				pkg.ErrorMapModelFailed,
				zap.Error(err),
				zap.Any(pkg.ErrorDatabaseFieldQuery, pm),
			)
			return nil, err
		}
		pms[i] = obj.(*billingpb.PaymentMethod)
	}

	if err = r.cache.Set(cachePaymentMethodAll, pms, 0); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, cachePaymentMethodAll),
			zap.Any(pkg.ErrorDatabaseFieldQuery, c),
		)
	}

	return pms, nil
}

func (r *paymentMethodRepository) GetByGroupAndCurrency(
	ctx context.Context,
	isProduction bool,
	group string,
	currency string,
) (*billingpb.PaymentMethod, error) {
	pm := &billingpb.PaymentMethod{}
	key := fmt.Sprintf(cachePaymentMethodGroup, group)
	err := r.cache.Get(key, pm)

	if err == nil {
		return pm, nil
	}

	field := fieldTestSettings

	if isProduction {
		field = fieldProductionSettings
	}

	mgo := &models.MgoPaymentMethod{}
	query := bson.M{
		"group_alias": group,
		field: bson.M{
			"$elemMatch": bson.M{
				"currency": currency,
			},
		},
	}
	err = r.db.Collection(collectionPaymentMethod).FindOne(ctx, query).Decode(mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentMethod),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	obj, err := r.mapper.MapMgoToObject(mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, mgo),
		)
		return nil, err
	}

	pm = obj.(*billingpb.PaymentMethod)

	if err := r.cache.Set(key, pm, 0); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorDatabaseFieldQuery, pm),
		)
	}

	return pm, nil
}

func (r *paymentMethodRepository) ListByOrder(
	ctx context.Context,
	order *billingpb.Order,
) ([]*billingpb.PaymentMethod, error) {
	objs := []*billingpb.PaymentMethod{}

	field := fieldTestSettings

	if order.IsProduction {
		field = fieldProductionSettings
	}

	key := fmt.Sprintf(cachePaymentMethodModeCurrencyMccCompany, field, order.Currency, order.MccCode, order.OperatingCompanyId)
	err := r.cache.Get(key, &objs)

	if err == nil {
		return objs, nil
	}

	query := bson.M{
		"is_active": true,
		field: bson.M{
			"$elemMatch": bson.M{
				"currency":             order.Currency,
				"mcc_code":             order.MccCode,
				"operating_company_id": order.OperatingCompanyId,
			},
		},
	}

	zap.S().Infow("Find payment methods", "query", query)
	cursor, err := r.db.Collection(collectionPaymentMethod).Find(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentMethod),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	var pms []*models.MgoPaymentMethod
	err = cursor.All(ctx, &pms)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentMethod),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	objs = make([]*billingpb.PaymentMethod, len(pms))
	for i, pm := range pms {
		obj, err := r.mapper.MapMgoToObject(pm)
		if err != nil {
			zap.L().Error(
				pkg.ErrorMapModelFailed,
				zap.Error(err),
				zap.Any(pkg.ErrorDatabaseFieldQuery, pm),
			)
			return nil, err
		}
		objs[i] = obj.(*billingpb.PaymentMethod)
	}

	err = r.cache.Set(key, objs, 0)

	if err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorDatabaseFieldQuery, objs),
		)
		return nil, err
	}

	return objs, nil
}

func (r *paymentMethodRepository) resetCaches(pm *billingpb.PaymentMethod) error {
	if err := r.cache.Delete(cachePaymentMethodAll); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "DELETE"),
			zap.String(pkg.ErrorCacheFieldKey, cachePaymentMethodAll),
		)
		return err
	}

	key := fmt.Sprintf(cachePaymentMethodId, pm.Id)

	if err := r.cache.Delete(key); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "DELETE"),
			zap.String(pkg.ErrorCacheFieldKey, key),
		)
		return err
	}

	key = fmt.Sprintf(cachePaymentMethodGroup, pm.Group)

	if err := r.cache.Delete(key); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "DELETE"),
			zap.String(pkg.ErrorCacheFieldKey, key),
		)
		return err
	}

	for _, param := range pm.TestSettings {
		key := fmt.Sprintf(cachePaymentMethodModeCurrencyMccCompany, fieldTestSettings, param.Currency, param.MccCode, param.OperatingCompanyId)

		if err := r.cache.Delete(key); err != nil {
			zap.L().Error(
				pkg.ErrorCacheQueryFailed,
				zap.Error(err),
				zap.String(pkg.ErrorCacheFieldCmd, "DELETE"),
				zap.String(pkg.ErrorCacheFieldKey, key),
			)
			return err
		}
	}

	for _, param := range pm.ProductionSettings {
		key := fmt.Sprintf(cachePaymentMethodModeCurrencyMccCompany, fieldProductionSettings, param.Currency, param.MccCode, param.OperatingCompanyId)

		if err := r.cache.Delete(key); err != nil {
			zap.L().Error(
				pkg.ErrorCacheQueryFailed,
				zap.Error(err),
				zap.String(pkg.ErrorCacheFieldCmd, "DELETE"),
				zap.String(pkg.ErrorCacheFieldKey, key),
			)
			return err
		}
	}

	return nil
}

func (r *paymentMethodRepository) FindByName(ctx context.Context, name string, sort []string) ([]*billingpb.PaymentMethod, error) {
	var pms []*billingpb.PaymentMethod

	query := bson.M{"is_active": true}

	if name != "" {
		query["name"] = primitive.Regex{Pattern: ".*" + name + ".*", Options: "i"}
	}

	opts := options.Find().SetSort(mongodb.ToSortOption(sort))
	cursor, err := r.db.Collection(collectionPaymentMethod).Find(ctx, query, opts)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentMethod),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	var mgo []*models.MgoPaymentMethod
	err = cursor.All(ctx, &mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaymentMethod),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	pms = make([]*billingpb.PaymentMethod, len(mgo))
	for i, pm := range mgo {
		obj, err := r.mapper.MapMgoToObject(pm)
		if err != nil {
			zap.L().Error(
				pkg.ErrorMapModelFailed,
				zap.Error(err),
				zap.Any(pkg.ErrorDatabaseFieldQuery, pm),
			)
			return nil, err
		}
		pms[i] = obj.(*billingpb.PaymentMethod)
	}

	return pms, nil
}

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
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
)

const (
	cachePriceGroupId     = "price_group:id:%s"
	cachePriceGroupAll    = "price_group:all"
	cachePriceGroupRegion = "price_group:region:%s"

	collectionPriceGroup = "price_group"
)

type priceGroupRepository repository

// NewPriceGroupRepository create and return an object for working with the price group repository.
// The returned object implements the PriceGroupRepositoryInterface interface.
func NewPriceGroupRepository(db mongodb.SourceInterface, cache database.CacheInterface) PriceGroupRepositoryInterface {
	s := &priceGroupRepository{db: db, cache: cache, mapper: models.NewPriceGroupMapper()}
	return s
}

func (r *priceGroupRepository) Insert(ctx context.Context, pg *billingpb.PriceGroup) error {
	mgo, err := r.mapper.MapObjectToMgo(pg)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, pg),
		)
		return err
	}

	_, err = r.db.Collection(collectionPriceGroup).InsertOne(ctx, mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPriceGroup),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, pg),
		)
		return err
	}

	if err := r.updateCache(pg); err != nil {
		return err
	}

	return nil
}

func (r priceGroupRepository) MultipleInsert(ctx context.Context, pg []*billingpb.PriceGroup) error {
	c := make([]interface{}, len(pg))

	for i, v := range pg {
		mgo, err := r.mapper.MapObjectToMgo(v)

		if err != nil {
			zap.L().Error(
				pkg.ErrorDatabaseMapModelFailed,
				zap.Error(err),
				zap.Any(pkg.ErrorDatabaseFieldQuery, pg),
			)
		}

		c[i] = mgo
	}

	_, err := r.db.Collection(collectionPriceGroup).InsertMany(ctx, c)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPriceGroup),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, c),
		)
		return err
	}

	return nil
}

func (r priceGroupRepository) Update(ctx context.Context, pg *billingpb.PriceGroup) error {
	oid, err := primitive.ObjectIDFromHex(pg.Id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPriceGroup),
			zap.String(pkg.ErrorDatabaseFieldQuery, pg.Id),
		)
		return err
	}

	mgo, err := r.mapper.MapObjectToMgo(pg)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, pg),
		)
		return err
	}

	filter := bson.M{"_id": oid}
	_, err = r.db.Collection(collectionPriceGroup).ReplaceOne(ctx, filter, mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPriceGroup),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationUpdate),
			zap.Any(pkg.ErrorDatabaseFieldQuery, pg),
		)
		return err
	}

	if err := r.updateCache(pg); err != nil {
		return err
	}

	return nil
}

func (r priceGroupRepository) GetById(ctx context.Context, id string) (*billingpb.PriceGroup, error) {
	var c = &billingpb.PriceGroup{}
	key := fmt.Sprintf(cachePriceGroupId, id)
	err := r.cache.Get(key, c)

	if err == nil {
		return c, nil
	}

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPriceGroup),
			zap.String(pkg.ErrorDatabaseFieldQuery, id),
		)
		return nil, err
	}

	var mgo = models.MgoPriceGroup{}
	query := bson.M{"_id": oid, "is_active": true}
	err = r.db.Collection(collectionPriceGroup).FindOne(ctx, query).Decode(&mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPriceGroup),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	obj, err := r.mapper.MapMgoToObject(&mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, mgo),
		)
		return nil, err
	}

	if err = r.cache.Set(key, obj.(*billingpb.PriceGroup), 0); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorDatabaseFieldQuery, obj),
		)
	}

	return obj.(*billingpb.PriceGroup), nil
}

func (r priceGroupRepository) GetByRegion(ctx context.Context, region string) (*billingpb.PriceGroup, error) {
	var c = &billingpb.PriceGroup{}
	key := fmt.Sprintf(cachePriceGroupRegion, region)

	if err := r.cache.Get(key, c); err == nil {
		return c, nil
	}

	var mgo = models.MgoPriceGroup{}
	query := bson.M{"region": region, "is_active": true}
	err := r.db.Collection(collectionPriceGroup).FindOne(ctx, query).Decode(&mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPriceGroup),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	obj, err := r.mapper.MapMgoToObject(&mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, mgo),
		)
		return nil, err
	}

	if err := r.cache.Set(key, obj.(*billingpb.PriceGroup), 0); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorDatabaseFieldQuery, c),
		)
	}

	return obj.(*billingpb.PriceGroup), nil
}

func (r priceGroupRepository) GetAll(ctx context.Context) ([]*billingpb.PriceGroup, error) {
	c := []*billingpb.PriceGroup{}

	if err := r.cache.Get(cachePriceGroupAll, &c); err == nil {
		return c, nil
	}

	query := bson.M{"is_active": true}
	cursor, err := r.db.Collection(collectionPriceGroup).Find(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPriceGroup),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	var mgoPriceGroups []*models.MgoPriceGroup
	err = cursor.All(ctx, &mgoPriceGroups)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPriceGroup),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	objs := make([]*billingpb.PriceGroup, len(mgoPriceGroups))

	for i, obj := range mgoPriceGroups {
		v, err := r.mapper.MapMgoToObject(obj)
		if err != nil {
			zap.L().Error(
				pkg.ErrorDatabaseMapModelFailed,
				zap.Error(err),
				zap.Any(pkg.ErrorDatabaseFieldQuery, obj),
			)
			return nil, err
		}
		objs[i] = v.(*billingpb.PriceGroup)
	}

	if err = r.cache.Set(cachePriceGroupAll, objs, 0); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, cachePriceGroupAll),
			zap.Any(pkg.ErrorDatabaseFieldQuery, c),
		)
	}

	return objs, nil
}

func (r priceGroupRepository) updateCache(pg *billingpb.PriceGroup) error {
	if pg != nil {
		if err := r.cache.Set(fmt.Sprintf(cachePriceGroupId, pg.Id), pg, 0); err != nil {
			zap.L().Error(
				pkg.ErrorCacheQueryFailed,
				zap.Error(err),
				zap.String(pkg.ErrorCacheFieldCmd, "SET"),
				zap.String(pkg.ErrorCacheFieldKey, fmt.Sprintf(cachePriceGroupId, pg.Id)),
				zap.Any(pkg.ErrorDatabaseFieldQuery, pg),
			)
			return err
		}

		if err := r.cache.Set(fmt.Sprintf(cachePriceGroupRegion, pg.Region), pg, 0); err != nil {
			zap.L().Error(
				pkg.ErrorCacheQueryFailed,
				zap.Error(err),
				zap.String(pkg.ErrorCacheFieldCmd, "SET"),
				zap.String(pkg.ErrorCacheFieldKey, fmt.Sprintf(cachePriceGroupRegion, pg.Region)),
				zap.Any(pkg.ErrorDatabaseFieldQuery, pg),
			)
			return err
		}
	}

	if err := r.cache.Delete(cachePriceGroupAll); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "DELETE"),
			zap.String(pkg.ErrorCacheFieldKey, cachePriceGroupAll),
		)
		return err
	}

	return nil
}

package repository

import (
	"context"
	"fmt"
	"github.com/paysuper/paysuper-billing-server/internal/database"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
)

const (
	collectionProduct = "product"

	cacheProductId = "product:id:%s"
)

type productRepository repository

// NewProductRepository create and return an object for working with the product repository.
// The returned object implements the ProductRepositoryInterface interface.
func NewProductRepository(db mongodb.SourceInterface, cache database.CacheInterface) ProductRepositoryInterface {
	s := &productRepository{db: db, cache: cache}
	return s
}

func (r *productRepository) Upsert(ctx context.Context, product *billingpb.Product) error {
	oid, err := primitive.ObjectIDFromHex(product.Id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionProduct),
			zap.String(pkg.ErrorDatabaseFieldQuery, product.Id),
		)
		return err
	}

	filter := bson.M{"_id": oid}
	opts := options.Replace().SetUpsert(true)
	_, err = r.db.Collection(collectionProduct).ReplaceOne(ctx, filter, product, opts)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationUpsert),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKeyProduct),
			zap.Any(pkg.ErrorDatabaseFieldQuery, product),
		)
		return err
	}

	if err := r.updateCache(product); err != nil {
		return err
	}

	return nil
}

func (r *productRepository) GetById(ctx context.Context, id string) (*billingpb.Product, error) {
	var c billingpb.Product
	key := fmt.Sprintf(cacheProductId, id)

	if err := r.cache.Get(key, c); err == nil {
		return &c, nil
	}

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionProduct),
			zap.String(pkg.ErrorDatabaseFieldQuery, id),
		)
		return nil, err
	}

	filter := bson.M{"_id": oid, "deleted": false}
	err = r.db.Collection(collectionProduct).FindOne(ctx, filter).Decode(&c)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionProduct),
			zap.Any(pkg.ErrorDatabaseFieldQuery, filter),
		)
		return nil, err
	}

	if err := r.cache.Set(key, c, 0); err != nil {
		zap.S().Errorf("Unable to set cache", "err", err.Error(), "key", key, "data", c)
	}

	return &c, nil
}

func (r *productRepository) CountByProjectSku(ctx context.Context, projectId string, sku string) (int64, error) {
	oid, err := primitive.ObjectIDFromHex(projectId)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionProduct),
			zap.String(pkg.ErrorDatabaseFieldQuery, projectId),
		)
		return int64(0), err
	}

	query := bson.M{"project_id": oid, "sku": sku, "deleted": false}
	count, err := r.db.Collection(collectionProduct).CountDocuments(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionProduct),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return 0, err
	}

	return count, nil
}

func (r *productRepository) Find(
	ctx context.Context,
	merchantId string,
	projectId string,
	sku string,
	name string,
	enabled int32,
	offset int64,
	limit int64,
) ([]*billingpb.Product, error) {
	merchantOid, err := primitive.ObjectIDFromHex(merchantId)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionProduct),
			zap.String(pkg.ErrorDatabaseFieldQuery, merchantId),
		)
		return nil, err
	}

	query := bson.M{"merchant_id": merchantOid, "deleted": false}

	if projectId != "" {
		query["project_id"], err = primitive.ObjectIDFromHex(projectId)

		if err != nil {
			zap.L().Error(
				pkg.ErrorDatabaseInvalidObjectId,
				zap.Error(err),
				zap.String(pkg.ErrorDatabaseFieldCollection, collectionProduct),
				zap.String(pkg.ErrorDatabaseFieldQuery, projectId),
			)
			return nil, err
		}
	}

	if sku != "" {
		query["sku"] = primitive.Regex{Pattern: sku, Options: "i"}
	}
	if name != "" {
		query["name"] = bson.M{"$elemMatch": bson.M{"value": primitive.Regex{Pattern: name, Options: "i"}}}
	}

	if enabled > 0 {
		if enabled == 1 {
			query["enabled"] = false
		} else {
			query["enabled"] = true
		}
	}

	opts := options.Find().
		SetLimit(limit).
		SetSkip(offset)
	cursor, err := r.db.Collection(collectionProduct).Find(ctx, query, opts)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionProduct),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	var list []*billingpb.Product
	err = cursor.All(ctx, &list)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionProduct),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	return list, nil
}

func (r *productRepository) FindCount(
	ctx context.Context,
	merchantId string,
	projectId string,
	sku string,
	name string,
	enabled int32,
) (int64, error) {
	merchantOid, err := primitive.ObjectIDFromHex(merchantId)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionProduct),
			zap.String(pkg.ErrorDatabaseFieldQuery, merchantId),
		)
		return int64(0), nil
	}

	query := bson.M{"merchant_id": merchantOid, "deleted": false}

	if projectId != "" {
		query["project_id"], err = primitive.ObjectIDFromHex(projectId)

		if err != nil {
			zap.L().Error(
				pkg.ErrorDatabaseInvalidObjectId,
				zap.Error(err),
				zap.String(pkg.ErrorDatabaseFieldCollection, collectionProduct),
				zap.String(pkg.ErrorDatabaseFieldQuery, projectId),
			)
			return int64(0), nil
		}
	}

	if sku != "" {
		query["sku"] = primitive.Regex{Pattern: sku, Options: "i"}
	}
	if name != "" {
		query["name"] = bson.M{"$elemMatch": bson.M{"value": primitive.Regex{Pattern: name, Options: "i"}}}
	}

	if enabled > 0 {
		if enabled == 1 {
			query["enabled"] = false
		} else {
			query["enabled"] = true
		}
	}

	count, err := r.db.Collection(collectionProduct).CountDocuments(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionProduct),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return 0, nil
	}

	return count, nil
}

func (r *productRepository) updateCache(product *billingpb.Product) error {
	key := fmt.Sprintf(cacheProductId, product.Id)
	if err := r.cache.Set(key, product, 0); err != nil {

		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorDatabaseFieldQuery, product),
		)
		return err
	}

	return nil
}

package repository

import (
	"context"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
)

const (
	collectionKeyProduct = "key_product"
)

type keyProductRepository repository

// NewKeyProductRepository create and return an object for working with the key product repository.
// The returned object implements the KeyProductRepositoryInterface interface.
func NewKeyProductRepository(db mongodb.SourceInterface) KeyProductRepositoryInterface {
	s := &keyProductRepository{db: db}
	return s
}

func (r *keyProductRepository) GetById(ctx context.Context, id string) (*billingpb.KeyProduct, error) {
	oid, _ := primitive.ObjectIDFromHex(id)
	query := bson.M{"_id": oid, "deleted": false}

	product := &billingpb.KeyProduct{}
	err := r.db.Collection(collectionKeyProduct).FindOne(ctx, query).Decode(product)

	return product, err
}

func (r *keyProductRepository) Upsert(ctx context.Context, keyProduct *billingpb.KeyProduct) error {
	oid, _ := primitive.ObjectIDFromHex(keyProduct.Id)
	opts := options.Replace().SetUpsert(true)
	filter := bson.M{"_id": oid}
	_, err := r.db.Collection(collectionKeyProduct).ReplaceOne(ctx, filter, keyProduct, opts)

	return err
}

func (r *keyProductRepository) Update(ctx context.Context, keyProduct *billingpb.KeyProduct) error {
	oid, _ := primitive.ObjectIDFromHex(keyProduct.Id)
	filter := bson.M{"_id": oid}
	_, err := r.db.Collection(collectionKeyProduct).ReplaceOne(ctx, filter, keyProduct)

	return err
}

func (r *keyProductRepository) CountByProjectIdSku(ctx context.Context, projectId, sku string) (int64, error) {
	projectOid, _ := primitive.ObjectIDFromHex(projectId)
	dupQuery := bson.M{
		"project_id": projectOid,
		"sku":        sku,
		"deleted":    false,
	}
	count, err := r.db.Collection(collectionKeyProduct).CountDocuments(ctx, dupQuery)

	if err != nil {
		return int64(0), err
	}

	return count, nil
}

func (r *keyProductRepository) FindByIdsProjectId(ctx context.Context, ids []string, projectId string) ([]*billingpb.KeyProduct, error) {
	idsLen := len(ids)
	items := make([]primitive.ObjectID, idsLen)

	for i, id := range ids {
		oid, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			continue
		}

		items[i] = oid
	}

	projectOid, _ := primitive.ObjectIDFromHex(projectId)
	query := bson.M{
		"_id":        bson.M{"$in": items},
		"enabled":    true,
		"deleted":    false,
		"project_id": projectOid,
	}
	cursor, err := r.db.Collection(collectionKeyProduct).Find(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKeyProduct),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	var list []*billingpb.KeyProduct
	err = cursor.All(ctx, &list)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKeyProduct),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	return list, nil
}

func (r *keyProductRepository) Find(
	ctx context.Context, merchantId, projectId, sku, name, enabled string, offset, limit int64,
) ([]*billingpb.KeyProduct, error) {
	merchantOid, _ := primitive.ObjectIDFromHex(merchantId)
	query := bson.M{"merchant_id": merchantOid, "deleted": false}

	if projectId != "" {
		query["project_id"], _ = primitive.ObjectIDFromHex(projectId)
	}

	if sku != "" {
		query["sku"] = primitive.Regex{Pattern: sku, Options: "i"}
	}
	if name != "" {
		query["name"] = bson.M{"$elemMatch": bson.M{"value": primitive.Regex{Pattern: name, Options: "i"}}}
	}

	if enabled == "true" {
		query["enabled"] = bson.M{"$eq": true}
	} else if enabled == "false" {
		query["enabled"] = bson.M{"$eq": false}
	}

	opts := options.Find().
		SetSkip(offset).
		SetLimit(limit)
	cursor, err := r.db.Collection(collectionKeyProduct).Find(ctx, query, opts)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKeyProduct),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	var items []*billingpb.KeyProduct
	err = cursor.All(ctx, &items)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKeyProduct),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	return items, nil
}

func (r *keyProductRepository) FindCount(ctx context.Context, merchantId, projectId, sku, name, enabled string) (int64, error) {
	merchantOid, _ := primitive.ObjectIDFromHex(merchantId)
	query := bson.M{"merchant_id": merchantOid, "deleted": false}

	if projectId != "" {
		query["project_id"], _ = primitive.ObjectIDFromHex(projectId)
	}

	if sku != "" {
		query["sku"] = primitive.Regex{Pattern: sku, Options: "i"}
	}
	if name != "" {
		query["name"] = bson.M{"$elemMatch": bson.M{"value": primitive.Regex{Pattern: name, Options: "i"}}}
	}

	if enabled == "true" {
		query["enabled"] = bson.M{"$eq": true}
	} else if enabled == "false" {
		query["enabled"] = bson.M{"$eq": false}
	}

	count, err := r.db.Collection(collectionKeyProduct).CountDocuments(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKeyProduct),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return int64(0), err
	}

	return count, nil
}

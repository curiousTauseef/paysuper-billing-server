package repository

import (
	"context"
	"github.com/paysuper/paysuper-billing-server/internal/repository/models"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
	"time"
)

const collectionKey = "key"

type keyRepository repository

// NewKeyRepository create and return an object for working with the key repository.
// The returned object implements the KeyRepositoryInterface interface.
func NewKeyRepository(db mongodb.SourceInterface) KeyRepositoryInterface {
	s := &keyRepository{db: db, mapper: models.NewKeyMapper()}
	return s
}

func (r *keyRepository) Insert(ctx context.Context, key *billingpb.Key) error {
	mgo, err := r.mapper.MapObjectToMgo(key)
	if err != nil {
		zap.L().Error(
			pkg.ErrorMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, key),
		)
		return err
	}

	_, err = r.db.Collection(collectionKey).InsertOne(ctx, mgo)

	if err != nil {
		return err
	}

	return nil
}

func (r *keyRepository) GetById(ctx context.Context, id string) (*billingpb.Key, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKey),
			zap.String(pkg.ErrorDatabaseFieldQuery, id),
		)
		return nil, err
	}

	mgo := &models.MgoKey{}
	filter := bson.M{"_id": oid}
	err = r.db.Collection(collectionKey).FindOne(ctx, filter).Decode(mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKey),
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

	return obj.(*billingpb.Key), nil
}

func (r *keyRepository) ReserveKey(ctx context.Context, keyProductId, platformId, orderId string, ttl int32) (*billingpb.Key, error) {
	oid, err := primitive.ObjectIDFromHex(keyProductId)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKey),
			zap.String(pkg.ErrorDatabaseFieldQuery, keyProductId),
		)
		return nil, err
	}

	orderOid, err := primitive.ObjectIDFromHex(orderId)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKey),
			zap.String(pkg.ErrorDatabaseFieldQuery, orderId),
		)
		return nil, err
	}

	mgo := &models.MgoKey{}
	duration := time.Second * time.Duration(ttl)

	query := bson.M{
		"key_product_id": oid,
		"platform_id":    platformId,
		"order_id":       nil,
	}
	update := bson.M{
		"$set": bson.M{
			"reserved_to": time.Now().UTC().Add(duration),
			"order_id":    orderOid,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err = r.db.Collection(collectionKey).FindOneAndUpdate(ctx, query, update, opts).Decode(mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKey),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
			zap.Any(pkg.ErrorDatabaseFieldOperationUpdate, update),
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

	return obj.(*billingpb.Key), nil
}

func (r *keyRepository) CancelById(ctx context.Context, id string) (*billingpb.Key, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKey),
			zap.String(pkg.ErrorDatabaseFieldQuery, id),
		)
		return nil, err
	}

	query := bson.M{"_id": oid}
	update := bson.M{
		"$set": bson.M{
			"reserved_to": time.Time{},
			"order_id":    nil,
		},
	}
	mgo := &models.MgoKey{}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err = r.db.Collection(collectionKey).FindOneAndUpdate(ctx, query, update, opts).Decode(mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKey),
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

	return obj.(*billingpb.Key), nil
}

func (r *keyRepository) FinishRedeemById(ctx context.Context, id string) (*billingpb.Key, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKey),
			zap.String(pkg.ErrorDatabaseFieldQuery, id),
		)
		return nil, err
	}

	query := bson.M{"_id": oid}
	update := bson.M{
		"$set": bson.M{
			"reserved_to": time.Time{},
			"redeemed_at": time.Now().UTC(),
		},
	}
	mgo := &models.MgoKey{}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err = r.db.Collection(collectionKey).FindOneAndUpdate(ctx, query, update, opts).Decode(mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKey),
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

	return obj.(*billingpb.Key), nil
}

func (r *keyRepository) CountKeysByProductPlatform(ctx context.Context, keyProductId string, platformId string) (int64, error) {
	oid, err := primitive.ObjectIDFromHex(keyProductId)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKey),
			zap.String(pkg.ErrorDatabaseFieldQuery, keyProductId),
		)
		return int64(0), err
	}

	query := bson.M{
		"key_product_id": oid,
		"platform_id":    platformId,
		"order_id":       nil,
	}

	count, err := r.db.Collection(collectionKey).CountDocuments(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKey),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return int64(0), err
	}

	return count, nil
}

func (r *keyRepository) FindUnfinished(ctx context.Context) ([]*billingpb.Key, error) {
	var keys []*models.MgoKey

	query := bson.M{
		"reserved_to": bson.M{
			"$gt": time.Time{},
			"$lt": time.Now().UTC(),
		},
	}

	cursor, err := r.db.Collection(collectionKey).Find(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKey),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	err = cursor.All(ctx, &keys)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionKey),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	result := make([]*billingpb.Key, len(keys))
	for i, key := range keys {
		obj, err := r.mapper.MapMgoToObject(key)
		if err != nil {
			zap.L().Error(
				pkg.ErrorMapModelFailed,
				zap.Error(err),
				zap.Any(pkg.ErrorDatabaseFieldQuery, key),
			)
			return nil, err
		}

		result[i] = obj.(*billingpb.Key)
	}

	return result, nil
}

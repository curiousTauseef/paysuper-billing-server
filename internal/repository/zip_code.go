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
	"regexp"
)

const (
	collectionZipCode = "zip_code"

	cacheZipCodeByZipAndCountry = "zip_code:zip_country:%s_%s"
)

type zipCodeRepository repository

// NewZipCodeRepository create and return an object for working with the zip codes repository.
// The returned object implements the ZipCodeRepositoryInterface interface.
func NewZipCodeRepository(db mongodb.SourceInterface, cache database.CacheInterface) ZipCodeRepositoryInterface {
	s := &zipCodeRepository{db: db, cache: cache, mapper: models.NewZipCodeMapper()}
	return s
}

func (h *zipCodeRepository) Insert(ctx context.Context, zipCode *billingpb.ZipCode) error {
	mgo, err := h.mapper.MapObjectToMgo(zipCode)
	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, zipCode),
		)
		return err
	}

	_, err = h.db.Collection(collectionZipCode).InsertOne(ctx, mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionZipCode),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, zipCode),
		)
		return err
	}

	var key = fmt.Sprintf(cacheZipCodeByZipAndCountry, zipCode.Zip, zipCode.Country)

	if err = h.cache.Set(key, zipCode, 0); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorDatabaseFieldQuery, zipCode),
		)
	}

	return nil
}

func (h *zipCodeRepository) GetByZipAndCountry(ctx context.Context, zip, country string) (*billingpb.ZipCode, error) {
	obj := &billingpb.ZipCode{}
	var key = fmt.Sprintf(cacheZipCodeByZipAndCountry, zip, country)

	if err := h.cache.Get(key, obj); err == nil {
		return obj, nil
	}

	query := bson.M{"zip": zip, "country": country}
	var mgo = models.MgoZipCode{}
	err := h.db.Collection(collectionZipCode).FindOne(ctx, query).Decode(&mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, cacheZipCodeByZipAndCountry),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	o, err := h.mapper.MapMgoToObject(&mgo)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseMapModelFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, mgo),
		)
		return nil, err
	}

	if err = h.cache.Set(key, o, 0); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorDatabaseFieldQuery, o),
		)
	}

	return o.(*billingpb.ZipCode), nil
}

func (h *zipCodeRepository) FindByZipAndCountry(ctx context.Context, zip, country string, offset, limit int64) ([]*billingpb.ZipCode, error) {
	query := bson.D{
		{"zip", primitive.Regex{Pattern: regexp.QuoteMeta(zip)}},
		{"country", country},
	}
	opts := options.Find().SetLimit(limit).SetSkip(offset)
	cursor, err := h.db.Collection(collectionZipCode).Find(ctx, query, opts)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionZipCode),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	var mgoZipCodes []*models.MgoZipCode
	err = cursor.All(ctx, &mgoZipCodes)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionZipCode),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	objs := make([]*billingpb.ZipCode, len(mgoZipCodes))

	for i, obj := range mgoZipCodes {
		v, err := h.mapper.MapMgoToObject(obj)
		if err != nil {
			zap.L().Error(
				pkg.ErrorDatabaseMapModelFailed,
				zap.Error(err),
				zap.Any(pkg.ErrorDatabaseFieldQuery, obj),
			)
			return nil, err
		}
		objs[i] = v.(*billingpb.ZipCode)
	}

	return objs, nil
}

func (h *zipCodeRepository) CountByZip(ctx context.Context, zip, country string) (int64, error) {
	query := bson.D{{"zip", primitive.Regex{Pattern: zip}}, {"country", country}}
	count, err := h.db.Collection(collectionZipCode).CountDocuments(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionZipCode),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return 0, err
	}

	return count, nil
}

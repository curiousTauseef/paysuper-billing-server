package repository

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-billing-server/internal/database"
	pkg2 "github.com/paysuper/paysuper-billing-server/internal/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
	"time"
)

const (
	collectionPayoutDocuments       = "payout_documents"
	collectionPayoutDocumentChanges = "payout_documents_changes"

	cacheKeyPayoutDocument         = "payout_document:id:%s"
	cacheKeyPayoutDocumentMerchant = "payout_document:id:%s:merchant:id:%s"
)

type payoutRepository repository

// NewPayoutRepository create and return an object for working with the payout repository.
// The returned object implements the PayoutRepositoryInterface interface.
func NewPayoutRepository(db mongodb.SourceInterface, cache database.CacheInterface) PayoutRepositoryInterface {
	s := &payoutRepository{db: db, cache: cache}
	return s
}

func (r *payoutRepository) Insert(ctx context.Context, pd *billingpb.PayoutDocument, ip, source string) error {
	_, err := r.db.Collection(collectionPayoutDocuments).InsertOne(ctx, pd)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldDocument, pd),
		)
		return err
	}

	if err = r.onPayoutDocumentChange(ctx, pd, ip, source); err != nil {
		return err
	}

	return r.updateCaches(pd)
}

func (r *payoutRepository) Update(ctx context.Context, pd *billingpb.PayoutDocument, ip, source string) error {
	oid, err := primitive.ObjectIDFromHex(pd.Id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
			zap.String(pkg.ErrorDatabaseFieldQuery, pd.Id),
		)
		return err
	}

	filter := bson.M{"_id": oid}
	_, err = r.db.Collection(collectionPayoutDocuments).ReplaceOne(ctx, filter, pd)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationUpdate),
			zap.Any(pkg.ErrorDatabaseFieldDocument, pd),
		)

		return err
	}

	if err = r.onPayoutDocumentChange(ctx, pd, ip, source); err != nil {
		return err
	}

	return r.updateCaches(pd)
}

func (r *payoutRepository) GetById(ctx context.Context, id string) (*billingpb.PayoutDocument, error) {
	var c billingpb.PayoutDocument
	key := fmt.Sprintf(cacheKeyPayoutDocument, id)

	if err := r.cache.Get(key, &c); err == nil {
		return &c, nil
	}

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
			zap.String(pkg.ErrorDatabaseFieldQuery, id),
		)
		return nil, err
	}

	filter := bson.M{"_id": oid}
	err = r.db.Collection(collectionPayoutDocuments).FindOne(ctx, filter).Decode(&c)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
			zap.String(pkg.ErrorDatabaseFieldDocumentId, id),
		)
		return nil, err
	}

	if err = r.updateCaches(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *payoutRepository) GetByIdMerchantId(ctx context.Context, id, merchantId string) (*billingpb.PayoutDocument, error) {
	var c billingpb.PayoutDocument
	key := fmt.Sprintf(cacheKeyPayoutDocumentMerchant, id, merchantId)

	if err := r.cache.Get(key, &c); err == nil {
		return &c, nil
	}

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
			zap.String(pkg.ErrorDatabaseFieldQuery, id),
		)
		return nil, err
	}

	merchantOid, err := primitive.ObjectIDFromHex(merchantId)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
			zap.String(pkg.ErrorDatabaseFieldQuery, merchantId),
		)
		return nil, err
	}

	query := bson.M{"_id": oid, "merchant_id": merchantOid}
	err = r.db.Collection(collectionPayoutDocuments).FindOne(ctx, query).Decode(&c)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	if err = r.updateCaches(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *payoutRepository) GetBalanceAmount(ctx context.Context, merchantId, currency string) (float64, error) {
	oid, err := primitive.ObjectIDFromHex(merchantId)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
			zap.String(pkg.ErrorDatabaseFieldQuery, merchantId),
		)
		return float64(0), err
	}

	payoutDocumentStatusActive := []string{
		pkg.PayoutDocumentStatusPending,
		pkg.PayoutDocumentStatusPaid,
	}
	query := []bson.M{
		{
			"$match": bson.M{
				"merchant_id": oid,
				"currency":    currency,
				"status":      bson.M{"$in": payoutDocumentStatusActive},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$currency",
				"amount": bson.M{"$sum": "$total_fees"},
			},
		},
	}

	res := &pkg2.BalanceQueryResItem{}
	cursor, err := r.db.Collection(collectionPayoutDocuments).Aggregate(ctx, query)

	if err != nil {
		if err != mongo.ErrNoDocuments {
			zap.L().Error(
				pkg.ErrorDatabaseQueryFailed,
				zap.Error(err),
				zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
				zap.Any(pkg.ErrorDatabaseFieldQuery, query),
			)
		}
		return 0, err
	}

	defer func() {
		if err := cursor.Close(ctx); err != nil {
			zap.L().Error(
				pkg.ErrorQueryCursorCloseFailed,
				zap.Error(err),
				zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
			)
		}
	}()

	if cursor.Next(ctx) {
		err = cursor.Decode(&res)
		if err != nil {
			zap.L().Error(
				pkg.ErrorQueryCursorExecutionFailed,
				zap.Error(err),
				zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
				zap.Any(pkg.ErrorDatabaseFieldQuery, query),
			)
			return 0, err
		}
	}

	return res.Amount, nil
}

func (r *payoutRepository) GetLast(ctx context.Context, merchantId, currency string) (*billingpb.PayoutDocument, error) {
	oid, err := primitive.ObjectIDFromHex(merchantId)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
			zap.String(pkg.ErrorDatabaseFieldQuery, merchantId),
		)
		return nil, err
	}

	payoutDocumentStatusActive := []string{
		pkg.PayoutDocumentStatusPending,
		pkg.PayoutDocumentStatusPaid,
	}
	query := bson.M{
		"merchant_id": oid,
		"currency":    currency,
		"status":      bson.M{"$in": payoutDocumentStatusActive},
	}

	var pd *billingpb.PayoutDocument
	sorts := bson.M{"created_at": -1}
	opts := options.FindOne().SetSort(sorts)
	err = r.db.Collection(collectionPayoutDocuments).FindOne(ctx, query, opts).Decode(&pd)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
			zap.Any(pkg.ErrorDatabaseFieldSorts, sorts),
		)
		return nil, err
	}

	return pd, nil
}

func (r *payoutRepository) FindCount(ctx context.Context, merchantId string, status []string, dateFrom, dateTo int64) (int64, error) {
	merchantOid, err := primitive.ObjectIDFromHex(merchantId)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
			zap.String(pkg.ErrorDatabaseFieldQuery, merchantId),
		)
		return int64(0), err
	}

	query := bson.M{"merchant_id": merchantOid}

	if len(status) > 0 {
		query["status"] = bson.M{"$in": status}
	}

	if dateFrom > 0 || dateTo > 0 {
		date := bson.M{}
		if dateFrom > 0 {
			date["$gte"] = time.Unix(dateFrom, 0)
		}
		if dateTo > 0 {
			date["$lte"] = time.Unix(dateTo, 0)
		}
		query["created_at"] = date
	}

	count, err := r.db.Collection(collectionPayoutDocuments).CountDocuments(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return int64(0), err
	}

	return count, nil
}

func (r *payoutRepository) Find(
	ctx context.Context,
	merchantId string,
	status []string,
	dateFrom, dateTo,
	offset, limit int64,
) ([]*billingpb.PayoutDocument, error) {
	merchantOid, err := primitive.ObjectIDFromHex(merchantId)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
			zap.String(pkg.ErrorDatabaseFieldQuery, merchantId),
		)
		return nil, err
	}

	query := bson.M{"merchant_id": merchantOid}

	if len(status) > 0 {
		query["status"] = bson.M{"$in": status}
	}

	if dateFrom > 0 || dateTo > 0 {
		date := bson.M{}
		if dateFrom > 0 {
			date["$gte"] = time.Unix(dateFrom, 0)
		}
		if dateTo > 0 {
			date["$lte"] = time.Unix(dateTo, 0)
		}
		query["created_at"] = date
	}

	opts := options.Find().
		SetSort(mongodb.ToSortOption([]string{"-_id"})).
		SetLimit(limit).
		SetSkip(offset)
	cursor, err := r.db.Collection(collectionPayoutDocuments).Find(ctx, query, opts)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
			zap.Any(pkg.ErrorDatabaseFieldLimit, limit),
			zap.Any(pkg.ErrorDatabaseFieldOffset, offset),
		)
		return nil, err
	}

	var pds []*billingpb.PayoutDocument
	err = cursor.All(ctx, &pds)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocuments),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
			zap.Any(pkg.ErrorDatabaseFieldLimit, limit),
			zap.Any(pkg.ErrorDatabaseFieldOffset, offset),
		)
		return nil, err
	}

	return pds, nil
}

func (r *payoutRepository) updateCaches(pd *billingpb.PayoutDocument) (err error) {
	key1 := fmt.Sprintf(cacheKeyPayoutDocument, pd.Id)
	key2 := fmt.Sprintf(cacheKeyPayoutDocumentMerchant, pd.Id, pd.MerchantId)

	err = r.cache.Set(key1, pd, 0)
	if err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key1),
			zap.Any(pkg.ErrorCacheFieldData, pd),
		)
		return
	}

	err = r.cache.Set(key2, pd, 0)
	if err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key2),
			zap.Any(pkg.ErrorCacheFieldData, pd),
		)
	}
	return
}

func (r *payoutRepository) onPayoutDocumentChange(
	ctx context.Context,
	document *billingpb.PayoutDocument,
	ip, source string,
) error {
	change := &billingpb.PayoutDocumentChanges{
		Id:               primitive.NewObjectID().Hex(),
		PayoutDocumentId: document.Id,
		Source:           source,
		Ip:               ip,
		CreatedAt:        ptypes.TimestampNow(),
	}

	b, err := json.Marshal(document)

	if err != nil {
		zap.L().Error(
			pkg.ErrorJsonMarshallingFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldDocument, document),
		)
		return err
	}

	hash := md5.New()
	hash.Write(b)
	change.Hash = hex.EncodeToString(hash.Sum(nil))

	_, err = r.db.Collection(collectionPayoutDocumentChanges).InsertOne(ctx, change)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayoutDocumentChanges),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldDocument, change),
		)
		return err
	}

	return nil
}

package repository

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/paysuper/paysuper-billing-server/internal/database"
	pkg2 "github.com/paysuper/paysuper-billing-server/internal/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
	"time"
)

const (
	collectionRoyaltyReport        = "royalty_report"
	collectionRoyaltyReportChanges = "royalty_report_changes"

	cacheKeyRoyaltyReport = "royalty_report:id:%s"
)

type royaltyReportRepository repository

// NewRoyaltyReportRepository create and return an object for working with the royalty report repository.
// The returned object implements the RoyaltyReportRepositoryInterface interface.
func NewRoyaltyReportRepository(db mongodb.SourceInterface, cache database.CacheInterface) RoyaltyReportRepositoryInterface {
	s := &royaltyReportRepository{db: db, cache: cache}
	return s
}

func (r *royaltyReportRepository) GetNonPayoutReports(
	ctx context.Context,
	merchantId, currency string,
) (result []*billingpb.RoyaltyReport, err error) {
	oid, err := primitive.ObjectIDFromHex(merchantId)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionRoyaltyReport),
			zap.String(pkg.ErrorDatabaseFieldQuery, merchantId),
		)
		return nil, err
	}

	royaltyReportsStatusActive := []string{
		billingpb.RoyaltyReportStatusPending,
		billingpb.RoyaltyReportStatusAccepted,
		billingpb.RoyaltyReportStatusDispute,
	}
	query := bson.M{
		"merchant_id":        oid,
		"currency":           currency,
		"status":             bson.M{"$in": royaltyReportsStatusActive},
		"payout_document_id": "",
	}

	sorts := bson.M{"period_from": 1}
	opts := options.Find().SetSort(sorts)
	cursor, err := r.db.Collection(collectionRoyaltyReport).Find(ctx, query, opts)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionRoyaltyReport),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
			zap.Any(pkg.ErrorDatabaseFieldSorts, sorts),
		)
		return
	}

	err = cursor.All(ctx, &result)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionRoyaltyReport),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return
	}

	return
}

func (r *royaltyReportRepository) GetByPayoutId(ctx context.Context, payoutId string) (result []*billingpb.RoyaltyReport, err error) {
	query := bson.M{
		"payout_document_id": payoutId,
	}

	sorts := bson.M{"period_from": 1}
	opts := options.Find().SetSort(sorts)
	cursor, err := r.db.Collection(collectionRoyaltyReport).Find(ctx, query, opts)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionRoyaltyReport),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
			zap.Any(pkg.ErrorDatabaseFieldSorts, sorts),
		)
		return
	}

	err = cursor.All(ctx, &result)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionRoyaltyReport),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
			zap.Any(pkg.ErrorDatabaseFieldSorts, sorts),
		)
		return
	}

	return
}

func (r *royaltyReportRepository) GetBalanceAmount(ctx context.Context, merchantId, currency string) (float64, error) {
	oid, err := primitive.ObjectIDFromHex(merchantId)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionRoyaltyReport),
			zap.String(pkg.ErrorDatabaseFieldQuery, merchantId),
		)
		return float64(0), err
	}

	royaltyReportsStatusForBalance := []string{
		billingpb.RoyaltyReportStatusAccepted,
		billingpb.RoyaltyReportStatusWaitForPayment,
		billingpb.RoyaltyReportStatusPaid,
	}
	query := []bson.M{
		{
			"$match": bson.M{
				"merchant_id": oid,
				"currency":    currency,
				"status":      bson.M{"$in": royaltyReportsStatusForBalance},
			},
		},
		{
			"$group": bson.M{
				"_id":               "currency",
				"payout_amount":     bson.M{"$sum": "$totals.payout_amount"},
				"correction_amount": bson.M{"$sum": "$totals.correction_amount"},
			},
		},
		{
			"$project": bson.M{
				"_id":    0,
				"amount": bson.M{"$subtract": []interface{}{"$payout_amount", "$correction_amount"}},
			},
		},
	}

	cursor, err := r.db.Collection(collectionRoyaltyReport).Aggregate(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionRoyaltyReport),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return 0, err
	}

	res := &pkg2.BalanceQueryResItem{}

	defer func() {
		err := cursor.Close(ctx)
		if err != nil {
			zap.L().Error(
				pkg.ErrorQueryCursorCloseFailed,
				zap.Error(err),
				zap.String(pkg.ErrorDatabaseFieldCollection, collectionRoyaltyReport),
			)
		}
	}()

	if cursor.Next(ctx) {
		err = cursor.Decode(&res)
		if err != nil {
			zap.L().Error(
				pkg.ErrorQueryCursorExecutionFailed,
				zap.Error(err),
				zap.String(pkg.ErrorDatabaseFieldCollection, collectionRoyaltyReport),
				zap.Any(pkg.ErrorDatabaseFieldQuery, query),
			)
			return 0, err
		}
	}

	return res.Amount, nil
}

func (r *royaltyReportRepository) GetReportExists(
	ctx context.Context,
	merchantId, currency string,
	from, to time.Time,
) (report *billingpb.RoyaltyReport) {
	oid, err := primitive.ObjectIDFromHex(merchantId)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionRoyaltyReport),
			zap.String(pkg.ErrorDatabaseFieldQuery, merchantId),
		)
		return nil
	}

	query := bson.M{
		"merchant_id": oid,
		"period_from": bson.M{"$gte": from},
		"period_to":   bson.M{"$lte": to},
		"currency":    currency,
	}
	err = r.db.Collection(collectionRoyaltyReport).FindOne(ctx, query).Decode(&report)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String("collection", collectionRoyaltyReport),
			zap.Any("query", query),
		)
		return nil
	}

	return report
}

func (r *royaltyReportRepository) Insert(ctx context.Context, rr *billingpb.RoyaltyReport, ip, source string) (err error) {
	_, err = r.db.Collection(collectionRoyaltyReport).InsertOne(ctx, rr)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionRoyaltyReport),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldDocument, rr),
		)
		return
	}

	if err = r.onRoyaltyReportChange(ctx, rr, ip, source); err != nil {
		return
	}

	key := fmt.Sprintf(cacheKeyRoyaltyReport, rr.Id)

	if err = r.cache.Set(key, rr, 0); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorCacheFieldData, rr),
		)
	}

	return
}

func (r *royaltyReportRepository) Update(ctx context.Context, rr *billingpb.RoyaltyReport, ip, source string) error {
	oid, err := primitive.ObjectIDFromHex(rr.Id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionRoyaltyReport),
			zap.String(pkg.ErrorDatabaseFieldQuery, rr.Id),
		)
		return nil
	}

	filter := bson.M{"_id": oid}
	_, err = r.db.Collection(collectionRoyaltyReport).ReplaceOne(ctx, filter, rr)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionRoyaltyReport),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationUpdate),
			zap.Any(pkg.ErrorDatabaseFieldDocument, rr),
		)

		return err
	}

	if err = r.onRoyaltyReportChange(ctx, rr, ip, source); err != nil {
		return err
	}

	key := fmt.Sprintf(cacheKeyRoyaltyReport, rr.Id)

	if err = r.cache.Set(key, rr, 0); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorCacheFieldData, rr),
		)
	}

	return nil
}

func (r *royaltyReportRepository) GetById(ctx context.Context, id string) (rr *billingpb.RoyaltyReport, err error) {
	var c billingpb.RoyaltyReport
	key := fmt.Sprintf(cacheKeyRoyaltyReport, id)

	if err := r.cache.Get(key, c); err == nil {
		return &c, nil
	}

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionRoyaltyReport),
			zap.String(pkg.ErrorDatabaseFieldQuery, id),
		)
		return
	}

	filter := bson.M{"_id": oid}
	err = r.db.Collection(collectionRoyaltyReport).FindOne(ctx, filter).Decode(&rr)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionRoyaltyReport),
			zap.String(pkg.ErrorDatabaseFieldDocumentId, id),
		)
		return
	}

	if err = r.cache.Set(key, rr, 0); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorCacheFieldData, rr),
		)
		// suppress error returning here
		err = nil
	}

	return
}

func (r *royaltyReportRepository) onRoyaltyReportChange(
	ctx context.Context,
	document *billingpb.RoyaltyReport,
	ip, source string,
) (err error) {
	change := &billingpb.RoyaltyReportChanges{
		Id:              primitive.NewObjectID().Hex(),
		RoyaltyReportId: document.Id,
		Source:          source,
		Ip:              ip,
	}

	b, err := json.Marshal(document)

	if err != nil {
		zap.L().Error(
			pkg.ErrorJsonMarshallingFailed,
			zap.Error(err),
			zap.Any("document", document),
		)
		return
	}
	hash := md5.New()
	hash.Write(b)
	change.Hash = hex.EncodeToString(hash.Sum(nil))

	_, err = r.db.Collection(collectionRoyaltyReportChanges).InsertOne(ctx, change)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionRoyaltyReportChanges),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldDocument, change),
		)
		return
	}

	return
}

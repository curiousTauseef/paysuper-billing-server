package repository

import (
	"context"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
	"time"
)

const (
	collectionPaylinkVisits = "paylink_visits"
)

type paylinkVisitRepository repository

// NewPaylinkVisitRepository create and return an object for working with the paylink visit repository.
// The returned object implements the PaylinkVisitRepositoryInterface interface.
func NewPaylinkVisitRepository(db mongodb.SourceInterface) PaylinkVisitRepositoryInterface {
	s := &paylinkVisitRepository{db: db}
	return s
}

func (r *paylinkVisitRepository) CountPaylinkVisits(ctx context.Context, id string, from, to int64) (int64, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaylinkVisits),
			zap.String(pkg.ErrorDatabaseFieldQuery, id),
		)
	}

	query := bson.M{"paylink_id": oid}

	if from > 0 || to > 0 {
		date := bson.M{}
		if from > 0 {
			date["$gte"] = time.Unix(from, 0)
		}
		if to > 0 {
			date["$lte"] = time.Unix(to, 0)
		}
		query["date"] = date
	}

	count, err := r.db.Collection(collectionPaylinkVisits).CountDocuments(ctx, query)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaylinkVisits),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationCount),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return int64(0), err
	}

	return count, nil
}

func (r *paylinkVisitRepository) IncrVisits(ctx context.Context, id string) (err error) {
	oid, _ := primitive.ObjectIDFromHex(id)
	visit := &paylinkVisits{PaylinkId: oid, Date: time.Now()}
	_, err = r.db.Collection(collectionPaylinkVisits).InsertOne(ctx, visit)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPaylinkVisits),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldDocument, visit),
		)
	}

	return
}

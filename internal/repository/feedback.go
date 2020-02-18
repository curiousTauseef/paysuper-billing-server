package repository

import (
	"context"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
)

const (
	collectionFeedback = "feedback"
)

type feedbackRepository repository

// NewFeedbackRepository create and return an object for working with the feedback repository.
// The returned object implements the FeedbackRepositoryInterface interface.
func NewFeedbackRepository(db mongodb.SourceInterface) FeedbackRepositoryInterface {
	s := &feedbackRepository{db: db}
	return s
}

func (r *feedbackRepository) Insert(ctx context.Context, obj *billingpb.PageReview) error {
	_, err := r.db.Collection(collectionFeedback).InsertOne(ctx, obj)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionFeedback),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, obj),
		)
		return err
	}

	return nil
}

func (r feedbackRepository) GetAll(ctx context.Context) ([]*billingpb.PageReview, error) {
	c := []*billingpb.PageReview{}

	cursor, err := r.db.Collection(collectionFeedback).Find(ctx, bson.M{})

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionFeedback),
		)
		return nil, err
	}

	err = cursor.All(ctx, &c)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionFeedback),
		)
		return nil, err
	}

	return c, nil
}

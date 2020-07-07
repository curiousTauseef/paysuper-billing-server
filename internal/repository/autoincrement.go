package repository

import (
	"context"
	"github.com/paysuper/paysuper-billing-server/internal/repository/models"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
	"time"
)

const (
	collectionAutoincrement = "autoincrement"
)

type autoincrementRepository repository

func NewAutoincrementRepository(db mongodb.SourceInterface) AutoincrementRepositoryInterface {
	return &autoincrementRepository{db: db}
}

func (m *autoincrementRepository) GatPayoutAutoincrementId(ctx context.Context) (int64, error) {
	doc, err := m.getAutoincrement(ctx, collectionPayoutDocuments)

	if err != nil {
		return 0, err
	}

	return doc.Counter, nil
}

func (m *autoincrementRepository) getAutoincrement(ctx context.Context, collection string) (*models.Autoincrement, error) {
	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)
	doc := new(models.Autoincrement)
	filter := bson.M{"collection": collection}
	update := bson.M{
		"$inc": bson.M{"counter": 1},
		"$set": bson.M{"updated_at": time.Now()},
	}
	err := m.db.Collection(collectionAutoincrement).FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionAutoincrement),
			zap.Any(pkg.ErrorDatabaseFieldQuery, filter),
			zap.Any(pkg.ErrorDatabaseFieldOperationUpdate, update),
		)
	}

	return doc, err
}

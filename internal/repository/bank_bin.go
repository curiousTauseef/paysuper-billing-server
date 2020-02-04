package repository

import (
	"context"
	intPkg "github.com/paysuper/paysuper-billing-server/internal/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
)

const (
	collectionBinData = "bank_bin"
)

type bankBinRepository repository

// NewBankBinRepository create and return an object for working with the bank bin repository.
// The returned object implements the BankBinRepositoryInterface interface.
func NewBankBinRepository(db mongodb.SourceInterface) BankBinRepositoryInterface {
	s := &bankBinRepository{db: db}
	return s
}

func (r *bankBinRepository) Insert(ctx context.Context, obj *intPkg.BinData) error {
	_, err := r.db.Collection(collectionBinData).InsertOne(ctx, obj)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionBinData),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, obj),
		)
		return err
	}

	return nil
}

func (r bankBinRepository) MultipleInsert(ctx context.Context, pg []*intPkg.BinData) error {
	c := make([]interface{}, len(pg))
	for i, v := range pg {
		c[i] = v
	}

	_, err := r.db.Collection(collectionBinData).InsertMany(ctx, c)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionBinData),
			zap.String(pkg.ErrorDatabaseFieldOperation, pkg.ErrorDatabaseFieldOperationInsert),
			zap.Any(pkg.ErrorDatabaseFieldQuery, c),
		)
		return err
	}

	return nil
}

func (r bankBinRepository) GetByBin(ctx context.Context, bin int32) (*intPkg.BinData, error) {
	var obj intPkg.BinData

	query := bson.M{"card_bin": bin}
	err := r.db.Collection(collectionBinData).FindOne(ctx, query).Decode(&obj)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionBinData),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, err
	}

	return &obj, nil
}

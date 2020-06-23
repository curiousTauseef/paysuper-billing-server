package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// PriceTableRepositoryInterface is abstraction layer for working with price table and representation in database.
type PriceTableRepositoryInterface interface {
	// Insert adds the price table to the collection.
	Insert(context.Context, *billingpb.PriceTable) error

	// GetByRegion returns the price table by region name.
	GetByRegion(context.Context, string) (*billingpb.PriceTable, error)
}

package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// PriceGroupRepositoryInterface is abstraction layer for working with price group and representation in database.
type PriceGroupRepositoryInterface interface {
	// Insert adds the price group to the collection.
	Insert(context.Context, *billingpb.PriceGroup) error

	// Insert adds the multiple price groups to the collection.
	MultipleInsert(context.Context, []*billingpb.PriceGroup) error

	// Update updates the price group in the collection.
	Update(context.Context, *billingpb.PriceGroup) error

	// GetById returns the price group by unique identity.
	GetById(context.Context, string) (*billingpb.PriceGroup, error)

	// GetByRegion returns the price group by region name.
	GetByRegion(context.Context, string) (*billingpb.PriceGroup, error)

	// GetByRegion returns all price groups.
	GetAll(context.Context) ([]*billingpb.PriceGroup, error)
}

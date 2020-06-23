package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// KeyProductRepositoryInterface is abstraction layer for working with key for product and representation in database.
type KeyProductRepositoryInterface interface {
	// Upsert add or update the key product to the collection.
	Upsert(context.Context, *billingpb.KeyProduct) error

	// Update updates the key product in the collection.
	Update(context.Context, *billingpb.KeyProduct) error

	// GetById returns the key product by unique identity.
	GetById(context.Context, string) (*billingpb.KeyProduct, error)

	// CountByProjectIdSku return count the key products by project id and sku.
	CountByProjectIdSku(context.Context, string, string) (int64, error)

	// FindByIdsProjectId find and return the key products by list identifiers and project id.
	FindByIdsProjectId(context.Context, []string, string) ([]*billingpb.KeyProduct, error)

	// Find returns list of key products by merchant id, project id, sku, name and enabled with pagination.
	Find(context.Context, string, string, string, string, string, int64, int64) ([]*billingpb.KeyProduct, error)

	// FindCount returns count of key products by merchant id, project id, sku, name and enabled with pagination.
	FindCount(context.Context, string, string, string, string, string) (int64, error)
}

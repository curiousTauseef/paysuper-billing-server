package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// ProductRepositoryInterface is abstraction layer for working with product and representation in database.
type ProductRepositoryInterface interface {
	// Upsert add or update the product to the collection.
	Upsert(ctx context.Context, product *billingpb.Product) error

	// GetById returns the product by unique identity.
	GetById(context.Context, string) (*billingpb.Product, error)

	// CountByProjectSku return count the products by project id and sku.
	CountByProjectSku(context.Context, string, string) (int64, error)

	// List returns list of products by merchant id, project id, sku, name and enabled with pagination.
	Find(context.Context, string, string, string, string, int32, int64, int64) ([]*billingpb.Product, error)

	// FindCount returns count of products by merchant id, project id, sku, name and enabled.
	FindCount(context.Context, string, string, string, string, int32) (int64, error)
}

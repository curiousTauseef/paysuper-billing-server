package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// KeyProductRepositoryInterface is abstraction layer for working with key for product and representation in database.
type KeyProductRepositoryInterface interface {
	Upsert(context.Context, *billingpb.KeyProduct) error
	Update(context.Context, *billingpb.KeyProduct) error
	GetById(context.Context, string) (*billingpb.KeyProduct, error)
	CountByProjectIdSku(context.Context, string, string) (int64, error)
	FindByIdsProjectId(context.Context, []string, string) ([]*billingpb.KeyProduct, error)
	Find(context.Context, string, string, string, string, string, int64, int64) ([]*billingpb.KeyProduct, error)
	FindCount(context.Context, string, string, string, string, string) (int64, error)
}

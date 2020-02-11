package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// PaylinkRepositoryInterface is abstraction layer for working with paylink and representation in database.
type PaylinkRepositoryInterface interface {
	// Insert adds paylink to the collection.
	Insert(context.Context, *billingpb.Paylink) error

	// Update updates the paylink in the collection.
	Update(context.Context, *billingpb.Paylink) error

	// Delete deletes the paylink in the collection.
	Delete(context.Context, *billingpb.Paylink) error

	// GetById returns the paylink by unique identifier.
	GetById(context.Context, string) (*billingpb.Paylink, error)

	// GetByIdAndMerchant returns the paylink by unique identifier and merchant id.
	GetByIdAndMerchant(context.Context, string, string) (*billingpb.Paylink, error)

	// UpdateTotalStat updates data from orders to paylink statistics.
	UpdateTotalStat(context.Context, *billingpb.Paylink) error

	// List returns list of paylinks by merchant id and project id with pagination.
	Find(context.Context, string, string, int64, int64) ([]*billingpb.Paylink, error)

	// FindCount returns count of paylinks by merchant id and project id.
	FindCount(context.Context, string, string) (int64, error)
}

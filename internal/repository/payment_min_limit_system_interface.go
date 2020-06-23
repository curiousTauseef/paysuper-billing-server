package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// PaymentMinLimitSystemRepositoryInterface is abstraction layer for working with payment min limit system and representation in database.
type PaymentMinLimitSystemRepositoryInterface interface {
	// MultipleInsert adds the multiple payment min limits to the collection.
	MultipleInsert(context.Context, []*billingpb.PaymentMinLimitSystem) error

	// Upsert add or update the payment min limit in the collection.
	Upsert(context.Context, *billingpb.PaymentMinLimitSystem) error

	// GetByCurrency returns the payment min limit by currency code.
	GetByCurrency(context.Context, string) (*billingpb.PaymentMinLimitSystem, error)

	// GetAll returns all payment min limits.
	GetAll(context.Context) ([]*billingpb.PaymentMinLimitSystem, error)
}

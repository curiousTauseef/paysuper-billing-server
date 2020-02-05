package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// PaymentSystemRepositoryInterface is abstraction layer for working with payment system and representation in database.
type PaymentSystemRepositoryInterface interface {
	// Insert adds the notify sales to the collection.
	Insert(context.Context, *billingpb.PaymentSystem) error

	// MultipleInsert adds the multiple payment systems to the collection.
	MultipleInsert(context.Context, []*billingpb.PaymentSystem) error

	// Update updates the payment system in the collection.
	Update(context.Context, *billingpb.PaymentSystem) error

	// GetById returns the payment system by unique identifier.
	GetById(context.Context, string) (*billingpb.PaymentSystem, error)
}

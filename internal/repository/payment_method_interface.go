package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// PaymentMethodRepositoryInterface is abstraction layer for working with payment method and representation in database.
type PaymentMethodRepositoryInterface interface {
	// Insert adds the payment method to the collection.
	Insert(context.Context, *billingpb.PaymentMethod) error

	// MultipleInsert adds the multiple payment method to the collection.
	MultipleInsert(context.Context, []*billingpb.PaymentMethod) error

	// Update updates the payment method in the collection.
	Update(context.Context, *billingpb.PaymentMethod) error

	// Delete deleting the payment method.
	Delete(context.Context, *billingpb.PaymentMethod) error

	// GetById returns the payment method by unique identity.
	GetById(context.Context, string) (*billingpb.PaymentMethod, error)

	// GetAll returns the list of payment method.
	GetAll(ctx context.Context) ([]*billingpb.PaymentMethod, error)

	// GetByGroupAndCurrency returns the payment method by group and currency code.
	GetByGroupAndCurrency(ctx context.Context, isProduction bool, group string, currency string) (*billingpb.PaymentMethod, error)

	// GetByGroupAndCurrency returns the list of payment methods in order properties.
	ListByOrder(ctx context.Context, order *billingpb.Order) ([]*billingpb.PaymentMethod, error)

	// FindByName find by name and returns the list of payment methods.
	FindByName(context.Context, string, []string) ([]*billingpb.PaymentMethod, error)
}

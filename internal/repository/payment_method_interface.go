package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// PaymentMethodRepositoryInterface is abstraction layer for working with payment method and representation in database.
type PaymentMethodRepositoryInterface interface {
	Insert(context.Context, *billingpb.PaymentMethod) error
	MultipleInsert(context.Context, []*billingpb.PaymentMethod) error
	Update(context.Context, *billingpb.PaymentMethod) error
	Delete(context.Context, *billingpb.PaymentMethod) error
	GetById(context.Context, string) (*billingpb.PaymentMethod, error)
	GetAll(ctx context.Context) ([]*billingpb.PaymentMethod, error)
	GetByGroupAndCurrency(ctx context.Context, isProduction bool, group string, currency string) (*billingpb.PaymentMethod, error)
	ListByOrder(ctx context.Context, order *billingpb.Order) ([]*billingpb.PaymentMethod, error)
	FindByName(context.Context, string, []string) ([]*billingpb.PaymentMethod, error)
}

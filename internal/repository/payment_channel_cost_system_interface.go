package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// PaymentChannelCostSystemRepositoryInterface is abstraction layer for working with payment channel cost
// for merchant and representation in database.
type PaymentChannelCostSystemRepositoryInterface interface {
	// Insert adds the payment channel cost to the collection.
	Insert(context.Context, *billingpb.PaymentChannelCostSystem) error

	// Insert adds the multiple payment channel costs to the collection.
	MultipleInsert(context.Context, []*billingpb.PaymentChannelCostSystem) error

	// Update updates the payment channel cost in the collection.
	Update(context.Context, *billingpb.PaymentChannelCostSystem) error

	// Delete вудуеу the payment channel cost in the collection.
	Delete(ctx context.Context, obj *billingpb.PaymentChannelCostSystem) error

	// GetById returns the payment channel cost by unique identity.
	GetById(context.Context, string) (*billingpb.PaymentChannelCostSystem, error)

	// GetAllForMerchant returns all payment channel costs for merchant.
	GetAll(context.Context) ([]*billingpb.PaymentChannelCostSystem, error)

	// Find returns the payment channel costs by name, region, country, mcc code and operating company id.
	Find(context.Context, string, string, string, string, string) (*billingpb.PaymentChannelCostSystem, error)
}

package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OrderRepositoryInterface is abstraction layer for working with order and representation in database.
type OrderRepositoryInterface interface {
	// Insert adds order to the collection.
	Insert(context.Context, *billingpb.Order) error

	// Update updates the order in the collection.
	Update(context.Context, *billingpb.Order) error

	// GetById returns a order by its identifier.
	GetById(context.Context, string) (*billingpb.Order, error)

	// GetByUuid returns a order by its public (uuid) identifier.
	GetByUuid(context.Context, string) (*billingpb.Order, error)

	// GetByUuid returns a order by its public (uuid) identifier and merchant identifier.
	GetByUuidAndMerchantId(ctx context.Context, uuid string, merchantId string) (*billingpb.Order, error)

	// GetByRefundReceiptNumber returns a order by its receipt number.
	GetByRefundReceiptNumber(context.Context, string) (*billingpb.Order, error)

	// UpdateOrderView updates orders into order view.
	UpdateOrderView(context.Context, []string) error

	// Return order by some conditions
	GetOneBy(ctx context.Context, filter bson.M) (*billingpb.Order, error)

	// Mark orders as included to royalty report.
	IncludeToRoyaltyReport(ctx context.Context, ordersIds []primitive.ObjectID, royaltyReportId string) error
}

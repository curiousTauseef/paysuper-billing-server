package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// KeyRepositoryInterface is abstraction layer for working with key and representation in database.
type KeyRepositoryInterface interface {
	// Insert add the key to the collection.
	Insert(context.Context, *billingpb.Key) error

	// GetById returns the key by unique identity.
	GetById(context.Context, string) (*billingpb.Key, error)

	// ReserveKey reserves the key for the specified time for the product, platform and order.
	ReserveKey(context.Context, string, string, string, int32) (*billingpb.Key, error)

	// CancelById cancels the key reserve.
	CancelById(context.Context, string) (*billingpb.Key, error)

	// FinishRedeemById marks the reserved key as successfully used.
	FinishRedeemById(context.Context, string) (*billingpb.Key, error)

	// CountKeysByProductPlatform returns the number of keys for the product and the specified platform.
	CountKeysByProductPlatform(context.Context, string, string) (int64, error)

	// FindUnfinished returns a list of keys that have not been revoked or used.
	FindUnfinished(context.Context) ([]*billingpb.Key, error)
}

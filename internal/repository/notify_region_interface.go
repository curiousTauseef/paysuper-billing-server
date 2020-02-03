package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

const (
	collectionNotifyNewRegion = "notify_new_region"
)

// NotifyRegionRepositoryInterface is abstraction layer for working with notify new region and representation in database.
type NotifyRegionRepositoryInterface interface {
	// Insert adds the notify new region to the collection.
	Insert(context.Context, *billingpb.NotifyUserNewRegion) error

	// FindByEmail returns the notify new region by email.
	FindByEmail(context.Context, string) ([]*billingpb.NotifyUserNewRegion, error)
}

package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// NotifySalesRepositoryInterface is abstraction layer for working with notify sales and representation in database.
type NotifySalesRepositoryInterface interface {
	// Insert adds the notify sales to the collection.
	Insert(context.Context, *billingpb.NotifyUserSales) error

	// FindByEmail returns the notify sales by email.
	FindByEmail(context.Context, string) ([]*billingpb.NotifyUserSales, error)
}

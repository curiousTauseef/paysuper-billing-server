package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

const (
	collectionNotification = "notification"
)

// NotificationRepositoryInterface is abstraction layer for working with notification and representation in database.
type NotificationRepositoryInterface interface {
	// Insert adds the price table to the collection.
	Insert(context.Context, *billingpb.Notification) error

	// Insert adds the price table to the collection.
	Update(context.Context, *billingpb.Notification) error

	// GetById returns the price table by unique identifier.
	GetById(context.Context, string) (*billingpb.Notification, error)

	// Find notifications by merchant, user id and type with pagination and sortable.
	Find(context.Context, string, string, int32, []string, int64, int64) ([]*billingpb.Notification, error)

	// Find return count of notifications by merchant, user id and type.
	FindCount(context.Context, string, string, int32) (int64, error)
}

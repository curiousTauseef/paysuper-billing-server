package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// FeedbackRepositoryInterface is abstraction layer for working with feedback and representation in database.
type FeedbackRepositoryInterface interface {
	// Insert adds the feedback to the collection.
	Insert(context.Context, *billingpb.PageReview) error

	// GetAll returns all feedback.
	GetAll(context.Context) ([]*billingpb.PageReview, error)
}

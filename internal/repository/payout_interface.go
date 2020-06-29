package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// PayoutRepositoryInterface is abstraction layer for working with payout and representation in database.
type PayoutRepositoryInterface interface {
	// Insert adds the payout to the collection.
	Insert(context.Context, *billingpb.PayoutDocument, string, string) error

	// Update updates the payout in the collection.
	Update(context.Context, *billingpb.PayoutDocument, string, string) error

	// GetById returns the payout by unique identity.
	GetById(context.Context, string) (*billingpb.PayoutDocument, error)

	// GetByIdMerchantId returns the payout by unique identity and merchant id.
	GetByIdMerchantId(context.Context, string, string) (*billingpb.PayoutDocument, error)

	// GetBalanceAmount returns balance amount by merchant id and currency code.
	GetBalanceAmount(context.Context, string, string) (float64, error)

	// GetLast returns the latest payout doc by merchant id and currency.
	GetLast(context.Context, string, string) (*billingpb.PayoutDocument, error)

	// Find payouts by merchant, statuses and dates from/to with pagination.
	Find(ctx context.Context, in *billingpb.GetPayoutDocumentsRequest) ([]*billingpb.PayoutDocument, error)

	// FindCount return count of payouts by merchant, statuses and dates from/to.
	FindCount(ctx context.Context, in *billingpb.GetPayoutDocumentsRequest) (int64, error)

	// Return all not paid payout invoices
	FindAll(ctx context.Context) ([]*billingpb.PayoutDocument, error)
}

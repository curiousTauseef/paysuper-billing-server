package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"time"
)

// RoyaltyReportRepositoryInterface is abstraction layer for working with royalty report and representation in database.
type RoyaltyReportRepositoryInterface interface {
	// Insert add the royalty report to the collection.
	Insert(ctx context.Context, document *billingpb.RoyaltyReport, ip, source string) error

	// Update updates the royalty report in the collection.
	Update(ctx context.Context, document *billingpb.RoyaltyReport, ip, source string) error

	// GetById returns the royalty report by unique identity.
	GetById(ctx context.Context, id string) (*billingpb.RoyaltyReport, error)

	// GetNonPayoutReports returns the royalty report by unique identity.
	GetNonPayoutReports(ctx context.Context, merchantId, currency string) ([]*billingpb.RoyaltyReport, error)

	// GetByPayoutId returns the royalty report by payout identity.
	GetByPayoutId(ctx context.Context, payoutId string) ([]*billingpb.RoyaltyReport, error)

	// GetBalanceAmount returns royalty balance amount for merchant and currency.
	GetBalanceAmount(ctx context.Context, merchantId, currency string) (float64, error)

	// GetReportExists returns exists a royalty reports by merchant id, currency and dates from/to.
	GetReportExists(ctx context.Context, merchantId, currency string, from, to time.Time) (report *billingpb.RoyaltyReport)
}

package repository

import (
	"context"
	"github.com/paysuper/paysuper-billing-server/internal/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"time"
)

// AccountingEntryRepositoryInterface is abstraction layer for working with accounting entry and representation in database.
type AccountingEntryRepositoryInterface interface {
	// Insert adds multiple accounting entries to the collection.
	MultipleInsert(context.Context, []*billingpb.AccountingEntry) error

	GetCorrectionsForRoyaltyReport(context.Context, string, string, time.Time, time.Time) ([]*billingpb.AccountingEntry, error)
	GetRollingReservesForRoyaltyReport(context.Context, string, string, []string, time.Time, time.Time) ([]*billingpb.AccountingEntry, error)
	GetRollingReserveForBalance(context.Context, string, string, []string, time.Time) ([]*pkg.ReserveQueryResItem, error)
	GetById(context.Context, string) (*billingpb.AccountingEntry, error)
	GetByTypeWithTaxes(context.Context, []string) ([]*billingpb.AccountingEntry, error)
	GetByObjectSource(context.Context, string, string, string) (*billingpb.AccountingEntry, error)
	ApplyObjectSource(context.Context, string, string, string, string, *billingpb.AccountingEntry) error
	GetDistinctBySourceId(context.Context) ([]string, error)
	FindBySource(context.Context, string, string) ([]*billingpb.AccountingEntry, error)
	FindByTypeCountryDates(context.Context, string, []string, time.Time, time.Time) ([]*billingpb.AccountingEntry, error)
	BulkWrite(context.Context, []*billingpb.AccountingEntry) error
}

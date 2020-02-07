package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"time"
)

// VatReportRepositoryInterface is abstraction layer for working with vat report and representation in database.
type VatReportRepositoryInterface interface {
	// Insert adds a vat report to the collection.
	Insert(context.Context, *billingpb.VatReport) error

	// Update updates a vat report in the collection.
	Update(context.Context, *billingpb.VatReport) error

	// GetById returns a vat report by unique identity.
	GetById(context.Context, string) (*billingpb.VatReport, error)

	// GetById returns list of a vat reports by country code.
	GetByCountry(context.Context, string, []string, int64, int64) ([]*billingpb.VatReport, error)

	// GetByStatus returns list of a vat reports by statuses.
	GetByStatus(context.Context, []string) ([]*billingpb.VatReport, error)

	// GetByCountryPeriod returns a vat report by country and period.
	GetByCountryPeriod(context.Context, string, time.Time, time.Time) (*billingpb.VatReport, error)
}

package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// OperatingCompanyRepositoryInterface is abstraction layer for working with operating company and representation in database.
type OperatingCompanyRepositoryInterface interface {
	// GetById returns the operating company by unique identifier.
	GetById(ctx context.Context, id string) (oc *billingpb.OperatingCompany, err error)

	// GetByPaymentCountry returns the operating company by country code.
	GetByPaymentCountry(ctx context.Context, countryCode string) (oc *billingpb.OperatingCompany, err error)

	// GetAll returns all operating companies.
	GetAll(ctx context.Context) (result []*billingpb.OperatingCompany, err error)

	// Upsert add or update the operating company to the collection.
	Upsert(ctx context.Context, oc *billingpb.OperatingCompany) (err error)

	// Exists is check exists the operating company by unique identifier.
	Exists(ctx context.Context, id string) bool
}

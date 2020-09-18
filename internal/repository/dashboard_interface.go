package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// DashboardRepositoryInterface is abstraction layer for working with dashboard report and representation in database.
type DashboardRepositoryInterface interface {
	GetMainReport(ctx context.Context, merchantId string, period string) (*billingpb.DashboardMainReport, error)
	GetRevenueDynamicsReport(ctx context.Context, merchantId string, period string) (*billingpb.DashboardRevenueDynamicReport, error)
	GetBaseReport(ctx context.Context,merchantId string, period string) (*billingpb.DashboardBaseReports, error)

	GetCustomersReport(ctx context.Context, merchantId string, period string) (*billingpb.DashboardCustomerReport, error)
	GetCustomerARPU(ctx context.Context, merchantId string, customerId string) (*billingpb.DashboardAmountItemWithChart, error)
}

package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// DashboardRepositoryInterface is abstraction layer for working with dashboard report and representation in database.
type DashboardRepositoryInterface interface {
	GetMainReport(context.Context, string, string) (*billingpb.DashboardMainReport, error)
	GetRevenueDynamicsReport(context.Context, string, string) (*billingpb.DashboardRevenueDynamicReport, error)
	GetBaseReport(context.Context, string, string) (*billingpb.DashboardBaseReports, error)
}

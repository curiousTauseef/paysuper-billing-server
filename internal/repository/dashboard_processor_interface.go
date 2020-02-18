package repository

import "context"

// DashboardProcessorRepositoryInterface is abstraction layer for working with dashboard report and representation in database.
type DashboardProcessorRepositoryInterface interface {
	ExecuteReport(context.Context, interface{}, func(context.Context, interface{}) (interface{}, error)) (interface{}, error)
	ExecuteGrossRevenueAndVatReports(context.Context, interface{}) (interface{}, error)
	ExecuteTotalTransactionsAndArpuReports(context.Context, interface{}) (interface{}, error)
	ExecuteRevenueDynamicReport(context.Context, interface{}) (interface{}, error)
	ExecuteRevenueByCountryReport(context.Context, interface{}) (interface{}, error)
	ExecuteSalesTodayReport(context.Context, interface{}) (interface{}, error)
	ExecuteSourcesReport(context.Context, interface{}) (interface{}, error)
}

package repository

import (
	"context"
	"github.com/paysuper/paysuper-billing-server/internal/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

// OrderViewRepositoryInterface is abstraction layer for working with view of order and representation in database.
type OrderViewRepositoryInterface interface {
	GetById(context.Context, string) (*billingpb.OrderViewPublic, error)
	CountTransactions(ctx context.Context, match bson.M) (n int64, err error)
	GetTransactionsPublic(ctx context.Context, match bson.M, limit, offset int64) (result []*billingpb.OrderViewPublic, err error)
	GetTransactionsPrivate(ctx context.Context, match bson.M, limit, offset int64) (result []*billingpb.OrderViewPrivate, err error)
	GetRoyaltySummary(ctx context.Context, merchantId, currency string, from, to time.Time) (items []*billingpb.RoyaltyReportProductSummaryItem, total *billingpb.RoyaltyReportProductSummaryItem, err error)
	GetOrderBy(ctx context.Context, id, uuid, merchantId string, receiver interface{}) (interface{}, error)
	GetPaylinkStat(ctx context.Context, paylinkId, merchantId string, from, to int64) (*billingpb.StatCommon, error)
	GetPaylinkStatByCountry(ctx context.Context, paylinkId, merchantId string, from, to int64) (result *billingpb.GroupStatCommon, err error)
	GetPaylinkStatByReferrer(ctx context.Context, paylinkId, merchantId string, from, to int64) (result *billingpb.GroupStatCommon, err error)
	GetPaylinkStatByDate(ctx context.Context, paylinkId, merchantId string, from, to int64) (result *billingpb.GroupStatCommon, err error)
	GetPaylinkStatByUtm(ctx context.Context, paylinkId, merchantId string, from, to int64) (result *billingpb.GroupStatCommon, err error)
	GetPublicByOrderId(ctx context.Context, merchantId string) (*billingpb.OrderViewPublic, error)
	GetVatSummary(context.Context, string, string, bool, time.Time, time.Time) ([]*pkg.VatReportQueryResItem, error)
	GetTurnoverSummary(context.Context, string, string, string, time.Time, time.Time) ([]*pkg.TurnoverQueryResItem, error)
	GetRoyaltyForMerchants(context.Context, []string, time.Time, time.Time) ([]*pkg.RoyaltyReportMerchant, error)
}

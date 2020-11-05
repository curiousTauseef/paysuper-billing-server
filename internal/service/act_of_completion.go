package service

import (
	"context"
	"github.com/jinzhu/now"
	"github.com/paysuper/paysuper-billing-server/pkg/errors"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"math"
	"time"
)

var (
	invalidActOfCompletionDateFrom = errors.NewBillingServerErrorMsg("aoc000001", "invalid start date the act of completion")
	invalidActOfCompletionDateTo   = errors.NewBillingServerErrorMsg("aoc000002", "invalid end date the act of completion")
	invalidActOfCompletionMerchant = errors.NewBillingServerErrorMsg("aoc000003", "invalid merchant identity the act of completion")
)

func (s *Service) GetActOfCompletion(
	ctx context.Context,
	req *billingpb.ActOfCompletionRequest,
	rsp *billingpb.ActOfCompletionResponse,
) error {
	dateFrom, err := time.Parse(billingpb.FilterDateFormat, req.DateFrom)

	if err != nil {
		rsp.Status = billingpb.ResponseStatusBadData
		rsp.Message = invalidActOfCompletionDateFrom

		return nil
	}

	dateTo, err := time.Parse(billingpb.FilterDateFormat, req.DateTo)

	if err != nil {
		rsp.Status = billingpb.ResponseStatusBadData
		rsp.Message = invalidActOfCompletionDateTo

		return nil
	}

	dateFrom = now.New(dateFrom).BeginningOfDay()
	dateTo = now.New(dateTo).EndOfDay()

	merchant, err := s.merchantRepository.GetById(ctx, req.MerchantId)
	if err != nil {
		rsp.Status = billingpb.ResponseStatusBadData
		rsp.Message = invalidActOfCompletionMerchant

		return nil
	}

	royaltyHandler := &royaltyHandler{
		Service: s,
		from:    dateFrom,
		to:      dateTo,
	}
	report, _, err := royaltyHandler.buildMerchantRoyaltyReportRoundedAmounts(ctx, merchant, false)
	if err != nil {
		return err
	}

	payoutAmount := report.Summary.ProductsTotal.GrossTotalAmount - report.Summary.ProductsTotal.TotalFees - report.Summary.ProductsTotal.TotalVat
	totalFeesAmount := payoutAmount + report.Totals.CorrectionAmount
	balanceAmount := payoutAmount + report.Totals.CorrectionAmount - report.Totals.RollingReserveAmount

	rsp.Status = billingpb.ResponseStatusOk
	rsp.Item = &billingpb.ActOfCompletionDocument{
		MerchantId:        merchant.Id,
		TotalFees:         math.Round(totalFeesAmount*100) / 100,
		Balance:           math.Round(balanceAmount*100) / 100,
		TotalTransactions: report.Totals.TransactionsCount,
	}

	return nil
}

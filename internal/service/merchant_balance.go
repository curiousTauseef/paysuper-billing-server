package service

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/errors"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"time"
)

var (
	errorMerchantPayoutCurrencyNotSet = errors.NewBillingServerErrorMsg("ba000001", "merchant payout currency not set")

	accountingEntriesForRollingReserve = []string{
		pkg.AccountingEntryTypeMerchantRollingReserveCreate,
		pkg.AccountingEntryTypeMerchantRollingReserveRelease,
	}
)

func (s *Service) GetMerchantBalance(
	ctx context.Context,
	req *billingpb.GetMerchantBalanceRequest,
	res *billingpb.GetMerchantBalanceResponse,
) error {
	var err error
	res.Status = billingpb.ResponseStatusOk

	res.Item, err = s.getMerchantBalance(ctx, req.MerchantId)
	if err == nil {
		return nil
	}

	if err == mongo.ErrNoDocuments {
		res.Item, err = s.updateMerchantBalance(ctx, req.MerchantId)
		if err != nil {
			if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
				res.Status = billingpb.ResponseStatusSystemError
				res.Message = e
				return nil
			}
			return err
		}
	} else {
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			res.Status = billingpb.ResponseStatusSystemError
			res.Message = e
			return nil
		}
		return err
	}

	return nil
}

func (s *Service) getMerchantBalance(ctx context.Context, merchantId string) (*billingpb.MerchantBalance, error) {
	merchant, err := s.merchantRepository.GetById(ctx, merchantId)
	if err != nil {
		return nil, merchantErrorNotFound
	}

	if merchant.GetPayoutCurrency() == "" {
		zap.L().Error(errorMerchantPayoutCurrencyNotSet.Error(), zap.String("merchant_id", merchantId))
		return nil, errorMerchantPayoutCurrencyNotSet
	}

	return s.merchantBalanceRepository.GetByIdAndCurrency(ctx, merchant.Id, merchant.GetPayoutCurrency())
}

func (s *Service) updateMerchantBalance(ctx context.Context, merchantId string) (*billingpb.MerchantBalance, error) {
	merchant, err := s.merchantRepository.GetById(ctx, merchantId)
	if err != nil {
		return nil, merchantErrorNotFound
	}

	if merchant.GetPayoutCurrency() == "" {
		zap.L().Error(errorMerchantPayoutCurrencyNotSet.Error(), zap.String("merchant_id", merchantId))
		return nil, errorMerchantPayoutCurrencyNotSet
	}

	debit, err := s.royaltyReportRepository.GetBalanceAmount(ctx, merchant.Id, merchant.GetPayoutCurrency())
	if err != nil {
		return nil, err
	}

	credit, err := s.payoutRepository.GetBalanceAmount(ctx, merchant.Id, merchant.GetPayoutCurrency())
	if err != nil {
		return nil, err
	}

	rr, err := s.getRollingReserveForBalance(ctx, merchantId, merchant.GetPayoutCurrency())
	if err != nil {
		return nil, err
	}

	balance := &billingpb.MerchantBalance{
		Id:             primitive.NewObjectID().Hex(),
		MerchantId:     merchantId,
		Currency:       merchant.GetPayoutCurrency(),
		Debit:          debit,
		Credit:         credit,
		RollingReserve: rr,
		CreatedAt:      ptypes.TimestampNow(),
	}

	balance.Total = balance.Debit - balance.Credit - balance.RollingReserve

	err = s.merchantBalanceRepository.Insert(ctx, balance)

	if err != nil {
		return nil, err
	}

	return balance, nil
}

func (s *Service) getRollingReserveForBalance(ctx context.Context, merchantId, currency string) (float64, error) {
	pd, err := s.payoutRepository.GetLast(ctx, merchantId, currency)
	if err != nil && err != mongo.ErrNoDocuments {
		return 0, err
	}

	createdAt := time.Time{}
	if pd != nil {
		createdAt, err = ptypes.Timestamp(pd.CreatedAt)
		if err != nil {
			return 0, err
		}
	}

	items, err := s.accountingRepository.GetRollingReserveForBalance(
		ctx, merchantId, currency, accountingEntriesForRollingReserve, createdAt,
	)

	if err != nil {
		return 0, err
	}

	if len(items) == 0 {
		return 0, nil
	}

	result := float64(0)

	for _, i := range items {
		// in case of rolling reserve release, result will have negative amount, it is ok
		if i.Type == pkg.AccountingEntryTypeMerchantRollingReserveRelease {
			result -= i.Amount
		} else {
			result += i.Amount
		}
	}

	return result, nil
}

package service

import (
	"context"
	"github.com/paysuper/paysuper-billing-server/internal/helper"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/currenciespb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
	"sort"
)

var (
	errorPaymentChannelMerchantGetAll           = newBillingServerErrorMsg("pcm000001", "can't get list of payment channel setting for merchant")
	errorPaymentChannelMerchantGet              = newBillingServerErrorMsg("pcm000002", "can't get payment channel setting for merchant")
	errorPaymentChannelMerchantSetFailed        = newBillingServerErrorMsg("pcm000003", "can't set payment channel setting for merchant")
	errorPaymentChannelMerchantDelete           = newBillingServerErrorMsg("pcm000004", "can't delete payment channel setting for merchant")
	errorPaymentChannelMerchantCurrency         = newBillingServerErrorMsg("pcm000005", "currency not supported")
	errorPaymentChannelMerchantCostAlreadyExist = newBillingServerErrorMsg("pcm000006", "cost with specified parameters already exist")
	errorCostMatchedToAmountNotFound            = newBillingServerErrorMsg("pcm000007", "cost matched to amount not found")
	errorPaymentChannelMccCode                  = newBillingServerErrorMsg("pcm000008", "mcc code not supported")
	errorCostMatchedNotFound                    = newBillingServerErrorMsg("pcm000009", "cost matched not found")
	errorMerchantTariffUpdate                   = newBillingServerErrorMsg("pcm000010", "can't update merchant tariffs")
)

func (s *Service) GetAllPaymentChannelCostMerchant(
	ctx context.Context,
	req *billingpb.PaymentChannelCostMerchantListRequest,
	res *billingpb.PaymentChannelCostMerchantListResponse,
) error {
	val, err := s.paymentChannelCostMerchantRepository.GetAllForMerchant(ctx, req.MerchantId)
	if err != nil {
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorPaymentChannelMerchantGetAll
		return nil
	}

	res.Status = billingpb.ResponseStatusOk
	res.Item = &billingpb.PaymentChannelCostMerchantList{Items: val}

	return nil
}

func (s *Service) GetPaymentChannelCostMerchant(
	ctx context.Context,
	req *billingpb.PaymentChannelCostMerchantRequest,
	res *billingpb.PaymentChannelCostMerchantResponse,
) error {
	val, err := s.getPaymentChannelCostMerchant(ctx, req)
	if err != nil {
		res.Status = billingpb.ResponseStatusNotFound
		res.Message = errorPaymentChannelMerchantGet
		return nil
	}

	res.Status = billingpb.ResponseStatusOk
	res.Item = val

	return nil
}

func (s *Service) SetPaymentChannelCostMerchant(
	ctx context.Context,
	req *billingpb.PaymentChannelCostMerchant,
	res *billingpb.PaymentChannelCostMerchantResponse,
) error {
	var err error

	merchant := &billingpb.Merchant{}
	if merchant, err = s.merchantRepository.GetById(ctx, req.MerchantId); err != nil {
		res.Status = billingpb.ResponseStatusNotFound
		res.Message = merchantErrorNotFound
		return nil
	}

	if req.Country != "" {
		country, err := s.country.GetByIsoCodeA2(ctx, req.Country)
		if err != nil {
			res.Status = billingpb.ResponseStatusNotFound
			res.Message = errorCountryNotFound
			return nil
		}
		req.Region = country.PayerTariffRegion
	} else {
		exists := s.country.IsTariffRegionSupported(req.Region)
		if !exists {
			res.Status = billingpb.ResponseStatusNotFound
			res.Message = errorCountryRegionNotExists
			return nil
		}
	}

	sCurr, err := s.curService.GetSettlementCurrencies(ctx, &currenciespb.EmptyRequest{})
	if err != nil {
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorPaymentChannelMerchantCurrency
		return nil
	}
	if !helper.Contains(sCurr.Currencies, req.PayoutCurrency) {
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorPaymentChannelMerchantCurrency
		return nil
	}
	if !helper.Contains(sCurr.Currencies, req.PsFixedFeeCurrency) {
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorPaymentChannelMerchantCurrency
		return nil
	}
	if !helper.Contains(sCurr.Currencies, req.MethodFixAmountCurrency) {
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorPaymentChannelMerchantCurrency
		return nil
	}
	if !helper.Contains(pkg.SupportedMccCodes, req.MccCode) {
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorPaymentChannelMccCode
		return nil
	}

	req.IsActive = true

	if req.Id != "" {
		val, err := s.paymentChannelCostMerchantRepository.GetById(ctx, req.Id)
		if err != nil {
			res.Status = billingpb.ResponseStatusSystemError
			res.Message = errorPaymentChannelMerchantSetFailed
			return nil
		}
		req.Id = val.Id
		req.MerchantId = val.MerchantId
		req.CreatedAt = val.CreatedAt
		err = s.paymentChannelCostMerchantRepository.Update(ctx, req)
	} else {
		req.Id = primitive.NewObjectID().Hex()
		err = s.paymentChannelCostMerchantRepository.Insert(ctx, req)
	}
	if err != nil {
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorPaymentChannelMerchantSetFailed

		if mongodb.IsDuplicate(err) {
			res.Status = billingpb.ResponseStatusBadData
			res.Message = errorPaymentChannelMerchantCostAlreadyExist
		}

		return nil
	}

	if err := s.merchantRepository.UpdateTariffs(ctx, merchant.Id, req); err != nil {
		zap.L().Error(pkg.MethodFinishedWithError, zap.Error(err))
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorMerchantTariffUpdate
		return nil
	}

	res.Status = billingpb.ResponseStatusOk
	res.Item = req

	return nil
}

func (s *Service) DeletePaymentChannelCostMerchant(
	ctx context.Context,
	req *billingpb.PaymentCostDeleteRequest,
	res *billingpb.ResponseError,
) error {
	pc, err := s.paymentChannelCostMerchantRepository.GetById(ctx, req.Id)
	if err != nil {
		res.Status = billingpb.ResponseStatusNotFound
		res.Message = errorCostRateNotFound
		return nil
	}
	err = s.paymentChannelCostMerchantRepository.Delete(ctx, pc)
	if err != nil {
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorPaymentChannelMerchantDelete
		return nil
	}

	res.Status = billingpb.ResponseStatusOk
	return nil
}

func (s *Service) getPaymentChannelCostMerchant(
	ctx context.Context,
	req *billingpb.PaymentChannelCostMerchantRequest,
) (*billingpb.PaymentChannelCostMerchant, error) {
	val, err := s.paymentChannelCostMerchantRepository.Find(
		ctx,
		req.MerchantId,
		req.Name,
		req.PayoutCurrency,
		req.Region,
		req.Country,
		req.MccCode,
	)

	if err != nil {
		return nil, err
	}

	if val == nil {
		return nil, errorCostMatchedNotFound
	}

	var matched []*kvIntFloat
	for _, set := range val {
		for k, i := range set.Set {
			if req.Amount >= i.MinAmount {
				matched = append(matched, &kvIntFloat{k, i.MinAmount})
			}
		}
		if len(matched) == 0 {
			continue
		}

		sort.Slice(matched, func(i, j int) bool {
			return matched[i].Value > matched[j].Value
		})
		return set.Set[matched[0].Key], nil
	}

	return nil, errorCostMatchedToAmountNotFound
}

package service

import (
	"context"
	"github.com/paysuper/paysuper-billing-server/internal/helper"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/errors"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/currenciespb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
)

const (
	collectionPaymentChannelCostSystem = "payment_channel_cost_system"
)

var (
	errorPaymentChannelSystemGetAll                    = errors.NewBillingServerErrorMsg("pcs000001", "can't get list of payment channel setting for system")
	errorPaymentChannelSystemGet                       = errors.NewBillingServerErrorMsg("pcs000002", "can't get payment channel setting for system")
	errorPaymentChannelSystemSetFailed                 = errors.NewBillingServerErrorMsg("pcs000003", "can't set payment channel setting for system")
	errorPaymentChannelSystemDelete                    = errors.NewBillingServerErrorMsg("pcs000004", "can't delete payment channel setting for system")
	errorPaymentChannelSystemCurrency                  = errors.NewBillingServerErrorMsg("pcs000005", "currency not supported")
	errorPaymentChannelSystemCostAlreadyExist          = errors.NewBillingServerErrorMsg("pcs000006", "cost with specified parameters already exist")
	errorPaymentChannelSystemMccCode                   = errors.NewBillingServerErrorMsg("pcs000007", "mcc code not supported")
	errorPaymentChannelSystemOperatingCompanyNotExists = errors.NewBillingServerErrorMsg("pcs000008", "operating company not exists")
)

func (s *Service) GetAllPaymentChannelCostSystem(
	ctx context.Context,
	req *billingpb.EmptyRequest,
	res *billingpb.PaymentChannelCostSystemListResponse,
) error {
	val, err := s.paymentChannelCostSystemRepository.GetAll(ctx)
	if err != nil {
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorPaymentChannelSystemGetAll
		return nil
	}

	res.Status = billingpb.ResponseStatusOk
	res.Item = &billingpb.PaymentChannelCostSystemList{Items: val}

	return nil
}

func (s *Service) GetPaymentChannelCostSystem(
	ctx context.Context,
	req *billingpb.PaymentChannelCostSystemRequest,
	res *billingpb.PaymentChannelCostSystemResponse,
) error {
	val, err := s.paymentChannelCostSystemRepository.Find(ctx, req.Name, req.Region, req.Country, req.MccCode, req.OperatingCompanyId)

	if err != nil {
		res.Status = billingpb.ResponseStatusNotFound
		res.Message = errorPaymentChannelSystemGet
		return nil
	}

	res.Status = billingpb.ResponseStatusOk
	res.Item = val

	return nil
}

func (s *Service) SetPaymentChannelCostSystem(
	ctx context.Context,
	req *billingpb.PaymentChannelCostSystem,
	res *billingpb.PaymentChannelCostSystemResponse,
) error {
	val, err := s.paymentChannelCostSystemRepository.Find(
		ctx,
		req.Name,
		req.Region,
		req.Country,
		req.MccCode,
		req.OperatingCompanyId,
	)

	if err != nil && err != mongo.ErrNoDocuments {
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorPaymentChannelSystemSetFailed
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

	req.IsActive = true

	sCurr, err := s.curService.GetSettlementCurrencies(ctx, &currenciespb.EmptyRequest{})
	if err != nil {
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorPaymentChannelSystemCurrency
		return nil
	}
	if !helper.Contains(sCurr.Currencies, req.FixAmountCurrency) {
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorPaymentChannelSystemCurrency
		return nil
	}
	if !helper.Contains(pkg.SupportedMccCodes, req.MccCode) {
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorPaymentChannelSystemMccCode
		return nil
	}
	if !s.operatingCompanyRepository.Exists(ctx, req.OperatingCompanyId) {
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorPaymentChannelSystemOperatingCompanyNotExists
		return nil
	}

	if val == nil {
		req.Id = primitive.NewObjectID().Hex()
		err = s.paymentChannelCostSystemRepository.Insert(ctx, req)
	} else {
		req.Id = val.Id
		req.CreatedAt = val.CreatedAt
		err = s.paymentChannelCostSystemRepository.Update(ctx, req)
	}
	if err != nil {
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorPaymentChannelSystemSetFailed

		if mongodb.IsDuplicate(err) {
			res.Status = billingpb.ResponseStatusBadData
			res.Message = errorPaymentChannelSystemCostAlreadyExist
		}

		return nil
	}

	res.Status = billingpb.ResponseStatusOk
	res.Item = req

	return nil
}

func (s *Service) DeletePaymentChannelCostSystem(
	ctx context.Context,
	req *billingpb.PaymentCostDeleteRequest,
	res *billingpb.ResponseError,
) error {
	pc, err := s.paymentChannelCostSystemRepository.GetById(ctx, req.Id)
	if err != nil {
		res.Status = billingpb.ResponseStatusNotFound
		res.Message = errorCostRateNotFound
		return nil
	}
	err = s.paymentChannelCostSystemRepository.Delete(ctx, pc)
	if err != nil {
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorPaymentChannelSystemDelete
		return nil
	}

	res.Status = billingpb.ResponseStatusOk
	return nil
}

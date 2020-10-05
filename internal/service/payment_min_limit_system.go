package service

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-billing-server/internal/helper"
	"github.com/paysuper/paysuper-billing-server/pkg/errors"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	errorPaymentMinLimitSystemCurrencyUnknown = errors.NewBillingServerErrorMsg("pmls0001", "payment min limit system currency unknown")
	errorPaymentMinLimitSystemNotFound        = errors.NewBillingServerErrorMsg("pmls0002", "payment min limit system not found")
	errorPaymentMinLimitSystemInvalidAmount   = errors.NewBillingServerErrorMsg("pmls0003", "payment min limit system amount invalid")
)

func (s *Service) GetPaymentMinLimitsSystem(
	ctx context.Context,
	req *billingpb.EmptyRequest,
	res *billingpb.GetPaymentMinLimitsSystemResponse,
) (err error) {
	res.Items, err = s.paymentMinLimitSystemRepository.GetAll(ctx)
	if err != nil {
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			res.Status = billingpb.ResponseStatusBadData
			res.Message = e
			return nil
		}
		return
	}

	res.Status = billingpb.ResponseStatusOk
	return
}

func (s *Service) SetPaymentMinLimitSystem(
	ctx context.Context,
	req *billingpb.PaymentMinLimitSystem,
	res *billingpb.EmptyResponseWithStatus,
) (err error) {
	pmls, err := s.paymentMinLimitSystemRepository.GetByCurrency(ctx, req.Currency)

	if err != nil && err != mongo.ErrNoDocuments {
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			res.Status = billingpb.ResponseStatusBadData
			res.Message = e
			return nil
		}
		return
	}

	if err == mongo.ErrNoDocuments || pmls == nil {
		if !helper.Contains(s.supportedCurrencies, req.Currency) {
			res.Status = billingpb.ResponseStatusBadData
			res.Message = errorPaymentMinLimitSystemCurrencyUnknown
			return nil
		}

		pmls = &billingpb.PaymentMinLimitSystem{
			Id:        primitive.NewObjectID().Hex(),
			Currency:  req.Currency,
			CreatedAt: ptypes.TimestampNow(),
		}
	}

	if req.Amount < 0 {
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorPaymentMinLimitSystemInvalidAmount
		return nil
	}

	pmls.Amount = req.Amount
	pmls.UpdatedAt = ptypes.TimestampNow()

	err = s.paymentMinLimitSystemRepository.Upsert(ctx, pmls)
	if err != nil {
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			res.Status = billingpb.ResponseStatusBadData
			res.Message = e
			return nil
		}
		return
	}

	res.Status = billingpb.ResponseStatusOk
	return
}

package service

import (
	"context"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/errors"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/recurringpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var (
	recurringErrorIncorrectCookie = errors.NewBillingServerErrorMsg("re000001", "customer cookie value is incorrect")
	recurringCustomerNotFound     = errors.NewBillingServerErrorMsg("re000002", "customer not found")
	recurringErrorUnknown         = errors.NewBillingServerErrorMsg("re000003", "unknown error")
	recurringSavedCardNotFount    = errors.NewBillingServerErrorMsg("re000005", "saved card for customer not found")
)

func (s *Service) DeleteSavedCard(
	ctx context.Context,
	req *billingpb.DeleteSavedCardRequest,
	rsp *billingpb.EmptyResponseWithStatus,
) error {
	customer, err := s.decryptBrowserCookie(req.Cookie)

	if err != nil {
		rsp.Status = billingpb.ResponseStatusBadData
		rsp.Message = recurringErrorIncorrectCookie
		return nil
	}

	if customer.CustomerId == "" && customer.VirtualCustomerId == "" {
		rsp.Status = billingpb.ResponseStatusNotFound
		rsp.Message = recurringCustomerNotFound
		return nil
	}

	if customer.CustomerId != "" {
		_, err = s.getCustomerById(ctx, customer.CustomerId)

		if err != nil {
			rsp.Status = billingpb.ResponseStatusNotFound
			rsp.Message = recurringCustomerNotFound
			return nil
		}
	}

	req1 := &recurringpb.DeleteSavedCardRequest{
		Id:    req.Id,
		Token: customer.CustomerId,
	}

	if req1.Token == "" {
		req1.Token = customer.VirtualCustomerId
	}

	rsp1, err := s.rep.DeleteSavedCard(ctx, req1)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(errorFieldService, recurringpb.PayOneRepositoryServiceName),
			zap.String(errorFieldMethod, "DeleteSavedCard"),
			zap.Any(errorFieldRequest, req),
		)

		rsp.Status = billingpb.ResponseStatusSystemError
		rsp.Message = recurringErrorUnknown
		return nil
	}

	if rsp1.Status != billingpb.ResponseStatusOk {
		rsp.Status = rsp1.Status

		if rsp.Status == billingpb.ResponseStatusSystemError {
			zap.L().Error(
				pkg.ErrorGrpcServiceCallFailed,
				zap.String(errorFieldService, recurringpb.PayOneRepositoryServiceName),
				zap.String(errorFieldMethod, "DeleteSavedCard"),
				zap.Any(errorFieldRequest, req),
				zap.Any(pkg.LogFieldResponse, rsp1),
			)

			rsp.Message = recurringErrorUnknown
		} else {
			rsp.Message = recurringSavedCardNotFount
		}

		return nil
	}

	rsp.Status = billingpb.ResponseStatusOk

	return nil
}

func (s *Service) GetSubscriptionOrders(ctx context.Context, req *billingpb.GetSubscriptionOrdersRequest, rsp *billingpb.GetSubscriptionOrdersResponse) error {
	customerId := req.CustomerId

	if len(req.Cookie) > 0 {
		browserCookie, err := s.decryptBrowserCookie(req.Cookie)
		if err != nil {
			zap.L().Error(
				"can't decrypt cookie",
				zap.Error(err),
				zap.Any(errorFieldRequest, req),
			)

			rsp.Status = billingpb.ResponseStatusForbidden
			rsp.Message = recurringCustomerNotFound
			return nil
		}

		if len(browserCookie.CustomerId) == 0 {
			zap.L().Error("customer_id is empty", zap.Any("browserCookie", browserCookie), zap.Any("req", req))
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = recurringCustomerNotFound
			return nil
		}

		customerId = browserCookie.CustomerId
	}

	req1 := &recurringpb.GetSubscriptionRequest{Id: req.SubscriptionId}
	rsp1, err := s.rep.GetSubscription(ctx, req1)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(errorFieldService, recurringpb.PayOneRepositoryServiceName),
			zap.String(errorFieldMethod, "GetSubscription"),
			zap.Any(errorFieldRequest, req),
		)

		rsp.Status = billingpb.ResponseStatusSystemError
		rsp.Message = recurringErrorUnknown
		return nil
	}

	if rsp1.Status != billingpb.ResponseStatusOk {
		rsp.Status = rsp1.Status

		if rsp.Status == billingpb.ResponseStatusSystemError {
			zap.L().Error(
				pkg.ErrorGrpcServiceCallFailed,
				zap.String(errorFieldService, recurringpb.PayOneRepositoryServiceName),
				zap.String(errorFieldMethod, "GetSubscription"),
				zap.Any(errorFieldRequest, req),
				zap.Any(pkg.LogFieldResponse, rsp1),
			)

			rsp.Message = recurringErrorUnknown
		} else {
			rsp.Message = recurringCustomerNotFound
		}

		return nil
	}

	if rsp1.Subscription.CustomerId != customerId && rsp1.Subscription.CustomerUuid != customerId {
		zap.L().Error("trying to get subscription without rights", zap.String("customer_id", customerId), zap.Any("subscription", rsp1.Subscription))
		rsp.Status = billingpb.ResponseStatusForbidden
		rsp.Message = recurringCustomerNotFound

		return nil
	}

	query := bson.M{
		"recurring_id": req.SubscriptionId,
	}
	opts := options.Find().SetSort(bson.M{"pm_order_close_date": -1})

	if req.Offset > 0 {
		opts = opts.SetSkip(int64(req.Offset))
	}

	if req.Limit > 0 {
		opts = opts.SetLimit(int64(req.Limit))
	}

	orders, err := s.orderViewRepository.GetManyBy(ctx, query, opts)
	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)

		rsp.Status = billingpb.ResponseStatusSystemError
		rsp.Message = recurringErrorUnknown
		return nil
	}

	items := make([]*billingpb.ShortOrder, len(orders))
	for i, order := range orders {
		items[i] = &billingpb.ShortOrder{
			Id:         order.Uuid,
			Amount:     float32(order.OrderCharge.AmountRounded),
			Currency:   order.OrderCharge.Currency,
			Date:       order.TransactionDate,
			CardNumber: order.GetCardNumber(),
		}
	}

	rsp.List = items
	rsp.Message = nil
	rsp.Status = billingpb.ResponseStatusOk
	return nil
}

func (s *Service) GetMerchantSubscriptions(ctx context.Context, req *billingpb.GetMerchantSubscriptionsRequest, rsp *billingpb.GetMerchantSubscriptionsResponse) error {
	req1 := &recurringpb.GetMerchantSubscriptionsRequest{
		MerchantId:  req.MerchantId,
		QuickFilter: req.QuickFilter,
		Limit:       req.Limit,
		Offset:      req.Offset,
	}
	res, err := s.rep.GetMerchantSubscriptions(ctx, req1)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(errorFieldService, recurringpb.PayOneRepositoryServiceName),
			zap.String(errorFieldMethod, "GetSubscription"),
			zap.Any(errorFieldRequest, req),
		)

		rsp.Status = billingpb.ResponseStatusSystemError
		rsp.Message = recurringErrorUnknown
		return nil
	}

	if res.Status != billingpb.ResponseStatusOk {
		rsp.Status = res.Status

		if rsp.Status == billingpb.ResponseStatusSystemError {
			zap.L().Error(
				pkg.ErrorGrpcServiceCallFailed,
				zap.String(errorFieldService, recurringpb.PayOneRepositoryServiceName),
				zap.String(errorFieldMethod, "GetMerchantSubscriptions"),
				zap.Any(errorFieldRequest, req),
				zap.Any(pkg.LogFieldResponse, res),
			)

			rsp.Message = recurringErrorUnknown
		} else {
			rsp.Message = recurringCustomerNotFound
		}

		return nil
	}

	items := make([]*billingpb.MerchantSubscription, len(res.List))

	projectNames := map[string]string{}

	for i, subscription := range res.List {
		items[i] = &billingpb.MerchantSubscription{
			Start: subscription.Start,
			Amount: subscription.Amount,
			Currency: subscription.Currency,
			Email: subscription.Email,
			SubscriptionId: subscription.SubscriptionId,
			ProjectId: subscription.ProjectId,
		}

		name, ok := projectNames[subscription.ProjectId]
		if !ok {
			project, err := s.project.GetById(ctx, subscription.ProjectId)
			if err != nil {
				zap.L().Error(
					pkg.ErrorDatabaseQueryFailed,
					zap.Error(err),
				)

				rsp.Status = billingpb.ResponseStatusSystemError
				rsp.Message = recurringErrorUnknown
				return nil
			}
			name = project.Name[DefaultLanguage]
			projectNames[subscription.ProjectId] = name
		}

		items[i].ProjectName = name
	}

	rsp.List = items
	rsp.Status = billingpb.ResponseStatusOk
	return nil
}

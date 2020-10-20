package service

import (
	"context"
	"fmt"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/errors"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/recurringpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var (
	recurringErrorIncorrectCookie    = errors.NewBillingServerErrorMsg("re000001", "customer cookie value is incorrect")
	recurringCustomerNotFound        = errors.NewBillingServerErrorMsg("re000002", "customer not found")
	recurringErrorUnknown            = errors.NewBillingServerErrorMsg("re000003", "unknown error")
	recurringSavedCardNotFount       = errors.NewBillingServerErrorMsg("re000005", "saved card for customer not found")
	recurringErrorDeleteSubscription = errors.NewBillingServerErrorMsg("re000006", "unable to delete subscription")
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

	if rsp1.Subscription.CustomerId != customerId && rsp1.Subscription.CustomerUuid != customerId ||
		(len(req.MerchantId) > 0 && rsp1.Subscription.MerchantId != req.MerchantId) {
		zap.L().Error("trying to get subscription without rights",
			zap.String("customer_id", customerId),
			zap.Any("subscription", rsp1.Subscription),
			zap.String("merchant_id", req.MerchantId))

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

	count, err := s.orderViewRepository.GetCountBy(ctx, query)

	items := make([]*billingpb.ShortOrder, len(orders))
	for i, order := range orders {
		name := ""
		if order.Project != nil {
			name, _ = order.Project.Name[DefaultLanguage]
		}

		items[i] = &billingpb.ShortOrder{
			Id:          order.Uuid,
			Amount:      float32(order.OrderCharge.AmountRounded),
			Currency:    order.OrderCharge.Currency,
			Date:        order.TransactionDate,
			CardNumber:  order.GetCardNumber(),
			ProductName: name,
		}
	}

	rsp.List = items
	rsp.Message = nil
	rsp.Status = billingpb.ResponseStatusOk
	rsp.Count = int32(count)

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
			Start:          subscription.Start,
			Amount:         subscription.Amount,
			Currency:       subscription.Currency,
			Email:          subscription.Email,
			SubscriptionId: subscription.SubscriptionId,
			ProjectId:      subscription.ProjectId,
			MaskedPan:      subscription.MaskedPan,
			CustomerId:     subscription.CustomerId,
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

	rsp.Count = res.Count
	rsp.List = items
	rsp.Status = billingpb.ResponseStatusOk
	return nil
}

func (s *Service) DeleteRecurringSubscription(
	ctx context.Context,
	req *billingpb.DeleteRecurringSubscriptionRequest,
	res *billingpb.EmptyResponseWithStatus,
) error {
	customerId := req.CustomerId

	if len(req.Cookie) > 0 {
		browserCookie, err := s.decryptBrowserCookie(req.Cookie)
		if err != nil {
			zap.L().Error(
				"can't decrypt cookie",
				zap.Error(err),
				zap.Any(errorFieldRequest, req),
			)

			res.Status = billingpb.ResponseStatusForbidden
			res.Message = recurringCustomerNotFound
			return nil
		}

		if len(browserCookie.CustomerId) == 0 {
			zap.L().Error(
				"customer_id is empty",
				zap.Any("browserCookie", browserCookie),
				zap.Any("req", req),
			)

			res.Status = billingpb.ResponseStatusBadData
			res.Message = recurringCustomerNotFound
			return nil
		}

		customerId = browserCookie.CustomerId
	}

	rsp, err := s.rep.GetSubscription(ctx, &recurringpb.GetSubscriptionRequest{
		Id: req.SubscriptionId,
	})

	if err != nil || rsp.Status != billingpb.ResponseStatusOk {
		if err == nil {
			err = fmt.Errorf(rsp.Message)
		}

		zap.L().Error(
			"Unable to get subscription",
			zap.Error(err),
			zap.Any("subscription_id", req.SubscriptionId),
		)

		res.Status = billingpb.ResponseStatusNotFound
		res.Message = orderErrorRecurringSubscriptionNotFound
		return nil
	}

	subscription := rsp.Subscription

	if customerId == "" || (subscription.CustomerId != customerId && subscription.CustomerUuid != customerId) {
		zap.L().Error(
			"trying to delete subscription without rights",
			zap.String("customer_id", customerId),
			zap.Any("subscription", subscription),
		)
		res.Status = billingpb.ResponseStatusForbidden
		res.Message = recurringCustomerNotFound

		return nil
	}

	if subscription.OrderId == primitive.NilObjectID.Hex() {
		zap.L().Error(
			"Order ID for subscription is empty. Skipping.",
			zap.Error(err),
			zap.Any("subscription", subscription),
		)
	} else {
		order, err := s.orderRepository.GetById(ctx, subscription.OrderId)

		if err != nil {
			res.Status = billingpb.ResponseStatusNotFound
			res.Message = orderErrorNotFound
			return nil
		}

		ps, err := s.paymentSystemRepository.GetById(ctx, order.PaymentMethod.PaymentSystemId)

		if err != nil {
			res.Status = billingpb.ResponseStatusNotFound
			res.Message = orderErrorPaymentSystemInactive
			return nil
		}

		h, err := s.paymentSystemGateway.GetGateway(ps.Handler)

		if err != nil {
			zap.L().Error(
				"Unable to get payment system gateway",
				zap.Error(err),
				zap.Any("subscription", subscription),
				zap.Any("payment_system", ps),
			)

			res.Status = billingpb.ResponseStatusSystemError
			res.Message = orderErrorPaymentSystemInactive
			return nil
		}

		err = h.DeleteRecurringSubscription(order, subscription)

		if err != nil {
			zap.L().Error(
				"Unable to delete subscription on payment system",
				zap.Error(err),
				zap.Any("subscription", subscription),
				zap.Any("payment_system", ps),
			)

			res.Status = billingpb.ResponseStatusSystemError
			res.Message = recurringErrorDeleteSubscription
			return nil
		}
	}

	resDelete, err := s.rep.DeleteSubscription(ctx, subscription)

	if err != nil || resDelete.Status != billingpb.ResponseStatusOk {
		if err == nil {
			err = fmt.Errorf(resDelete.Message)
		}

		zap.L().Error(
			"Unable to delete subscription on recurring service",
			zap.Error(err),
			zap.Any("subscription", subscription),
		)

		res.Status = billingpb.ResponseStatusSystemError
		res.Message = recurringErrorDeleteSubscription
		return nil
	}

	res.Status = billingpb.ResponseStatusOk

	return nil
}

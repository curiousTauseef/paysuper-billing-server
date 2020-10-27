package service

import (
	"context"
	"fmt"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/errors"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/recurringpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var (
	recurringErrorIncorrectCookie      = errors.NewBillingServerErrorMsg("re000001", "customer cookie value is incorrect")
	recurringCustomerNotFound          = errors.NewBillingServerErrorMsg("re000002", "customer not found")
	recurringErrorUnknown              = errors.NewBillingServerErrorMsg("re000003", "unknown error")
	recurringSavedCardNotFount         = errors.NewBillingServerErrorMsg("re000005", "saved card for customer not found")
	recurringErrorDeleteSubscription   = errors.NewBillingServerErrorMsg("re000006", "unable to delete subscription")
	recurringErrorIdentifierNotFound   = errors.NewBillingServerErrorMsg("re000007", "identifier for subscription is not found")
	recurringErrorSubscriptionNotFound = errors.NewBillingServerErrorMsg("re000008", "subscription not found")
	recurringErrorAccessDeny           = errors.NewBillingServerErrorMsg("re000009", "subscription access denied")
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

func (s *Service) GetSubscriptionOrders(
	ctx context.Context,
	req *billingpb.GetSubscriptionOrdersRequest,
	rsp *billingpb.GetSubscriptionOrdersResponse,
) error {
	var customerId string

	browserCookie, err := s.findAndParseBrowserCookie(req.Cookie)

	if err != nil {
		rsp.Status = billingpb.ResponseStatusForbidden
		rsp.Message = recurringCustomerNotFound
		return nil
	}

	if browserCookie != nil {
		customerId = browserCookie.CustomerId
	}

	req1 := &recurringpb.GetSubscriptionRequest{Id: req.Id}
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
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.String(errorFieldService, recurringpb.PayOneRepositoryServiceName),
			zap.String(errorFieldMethod, "GetSubscription"),
			zap.Any(errorFieldRequest, req),
			zap.Any(pkg.LogFieldResponse, rsp1),
		)

		rsp.Status = rsp1.Status
		rsp.Message = recurringErrorSubscriptionNotFound
		return nil
	}

	if err = s.checkSubscriptionPermission(customerId, req.MerchantId, rsp1.Subscription); err != nil {
		rsp.Status = billingpb.ResponseStatusForbidden
		rsp.Message = recurringErrorAccessDeny
		return nil
	}

	query := bson.M{
		"recurring_id": rsp1.Subscription.Id,
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

	items := make([]*billingpb.SubscriptionOrder, len(orders))
	for i, order := range orders {
		var name []string

		if order.Items != nil {
			for _, item := range order.Items {
				name = append(name, item.Name)
			}
		}

		items[i] = &billingpb.SubscriptionOrder{
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

func (s *Service) GetSubscription(
	ctx context.Context,
	req *billingpb.GetSubscriptionRequest,
	rsp *billingpb.GetSubscriptionResponse,
) error {
	var customerId string

	browserCookie, err := s.findAndParseBrowserCookie(req.Cookie)

	if err != nil {
		rsp.Status = billingpb.ResponseStatusForbidden
		rsp.Message = recurringCustomerNotFound
		return nil
	}

	if browserCookie != nil {
		customerId = browserCookie.CustomerId
	}

	req1 := &recurringpb.GetSubscriptionRequest{Id: req.Id}
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
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.String(errorFieldService, recurringpb.PayOneRepositoryServiceName),
			zap.String(errorFieldMethod, "GetSubscription"),
			zap.Any(errorFieldRequest, req),
			zap.Any(pkg.LogFieldResponse, rsp1),
		)

		rsp.Status = rsp1.Status
		rsp.Message = recurringErrorSubscriptionNotFound
		return nil
	}

	if err = s.checkSubscriptionPermission(customerId, req.MerchantId, rsp1.Subscription); err != nil {
		rsp.Status = billingpb.ResponseStatusForbidden
		rsp.Message = recurringErrorAccessDeny
		return nil
	}

	rsp.Message = nil
	rsp.Status = billingpb.ResponseStatusOk
	rsp.Subscription = s.mapRecurringToBilling(rsp1.Subscription)

	return nil
}

func (s *Service) DeleteRecurringSubscription(
	ctx context.Context,
	req *billingpb.DeleteRecurringSubscriptionRequest,
	res *billingpb.EmptyResponseWithStatus,
) error {
	var customerId string

	browserCookie, err := s.findAndParseBrowserCookie(req.Cookie)

	if err != nil {
		res.Status = billingpb.ResponseStatusForbidden
		res.Message = recurringCustomerNotFound
		return nil
	}

	if browserCookie != nil {
		customerId = browserCookie.CustomerId
	}

	rsp, err := s.rep.GetSubscription(ctx, &recurringpb.GetSubscriptionRequest{
		Id: req.Id,
	})

	if err != nil || rsp.Status != billingpb.ResponseStatusOk {
		if err == nil {
			err = fmt.Errorf(rsp.Message)
		}

		zap.L().Error(
			"Unable to get subscription",
			zap.Error(err),
			zap.Any("subscription_id", req.Id),
		)

		res.Status = billingpb.ResponseStatusNotFound
		res.Message = orderErrorRecurringSubscriptionNotFound
		return nil
	}

	subscription := rsp.Subscription

	if err = s.checkSubscriptionPermission(customerId, req.MerchantId, subscription); err != nil {
		res.Status = billingpb.ResponseStatusForbidden
		res.Message = recurringErrorAccessDeny
		return nil
	}

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

func (s *Service) FindSubscriptions(ctx context.Context, req *billingpb.FindSubscriptionsRequest, rsp *billingpb.FindSubscriptionsResponse) error {
	var customerId string

	browserCookie, err := s.findAndParseBrowserCookie(req.Cookie)

	if err != nil {
		rsp.Status = billingpb.ResponseStatusForbidden
		rsp.Message = recurringCustomerNotFound
		return nil
	}

	if browserCookie != nil {
		customerId = browserCookie.CustomerId
	}

	if customerId == "" && req.MerchantId == "" {
		zap.L().Error(
			"unable to identify performer for find subscriptions",
			zap.String("cookie", req.Cookie),
			zap.String("merchant_id", req.MerchantId),
		)
		rsp.Status = billingpb.ResponseStatusForbidden
		rsp.Message = recurringErrorAccessDeny
		return nil
	}

	reqFind := &recurringpb.FindSubscriptionsRequest{
		MerchantId:  req.MerchantId,
		CustomerId:  customerId,
		QuickFilter: req.QuickFilter,
		Offset:      req.Offset,
		Limit:       req.Limit,
	}
	rsp1, err := s.rep.FindSubscriptions(ctx, reqFind)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(errorFieldService, recurringpb.PayOneRepositoryServiceName),
			zap.String(errorFieldMethod, "FindSubscriptions"),
			zap.Any(errorFieldRequest, reqFind),
		)

		rsp.Status = billingpb.ResponseStatusSystemError
		rsp.Message = recurringErrorUnknown
		return nil
	}

	rsp.Count = rsp1.Count
	rsp.List = make([]*billingpb.RecurringSubscription, len(rsp1.List))

	for i, subscription := range rsp1.List {
		rsp.List[i] = s.mapRecurringToBilling(subscription)
	}

	rsp.Status = billingpb.ResponseStatusOk

	return nil
}

func (s *Service) mapRecurringToBilling(sub *recurringpb.Subscription) *billingpb.RecurringSubscription {
	rSub := &billingpb.RecurringSubscription{
		Id:            sub.Id,
		CustomerId:    sub.CustomerId,
		CustomerEmail: sub.CustomerInfo.Email,
		Period:        sub.Period,
		MerchantId:    sub.MerchantId,
		ProjectId:     sub.ProjectId,
		Amount:        sub.Amount,
		TotalAmount:   sub.TotalAmount,
		Currency:      sub.Currency,
		IsActive:      sub.IsActive,
		MaskedPan:     sub.MaskedPan,
		ExpireAt:      sub.ExpireAt,
		CreatedAt:     sub.CreatedAt,
		LastPaymentAt: sub.LastPaymentAt,
		ProjectName:   sub.ProjectName,
	}

	if sub.CustomerInfo != nil {
		rSub.CustomerEmail = sub.CustomerInfo.Email
	}

	return rSub
}

func (s *Service) findAndParseBrowserCookie(cookie string) (*BrowserCookieCustomer, error) {
	if cookie != "" {
		browserCookie, err := s.decryptBrowserCookie(cookie)
		if err != nil {
			zap.L().Error(
				"can't decrypt cookie",
				zap.Error(err),
				zap.Any(errorFieldRequest, cookie),
			)

			return nil, err
		}

		if len(browserCookie.CustomerId) == 0 {
			zap.L().Error(
				"customer_id is empty",
				zap.Any("browserCookie", browserCookie),
				zap.Any("cookie", cookie),
			)

			return nil, err
		}

		return browserCookie, nil
	}

	return nil, nil
}

func (s *Service) checkSubscriptionPermission(customerId, merchantId string, subscription *recurringpb.Subscription) error {
	if customerId == "" && merchantId == "" {
		zap.L().Error(
			"unable to identify performer for delete subscription",
			zap.String("customerId", customerId),
			zap.String("merchant_id", merchantId),
			zap.Any("subscription", subscription),
		)

		return recurringErrorAccessDeny
	}

	if customerId != "" && customerId != subscription.CustomerId {
		zap.L().Error(
			"trying to get subscription for another customer",
			zap.String("customer_id", customerId),
			zap.Any("subscription", subscription),
		)

		return recurringErrorAccessDeny
	}

	if merchantId != "" && merchantId != subscription.MerchantId {
		zap.L().Error(
			"trying to get subscription for another customer",
			zap.String("merchant_id", merchantId),
			zap.Any("subscription", subscription),
		)

		return recurringErrorAccessDeny
	}

	return nil
}

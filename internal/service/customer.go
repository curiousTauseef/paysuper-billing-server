package service

import (
	"context"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/errors"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/recurringpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"math"
	"net"
)

var (
	errorCustomerOrderTypeNotSupport = errors.NewBillingServerErrorMsg("cm000001", "order type is not support")
	errorCustomerUnknown             = errors.NewBillingServerErrorMsg("cm000002", "unknown error")
)

func (s *Service) SetCustomerPaymentActivity(
	ctx context.Context,
	req *billingpb.SetCustomerPaymentActivityRequest,
	rsp *billingpb.EmptyResponseWithStatus,
) error {
	if req.Type != billingpb.OrderTypeOrder && req.Type != billingpb.OrderTypeRefund {
		rsp.Status = billingpb.ResponseStatusBadData
		rsp.Message = errorCustomerOrderTypeNotSupport
		return nil
	}

	customer, err := s.customerRepository.GetById(ctx, req.CustomerId)

	if err != nil {
		rsp.Status = billingpb.ResponseStatusNotFound
		rsp.Message = customerNotFound
		return nil
	}

	s.mx.Lock()
	defer s.mx.Unlock()
	paymentActivity, ok := customer.PaymentActivity[req.MerchantId]

	if !ok {
		paymentActivity = &billingpb.PaymentActivityItem{
			Count: &billingpb.PaymentActivityItemCount{},
			LastTxnAt: &billingpb.PaymentActivityItemLastTxnAt{
				Payment: &timestamppb.Timestamp{},
				Refund:  &timestamppb.Timestamp{},
			},
			Revenue: &billingpb.PaymentActivityItemRevenue{},
		}
	}

	if req.Type == billingpb.OrderTypeOrder {
		paymentActivity.Count.Payment++
		paymentActivity.LastTxnAt.Payment = req.ProcessingAt
		paymentActivity.Revenue.Payment = req.Amount
		paymentActivity.Revenue.Total += req.Amount
	} else {
		paymentActivity.Count.Refund++
		paymentActivity.LastTxnAt.Refund = req.ProcessingAt
		paymentActivity.Revenue.Refund = req.Amount
		paymentActivity.Revenue.Total -= req.Amount
	}

	paymentActivity.Revenue.Payment = math.Round(paymentActivity.Revenue.Payment*100) / 100
	paymentActivity.Revenue.Refund = math.Round(paymentActivity.Revenue.Refund*100) / 100
	paymentActivity.Revenue.Total = math.Round(paymentActivity.Revenue.Total*100) / 100

	customer.PaymentActivity[req.MerchantId] = paymentActivity
	err = s.customerRepository.Update(ctx, customer)

	if err != nil {
		rsp.Status = billingpb.ResponseStatusSystemError
		rsp.Message = errorCustomerUnknown
		return nil
	}

	rsp.Status = billingpb.ResponseStatusOk
	return nil
}

func (s *Service) MigrateCustomers(ctx context.Context) error {
	customers, err := s.customerRepository.FindAll(ctx)

	if err != nil {
		return err
	}

	for _, customer := range customers {
		if len(customer.Ip) > 0 {
			ip := net.IP(customer.Ip).String()

			if customer.IpString == "" {
				customer.IpString = ip
			}

			if customer.Address == nil {
				if address, err := s.getAddressByIp(ctx, ip); err == nil {
					customer.Address = address
				}
			}

			exists := false

			for key, val := range customer.IpHistory {
				hIp := net.IP(val.Ip).String()

				if val.IpString == "" {
					customer.IpHistory[key].IpString = hIp
				}

				if val.Address == nil {
					if address, err := s.getAddressByIp(ctx, hIp); err == nil {
						customer.IpHistory[key].Address = address
					}
				}

				if hIp == ip {
					exists = true
				}
			}

			if !exists {
				it := &billingpb.CustomerIpHistory{
					Ip:        customer.Ip,
					CreatedAt: customer.CreatedAt,
					IpString:  ip,
					Address:   customer.Address,
				}
				customer.IpHistory = append(customer.IpHistory, it)
			}
		}

		if customer.Address != nil && len(customer.AddressHistory) <= 0 {
			it := &billingpb.CustomerAddressHistory{
				Country:    customer.Address.Country,
				City:       customer.Address.City,
				State:      customer.Address.State,
				PostalCode: customer.Address.PostalCode,
				CreatedAt:  customer.CreatedAt,
			}
			customer.AddressHistory = append(customer.AddressHistory, it)
		}

		if customer.Uuid == "" {
			customer.Uuid = uuid.New().String()
		}

		if err := s.customerRepository.Update(ctx, customer); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) UpdateCustomersPayments(ctx context.Context) error {
	customers, err := s.customerRepository.FindAll(ctx)

	if err != nil {
		zap.L().Error("can't get customers", zap.Error(err))
		return err
	}

	merchants, err := s.merchantRepository.GetAll(ctx)
	if err != nil {
		zap.L().Error("can't get merchants", zap.Error(err))
		return err
	}

	for _, customer := range customers {
		for _, merchant := range merchants {
			oid, err := primitive.ObjectIDFromHex(merchant.Id)
			if err != nil {
				zap.L().Error("can't create objectId from merchant_id", zap.Error(err), zap.String("merchant_id", merchant.Id), zap.String("customer_id", customer.Id))
				continue
			}

			opts := options.Find()
			opts.SetSort(bson.D{{"pm_order_close_date", 1}})
			orders, err := s.orderViewRepository.GetManyBy(ctx, bson.M{"project.merchant_id": oid}, opts)
			if len(orders) > 0 {
				paymentActivity, ok := customer.PaymentActivity[merchant.Id]
				lastOrder := orders[0]

				if !ok {
					paymentActivity = &billingpb.PaymentActivityItem{
						Count: &billingpb.PaymentActivityItemCount{},
						LastTxnAt: &billingpb.PaymentActivityItemLastTxnAt{
							Payment: &timestamppb.Timestamp{},
							Refund:  &timestamppb.Timestamp{},
						},
						Revenue: &billingpb.PaymentActivityItemRevenue{},
					}
					customer.PaymentActivity[merchant.Id] = paymentActivity
				}

				if lastOrder.TransactionDate != nil && lastOrder.TransactionDate.Nanos != 0 {
					paymentActivity.LastTxnAt.Payment = lastOrder.TransactionDate
				}

				for _, order := range orders {
					if order.Refund != nil && order.Refund.Amount > 0 {
						paymentActivity.Count.Refund++
						paymentActivity.Revenue.Refund += order.GrossRevenue.Amount
					} else if order.Cancellation == nil {
						paymentActivity.Count.Payment++
						paymentActivity.Revenue.Payment += order.GrossRevenue.Amount
					}
				}
			}
		}

		err = s.customerRepository.Update(ctx, customer)
		if err != nil {
			zap.L().Error("can't update customer in migration", zap.Error(err), zap.String("customer_id", customer.Id))
		}
	}

	zap.L().Info("migration ended", zap.Int("customers", len(customers)))

	return nil
}

func (s *Service) GetCustomerInfo(ctx context.Context, req *billingpb.GetCustomerInfoRequest, rsp *billingpb.GetCustomerInfoResponse) error {
	rsp.Status = billingpb.ResponseStatusSystemError

	opts := options.Find()
	opts.SetLimit(1)

	customers, err := s.customerRepository.FindBy(ctx, bson.M{
		"uuid": req.UserId,
		"payment_activity": bson.M{
			"$elemMatch": bson.M{
				"merchant_id": req.MerchantId,
			},
		},
	}, opts)

	if len(customers) == 0 {
		rsp.Status = billingpb.ResponseStatusBadData
		zap.L().Error("no customers", zap.Any("req", req))
		return nil
	}

	if err != nil {
		zap.L().Error("can't get customer", zap.Error(err))
		return nil
	}

	customer := customers[0]

	// Protect. Do not send info that not needed
	customer.PaymentActivity = nil
	customer.Identity = nil

	rsp.Item = customer
	rsp.Status = billingpb.ResponseStatusOk

	return nil
}
func (s *Service) GetCustomerList(ctx context.Context, req *billingpb.ListCustomersRequest, rsp *billingpb.ListCustomersResponse) error {
	query := bson.M{}
	rsp.Status = billingpb.ResponseStatusSystemError

	if len(req.MerchantId) > 0 {
		query["payment_activity"] = bson.M{
			"$elemMatch": bson.M{
				"merchant_id": req.MerchantId,
			},
		}
	}

	if len(req.Country) > 0 {
		query["address.country"] = req.Country
	}

	if len(req.Phone) > 0 {
		query["phone"] = req.Phone
	}

	if len(req.Email) > 0 {
		query["email"] = req.Email
	}

	if len(req.ExternalId) > 0 {
		query["external_id"] = req.ExternalId
	}

	if len(req.Language) > 0 {
		query["locale_history"] = bson.M{
			"$elemMatch": bson.M{
				"value": req.Language,
			},
		}
	}

	if len(req.Name) > 0 {
		query["name"] = req.Name
	}

	opts := options.Find()
	opts = opts.SetLimit(req.Limit)

	customers, err := s.customerRepository.FindBy(ctx, query)
	if err != nil {
		zap.L().Error("can't get customers", zap.Error(err), zap.Any("req", req))
		return err
	}

	zap.L().Info("GetCustomerList", zap.Any("req", req), zap.Any("query", query), zap.Int("customers", len(customers)))

	result := make([]*billingpb.ShortCustomerInfo, len(customers))
	for i, customer := range customers {
		shortCustomer := &billingpb.ShortCustomerInfo{
			Id:         customer.Uuid,
			ExternalId: customer.ExternalId,
			Language:   customer.Locale,
			LastOrder:  &timestamp.Timestamp{},
		}

		if customer.Address != nil {
			shortCustomer.Country = customer.Address.Country
		}

		for key, activityItem := range customer.PaymentActivity {
			if (len(req.MerchantId) > 0 && key == req.MerchantId) || len(req.MerchantId) == 0 {
				if activityItem.Count != nil {
					shortCustomer.Orders += activityItem.Count.Payment
				}
				if activityItem.Revenue != nil {
					shortCustomer.Revenue += activityItem.Revenue.Payment
				}
				if activityItem.LastTxnAt != nil && activityItem.LastTxnAt.Payment != nil && shortCustomer.LastOrder.Seconds < activityItem.LastTxnAt.Payment.Seconds {
					shortCustomer.LastOrder = activityItem.LastTxnAt.Payment
				}
			}
		}

		result[i] = shortCustomer
	}

	if len(result) == 0 {
		result = []*billingpb.ShortCustomerInfo{}
	}

	rsp.Items = result
	rsp.Status = billingpb.ResponseStatusOk

	return nil
}

func (s *Service) DeserializeCookie(ctx context.Context, req *billingpb.DeserializeCookieRequest, rsp *billingpb.DeserializeCookieResponse) error {
	browserCookie, err := s.decryptBrowserCookie(req.Cookie)
	rsp.Status = billingpb.ResponseStatusBadData

	if err != nil {
		return nil
	}

	rsp.Item = &billingpb.BrowserCookie{
		CustomerId:        browserCookie.CustomerId,
		VirtualCustomerId: browserCookie.VirtualCustomerId,
		Ip:                browserCookie.Ip,
		IpCountry:         browserCookie.IpCountry,
		SelectedCountry:   browserCookie.SelectedCountry,
		UserAgent:         browserCookie.UserAgent,
		AcceptLanguage:    browserCookie.AcceptLanguage,
		SessionCount:      browserCookie.SessionCount,
	}

	rsp.Status = billingpb.ResponseStatusOk
	return nil
}

func (s *Service) DeleteCustomerCard(ctx context.Context, req *billingpb.DeleteCustomerCardRequest, rsp *billingpb.EmptyResponseWithStatus) error {
	rsp.Status = billingpb.ResponseStatusBadData
	rsp.Message = recurringErrorUnknown

	req1 := &recurringpb.DeleteSavedCardRequest{
		Id:    req.Id,
		Token: req.CustomerId,
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
	rsp.Message = nil

	return nil
}

func (s *Service) GetCustomerSubscription(ctx context.Context, req *billingpb.GetSubscriptionRequest, rsp *billingpb.GetSubscriptionResponse) error {
	rsp.Status = billingpb.ResponseStatusBadData
	rsp.Message = recurringErrorUnknown

	req1 := &recurringpb.GetSubscriptionRequest{
		Id: req.Id,
	}

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
			rsp.Message = recurringSavedCardNotFount
		}

		return nil
	}

	if len(req.Cookie) > 0 {
		browserCookie, err := s.decryptBrowserCookie(req.Cookie)
		if err != nil {
			zap.L().Error(
				"can't decrypt cookie",
				zap.Error(err),
				zap.Any(errorFieldRequest, req),
			)

			rsp.Status = billingpb.ResponseStatusForbidden
			rsp.Message = recurringErrorUnknown
			return nil
		}

		if rsp1.Subscription.CustomerId != browserCookie.CustomerId {
			zap.L().Error("trying to get subscription for another customer", zap.String("customer_id", browserCookie.CustomerId), zap.String("subscription_customer_id", rsp1.Subscription.CustomerId), zap.String("subscription_id", rsp1.Subscription.Id))
			rsp.Status = billingpb.ResponseStatusForbidden
			rsp.Message = recurringSavedCardNotFount
			return nil
		}

	} else {
		if rsp1.Subscription.CustomerUuid != req.CustomerId {
			zap.L().Error("trying to get subscription for another customer", zap.String("customer_id", req.CustomerId), zap.String("subscription_customer_id", rsp1.Subscription.CustomerUuid), zap.String("subscription_id", rsp1.Subscription.Id))
			rsp.Status = billingpb.ResponseStatusForbidden
			rsp.Message = recurringSavedCardNotFount
			return nil
		}
	}

	rsp.Status = billingpb.ResponseStatusOk
	rsp.Subscription = s.mapRecurringToBilling(rsp1.Subscription)

	return nil
}

func (s *Service) mapRecurringToBilling(sub *recurringpb.Subscription) *billingpb.Subscription {
	cust := sub.CustomerInfo

	var info billingpb.CustomerInfo
	if cust != nil {
		info = billingpb.CustomerInfo{
			ExternalId: cust.ExternalId,
			Email:      cust.Email,
			Phone:      cust.Phone,
		}
	}

	return &billingpb.Subscription{
		Id:                    sub.Id,
		CustomerId:            sub.CustomerId,
		CustomerUuid:          sub.CustomerUuid,
		Period:                sub.Period,
		MerchantId:            sub.MerchantId,
		ProjectId:             sub.ProjectId,
		Amount:                sub.Amount,
		Currency:              sub.Currency,
		ItemType:              sub.ItemType,
		ItemList:              sub.ItemList,
		IsActive:              sub.IsActive,
		MaskedPan:             sub.MaskedPan,
		CardpaySubscriptionId: sub.CardpaySubscriptionId,
		CustomerInfo:          &info,
		ExpireAt:              sub.ExpireAt,
		LastPaymentAt:         sub.LastPaymentAt,
		CreatedAt:             sub.CreatedAt,
		UpdatedAt:             sub.UpdatedAt,
	}
}

func (s *Service) mapBillingCustomerToRecurring(customer *billingpb.CustomerInfo) *recurringpb.CustomerInfo {
	if customer == nil {
		return nil
	}

	return &recurringpb.CustomerInfo{
		ExternalId: customer.ExternalId,
		Email:      customer.Email,
		Phone:      customer.Phone,
	}
}

func (s *Service) FindSubscriptions(ctx context.Context, req *billingpb.FindSubscriptionsRequest, rsp *billingpb.FindSubscriptionsResponse) error {
	rsp.Status = billingpb.ResponseStatusBadData
	rsp.Message = recurringErrorUnknown

	req1 := &recurringpb.FindSubscriptionsRequest{
	}

	if len(req.Cookie) > 0 {
		browserCookie, err := s.decryptBrowserCookie(req.Cookie)
		if err != nil {
			zap.L().Error(
				"can't decrypt cookie",
				zap.Error(err),
				zap.Any(errorFieldRequest, req),
			)

			rsp.Status = billingpb.ResponseStatusForbidden
			rsp.Message = recurringErrorUnknown
			return nil
		}
		req1.CustomerId = browserCookie.CustomerId
	} else {
		req1.CustomerUuid = req.CustomerId
	}

	rsp1, err := s.rep.FindSubscriptions(ctx, req1)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(errorFieldService, recurringpb.PayOneRepositoryServiceName),
			zap.String(errorFieldMethod, "FindSubscriptions"),
			zap.Any(errorFieldRequest, req),
		)

		rsp.Status = billingpb.ResponseStatusSystemError
		rsp.Message = recurringErrorUnknown
		return nil
	}

	rsp.Status = billingpb.ResponseStatusOk

	rsp.List = make([]*billingpb.Subscription, len(rsp1.List))
	for i, subscription := range rsp1.List {
		rsp.List[i] = s.mapRecurringToBilling(subscription)
	}

	return nil
}

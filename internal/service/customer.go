package service

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net"
)

var (
	errorCustomerOrderTypeNotSupport = newBillingServerErrorMsg("cm000001", "order type is not support")
	errorCustomerUnknown             = newBillingServerErrorMsg("cm000002", "unknown error")
)

const (
	roundKeyCustomerPaymentActivityRevenuePayment = "customer_payment_activity_revenue_payment"
	roundKeyCustomerPaymentActivityRevenueRefund  = "customer_payment_activity_revenue_refund"
	roundKeyCustomerPaymentActivityRevenueTotal   = "customer_payment_activity_revenue_total"
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
		paymentActivity.Revenue.Payment += req.Amount
		paymentActivity.Revenue.Total += req.Amount
	} else {
		paymentActivity.Count.Refund++
		paymentActivity.LastTxnAt.Refund = req.ProcessingAt
		paymentActivity.Revenue.Refund += req.Amount
		paymentActivity.Revenue.Total -= req.Amount
	}

	paymentActivity.Revenue.Payment, err = s.round(roundKeyCustomerPaymentActivityRevenuePayment, paymentActivity.Revenue.Payment)

	if err != nil {
		rsp.Status = billingpb.ResponseStatusSystemError
		rsp.Message = errorCustomerUnknown
		return nil
	}

	paymentActivity.Revenue.Refund, err = s.round(roundKeyCustomerPaymentActivityRevenueRefund, paymentActivity.Revenue.Refund)

	if err != nil {
		rsp.Status = billingpb.ResponseStatusSystemError
		rsp.Message = errorCustomerUnknown
		return nil
	}

	paymentActivity.Revenue.Total, err = s.round(roundKeyCustomerPaymentActivityRevenueTotal, paymentActivity.Revenue.Total)

	if err != nil {
		rsp.Status = billingpb.ResponseStatusSystemError
		rsp.Message = errorCustomerUnknown
		return nil
	}

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
					if order.Refund != nil &&  order.Refund.Amount > 0 {
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
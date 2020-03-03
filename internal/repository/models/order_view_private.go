package models

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type MgoOrderViewPrivate struct {
	Id                                         primitive.ObjectID                       `bson:"_id" json:"-" faker:"objectId"`
	Uuid                                       string                                   `bson:"uuid" json:"uuid"`
	TotalPaymentAmount                         float64                                  `bson:"total_payment_amount" json:"total_payment_amount"`
	Currency                                   string                                   `bson:"currency" json:"currency"`
	Project                                    *MgoOrderProject                         `bson:"project" json:"project"`
	CreatedAt                                  time.Time                                `bson:"created_at" json:"created_at"`
	Transaction                                string                                   `bson:"pm_order_id" json:"transaction"`
	PaymentMethod                              *MgoOrderPaymentMethod                   `bson:"payment_method" json:"payment_method"`
	CountryCode                                string                                   `bson:"country_code" json:"country_code"`
	MerchantId                                 primitive.ObjectID                       `bson:"merchant_id" json:"merchant_id" faker:"objectId"`
	Locale                                     string                                   `bson:"locale" json:"locale"`
	Status                                     string                                   `bson:"status" json:"status"`
	TransactionDate                            time.Time                                `bson:"pm_order_close_date" json:"transaction_date"`
	User                                       *billingpb.OrderUser                     `bson:"user" json:"user"`
	BillingAddress                             *billingpb.OrderBillingAddress           `bson:"billing_address" json:"billing_address"`
	Type                                       string                                   `bson:"type" json:"type"`
	IsVatDeduction                             bool                                     `bson:"is_vat_deduction" json:"is_vat_deduction"`
	PaymentGrossRevenueLocal                   *billingpb.OrderViewMoney                `bson:"payment_gross_revenue_local" json:"payment_gross_revenue_local"`
	PaymentGrossRevenueOrigin                  *billingpb.OrderViewMoney                `bson:"payment_gross_revenue_origin" json:"payment_gross_revenue_origin"`
	PaymentGrossRevenue                        *billingpb.OrderViewMoney                `bson:"payment_gross_revenue" json:"payment_gross_revenue"`
	PaymentTaxFee                              *billingpb.OrderViewMoney                `bson:"payment_tax_fee" json:"payment_tax_fee"`
	PaymentTaxFeeLocal                         *billingpb.OrderViewMoney                `bson:"payment_tax_fee_local" json:"payment_tax_fee_local"`
	PaymentTaxFeeOrigin                        *billingpb.OrderViewMoney                `bson:"payment_tax_fee_origin" json:"payment_tax_fee_origin"`
	PaymentTaxFeeCurrencyExchangeFee           *billingpb.OrderViewMoney                `bson:"payment_tax_fee_currency_exchange_fee" json:"payment_tax_fee_currency_exchange_fee"`
	PaymentTaxFeeTotal                         *billingpb.OrderViewMoney                `bson:"payment_tax_fee_total" json:"payment_tax_fee_total"`
	PaymentGrossRevenueFx                      *billingpb.OrderViewMoney                `bson:"payment_gross_revenue_fx" json:"payment_gross_revenue_fx"`
	PaymentGrossRevenueFxTaxFee                *billingpb.OrderViewMoney                `bson:"payment_gross_revenue_fx_tax_fee" json:"payment_gross_revenue_fx_tax_fee"`
	PaymentGrossRevenueFxProfit                *billingpb.OrderViewMoney                `bson:"payment_gross_revenue_fx_profit" json:"payment_gross_revenue_fx_profit"`
	GrossRevenue                               *billingpb.OrderViewMoney                `bson:"gross_revenue" json:"gross_revenue"`
	TaxFee                                     *billingpb.OrderViewMoney                `bson:"tax_fee" json:"tax_fee"`
	TaxFeeCurrencyExchangeFee                  *billingpb.OrderViewMoney                `bson:"tax_fee_currency_exchange_fee" json:"tax_fee_currency_exchange_fee"`
	TaxFeeTotal                                *billingpb.OrderViewMoney                `bson:"tax_fee_total" json:"tax_fee_total"`
	MethodFeeTotal                             *billingpb.OrderViewMoney                `bson:"method_fee_total" json:"method_fee_total"`
	MethodFeeTariff                            *billingpb.OrderViewMoney                `bson:"method_fee_tariff" json:"method_fee_tariff"`
	PaysuperMethodFeeTariffSelfCost            *billingpb.OrderViewMoney                `bson:"paysuper_method_fee_tariff_self_cost" json:"paysuper_method_fee_tariff_self_cost"`
	PaysuperMethodFeeProfit                    *billingpb.OrderViewMoney                `bson:"paysuper_method_fee_profit" json:"paysuper_method_fee_profit"`
	MethodFixedFeeTariff                       *billingpb.OrderViewMoney                `bson:"method_fixed_fee_tariff" json:"method_fixed_fee_tariff"`
	PaysuperMethodFixedFeeTariffFxProfit       *billingpb.OrderViewMoney                `bson:"paysuper_method_fixed_fee_tariff_fx_profit" json:"paysuper_method_fixed_fee_tariff_fx_profit"`
	PaysuperMethodFixedFeeTariffSelfCost       *billingpb.OrderViewMoney                `bson:"paysuper_method_fixed_fee_tariff_self_cost" json:"paysuper_method_fixed_fee_tariff_self_cost"`
	PaysuperMethodFixedFeeTariffTotalProfit    *billingpb.OrderViewMoney                `bson:"paysuper_method_fixed_fee_tariff_total_profit" json:"paysuper_method_fixed_fee_tariff_total_profit"`
	PaysuperFixedFee                           *billingpb.OrderViewMoney                `bson:"paysuper_fixed_fee" json:"paysuper_fixed_fee"`
	PaysuperFixedFeeFxProfit                   *billingpb.OrderViewMoney                `bson:"paysuper_fixed_fee_fx_profit" json:"paysuper_fixed_fee_fx_profit"`
	FeesTotal                                  *billingpb.OrderViewMoney                `bson:"fees_total" json:"fees_total"`
	FeesTotalLocal                             *billingpb.OrderViewMoney                `bson:"fees_total_local" json:"fees_total_local"`
	NetRevenue                                 *billingpb.OrderViewMoney                `bson:"net_revenue" json:"net_revenue"`
	PaysuperMethodTotalProfit                  *billingpb.OrderViewMoney                `bson:"paysuper_method_total_profit" json:"paysuper_method_total_profit"`
	PaysuperTotalProfit                        *billingpb.OrderViewMoney                `bson:"paysuper_total_profit" json:"paysuper_total_profit"`
	PaymentRefundGrossRevenueLocal             *billingpb.OrderViewMoney                `bson:"payment_refund_gross_revenue_local" json:"payment_refund_gross_revenue_local"`
	PaymentRefundGrossRevenueOrigin            *billingpb.OrderViewMoney                `bson:"payment_refund_gross_revenue_origin" json:"payment_refund_gross_revenue_origin"`
	PaymentRefundGrossRevenue                  *billingpb.OrderViewMoney                `bson:"payment_refund_gross_revenue" json:"payment_refund_gross_revenue"`
	PaymentRefundTaxFee                        *billingpb.OrderViewMoney                `bson:"payment_refund_tax_fee" json:"payment_refund_tax_fee"`
	PaymentRefundTaxFeeLocal                   *billingpb.OrderViewMoney                `bson:"payment_refund_tax_fee_local" json:"payment_refund_tax_fee_local"`
	PaymentRefundTaxFeeOrigin                  *billingpb.OrderViewMoney                `bson:"payment_refund_tax_fee_origin" json:"payment_refund_tax_fee_origin"`
	PaymentRefundFeeTariff                     *billingpb.OrderViewMoney                `bson:"payment_refund_fee_tariff" json:"payment_refund_fee_tariff"`
	MethodRefundFixedFeeTariff                 *billingpb.OrderViewMoney                `bson:"method_refund_fixed_fee_tariff" json:"method_refund_fixed_fee_tariff"`
	RefundGrossRevenue                         *billingpb.OrderViewMoney                `bson:"refund_gross_revenue" json:"refund_gross_revenue"`
	RefundGrossRevenueFx                       *billingpb.OrderViewMoney                `bson:"refund_gross_revenue_fx" json:"refund_gross_revenue_fx"`
	MethodRefundFeeTariff                      *billingpb.OrderViewMoney                `bson:"method_refund_fee_tariff" json:"method_refund_fee_tariff"`
	PaysuperMethodRefundFeeTariffProfit        *billingpb.OrderViewMoney                `bson:"paysuper_method_refund_fee_tariff_profit" json:"paysuper_method_refund_fee_tariff_profit"`
	PaysuperMethodRefundFixedFeeTariffSelfCost *billingpb.OrderViewMoney                `bson:"paysuper_method_refund_fixed_fee_tariff_self_cost" json:"paysuper_method_refund_fixed_fee_tariff_self_cost"`
	MerchantRefundFixedFeeTariff               *billingpb.OrderViewMoney                `bson:"merchant_refund_fixed_fee_tariff" json:"merchant_refund_fixed_fee_tariff"`
	PaysuperMethodRefundFixedFeeTariffProfit   *billingpb.OrderViewMoney                `bson:"paysuper_method_refund_fixed_fee_tariff_profit" json:"paysuper_method_refund_fixed_fee_tariff_profit"`
	RefundTaxFee                               *billingpb.OrderViewMoney                `bson:"refund_tax_fee" json:"refund_tax_fee"`
	RefundTaxFeeCurrencyExchangeFee            *billingpb.OrderViewMoney                `bson:"refund_tax_fee_currency_exchange_fee" json:"refund_tax_fee_currency_exchange_fee"`
	PaysuperRefundTaxFeeCurrencyExchangeFee    *billingpb.OrderViewMoney                `bson:"paysuper_refund_tax_fee_currency_exchange_fee" json:"paysuper_refund_tax_fee_currency_exchange_fee"`
	RefundTaxFeeTotal                          *billingpb.OrderViewMoney                `bson:"refund_tax_fee_total" json:"refund_tax_fee_total"`
	RefundReverseRevenue                       *billingpb.OrderViewMoney                `bson:"refund_reverse_revenue" json:"refund_reverse_revenue"`
	RefundFeesTotal                            *billingpb.OrderViewMoney                `bson:"refund_fees_total" json:"refund_fees_total"`
	RefundFeesTotalLocal                       *billingpb.OrderViewMoney                `bson:"refund_fees_total_local" json:"refund_fees_total_local"`
	PaysuperRefundTotalProfit                  *billingpb.OrderViewMoney                `bson:"paysuper_refund_total_profit" json:"paysuper_refund_total_profit"`
	Issuer                                     *billingpb.OrderIssuer                   `bson:"issuer"`
	Items                                      []*MgoOrderItem                          `bson:"items"`
	MerchantPayoutCurrency                     string                                   `bson:"merchant_payout_currency"`
	ParentOrder                                *billingpb.ParentOrder                   `bson:"parent_order"`
	Refund                                     *MgoOrderNotificationRefund              `bson:"refund"`
	Cancellation                               *billingpb.OrderNotificationCancellation `bson:"cancellation"`
	MccCode                                    string                                   `bson:"mcc_code"`
	OperatingCompanyId                         string                                   `bson:"operating_company_id"`
	IsHighRisk                                 bool                                     `bson:"is_high_risk"`
	RefundAllowed                              bool                                     `bson:"refund_allowed"`
	VatPayer                                   string                                   `bson:"vat_payer"`
	IsProduction                               bool                                     `bson:"is_production"`
	TaxRate                                    float64                                  `bson:"tax_rate"`
	MerchantInfo                               *billingpb.OrderViewMerchantInfo         `bson:"merchant_info"`
	OrderChargeBeforeVat                       *billingpb.OrderViewMoney                `bson:"order_charge_before_vat"`
}

type orderViewPrivateMapper struct {
}

func NewOrderViewPrivateMapper() Mapper {
	return &orderViewPrivateMapper{}
}

func (o *orderViewPrivateMapper) MapObjectToMgo(obj interface{}) (interface{}, error) {
	panic("not supported")
}

func (o *orderViewPrivateMapper) MapMgoToObject(obj interface{}) (interface{}, error) {
	var err error
	decoded := obj.(*MgoOrderViewPrivate)
	m := &billingpb.OrderViewPrivate{}

	m.Id = decoded.Id.Hex()
	m.Uuid = decoded.Uuid
	m.TotalPaymentAmount = decoded.TotalPaymentAmount
	m.Currency = decoded.Currency
	m.Project = getOrderProject(decoded.Project)
	m.Transaction = decoded.Transaction
	m.PaymentMethod = getPaymentMethodOrder(decoded.PaymentMethod)
	m.CountryCode = decoded.CountryCode
	m.MerchantId = decoded.MerchantId.Hex()
	m.Locale = decoded.Locale
	m.Status = decoded.Status
	m.User = decoded.User
	m.BillingAddress = decoded.BillingAddress
	m.Type = decoded.Type
	m.Issuer = decoded.Issuer
	m.MerchantPayoutCurrency = decoded.MerchantPayoutCurrency
	m.IsVatDeduction = decoded.IsVatDeduction
	m.ParentOrder = decoded.ParentOrder
	m.Cancellation = decoded.Cancellation

	if decoded.Refund != nil {
		m.Refund = &billingpb.OrderNotificationRefund{
			Amount:        decoded.Refund.Amount,
			Currency:      decoded.Refund.Currency,
			Reason:        decoded.Refund.Reason,
			Code:          decoded.Refund.Code,
			ReceiptNumber: decoded.Refund.ReceiptNumber,
			ReceiptUrl:    decoded.Refund.ReceiptUrl,
		}
	}

	m.PaymentGrossRevenueLocal = getOrderViewMoney(decoded.PaymentGrossRevenueLocal)
	m.PaymentGrossRevenueOrigin = getOrderViewMoney(decoded.PaymentGrossRevenueOrigin)
	m.PaymentGrossRevenue = getOrderViewMoney(decoded.PaymentGrossRevenue)
	m.PaymentTaxFee = getOrderViewMoney(decoded.PaymentTaxFee)
	m.PaymentTaxFeeLocal = getOrderViewMoney(decoded.PaymentTaxFeeLocal)
	m.PaymentTaxFeeOrigin = getOrderViewMoney(decoded.PaymentTaxFeeOrigin)
	m.PaymentTaxFeeCurrencyExchangeFee = getOrderViewMoney(decoded.PaymentTaxFeeCurrencyExchangeFee)
	m.PaymentTaxFeeTotal = getOrderViewMoney(decoded.PaymentTaxFeeTotal)
	m.PaymentGrossRevenueFx = getOrderViewMoney(decoded.PaymentGrossRevenueFx)
	m.PaymentGrossRevenueFxTaxFee = getOrderViewMoney(decoded.PaymentGrossRevenueFxTaxFee)
	m.PaymentGrossRevenueFxProfit = getOrderViewMoney(decoded.PaymentGrossRevenueFxProfit)
	m.GrossRevenue = getOrderViewMoney(decoded.GrossRevenue)
	m.TaxFee = getOrderViewMoney(decoded.TaxFee)
	m.TaxFeeCurrencyExchangeFee = getOrderViewMoney(decoded.TaxFeeCurrencyExchangeFee)
	m.TaxFeeTotal = getOrderViewMoney(decoded.TaxFeeTotal)
	m.MethodFeeTotal = getOrderViewMoney(decoded.MethodFeeTotal)
	m.MethodFeeTariff = getOrderViewMoney(decoded.MethodFeeTariff)
	m.PaysuperMethodFeeTariffSelfCost = getOrderViewMoney(decoded.PaysuperMethodFeeTariffSelfCost)
	m.PaysuperMethodFeeProfit = getOrderViewMoney(decoded.PaysuperMethodFeeProfit)
	m.MethodFixedFeeTariff = getOrderViewMoney(decoded.MethodFixedFeeTariff)
	m.PaysuperMethodFixedFeeTariffFxProfit = getOrderViewMoney(decoded.PaysuperMethodFixedFeeTariffFxProfit)
	m.PaysuperMethodFixedFeeTariffSelfCost = getOrderViewMoney(decoded.PaysuperMethodFixedFeeTariffSelfCost)
	m.PaysuperMethodFixedFeeTariffTotalProfit = getOrderViewMoney(decoded.PaysuperMethodFixedFeeTariffTotalProfit)
	m.PaysuperFixedFee = getOrderViewMoney(decoded.PaysuperFixedFee)
	m.PaysuperFixedFeeFxProfit = getOrderViewMoney(decoded.PaysuperFixedFeeFxProfit)
	m.FeesTotal = getOrderViewMoney(decoded.FeesTotal)
	m.FeesTotalLocal = getOrderViewMoney(decoded.FeesTotalLocal)
	m.NetRevenue = getOrderViewMoney(decoded.NetRevenue)
	m.PaysuperMethodTotalProfit = getOrderViewMoney(decoded.PaysuperMethodTotalProfit)
	m.PaysuperTotalProfit = getOrderViewMoney(decoded.PaysuperTotalProfit)
	m.PaymentRefundGrossRevenueLocal = getOrderViewMoney(decoded.PaymentRefundGrossRevenueLocal)
	m.PaymentRefundGrossRevenueOrigin = getOrderViewMoney(decoded.PaymentRefundGrossRevenueOrigin)
	m.PaymentRefundGrossRevenue = getOrderViewMoney(decoded.PaymentRefundGrossRevenue)
	m.PaymentRefundTaxFee = getOrderViewMoney(decoded.PaymentRefundTaxFee)
	m.PaymentRefundTaxFeeLocal = getOrderViewMoney(decoded.PaymentRefundTaxFeeLocal)
	m.PaymentRefundTaxFeeOrigin = getOrderViewMoney(decoded.PaymentRefundTaxFeeOrigin)
	m.PaymentRefundFeeTariff = getOrderViewMoney(decoded.PaymentRefundFeeTariff)
	m.MethodRefundFixedFeeTariff = getOrderViewMoney(decoded.MethodRefundFixedFeeTariff)
	m.RefundGrossRevenue = getOrderViewMoney(decoded.RefundGrossRevenue)
	m.RefundGrossRevenueFx = getOrderViewMoney(decoded.RefundGrossRevenueFx)
	m.MethodRefundFeeTariff = getOrderViewMoney(decoded.MethodRefundFeeTariff)
	m.PaysuperMethodRefundFeeTariffProfit = getOrderViewMoney(decoded.PaysuperMethodRefundFeeTariffProfit)
	m.PaysuperMethodRefundFixedFeeTariffSelfCost = getOrderViewMoney(decoded.PaysuperMethodRefundFixedFeeTariffSelfCost)
	m.MerchantRefundFixedFeeTariff = getOrderViewMoney(decoded.MerchantRefundFixedFeeTariff)
	m.PaysuperMethodRefundFixedFeeTariffProfit = getOrderViewMoney(decoded.PaysuperMethodRefundFixedFeeTariffProfit)
	m.RefundTaxFee = getOrderViewMoney(decoded.RefundTaxFee)
	m.RefundTaxFeeCurrencyExchangeFee = getOrderViewMoney(decoded.RefundTaxFeeCurrencyExchangeFee)
	m.PaysuperRefundTaxFeeCurrencyExchangeFee = getOrderViewMoney(decoded.PaysuperRefundTaxFeeCurrencyExchangeFee)
	m.RefundTaxFeeTotal = getOrderViewMoney(decoded.RefundTaxFeeTotal)
	m.RefundReverseRevenue = getOrderViewMoney(decoded.RefundReverseRevenue)
	m.RefundFeesTotal = getOrderViewMoney(decoded.RefundFeesTotal)
	m.RefundFeesTotalLocal = getOrderViewMoney(decoded.RefundFeesTotalLocal)
	m.PaysuperRefundTotalProfit = getOrderViewMoney(decoded.PaysuperRefundTotalProfit)
	m.Items = getOrderViewItems(decoded.Items)
	m.MccCode = decoded.MccCode
	m.OperatingCompanyId = decoded.OperatingCompanyId
	m.IsHighRisk = decoded.IsHighRisk
	m.RefundAllowed = decoded.RefundAllowed
	m.VatPayer = decoded.VatPayer
	m.IsProduction = decoded.IsProduction
	m.TaxRate = decoded.TaxRate
	m.MerchantInfo = decoded.MerchantInfo
	m.OrderChargeBeforeVat = decoded.OrderChargeBeforeVat

	m.CreatedAt, err = ptypes.TimestampProto(decoded.CreatedAt)
	if err != nil {
		return nil, err
	}

	m.TransactionDate, err = ptypes.TimestampProto(decoded.TransactionDate)
	if err != nil {
		return nil, err
	}

	return m, nil
}

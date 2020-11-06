package service

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	geoip "github.com/ProtocolONE/geoip-service/pkg/proto"
	"github.com/golang/protobuf/jsonpb"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/paysuper/paysuper-billing-server/internal/helper"
	"github.com/paysuper/paysuper-billing-server/internal/payment_system"
	intPkg "github.com/paysuper/paysuper-billing-server/internal/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg"
	errors2 "github.com/paysuper/paysuper-billing-server/pkg/errors"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/currenciespb"
	"github.com/paysuper/paysuper-proto/go/postmarkpb"
	"github.com/paysuper/paysuper-proto/go/recurringpb"
	"github.com/paysuper/paysuper-proto/go/taxpb"
	tools "github.com/paysuper/paysuper-tools/number"
	stringTools "github.com/paysuper/paysuper-tools/string"
	"github.com/streadway/amqp"
	"github.com/ttacon/libphonenumber"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	paymentRequestIncorrect             = "payment request has incorrect format"
	callbackRequestIncorrect            = "callback request has incorrect format"
	callbackHandlerIncorrect            = "unknown callback type"
	orderErrorPublishNotificationFailed = "publish order notification failed"
	orderErrorUpdateOrderDataFailed     = "update order data failed"
	brokerPublicationFailed             = "message publication to broker failed"
	subscriptionUpdateFailed            = "unable to update recurring subscription"

	defaultExpireDateToFormInput = 30
	cookieCounterUpdateTime      = 1800

	taxTypeVat      = "vat"
	taxTypeSalesTax = "sales_tax"

	defaultPaymentFormOpeningMode = "embed"
)

var (
	orderErrorProjectIdIncorrect                              = errors2.NewBillingServerErrorMsg("fm000001", "project identifier is incorrect")
	orderErrorProjectNotFound                                 = errors2.NewBillingServerErrorMsg("fm000002", "project with specified identifier not found")
	orderErrorProjectInactive                                 = errors2.NewBillingServerErrorMsg("fm000003", "project with specified identifier is inactive")
	orderErrorProjectMerchantInactive                         = errors2.NewBillingServerErrorMsg("fm000004", "merchant for project with specified identifier is inactive")
	orderErrorPaymentMethodNotAllowed                         = errors2.NewBillingServerErrorMsg("fm000005", "payment Method not available for project")
	orderErrorPaymentMethodNotFound                           = errors2.NewBillingServerErrorMsg("fm000006", "payment Method with specified identifier not found")
	orderErrorPaymentMethodInactive                           = errors2.NewBillingServerErrorMsg("fm000007", "payment Method with specified identifier is inactive")
	orderErrorConvertionCurrency                              = errors2.NewBillingServerErrorMsg("fm000008", "currency conversion error")
	orderErrorPaymentMethodEmptySettings                      = errors2.NewBillingServerErrorMsg("fm000009", "payment Method setting for project is empty")
	orderErrorPaymentSystemInactive                           = errors2.NewBillingServerErrorMsg("fm000010", "payment system for specified payment Method is inactive")
	orderErrorPayerRegionUnknown                              = errors2.NewBillingServerErrorMsg("fm000011", "payer region can't be found")
	orderErrorDynamicNotifyUrlsNotAllowed                     = errors2.NewBillingServerErrorMsg("fm000013", "dynamic verify url or notify url not allowed for project")
	orderErrorDynamicRedirectUrlsNotAllowed                   = errors2.NewBillingServerErrorMsg("fm000014", "dynamic payer redirect urls not allowed for project")
	orderErrorCurrencyNotFound                                = errors2.NewBillingServerErrorMsg("fm000015", "currency received from request not found")
	orderErrorAmountLowerThanMinAllowed                       = errors2.NewBillingServerErrorMsg("fm000016", "order amount is lower than min allowed payment amount for project")
	orderErrorAmountGreaterThanMaxAllowed                     = errors2.NewBillingServerErrorMsg("fm000017", "order amount is greater than max allowed payment amount for project")
	orderErrorAmountLowerThanMinAllowedPaymentMethod          = errors2.NewBillingServerErrorMsg("fm000018", "order amount is lower than min allowed payment amount for payment Method")
	orderErrorAmountGreaterThanMaxAllowedPaymentMethod        = errors2.NewBillingServerErrorMsg("fm000019", "order amount is greater than max allowed payment amount for payment Method")
	orderErrorCanNotCreate                                    = errors2.NewBillingServerErrorMsg("fm000020", "order can't create. try request later")
	orderErrorNotFound                                        = errors2.NewBillingServerErrorMsg("fm000021", "order with specified identifier not found")
	orderErrorOrderCreatedAnotherProject                      = errors2.NewBillingServerErrorMsg("fm000022", "order created for another project")
	orderErrorFormInputTimeExpired                            = errors2.NewBillingServerErrorMsg("fm000023", "time to enter date on payment form expired")
	orderErrorCurrencyIsRequired                              = errors2.NewBillingServerErrorMsg("fm000024", "parameter currency in create order request is required")
	orderErrorUnknown                                         = errors2.NewBillingServerErrorMsg("fm000025", "unknown error. try request later")
	orderCountryPaymentRestrictedError                        = errors2.NewBillingServerErrorMsg("fm000027", "payments from your country are not allowed")
	orderGetSavedCardError                                    = errors2.NewBillingServerErrorMsg("fm000028", "saved card data with specified identifier not found")
	orderErrorCountryByPaymentAccountNotFound                 = errors2.NewBillingServerErrorMsg("fm000029", "information about user country can't be found")
	orderErrorPaymentAccountIncorrect                         = errors2.NewBillingServerErrorMsg("fm000030", "account in payment system is incorrect")
	orderErrorProductsEmpty                                   = errors2.NewBillingServerErrorMsg("fm000031", "products set is empty")
	orderErrorProductsInvalid                                 = errors2.NewBillingServerErrorMsg("fm000032", "some products in set are invalid or inactive")
	orderErrorNoProductsCommonCurrency                        = errors2.NewBillingServerErrorMsg("fm000033", "no common prices neither in requested currency nor in default currency")
	orderErrorNoNameInDefaultLanguage                         = errors2.NewBillingServerErrorMsg("fm000034", "no name in default language %s")
	orderErrorNoNameInRequiredLanguage                        = errors2.NewBillingServerErrorMsg("fm000035", "no name in required language %s")
	orderErrorNoDescriptionInDefaultLanguage                  = errors2.NewBillingServerErrorMsg("fm000036", "no description in default language %s")
	orderErrorNoDescriptionInRequiredLanguage                 = errors2.NewBillingServerErrorMsg("fm000037", "no description in required language %s")
	orderErrorProjectMerchantNotFound                         = errors2.NewBillingServerErrorMsg("fm000038", "merchant for project with specified identifier not found")
	orderErrorRecurringCardNotOwnToUser                       = errors2.NewBillingServerErrorMsg("fm000039", "you can't use not own bank card for payment")
	orderErrorNotRestricted                                   = errors2.NewBillingServerErrorMsg("fm000040", "order country not restricted")
	orderErrorEmailRequired                                   = errors2.NewBillingServerErrorMsg("fm000041", "email is required")
	orderErrorCreatePaymentRequiredFieldIdNotFound            = errors2.NewBillingServerErrorMsg("fm000042", "required field with order identifier not found")
	orderErrorCreatePaymentRequiredFieldPaymentMethodNotFound = errors2.NewBillingServerErrorMsg("fm000043", "required field with payment Method identifier not found")
	orderErrorCreatePaymentRequiredFieldEmailNotFound         = errors2.NewBillingServerErrorMsg("fm000044", "required field \"email\" not found")
	orderErrorCreatePaymentRequiredFieldUserCountryNotFound   = errors2.NewBillingServerErrorMsg("fm000045", "user country is required")
	orderErrorCreatePaymentRequiredFieldUserZipNotFound       = errors2.NewBillingServerErrorMsg("fm000046", "user zip is required")
	orderErrorOrderAlreadyComplete                            = errors2.NewBillingServerErrorMsg("fm000047", "order with specified identifier paid early")
	orderErrorSignatureInvalid                                = errors2.NewBillingServerErrorMsg("fm000048", "request signature is invalid")
	orderErrorProductsPrice                                   = errors2.NewBillingServerErrorMsg("fm000051", "can't get product price")
	orderErrorCheckoutWithoutProducts                         = errors2.NewBillingServerErrorMsg("fm000052", "order products not specified")
	orderErrorCheckoutWithoutAmount                           = errors2.NewBillingServerErrorMsg("fm000053", "order amount not specified")
	orderErrorUnknownType                                     = errors2.NewBillingServerErrorMsg("fm000055", "unknown type of order")
	orderErrorMerchantBadTariffs                              = errors2.NewBillingServerErrorMsg("fm000056", "merchant don't have tariffs")
	orderErrorReceiptNotEquals                                = errors2.NewBillingServerErrorMsg("fm000057", "receipts not equals")
	orderErrorDuringFormattingCurrency                        = errors2.NewBillingServerErrorMsg("fm000058", "error during formatting currency")
	orderErrorDuringFormattingDate                            = errors2.NewBillingServerErrorMsg("fm000059", "error during formatting date")
	orderErrorMerchantForOrderNotFound                        = errors2.NewBillingServerErrorMsg("fm000060", "merchant for order not found")
	orderErrorNoPlatforms                                     = errors2.NewBillingServerErrorMsg("fm000062", "no available platforms")
	orderCountryPaymentRestricted                             = errors2.NewBillingServerErrorMsg("fm000063", "payments from your country are not allowed")
	orderErrorCostsRatesNotFound                              = errors2.NewBillingServerErrorMsg("fm000064", "settings to calculate commissions for order not found")
	orderErrorVirtualCurrencyNotFilled                        = errors2.NewBillingServerErrorMsg("fm000065", "virtual currency is not filled")
	orderErrorVirtualCurrencyFracNotSupported                 = errors2.NewBillingServerErrorMsg("fm000066", "fractional numbers is not supported for this virtual currency")
	orderErrorVirtualCurrencyLimits                           = errors2.NewBillingServerErrorMsg("fm000067", "amount of order is more than max amount or less than minimal amount for virtual currency")
	orderErrorCheckoutWithProducts                            = errors2.NewBillingServerErrorMsg("fm000069", "request to processing simple payment can't contain products list")
	orderErrorMerchantDoNotHaveBanking                        = errors2.NewBillingServerErrorMsg("fm000071", "merchant don't have completed banking info")
	orderErrorMerchantUserAccountNotChecked                   = errors2.NewBillingServerErrorMsg("fm000073", "failed to check user account")
	orderErrorAmountLowerThanMinLimitSystem                   = errors2.NewBillingServerErrorMsg("fm000074", "order amount is lower than min system limit")
	orderErrorAlreadyProcessed                                = errors2.NewBillingServerErrorMsg("fm000075", "order is already processed")
	orderErrorDontHaveReceiptUrl                              = errors2.NewBillingServerErrorMsg("fm000076", "processed order don't have receipt url")
	orderErrorWrongPrivateStatus                              = errors2.NewBillingServerErrorMsg("fm000077", "order has wrong private status and cannot be recreated")
	orderCountryChangeRestrictedError                         = errors2.NewBillingServerErrorMsg("fm000078", "change country is not allowed")
	orderErrorVatPayerUnknown                                 = errors2.NewBillingServerErrorMsg("fm000079", "vat payer unknown")
	orderErrorCookieIsEmpty                                   = errors2.NewBillingServerErrorMsg("fm000080", "can't get payment cookie")
	orderErrorCookieInvalid                                   = errors2.NewBillingServerErrorMsg("fm000081", "unable to read payment cookie")
	orderErrorRecurringNotAllowed                             = errors2.NewBillingServerErrorMsg("fm000082", "payment method not allowed recurring payments")
	orderErrorRecurringDateEndInvalid                         = errors2.NewBillingServerErrorMsg("fm000083", "invalid the end date of recurring payments")
	orderErrorRecurringDateEndOutOfRange                      = errors2.NewBillingServerErrorMsg("fm000084", "subscription period cannot be less than the selected period and more than one year")
	orderErrorRecurringInvalidPeriod                          = errors2.NewBillingServerErrorMsg("fm000085", "recurring period subscription is invalid")
	orderErrorRecurringSubscriptionNotFound                   = errors2.NewBillingServerErrorMsg("fm000086", "recurring subscription not found")
	orderErrorRecurringUnableToAdd                            = errors2.NewBillingServerErrorMsg("fm000087", "unable to add recurring subscription")
	orderErrorRecurringUnableToUpdate                         = errors2.NewBillingServerErrorMsg("fm000088", "unable to update recurring subscription")

	virtualCurrencyPayoutCurrencyMissed = errors2.NewBillingServerErrorMsg("vc000001", "virtual currency don't have price in merchant payout currency")

	paymentSystemPaymentProcessingSuccessStatus = "PAYMENT_SYSTEM_PROCESSING_SUCCESS"

	possiblePaymentFormOpeningModes = map[string]bool{"embed": true, "iframe": true, "standalone": true}
)

type orderCreateRequestProcessorChecked struct {
	id                      string
	project                 *billingpb.Project
	merchant                *billingpb.Merchant
	currency                string
	amount                  float64
	paymentMethod           *billingpb.PaymentMethod
	products                []string
	items                   []*billingpb.OrderItem
	metadata                map[string]string
	privateMetadata         map[string]string
	user                    *billingpb.OrderUser
	virtualAmount           float64
	mccCode                 string
	operatingCompanyId      string
	priceGroup              *billingpb.PriceGroup
	isCurrencyPredefined    bool
	isBuyForVirtualCurrency bool
	recurringPeriod         string
	recurringInterval       int32
	recurringDateEnd        string
}

type OrderCreateRequestProcessor struct {
	*Service
	checked *orderCreateRequestProcessorChecked
	request *billingpb.OrderCreateRequest
	ctx     context.Context
}

type PaymentFormProcessor struct {
	service *Service
	order   *billingpb.Order
	request *billingpb.PaymentFormJsonDataRequest
}

type PaymentCreateProcessor struct {
	service        *Service
	data           map[string]string
	ip             string
	acceptLanguage string
	userAgent      string
	checked        struct {
		order         *billingpb.Order
		project       *billingpb.Project
		paymentMethod *billingpb.PaymentMethod
	}
}

func (s *Service) OrderCreateByPaylink(
	ctx context.Context,
	req *billingpb.OrderCreateByPaylink,
	rsp *billingpb.OrderCreateProcessResponse,
) error {
	pl, err := s.paylinkRepository.GetById(ctx, req.PaylinkId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			rsp.Status = billingpb.ResponseStatusNotFound
			rsp.Message = errorPaylinkNotFound
			return nil
		}
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	if pl.GetIsExpired() == true {
		rsp.Status = billingpb.ResponseStatusGone
		rsp.Message = errorPaylinkExpired
		return nil
	}

	oReq := &billingpb.OrderCreateRequest{
		ProjectId: pl.ProjectId,
		User: &billingpb.OrderUser{
			Ip: req.PayerIp,
		},
		Products: pl.Products,
		PrivateMetadata: map[string]string{
			"PaylinkId": pl.Id,
		},
		Type:                pl.ProductsType,
		IssuerUrl:           req.IssuerUrl,
		IsEmbedded:          req.IsEmbedded,
		IssuerReferenceType: pkg.OrderIssuerReferenceTypePaylink,
		IssuerReference:     pl.Id,
		UtmSource:           req.UtmSource,
		UtmMedium:           req.UtmMedium,
		UtmCampaign:         req.UtmCampaign,
		Cookie:              req.Cookie,
	}

	err = s.OrderCreateProcess(ctx, oReq, rsp)
	if err != nil {
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	return nil
}

func (s *Service) OrderCreateProcess(
	ctx context.Context,
	req *billingpb.OrderCreateRequest,
	rsp *billingpb.OrderCreateProcessResponse,
) error {
	rsp.Status = billingpb.ResponseStatusOk

	processor := &OrderCreateRequestProcessor{
		Service: s,
		request: req,
		checked: &orderCreateRequestProcessorChecked{},
		ctx:     ctx,
	}

	if req.Token != "" {
		err := processor.processCustomerToken(ctx)

		if err != nil {
			zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
			if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
				rsp.Status = billingpb.ResponseStatusBadData
				rsp.Message = e
				return nil
			}
			return err
		}
	} else {
		if req.ProjectId == "" {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = orderErrorProjectIdIncorrect
			return nil
		}

		_, err := primitive.ObjectIDFromHex(req.ProjectId)

		if err != nil {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = orderErrorProjectIdIncorrect
			return nil
		}
	}

	if err := processor.processProject(); err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	if err := processor.processMerchant(); err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	if processor.checked.mccCode == "" {
		processor.checked.mccCode = processor.checked.merchant.MccCode
	}
	if processor.checked.operatingCompanyId == "" {
		processor.checked.operatingCompanyId = processor.checked.merchant.OperatingCompanyId
	}

	if req.Signature != "" || processor.checked.project.SignatureRequired == true {
		if err := processor.processSignature(); err != nil {
			zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
			if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
				rsp.Status = billingpb.ResponseStatusBadData
				rsp.Message = e
				return nil
			}
			return err
		}
	}

	switch req.Type {
	case pkg.OrderType_simple, pkg.OrderTypeVirtualCurrency:
		if req.Products != nil {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = orderErrorCheckoutWithProducts
			return nil
		}

		if req.Amount <= 0 {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = orderErrorCheckoutWithoutAmount
			return nil
		}
		break
	case pkg.OrderType_product, pkg.OrderType_key:
		if req.Amount > float64(0) {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = orderErrorCheckoutWithoutProducts
			return nil
		}
		break
	default:
		rsp.Status = billingpb.ResponseStatusBadData
		rsp.Message = orderErrorUnknownType
		return nil
	}

	if req.User != nil {
		err := processor.processUserData()

		if err != nil {
			zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
			if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
				rsp.Status = billingpb.ResponseStatusBadData
				rsp.Message = e
				return nil
			}
			return err
		}
	}

	if processor.checked.user != nil && processor.checked.user.Ip != "" && !processor.checked.user.HasAddress() {
		err := processor.processPayerIp(ctx)

		if err != nil {
			zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
			if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
				rsp.Status = billingpb.ResponseStatusBadData
				rsp.Message = e
				return nil
			}
			return err
		}

		// try to restore country change from cookie
		if req.Cookie != "" {
			decryptedBrowserCustomer, err := s.decryptBrowserCookie(req.Cookie)

			if err == nil &&
				processor.checked.user.Address != nil &&
				processor.checked.user.Address.Country == decryptedBrowserCustomer.IpCountry &&
				processor.checked.user.Ip == decryptedBrowserCustomer.Ip &&
				decryptedBrowserCustomer.SelectedCountry != "" {

				processor.checked.user.Address = &billingpb.OrderBillingAddress{
					Country: decryptedBrowserCustomer.SelectedCountry,
				}
			}
		}
	}

	err := processor.processCurrency(req.Type)
	if err != nil {
		zap.L().Error("process currency failed", zap.Error(err))
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	switch req.Type {
	case pkg.OrderType_simple:
		if req.Amount != 0 {
			processor.processAmount()
		}
		break
	case pkg.OrderTypeVirtualCurrency:
		err := processor.processVirtualCurrency(ctx)
		if err != nil {
			zap.L().Error(
				pkg.MethodFinishedWithError,
				zap.Error(err),
			)

			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = err.(*billingpb.ResponseErrorMessage)
			return nil
		}
		break
	case pkg.OrderType_product:
		if err := processor.processPaylinkProducts(ctx); err != nil {
			if pid := req.PrivateMetadata["PaylinkId"]; pid != "" {
				s.notifyPaylinkError(ctx, pid, err, req, nil)
			}
			zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
			if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
				rsp.Status = billingpb.ResponseStatusBadData
				rsp.Message = e
				return nil
			}

			if err == billingpb.ProductNoPriceInCurrencyError {
				rsp.Status = billingpb.ResponseStatusBadData
				rsp.Message = productNoPriceInCurrencyError
				return nil
			}

			return err
		}
		break
	case pkg.OrderType_key:
		if err := processor.processPaylinkKeyProducts(); err != nil {
			if pid := req.PrivateMetadata["PaylinkId"]; pid != "" {
				s.notifyPaylinkError(ctx, pid, err, req, nil)
			}
			zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
			if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
				rsp.Status = billingpb.ResponseStatusBadData
				rsp.Message = e
				return nil
			}
			return err
		}
		break
	}

	if req.PaymentMethod != "" {
		pm, err := s.paymentMethodRepository.GetByGroupAndCurrency(
			ctx,
			processor.checked.project.IsProduction(),
			req.PaymentMethod,
			processor.checked.currency,
		)

		if err != nil {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = orderErrorPaymentMethodNotFound
			return nil
		}

		if err := processor.processPaymentMethod(pm); err != nil {
			zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
			if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
				rsp.Status = billingpb.ResponseStatusBadData
				rsp.Message = e
				return nil
			}
			return err
		}
	}

	if req.Type == pkg.OrderType_simple {
		if err := processor.processLimitAmounts(); err != nil {
			zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
			if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
				rsp.Status = billingpb.ResponseStatusBadData
				rsp.Message = e
				return nil
			}
			return err
		}
	}

	if req.RecurringPeriod != "" {
		if err := processor.processRecurringSettings(); err != nil {
			zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
			if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
				rsp.Status = billingpb.ResponseStatusBadData
				rsp.Message = e
				return nil
			}
			return err
		}
	}

	processor.processMetadata()
	processor.processPrivateMetadata()

	order, err := processor.prepareOrder()

	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	if err = s.orderRepository.Insert(ctx, order); err != nil {
		rsp.Status = billingpb.ResponseStatusBadData
		rsp.Message = orderErrorCanNotCreate
		return nil
	}

	rsp.Item = order

	return nil
}

func (s *Service) PaymentFormJsonDataProcess(
	ctx context.Context,
	req *billingpb.PaymentFormJsonDataRequest,
	rsp *billingpb.PaymentFormJsonDataResponse,
) error {
	rsp.Status = billingpb.ResponseStatusOk
	rsp.Item = &billingpb.PaymentFormJsonData{}

	order, err := s.getOrderByUuidToForm(ctx, req.OrderId)

	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	rsp.Item.Type = order.ProductType

	if order.IsDeclinedByCountry() {
		rsp.Status = billingpb.ResponseStatusBadData
		rsp.Message = orderCountryPaymentRestrictedError
		return nil
	}

	if order.PrivateStatus != recurringpb.OrderStatusNew && order.PrivateStatus != recurringpb.OrderStatusPaymentSystemComplete {
		if len(order.ReceiptUrl) == 0 {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = orderErrorDontHaveReceiptUrl
			return nil
		}
		s.fillPaymentFormJsonData(order, rsp)
		rsp.Item.IsAlreadyProcessed = true
		rsp.Item.ReceiptUrl = order.ReceiptUrl
		return nil
	}

	p := &PaymentFormProcessor{service: s, order: order, request: req}
	p1 := &OrderCreateRequestProcessor{
		Service: s,
		checked: &orderCreateRequestProcessorChecked{
			user: &billingpb.OrderUser{
				Ip:      req.Ip,
				Address: &billingpb.OrderBillingAddress{},
			},
		},
		ctx: ctx,
	}

	if !order.User.HasAddress() && p1.checked.user.Ip != "" {
		err = p1.processPayerIp(ctx)

		if err != nil {
			zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
			if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
				rsp.Status = billingpb.ResponseStatusSystemError
				rsp.Message = e
				return nil
			}
			return err
		}

		order.User.Ip = p1.checked.user.Ip
		order.User.Address = &billingpb.OrderBillingAddress{
			Country:    p1.checked.user.Address.Country,
			City:       p1.checked.user.Address.City,
			PostalCode: p1.checked.user.Address.PostalCode,
			State:      p1.checked.user.Address.State,
		}
	}

	loc, _ := s.getCountryFromAcceptLanguage(req.Locale)
	isIdentified := helper.IsIdentified(order.User.Id)
	browserCustomer := &BrowserCookieCustomer{
		Ip:             req.Ip,
		UserAgent:      req.UserAgent,
		AcceptLanguage: req.Locale,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	var customer *billingpb.Customer

	if isIdentified == true {
		customer, err = s.processCustomerData(ctx, order.User.Id, order, req, browserCustomer, loc)

		if err == nil {
			browserCustomer.CustomerId = customer.Id
		}
	} else {
		if req.Cookie != "" {
			decryptedBrowserCustomer, err := s.decryptBrowserCookie(req.Cookie)

			if err == nil {
				if (time.Now().Unix() - decryptedBrowserCustomer.UpdatedAt.Unix()) <= cookieCounterUpdateTime {
					decryptedBrowserCustomer.SessionCount++
				}

				if decryptedBrowserCustomer.CustomerId != "" {
					customer, err = s.processCustomerData(
						ctx,
						decryptedBrowserCustomer.CustomerId,
						order,
						req,
						decryptedBrowserCustomer,
						loc,
					)

					if err != nil {
						zap.L().Error("Customer by identifier in browser cookie not processed", zap.Error(err))
					}

					if customer != nil {
						browserCustomer = decryptedBrowserCustomer
						order.User.Id = customer.Id
						order.User.TechEmail = customer.TechEmail
						order.User.Uuid = customer.Uuid
					} else {
						browserCustomer.VirtualCustomerId = s.getTokenString(s.cfg.Length)
					}
				} else {
					if decryptedBrowserCustomer.VirtualCustomerId == "" {
						browserCustomer.VirtualCustomerId = s.getTokenString(s.cfg.Length)
					} else {
						browserCustomer.VirtualCustomerId = decryptedBrowserCustomer.VirtualCustomerId
					}
				}

				// restore user address from cookie, if it was changed manually
				if order.User.Address != nil &&
					order.User.Address.Country == decryptedBrowserCustomer.IpCountry &&
					order.User.Ip == decryptedBrowserCustomer.Ip &&
					decryptedBrowserCustomer.SelectedCountry != "" {

					order.User.Address = &billingpb.OrderBillingAddress{
						Country: decryptedBrowserCustomer.SelectedCountry,
					}
				}

			} else {
				browserCustomer.VirtualCustomerId = s.getTokenString(s.cfg.Length)
			}
		} else {
			browserCustomer.VirtualCustomerId = s.getTokenString(s.cfg.Length)
		}

		if order.User.Id == "" {
			order.User.Id = browserCustomer.VirtualCustomerId
		}

		if order.User.TechEmail == "" {
			order.User.TechEmail = order.User.Id + pkg.TechEmailDomain
		}
	}

	if order.PaymentMethod != nil && order.PaymentMethod.RecurringAllowed && customer != nil && customer.Id != "" {
		req := &recurringpb.FindSubscriptionsRequest{
			CustomerId: customer.Id,
		}

		subscriptions, err := s.rep.FindSubscriptions(ctx, req)

		if err != nil {
			zap.L().Error(
				pkg.ErrorGrpcServiceCallFailed,
				zap.Error(err),
				zap.String(errorFieldService, recurringpb.PayOneRepositoryServiceName),
				zap.String(errorFieldMethod, "FindSubscriptions"),
			)
		}

		if len(subscriptions.List) > 0 {
			rsp.Item.RecurringManagementUrl = fmt.Sprintf("%s/subscriptions", s.cfg.CheckoutUrl)
		}
	}

	if order.User.Locale == "" && loc != "" && loc != order.User.Locale {
		order.User.Locale = loc
	}

	restricted, err := s.applyCountryRestriction(ctx, order, order.GetCountry())
	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusSystemError
			rsp.Message = e
			return nil
		}
		return err
	}
	if restricted {
		rsp.Status = billingpb.ResponseStatusForbidden
		rsp.Message = orderCountryPaymentRestricted
		rsp.Item.Id = order.Uuid
		return nil
	}

	switch order.ProductType {
	case pkg.OrderType_product:
		err = s.ProcessOrderProducts(ctx, order)
		break
	case pkg.OrderType_key:
		rsp.Item.Platforms, err = s.ProcessOrderKeyProducts(ctx, order)
	case pkg.OrderTypeVirtualCurrency:
		err = s.ProcessOrderVirtualCurrency(ctx, order)
	}

	if err != nil {
		if pid := order.PrivateMetadata["PaylinkId"]; pid != "" {
			s.notifyPaylinkError(ctx, pid, err, req, order)
		}
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusSystemError
			rsp.Message = e
			return nil
		}
		return err
	}

	if order.Issuer == nil {
		order.Issuer = &billingpb.OrderIssuer{
			Embedded: req.IsEmbedded,
		}
	}
	if order.Issuer.Url == "" {
		order.Issuer.Url = req.Referer
		order.Issuer.ReferrerHost = getHostFromUrl(req.Referer)
	}
	if order.Issuer.ReferenceType == "" {
		order.Issuer.ReferenceType = req.IssuerReferenceType
	}
	if order.Issuer.Reference == "" {
		order.Issuer.Reference = req.IssuerReference
	}
	if order.Issuer.UtmSource == "" {
		order.Issuer.UtmSource = req.UtmSource
	}
	if order.Issuer.UtmCampaign == "" {
		order.Issuer.UtmCampaign = req.UtmCampaign
	}
	if order.Issuer.UtmMedium == "" {
		order.Issuer.UtmMedium = req.UtmMedium
	}
	order.Issuer.ReferrerHost = getHostFromUrl(order.Issuer.Url)

	err = p1.processOrderVat(order)
	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error(), "method", "processOrderVat")
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusSystemError
			rsp.Message = e
			return nil
		}
		return err
	}

	err = s.updateOrder(ctx, order)

	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusSystemError
			rsp.Message = e
			return nil
		}
		return err
	}

	pms, err := p.processRenderFormPaymentMethods(ctx)

	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusSystemError
			rsp.Message = e

			if e == orderErrorPaymentMethodNotAllowed {
				rsp.Status = billingpb.ResponseStatusNotFound
			}

			return nil
		}
		return err
	}

	s.fillPaymentFormJsonData(order, rsp)
	rsp.Item.PaymentMethods = pms

	rsp.Item.VatInChargeCurrency = tools.FormatAmount(order.GetTaxAmountInChargeCurrency())
	rsp.Item.VatRate = tools.ToPrecise(order.Tax.Rate)

	cookie, err := s.generateBrowserCookie(browserCustomer)

	if err == nil {
		rsp.Cookie = cookie
	}

	return nil
}

func (s *Service) fillPaymentFormJsonData(order *billingpb.Order, rsp *billingpb.PaymentFormJsonDataResponse) {
	projectName, ok := order.Project.Name[order.User.Locale]

	if !ok {
		projectName = order.Project.Name[DefaultLanguage]
	}

	expire := time.Now().Add(time.Minute * 30).Unix()

	rsp.Item.Id = order.Uuid
	rsp.Item.Account = order.User.ExternalId
	rsp.Item.Description = order.Description
	rsp.Item.HasVat = order.Tax.Amount > 0
	rsp.Item.Vat = order.Tax.Amount
	rsp.Item.Currency = order.Currency
	rsp.Item.Project = &billingpb.PaymentFormJsonDataProject{
		Id:               order.Project.Id,
		Name:             projectName,
		UrlSuccess:       order.Project.UrlSuccess,
		UrlFail:          order.Project.UrlFail,
		RedirectSettings: order.Project.RedirectSettings,
	}
	rsp.Item.Token = s.centrifugoPaymentForm.GetChannelToken(order.Uuid, expire)
	rsp.Item.Amount = order.OrderAmount
	rsp.Item.TotalAmount = order.TotalPaymentAmount
	rsp.Item.ChargeCurrency = order.ChargeCurrency
	rsp.Item.ChargeAmount = order.ChargeAmount
	rsp.Item.Items = order.Items
	rsp.Item.Email = order.User.Email
	rsp.Item.UserAddressDataRequired = order.UserAddressDataRequired

	if order.CountryRestriction != nil {
		rsp.Item.CountryPaymentsAllowed = order.CountryRestriction.PaymentsAllowed
		rsp.Item.CountryChangeAllowed = order.CountryRestriction.ChangeAllowed
	} else {
		rsp.Item.CountryPaymentsAllowed = true
		rsp.Item.CountryChangeAllowed = true
	}

	rsp.Item.UserIpData = &billingpb.UserIpData{
		Country: order.User.Address.Country,
		City:    order.User.Address.City,
		Zip:     order.User.Address.PostalCode,
	}
	rsp.Item.Lang = order.User.Locale
	rsp.Item.VatPayer = order.VatPayer
	rsp.Item.IsProduction = order.IsProduction

	if order.RecurringSettings != nil {
		rsp.Item.RecurringSettings = order.RecurringSettings
	}
}

func (s *Service) PaymentCreateProcess(
	ctx context.Context,
	req *billingpb.PaymentCreateRequest,
	rsp *billingpb.PaymentCreateResponse,
) error {
	processor := &PaymentCreateProcessor{
		service:        s,
		data:           req.Data,
		ip:             req.Ip,
		acceptLanguage: req.AcceptLanguage,
		userAgent:      req.UserAgent,
	}

	if req.Cookie == "" {
		rsp.Message = orderErrorCookieIsEmpty
		rsp.Status = billingpb.ResponseStatusBadData
		return nil
	}

	decryptedBrowserCustomer, err := s.decryptBrowserCookie(req.Cookie)
	if err != nil {
		rsp.Message = orderErrorCookieInvalid
		rsp.Status = billingpb.ResponseStatusBadData
		return nil
	}

	err = processor.processPaymentFormData(ctx)
	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	order := processor.checked.order

	decryptedBrowserCustomer.CustomerId = order.User.Id
	cookie, err := s.generateBrowserCookie(decryptedBrowserCustomer)

	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
	}

	rsp.Cookie = cookie

	if !order.CountryRestriction.PaymentsAllowed {
		rsp.Message = orderCountryPaymentRestrictedError
		rsp.Status = billingpb.ResponseStatusForbidden
		return nil
	}

	if order.ProductType == pkg.OrderType_product {
		err = s.ProcessOrderProducts(ctx, order)
	} else if order.ProductType == pkg.OrderType_key {
		// We should reserve keys only before payment
		if _, err = s.ProcessOrderKeyProducts(ctx, order); err == nil {
			err = processor.reserveKeysForOrder(ctx, order)
		}
	} else if order.ProductType == pkg.OrderTypeVirtualCurrency {
		err = s.ProcessOrderVirtualCurrency(ctx, order)
	}

	if err != nil {
		if pid := order.PrivateMetadata["PaylinkId"]; pid != "" {
			s.notifyPaylinkError(ctx, pid, err, req, order)
		}

		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	p1 := &OrderCreateRequestProcessor{Service: s, ctx: ctx}
	err = p1.processOrderVat(order)
	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error(), "method", "processOrderVat")
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusSystemError
			rsp.Message = e
			return nil
		}
		return err
	}

	if req.Ip != "" {
		address, err := s.getAddressByIp(ctx, req.Ip)
		if err == nil {
			order.PaymentIpCountry = address.Country
		}
	}

	ps, err := s.paymentSystemRepository.GetById(ctx, processor.checked.paymentMethod.PaymentSystemId)
	if err != nil {
		rsp.Message = orderErrorPaymentSystemInactive
		rsp.Status = billingpb.ResponseStatusBadData

		return nil
	}

	order.PaymentMethod = &billingpb.PaymentMethodOrder{
		Id:               processor.checked.paymentMethod.Id,
		Name:             processor.checked.paymentMethod.Name,
		PaymentSystemId:  ps.Id,
		Group:            processor.checked.paymentMethod.Group,
		ExternalId:       processor.checked.paymentMethod.ExternalId,
		Handler:          ps.Handler,
		RefundAllowed:    processor.checked.paymentMethod.RefundAllowed,
		RecurringAllowed: processor.checked.paymentMethod.RecurringAllowed,
	}

	err = s.setOrderChargeAmountAndCurrency(ctx, order)
	if err != nil {
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	methodName, err := order.GetCostPaymentMethodName()
	if err != nil {
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}

		return err
	}

	order.PaymentMethod.Params, err = s.getPaymentSettings(
		processor.checked.paymentMethod,
		order.ChargeCurrency,
		order.MccCode,
		order.OperatingCompanyId,
		methodName,
		order.IsProduction,
	)

	if err != nil {
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}

		return err
	}

	if _, ok := order.PaymentRequisites[billingpb.PaymentCreateFieldRecurringId]; ok {
		req.Data[billingpb.PaymentCreateFieldRecurringId] = order.PaymentRequisites[billingpb.PaymentCreateFieldRecurringId]
		delete(order.PaymentRequisites, billingpb.PaymentCreateFieldRecurringId)
	}

	merchant, err := s.merchantRepository.GetById(ctx, order.GetMerchantId())
	if err != nil {
		return merchantErrorNotFound
	}
	order.MccCode = merchant.MccCode
	order.IsHighRisk = merchant.IsHighRisk()

	order.OperatingCompanyId, err = s.getOrderOperatingCompanyId(ctx, order.GetCountry(), merchant)
	if err != nil {
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	err = s.updateOrder(ctx, order)

	if err != nil {
		zap.L().Error(
			"s.updateOrder Method failed",
			zap.Error(err),
			zap.Any("order", order),
		)
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusSystemError
			rsp.Message = e
			return nil
		} else {
			rsp.Message = orderErrorUnknown
			rsp.Status = billingpb.ResponseStatusSystemError
		}
		return nil
	}

	if !s.hasPaymentCosts(ctx, order) {
		rsp.Status = billingpb.ResponseStatusBadData
		rsp.Message = orderErrorCostsRatesNotFound
		return nil
	}

	h, err := s.paymentSystemGateway.GetGateway(order.PaymentMethod.Handler)

	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusSystemError
			rsp.Message = e
			return nil
		}
		return err
	}

	var url string

	if order.PaymentMethod.RecurringAllowed && order.RecurringSettings != nil && order.User.Uuid != "" {
		var subscription *recurringpb.Subscription

		subscription, url, err = s.addRecurringSubscription(ctx, order, h, req.Data)

		if err != nil {
			fmt.Println(err)
			if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
				rsp.Status = billingpb.ResponseStatusSystemError
				rsp.Message = e
				return nil
			}
			return err
		}

		order.Recurring = true
		order.RecurringId = subscription.Id
		rsp.RecurringExpireDate = subscription.ExpireAt
	} else {
		url, err = h.CreatePayment(
			order,
			s.cfg.GetRedirectUrlSuccess(nil),
			s.cfg.GetRedirectUrlFail(nil),
			req.Data,
		)

		if err != nil {
			zap.L().Error(
				"h.CreatePayment Method failed",
				zap.Error(err),
				zap.Any("order", order),
			)
			if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
				rsp.Status = billingpb.ResponseStatusSystemError
				rsp.Message = e
				return nil
			} else {
				rsp.Message = orderErrorUnknown
				rsp.Status = billingpb.ResponseStatusBadData
			}
			return nil
		}
	}

	err = s.updateOrder(ctx, order)
	if err != nil {
		zap.S().Errorf("Order create in payment system failed", "err", err.Error(), "order", order)

		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusSystemError
			rsp.Message = e
			return nil
		}
		return err
	}

	rsp.Status = billingpb.ResponseStatusOk
	rsp.RedirectUrl = url
	rsp.NeedRedirect = true

	if _, ok := req.Data[billingpb.PaymentCreateFieldRecurringId]; ok && url == "" {
		rsp.NeedRedirect = false
	}

	return nil
}

func (s *Service) getOrderOperatingCompanyId(
	ctx context.Context,
	orderCountry string,
	merchant *billingpb.Merchant,
) (string, error) {
	orderOperatingCompany, err := s.operatingCompanyRepository.GetByPaymentCountry(ctx, orderCountry)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return merchant.OperatingCompanyId, nil
		}

		return "", err
	}
	return orderOperatingCompany.Id, nil
}

func (s *Service) PaymentCallbackProcess(
	ctx context.Context,
	req *billingpb.PaymentNotifyRequest,
	rsp *billingpb.PaymentNotifyResponse,
) error {
	order, err := s.getOrderById(ctx, req.OrderId)

	if err != nil {
		return orderErrorNotFound
	}

	var data protobuf.Message

	ps, err := s.paymentSystemRepository.GetById(ctx, order.PaymentMethod.PaymentSystemId)
	if err != nil {
		return orderErrorPaymentSystemInactive
	}

	switch ps.Handler {
	case billingpb.PaymentSystemHandlerCardPay, PaymentSystemHandlerCardPayMock:
		data = &billingpb.CardPayPaymentCallback{}
		err := json.Unmarshal(req.Request, data)

		if err != nil {
			return errors.New(paymentRequestIncorrect)
		}
		break
	default:
		return orderErrorPaymentMethodNotFound
	}

	h, err := s.paymentSystemGateway.GetGateway(ps.Handler)

	if err != nil {
		return err
	}

	pErr := h.ProcessPayment(order, data, string(req.Request), req.Signature)

	if pErr != nil {
		pErr, _ := pErr.(*billingpb.ResponseError)

		rsp.Error = pErr.Error()
		rsp.Status = pErr.Status

		zap.L().Error(
			"error on ProcessPayment method",
			zap.Error(err),
			zap.Any("request", req.Request),
		)

		if pErr.Status == pkg.StatusTemporary {
			return nil
		}
	}

	switch order.PaymentMethod.ExternalId {
	case recurringpb.PaymentSystemGroupAliasBankCard:
		if err := s.fillPaymentDataCard(order); err != nil {
			return err
		}
		break
	case recurringpb.PaymentSystemGroupAliasQiwi,
		recurringpb.PaymentSystemGroupAliasWebMoney,
		recurringpb.PaymentSystemGroupAliasNeteller,
		recurringpb.PaymentSystemGroupAliasAlipay:
		if err := s.fillPaymentDataEwallet(order); err != nil {
			return err
		}
		break
	case recurringpb.PaymentSystemGroupAliasBitcoin:
		if err := s.fillPaymentDataCrypto(order); err != nil {
			return err
		}
		break
	}

	merchant, err := s.merchantRepository.GetById(ctx, order.GetMerchantId())
	if err != nil {
		return err
	}

	if order.IsProduction {
		zap.L().Info("debug info", zap.Any("merchant_first_payment_at", merchant.FirstPaymentAt))
		if merchant.FirstPaymentAt == nil || merchant.FirstPaymentAt.Seconds <= 0 {
			currentTimeOrder := order.PaymentMethodOrderClosedAt
			merchant.FirstPaymentAt = currentTimeOrder
			order.Project.FirstPaymentAt = currentTimeOrder
			err = s.merchantRepository.Update(ctx, merchant)
			if err != nil {
				zap.L().Error("can't update first_payment_at field", zap.Error(err), zap.String("merchant_id", merchant.Id), zap.Any("time", currentTimeOrder))
				return err
			}
		} else {
			if order.Project.FirstPaymentAt == nil || merchant.FirstPaymentAt.Seconds <= 0 {
				order.Project.FirstPaymentAt = merchant.FirstPaymentAt
			}
		}
	}

	var subscription *recurringpb.Subscription

	if h.IsSubscriptionCallback(data) {
		subscriptionRsp, err := s.rep.GetSubscription(ctx, &recurringpb.GetSubscriptionRequest{Id: order.RecurringId})

		if err != nil || subscriptionRsp.Status != billingpb.ResponseStatusOk {
			if err == nil {
				err = errors.New(subscriptionRsp.Message)
			}

			zap.L().Error(
				pkg.MethodFinishedWithError,
				zap.String("Method", "GetSubscription"),
				zap.Error(err),
				zap.String("orderId", order.Id),
				zap.String("subscriptionId", order.RecurringId),
			)
			return err
		}

		subscription = subscriptionRsp.Subscription

		if subscription.LastPaymentAt != nil {
			newOrder := new(billingpb.Order)
			err = copier.Copy(&newOrder, &order)

			if err != nil {
				zap.S().Error(
					"Copy order to new structure order by recurring subscription failed",
					zap.Error(err),
					zap.Any("order", order),
				)
				return err
			}

			newOrder.Id = primitive.NewObjectID().Hex()
			newOrder.Uuid = uuid.New().String()
			newOrder.ReceiptId = uuid.New().String()
			newOrder.CreatedAt = ptypes.TimestampNow()
			newOrder.UpdatedAt = ptypes.TimestampNow()
			newOrder.Canceled = false
			newOrder.CanceledAt = nil
			newOrder.ReceiptUrl = ""
			newOrder.RoyaltyReportId = ""
			newOrder.PrivateStatus = recurringpb.OrderStatusNew
			newOrder.ParentOrder = &billingpb.ParentOrder{
				Id:   order.Id,
				Uuid: order.Uuid,
			}

			if err = s.orderRepository.Insert(ctx, newOrder); err != nil {
				return err
			}

			newOrder.PrivateStatus = order.PrivateStatus
			order = newOrder
		}
	}

	err = s.updateOrder(ctx, order)

	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = pkg.StatusErrorSystem
			rsp.Error = e.Message
			return nil
		}
		return err
	}

	if pErr == nil {
		if h.IsSubscriptionCallback(data) && subscription != nil {
			if order.PrivateStatus != recurringpb.OrderStatusPaymentSystemComplete {
				err = h.DeleteRecurringSubscription(order, subscription)

				if err != nil {
					zap.L().Error(
						pkg.MethodFinishedWithError,
						zap.String("Method", "DeleteRecurringSubscription"),
						zap.Error(err),
						zap.String("orderId", order.Id),
						zap.String("subscriptionId", order.RecurringId),
						zap.String("planId", subscription.CardpaySubscriptionId),
					)
					return err
				}
			} else {
				t := time.Now().UTC()
				latestPayment, _ := ptypes.TimestampProto(time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location()))

				subscription.IsActive = true
				subscription.LastPaymentAt = latestPayment
				subscription.TotalAmount += subscription.Amount
				updateRsp, err := s.rep.UpdateSubscription(ctx, subscription)

				if err != nil || updateRsp.Status != billingpb.ResponseStatusOk {
					zap.L().Error(
						pkg.MethodFinishedWithError,
						zap.String("Method", "UpdateSubscription"),
						zap.Error(err),
						zap.String("orderId", order.Id),
						zap.String("subscriptionId", order.RecurringId),
						zap.String("planId", subscription.CardpaySubscriptionId),
						zap.Any("update_response", updateRsp),
					)

					return errors.New(subscriptionUpdateFailed)
				}
			}
		}

		if order.PrivateStatus == recurringpb.OrderStatusPaymentSystemComplete {
			err = s.paymentSystemPaymentCallbackComplete(ctx, order)

			if err != nil {
				rsp.Status = pkg.StatusErrorSystem
				rsp.Error = err.Error()
				return nil
			}
		}

		err = s.onPaymentNotify(ctx, order)

		if err != nil {
			zap.L().Error(
				pkg.MethodFinishedWithError,
				zap.String("Method", "onPaymentNotify"),
				zap.Error(err),
				zap.String("orderId", order.Id),
				zap.String("orderUuid", order.Uuid),
			)

			if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
				rsp.Status = pkg.StatusErrorSystem
				rsp.Error = e.Message
				return nil
			}
			return err
		}

		if order.PrivateStatus == recurringpb.OrderStatusPaymentSystemComplete {
			s.sendMailWithReceipt(ctx, order)
		}

		if h.CanSaveCard(data) {
			s.saveRecurringCard(ctx, order, h.GetRecurringId(data))
		}

		rsp.Status = pkg.StatusOK
	}

	return nil
}

func (s *Service) PaymentFormLanguageChanged(
	ctx context.Context,
	req *billingpb.PaymentFormUserChangeLangRequest,
	rsp *billingpb.PaymentFormDataChangeResponse,
) error {
	order, err := s.getOrderByUuidToForm(ctx, req.OrderId)

	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	rsp.Status = billingpb.ResponseStatusOk
	rsp.Item = order.GetPaymentFormDataChangeResult()

	if order.User.Locale == req.Lang {
		return nil
	}

	if helper.IsIdentified(order.User.Id) == true {
		s.updateCustomerFromRequestLocale(ctx, order, req.Ip, req.AcceptLanguage, req.UserAgent, req.Lang)
	}

	order.User.Locale = req.Lang

	if order.ProductType == pkg.OrderType_product {
		err = s.ProcessOrderProducts(ctx, order)
	} else if order.ProductType == pkg.OrderType_key {
		_, err = s.ProcessOrderKeyProducts(ctx, order)
	}

	if err != nil {
		if pid := order.PrivateMetadata["PaylinkId"]; pid != "" {
			s.notifyPaylinkError(ctx, pid, err, req, order)
		}
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	err = s.updateOrder(ctx, order)

	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusSystemError
			rsp.Message = e
			return nil
		}
		return err
	}

	rsp.Item = order.GetPaymentFormDataChangeResult()

	return nil
}

func (s *Service) PaymentFormPaymentAccountChanged(
	ctx context.Context,
	req *billingpb.PaymentFormUserChangePaymentAccountRequest,
	rsp *billingpb.PaymentFormDataChangeResponse,
) error {
	order, err := s.getOrderByUuidToForm(ctx, req.OrderId)

	if err != nil {
		rsp.Status = billingpb.ResponseStatusBadData
		rsp.Message = err.(*billingpb.ResponseErrorMessage)
		return nil
	}

	project, err := s.project.GetById(ctx, order.Project.Id)
	if err != nil {
		return orderErrorProjectNotFound
	}
	if project.IsDeleted() == true {
		return orderErrorProjectInactive
	}

	pm, err := s.paymentMethodRepository.GetById(ctx, req.MethodId)

	if err != nil {
		rsp.Status = billingpb.ResponseStatusBadData
		rsp.Message = orderErrorPaymentMethodNotFound
		return nil
	}

	ps, err := s.paymentSystemRepository.GetById(ctx, pm.PaymentSystemId)
	if err != nil {
		rsp.Message = orderErrorPaymentSystemInactive
		rsp.Status = billingpb.ResponseStatusBadData

		return nil
	}

	regex := pm.AccountRegexp

	if pm.ExternalId == recurringpb.PaymentSystemGroupAliasBankCard {
		regex = "^\\d{6}(.*)\\d{4}$"
	}

	match, err := regexp.MatchString(regex, req.Account)

	if match == false || err != nil {
		rsp.Status = billingpb.ResponseStatusBadData
		rsp.Message = orderErrorPaymentAccountIncorrect
		return nil
	}

	rsp.Status = billingpb.ResponseStatusOk

	switch pm.ExternalId {
	case recurringpb.PaymentSystemGroupAliasBankCard:
		data := s.getBinData(ctx, req.Account)

		if data == nil {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = orderErrorCountryByPaymentAccountNotFound
			return nil
		}

		if order.PaymentRequisites == nil {
			order.PaymentRequisites = make(map[string]string)
		}
		order.PaymentRequisites[billingpb.PaymentCreateBankCardFieldBrand] = data.CardBrand
		order.PaymentRequisites[billingpb.PaymentCreateBankCardFieldIssuerCountryIsoCode] = data.BankCountryIsoCode

		break

	case recurringpb.PaymentSystemGroupAliasQiwi:
		req.Account = "+" + req.Account
		num, err := libphonenumber.Parse(req.Account, CountryCodeUSA)

		if err != nil || num.CountryCode == nil {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = orderErrorPaymentAccountIncorrect
			return nil
		}

		_, ok := pkg.CountryPhoneCodes[*num.CountryCode]
		if !ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = orderErrorCountryByPaymentAccountNotFound
			return nil
		}
		break
	}

	order.PaymentMethod = &billingpb.PaymentMethodOrder{
		Id:              pm.Id,
		Name:            pm.Name,
		PaymentSystemId: ps.Id,
		Group:           pm.Group,
		ExternalId:      pm.ExternalId,
		Handler:         ps.Handler,
		RefundAllowed:   pm.RefundAllowed,
	}

	err = s.setOrderChargeAmountAndCurrency(ctx, order)
	if err != nil {
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	methodName, err := order.GetCostPaymentMethodName()
	if err != nil {
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}

		return err
	}

	order.PaymentMethod.Params, err = s.getPaymentSettings(
		pm,
		order.ChargeCurrency,
		order.MccCode,
		order.OperatingCompanyId,
		methodName,
		order.IsProduction,
	)

	if err != nil {
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}

		return err
	}

	if !s.hasPaymentCosts(ctx, order) {
		rsp.Status = billingpb.ResponseStatusBadData
		rsp.Message = orderErrorCostsRatesNotFound
		return nil
	}

	err = s.updateOrder(ctx, order)

	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusSystemError
			rsp.Message = e
			return nil
		}
		return err
	}

	rsp.Item = order.GetPaymentFormDataChangeResult()
	return nil
}

func (s *Service) ProcessBillingAddress(
	ctx context.Context,
	req *billingpb.ProcessBillingAddressRequest,
	rsp *billingpb.ProcessBillingAddressResponse,
) error {
	var err error
	var zip *billingpb.ZipCode

	order, err := s.getOrderByUuidToForm(ctx, req.OrderId)

	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	initialCountry := order.GetCountry()

	billingAddress := &billingpb.OrderBillingAddress{
		Country: req.Country,
	}

	if req.Country == CountryCodeUSA && req.Zip != "" {
		billingAddress.PostalCode = req.Zip

		zip, err = s.zipCodeRepository.GetByZipAndCountry(ctx, req.Zip, req.Country)

		if err == nil && zip != nil {
			billingAddress.Country = zip.Country
			billingAddress.PostalCode = zip.Zip
			billingAddress.City = zip.City
			billingAddress.State = zip.State.Code
		}
	}

	if !order.CountryChangeAllowed() && initialCountry != billingAddress.Country {
		rsp.Status = billingpb.ResponseStatusForbidden
		rsp.Message = orderCountryChangeRestrictedError
		return nil
	}

	order.BillingAddress = billingAddress

	restricted, err := s.applyCountryRestriction(ctx, order, billingAddress.Country)
	if err != nil {
		zap.L().Error(
			"s.applyCountryRestriction Method failed",
			zap.Error(err),
			zap.Any("order", order),
		)
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusSystemError
			rsp.Message = e
			return nil
		} else {
			rsp.Message = orderErrorUnknown
			rsp.Status = billingpb.ResponseStatusSystemError
		}
		return nil
	}
	if restricted {
		rsp.Status = billingpb.ResponseStatusForbidden
		rsp.Message = orderCountryPaymentRestrictedError
		return nil
	}

	// save user replace country rule to cookie - start
	cookie := ""
	customer := &BrowserCookieCustomer{
		CreatedAt: time.Now(),
	}
	if req.Cookie != "" {
		customer, err = s.decryptBrowserCookie(req.Cookie)
		if err != nil || customer == nil {
			customer = &BrowserCookieCustomer{
				CreatedAt: time.Now(),
			}
		}
	}

	address, err := s.getAddressByIp(ctx, req.Ip)
	if err == nil {
		customer.Ip = req.Ip
		customer.IpCountry = address.Country
		customer.SelectedCountry = billingAddress.Country
		customer.UpdatedAt = time.Now()

		cookie, err = s.generateBrowserCookie(customer)
		if err != nil {
			cookie = ""
		}
	}
	// save user replace country rule to cookie - end

	if order.ProductType == pkg.OrderType_product {
		err = s.ProcessOrderProducts(ctx, order)
	} else if order.ProductType == pkg.OrderType_key {
		_, err = s.ProcessOrderKeyProducts(ctx, order)
	}

	if err != nil {
		if pid := order.PrivateMetadata["PaylinkId"]; pid != "" {
			s.notifyPaylinkError(ctx, pid, err, req, order)
		}
		return err
	}

	processor := &OrderCreateRequestProcessor{Service: s, ctx: ctx}
	err = processor.processOrderVat(order)
	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error(), "method", "processOrderVat")
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusSystemError
			rsp.Message = e
			return nil
		}
		return err
	}

	err = s.setOrderChargeAmountAndCurrency(ctx, order)
	if err != nil {
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	methodName, _ := order.GetCostPaymentMethodName()
	if methodName != "" && !s.hasPaymentCosts(ctx, order) {
		rsp.Status = billingpb.ResponseStatusBadData
		rsp.Message = orderErrorCostsRatesNotFound
		return nil
	}

	order.BillingCountryChangedByUser = order.BillingCountryChangedByUser == true || initialCountry != order.GetCountry()

	err = s.updateOrder(ctx, order)

	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusSystemError
			rsp.Message = e
			return nil
		}
		return err
	}

	rsp.Status = billingpb.ResponseStatusOk
	rsp.Cookie = cookie
	rsp.Item = &billingpb.ProcessBillingAddressResponseItem{
		HasVat:               order.Tax.Rate > 0,
		VatRate:              tools.ToPrecise(order.Tax.Rate),
		Vat:                  order.Tax.Amount,
		VatInChargeCurrency:  s.FormatAmount(order.GetTaxAmountInChargeCurrency(), order.Currency),
		Amount:               order.OrderAmount,
		TotalAmount:          order.TotalPaymentAmount,
		Currency:             order.Currency,
		ChargeCurrency:       order.ChargeCurrency,
		ChargeAmount:         order.ChargeAmount,
		Items:                order.Items,
		CountryChangeAllowed: order.CountryChangeAllowed(),
	}

	return nil
}

func (s *Service) saveRecurringCard(ctx context.Context, order *billingpb.Order, recurringId string) {
	req := &recurringpb.SavedCardRequest{
		Token:      order.User.Id,
		ProjectId:  order.Project.Id,
		MerchantId: order.Project.MerchantId,
		MaskedPan:  order.PaymentMethodTxnParams[billingpb.PaymentCreateFieldPan],
		CardHolder: order.PaymentMethodTxnParams[billingpb.PaymentCreateFieldHolder],
		Expire: &recurringpb.CardExpire{
			Month: order.PaymentRequisites[billingpb.PaymentCreateFieldMonth],
			Year:  order.PaymentRequisites[billingpb.PaymentCreateFieldYear],
		},
		RecurringId: recurringId,
	}

	_, err := s.rep.InsertSavedCard(ctx, req)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(errorFieldService, recurringpb.PayOneRepositoryServiceName),
			zap.String(errorFieldMethod, "InsertSavedCard"),
		)
	} else {
		order, err := s.orderRepository.GetById(ctx, order.Id)

		if err != nil {
			zap.L().Error(
				"Failed to refresh order data",
				zap.Error(err),
				zap.String("uuid", order.Uuid),
			)
			return
		}

		order.PaymentRequisites["saved"] = "1"
		err = s.updateOrder(ctx, order)

		if err != nil {
			zap.L().Error(
				"Failed to update order after save recurruing card",
				zap.Error(err),
			)
		}
	}
}

func (s *Service) updateOrder(ctx context.Context, order *billingpb.Order) error {
	ps := order.GetPublicStatus()

	zap.S().Debug("[updateOrder] updating order", "order_id", order.Id, "status", ps)

	originalOrder, _ := s.getOrderById(ctx, order.Id)

	statusChanged := false
	if originalOrder != nil {
		ops := originalOrder.GetPublicStatus()
		zap.S().Debug("[updateOrder] no original order status", "order_id", order.Id, "status", ops)
		statusChanged = ops != ps
	} else {
		zap.S().Debug("[updateOrder] no original order found", "order_id", order.Id)
	}

	needReceipt := statusChanged && (ps == recurringpb.OrderPublicStatusChargeback || ps == recurringpb.OrderPublicStatusRefunded || ps == recurringpb.OrderPublicStatusProcessed)

	if needReceipt {
		switch order.Type {
		case pkg.OrderTypeRefund:
			order.ReceiptUrl = s.cfg.GetReceiptRefundUrl(order.Uuid, order.ReceiptId)
		case pkg.OrderTypeOrder:
			order.ReceiptUrl = s.cfg.GetReceiptPurchaseUrl(order.Uuid, order.ReceiptId)
		}
	}

	if err := s.orderRepository.Update(ctx, order); err != nil {
		if err == mongo.ErrNoDocuments {
			return orderErrorNotFound
		}
		return orderErrorUnknown
	}

	zap.S().Debug("[updateOrder] updating order success", "order_id", order.Id, "status_changed", statusChanged, "type", order.ProductType)

	if order.ProductType == pkg.OrderType_key {
		s.orderNotifyKeyProducts(ctx, order)
	}

	if statusChanged && order.NeedCallbackNotification() {
		s.orderNotifyMerchant(ctx, order)
	}

	return nil
}

func (s *Service) orderNotifyKeyProducts(ctx context.Context, order *billingpb.Order) {
	zap.S().Debug("[orderNotifyKeyProducts] called", "order_id", order.Id, "status", order.GetPublicStatus(), "is product notified: ", order.IsKeyProductNotified)

	if order.IsKeyProductNotified {
		return
	}

	keys := order.Keys
	var err error
	switch order.GetPublicStatus() {
	case recurringpb.OrderPublicStatusCanceled, recurringpb.OrderPublicStatusRejected:
		for _, key := range keys {
			zap.S().Infow("[orderNotifyKeyProducts] trying to cancel reserving key", "order_id", order.Id, "key", key)
			rsp := &billingpb.EmptyResponseWithStatus{}
			err = s.CancelRedeemKeyForOrder(ctx, &billingpb.KeyForOrderRequest{KeyId: key}, rsp)
			if err != nil {
				zap.S().Error("internal error during canceling reservation for key", "err", err, "key", key)
				continue
			}
			if rsp.Status != billingpb.ResponseStatusOk {
				zap.S().Error("could not cancel reservation for key", "key", key, "message", rsp.Message)
				continue
			}
		}
		order.IsKeyProductNotified = true
		break
	case recurringpb.OrderPublicStatusProcessed:
		for _, key := range keys {
			zap.S().Infow("[orderNotifyKeyProducts] trying to finish reserving key", "order_id", order.Id, "key", key)
			rsp := &billingpb.GetKeyForOrderRequestResponse{}
			err = s.FinishRedeemKeyForOrder(ctx, &billingpb.KeyForOrderRequest{KeyId: key}, rsp)
			if err != nil {
				zap.S().Errorw("internal error during finishing reservation for key", "err", err, "key", key)
				continue
			}
			if rsp.Status != billingpb.ResponseStatusOk {
				zap.S().Errorw("could not finish reservation for key", "key", key, "message", rsp.Message)
				continue
			}

			s.sendMailWithCode(ctx, order, rsp.Key)
		}
		order.IsKeyProductNotified = true
		break
	}
}

func (s *Service) sendMailWithReceipt(ctx context.Context, order *billingpb.Order) {
	payload, err := s.getPayloadForReceipt(ctx, order)
	if err != nil {
		zap.L().Error("get order receipt object failed", zap.Error(err))
		return
	}

	zap.S().Infow("sending receipt to broker", "order_id", order.Id, "topic", postmarkpb.PostmarkSenderTopicName)
	err = s.postmarkBroker.Publish(postmarkpb.PostmarkSenderTopicName, payload, amqp.Table{})
	if err != nil {
		zap.S().Errorw(
			"Publication receipt to user email queue is failed",
			"err", err, "email", order.ReceiptEmail, "order_id", order.Id, "topic", postmarkpb.PostmarkSenderTopicName)
	}
}

func (s *Service) getPayloadForReceipt(ctx context.Context, order *billingpb.Order) (*postmarkpb.Payload, error) {
	template := s.cfg.EmailTemplates.SuccessTransaction
	if order.Type == pkg.OrderTypeRefund {
		template = s.cfg.EmailTemplates.RefundTransaction
	}

	receipt, err := s.getOrderReceiptObject(ctx, order)
	if err != nil {
		return nil, err
	}

	var items []*structpb.Value
	if receipt.Items != nil && len(receipt.Items) > 0 {
		for _, item := range receipt.Items {
			item := &structpb.Value{
				Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"name": {
								Kind: &structpb.Value_StringValue{StringValue: item.Name},
							},
							"is-simple": {
								Kind: &structpb.Value_BoolValue{BoolValue: false},
							},
							"price": {
								Kind: &structpb.Value_StringValue{StringValue: item.Price},
							},
						},
					},
				},
			}

			items = append(items, item)
		}
	} else {
		item := &structpb.Value{
			Kind: &structpb.Value_StructValue{
				StructValue: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"name": {
							Kind: &structpb.Value_StringValue{StringValue: "In-game purchase"},
						},
						"is-simple": {
							Kind: &structpb.Value_BoolValue{BoolValue: true},
						},
						"price": {
							Kind: &structpb.Value_StringValue{StringValue: receipt.TotalPrice},
						},
					},
				},
			},
		}

		items = append(items, item)
	}

	// set receipt items to nil to omit this field in jsonpb Marshal result
	// otherwise we get an Unmarshal error on attempt to marshal array to string
	receipt.Items = nil

	march := &jsonpb.Marshaler{}
	var buf bytes.Buffer
	err = march.Marshal(&buf, receipt)
	if err != nil {
		return nil, err
	}

	templateModel := make(map[string]string)
	err = json.Unmarshal(buf.Bytes(), &templateModel)
	if err != nil {
		return nil, err
	}

	// removing empty "platform" key (if any), for easy email template condition
	if v, ok := templateModel["platform"]; ok && v == "" {
		delete(templateModel, "platform")
	}

	templateModel["current_year"] = time.Now().UTC().Format("2006")

	// pass subtotal row info as struct, for email template condition
	var subTotal *structpb.Value
	if receipt.TotalAmount != receipt.TotalCharge {
		subTotal = &structpb.Value{
			Kind: &structpb.Value_StructValue{
				StructValue: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"totalAmount": {
							Kind: &structpb.Value_StringValue{StringValue: receipt.TotalAmount},
						},
						"totalCharge": {
							Kind: &structpb.Value_StringValue{StringValue: receipt.TotalCharge},
						},
					},
				},
			},
		}
	}

	// pass vat row info as struct, for email template condition
	var vat *structpb.Value
	if receipt.VatRate != "0%" {
		vat = &structpb.Value{
			Kind: &structpb.Value_StructValue{
				StructValue: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"rate": {
							Kind: &structpb.Value_StringValue{StringValue: receipt.VatRate},
						},
						"amount": {
							Kind: &structpb.Value_StringValue{StringValue: receipt.VatInOrderCurrency},
						},
						"including": {
							Kind: &structpb.Value_BoolValue{BoolValue: receipt.VatPayer == billingpb.VatPayerSeller},
						},
					},
				},
			},
		}
	}

	payload := &postmarkpb.Payload{
		TemplateAlias: template,
		TemplateModel: templateModel,
		To:            order.ReceiptEmail,
	}

	fields := map[string]*structpb.Value{
		"items": {
			Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{
				Values: items,
			}},
		},
		"showSummary": {
			Kind: &structpb.Value_BoolValue{BoolValue: receipt.Items != nil && len(receipt.Items) > 1},
		},
	}

	if subTotal != nil {
		fields["subTotal"] = subTotal
	}

	if vat != nil {
		fields["vat"] = vat
	}

	payload.TemplateObjectModel = &structpb.Struct{
		Fields: fields,
	}

	return payload, nil
}

func (s *Service) sendMailWithCode(_ context.Context, order *billingpb.Order, key *billingpb.Key) {
	platformIconUrl := ""
	activationInstructionUrl := ""
	platformName := ""

	if platform, ok := availablePlatforms[order.PlatformId]; ok {
		platformIconUrl = platform.Icon
		activationInstructionUrl = platform.ActivationInstructionUrl
		platformName = platform.Name
	}

	for _, item := range order.Items {
		if item.Id == key.KeyProductId {
			item.Code = key.Code
			payload := &postmarkpb.Payload{
				TemplateAlias: s.cfg.EmailTemplates.ActivationGameKey,
				TemplateModel: map[string]string{
					"code":                       key.Code,
					"platform_icon":              platformIconUrl,
					"product_name":               item.Name,
					"activation_instruction_url": activationInstructionUrl,
					"platform_name":              platformName,
					"receipt_url":                order.ReceiptUrl,
					"current_year":               time.Now().UTC().Format("2006"),
				},
				To: order.ReceiptEmail,
			}

			if len(item.Images) > 0 {
				payload.TemplateModel["product_image"] = item.Images[0]
			}

			err := s.postmarkBroker.Publish(postmarkpb.PostmarkSenderTopicName, payload, amqp.Table{})
			if err != nil {
				zap.S().Errorw(
					"Publication activation code to user email queue is failed",
					"err", err, "email", order.ReceiptEmail, "order_id", order.Id, "key_id", key.Id)

			} else {
				zap.S().Infow("Sent payload to broker", "email", order.ReceiptEmail, "order_id", order.Id, "key_id", key.Id, "topic", postmarkpb.PostmarkSenderTopicName)
			}
			return
		}
	}

	zap.S().Errorw("Mail not sent because no items found for key", "order_id", order.Id, "key_id", key.Id, "email", order.ReceiptEmail)
}

func (s *Service) orderNotifyMerchant(ctx context.Context, order *billingpb.Order) {
	zap.S().Debug("[orderNotifyMerchant] try to send notify merchant to rmq", "order_id", order.Id, "status", order.GetPublicStatus())

	err := s.broker.Publish(recurringpb.PayOneTopicNotifyPaymentName, order, amqp.Table{"x-retry-count": int32(0)})
	if err != nil {
		zap.S().Debug("[orderNotifyMerchant] send notify merchant to rmq failed", "order_id", order.Id)
		s.logError(orderErrorPublishNotificationFailed, []interface{}{
			"err", err.Error(), "order", order, "topic", recurringpb.PayOneTopicNotifyPaymentName,
		})
	} else {
		zap.S().Debug("[orderNotifyMerchant] send notify merchant to rmq failed", "order_id", order.Id)
	}
	order.SetNotificationStatus(order.GetPublicStatus(), err == nil)

	if err = s.orderRepository.Update(ctx, order); err != nil {
		zap.S().Debug("[orderNotifyMerchant] notification status update failed", "order_id", order.Id)
		s.logError(orderErrorUpdateOrderDataFailed, []interface{}{"error", err.Error(), "order", order})
	} else {
		zap.S().Debug("[orderNotifyMerchant] notification status updated successfully", "order_id", order.Id)
	}
}

func (s *Service) getOrderById(ctx context.Context, id string) (order *billingpb.Order, err error) {
	order, err = s.orderRepository.GetById(ctx, id)

	if order == nil {
		return order, orderErrorNotFound
	}

	return
}

func (s *Service) getOrderByUuid(ctx context.Context, uuid string) (order *billingpb.Order, err error) {
	order, err = s.orderRepository.GetByUuid(ctx, uuid)

	if err != nil && err != mongo.ErrNoDocuments {
		zap.S().Errorf("Order not found in payment create process", "err", err.Error(), "uuid", uuid)
	}

	if order == nil {
		return order, orderErrorNotFound
	}

	return
}

func (s *Service) getOrderByUuidToForm(ctx context.Context, uuid string) (*billingpb.Order, error) {
	order, err := s.getOrderByUuid(ctx, uuid)

	if err != nil {
		return nil, orderErrorNotFound
	}

	if order.HasEndedStatus() == true {
		return nil, orderErrorOrderAlreadyComplete
	}

	if order.FormInputTimeIsEnded() == true {
		return nil, orderErrorFormInputTimeExpired
	}

	return order, nil
}

func (s *Service) getBinData(ctx context.Context, pan string) (data *intPkg.BinData) {
	if len(pan) < 6 {
		zap.S().Errorf("Incorrect PAN to get BIN data", "pan", pan)
		return
	}

	i, err := strconv.ParseInt(pan[:6], 10, 32)

	if err != nil {
		zap.S().Errorf("Parse PAN to int failed", "error", err.Error(), "pan", pan)
		return
	}

	data, err = s.bankBinRepository.GetByBin(ctx, int32(i))

	if err != nil {
		zap.S().Errorf("Query to get bank card BIN data failed", "error", err.Error(), "pan", pan)
		return
	}

	return
}

func (v *OrderCreateRequestProcessor) prepareOrder() (*billingpb.Order, error) {
	id := primitive.NewObjectID().Hex()
	amount := v.FormatAmount(v.checked.amount, v.checked.currency)

	if (v.request.UrlVerify != "" || v.request.UrlNotify != "") && v.checked.project.AllowDynamicNotifyUrls == false {
		return nil, orderErrorDynamicNotifyUrlsNotAllowed
	}

	if (v.request.UrlSuccess != "" || v.request.UrlFail != "") && v.checked.project.AllowDynamicRedirectUrls == false {
		return nil, orderErrorDynamicRedirectUrlsNotAllowed
	}

	order := &billingpb.Order{
		Id:   id,
		Type: pkg.OrderTypeOrder,
		Project: &billingpb.ProjectOrder{
			Id:                      v.checked.project.Id,
			Name:                    v.checked.project.Name,
			UrlSuccess:              v.checked.project.UrlRedirectSuccess,
			UrlFail:                 v.checked.project.UrlRedirectFail,
			SendNotifyEmail:         v.checked.project.SendNotifyEmail,
			NotifyEmails:            v.checked.project.NotifyEmails,
			SecretKey:               v.checked.project.SecretKey,
			UrlCheckAccount:         v.checked.project.UrlCheckAccount,
			UrlProcessPayment:       v.checked.project.UrlProcessPayment,
			UrlChargebackPayment:    v.checked.project.UrlChargebackPayment,
			UrlCancelPayment:        v.checked.project.UrlCancelPayment,
			UrlRefundPayment:        v.checked.project.UrlRefundPayment,
			UrlFraudPayment:         v.checked.project.UrlFraudPayment,
			CallbackProtocol:        v.checked.project.CallbackProtocol,
			MerchantId:              v.checked.merchant.Id,
			Status:                  v.checked.project.Status,
			MerchantRoyaltyCurrency: v.checked.merchant.GetPayoutCurrency(),
			RedirectSettings:        v.checked.project.RedirectSettings,
			FirstPaymentAt:          v.checked.merchant.FirstPaymentAt,
		},
		Description:   fmt.Sprintf(pkg.OrderDefaultDescription, id),
		PrivateStatus: recurringpb.OrderStatusNew,
		CreatedAt:     ptypes.TimestampNow(),
		IsJsonRequest: v.request.IsJson,

		Uuid:               uuid.New().String(),
		ReceiptId:          uuid.New().String(),
		User:               v.checked.user,
		OrderAmount:        amount,
		TotalPaymentAmount: amount,
		ChargeAmount:       amount,
		Currency:           v.checked.currency,
		ChargeCurrency:     v.checked.currency,
		Products:           v.checked.products,
		Items:              v.checked.items,
		Metadata:           v.checked.metadata,
		PrivateMetadata:    v.checked.privateMetadata,
		Issuer: &billingpb.OrderIssuer{
			Url:           v.request.IssuerUrl,
			Embedded:      v.request.IsEmbedded,
			ReferenceType: v.request.IssuerReferenceType,
			Reference:     v.request.IssuerReference,
			UtmSource:     v.request.UtmSource,
			UtmCampaign:   v.request.UtmCampaign,
			UtmMedium:     v.request.UtmMedium,
			ReferrerHost:  getHostFromUrl(v.request.IssuerUrl),
		},
		CountryRestriction: &billingpb.CountryRestriction{
			IsoCodeA2:       "",
			PaymentsAllowed: true,
			ChangeAllowed:   true,
		},
		PlatformId:              v.request.PlatformId,
		ProductType:             v.request.Type,
		IsBuyForVirtualCurrency: v.checked.isBuyForVirtualCurrency,
		MccCode:                 v.checked.merchant.MccCode,
		OperatingCompanyId:      v.checked.merchant.OperatingCompanyId,
		IsHighRisk:              v.checked.merchant.IsHighRisk(),
		TestingCase:             v.request.TestingCase,
		IsCurrencyPredefined:    v.checked.isCurrencyPredefined,
		VatPayer:                v.checked.project.VatPayer,
		IsProduction:            v.checked.project.IsProduction(),
		MerchantInfo: &billingpb.OrderViewMerchantInfo{
			CompanyName:     v.checked.merchant.GetCompanyName(),
			AgreementNumber: v.checked.merchant.AgreementNumber,
		},
	}

	if v.checked.virtualAmount > 0 {
		order.VirtualCurrencyAmount = v.checked.virtualAmount
	}

	if order.User == nil {
		order.User = &billingpb.OrderUser{
			Object:     pkg.ObjectTypeUser,
			ExternalId: v.request.Account,
		}
	} else {
		if order.User.Address != nil {
			err := v.processOrderVat(order)
			if err != nil {
				zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error(), "method", "processOrderVat")
				return nil, err
			}

			restricted, err := v.applyCountryRestriction(v.ctx, order, order.GetCountry())
			if err != nil {
				return nil, err
			}
			if restricted {
				return nil, orderCountryPaymentRestrictedError
			}
		}
	}

	if v.request.Description != "" {
		order.Description = v.request.Description
	}

	if v.request.UrlSuccess != "" {
		order.Project.UrlSuccess = v.request.UrlSuccess
	}

	if v.request.UrlFail != "" {
		order.Project.UrlFail = v.request.UrlFail
	}

	if v.checked.paymentMethod != nil {
		ps, err := v.paymentSystemRepository.GetById(v.ctx, v.checked.paymentMethod.PaymentSystemId)
		if err != nil {
			return nil, err
		}

		order.PaymentMethod = &billingpb.PaymentMethodOrder{
			Id:               v.checked.paymentMethod.Id,
			Name:             v.checked.paymentMethod.Name,
			PaymentSystemId:  ps.Id,
			Group:            v.checked.paymentMethod.Group,
			RecurringAllowed: v.checked.paymentMethod.RecurringAllowed,
		}

		methodName, err := order.GetCostPaymentMethodName()
		if err == nil {
			order.PaymentMethod.Params, err = v.getPaymentSettings(
				v.checked.paymentMethod,
				v.checked.currency,
				v.checked.mccCode,
				v.checked.operatingCompanyId,
				methodName,
				order.IsProduction,
			)

			if err != nil {
				return nil, err
			}
		}
	}

	order.FormMode = defaultPaymentFormOpeningMode

	if _, ok := possiblePaymentFormOpeningModes[v.request.FormMode]; ok {
		order.FormMode = v.request.FormMode
	}

	order.ExpireDateToFormInput, _ = ptypes.TimestampProto(time.Now().Add(time.Minute * defaultExpireDateToFormInput))

	if v.checked.recurringPeriod != "" {
		order.RecurringSettings = &billingpb.OrderRecurringSettings{
			Period:   v.checked.recurringPeriod,
			Interval: v.checked.recurringInterval,
			DateEnd:  v.checked.recurringDateEnd,
		}
	}

	return order, nil
}

func (v *OrderCreateRequestProcessor) processMerchant() error {
	if v.checked.merchant.Banking == nil || v.checked.merchant.Banking.Currency == "" {
		return orderErrorMerchantDoNotHaveBanking
	}

	if v.checked.merchant.HasTariff() == false {
		return orderErrorMerchantBadTariffs
	}

	return nil
}

func (v *OrderCreateRequestProcessor) processProject() error {
	project, err := v.project.GetById(v.ctx, v.request.ProjectId)

	if err != nil {
		zap.S().Errorw("Order create get project error", "err", err, "request", v.request)
		return orderErrorProjectNotFound
	}

	if project.IsDeleted() == true {
		return orderErrorProjectInactive
	}

	if project.MerchantId == "" {
		return orderErrorProjectMerchantNotFound
	}

	_, err = primitive.ObjectIDFromHex(project.MerchantId)

	if err != nil {
		return orderErrorProjectMerchantNotFound
	}

	merchant, err := v.merchantRepository.GetById(v.ctx, project.MerchantId)
	if err != nil {
		return orderErrorProjectMerchantNotFound
	}

	if merchant.IsDeleted() == true {
		return orderErrorProjectMerchantInactive
	}

	v.checked.project = project
	v.checked.merchant = merchant

	if v.request.ButtonCaption != "" {
		v.checked.setRedirectButtonCaption(v.request.ButtonCaption)
	}

	return nil
}

func (v *OrderCreateRequestProcessor) processCurrency(orderType string) error {
	if v.request.Currency != "" {
		if !helper.Contains(v.supportedCurrencies, v.request.Currency) {
			return orderErrorCurrencyNotFound
		}

		v.checked.currency = v.request.Currency
		v.checked.isCurrencyPredefined = true

		pricegroup, err := v.priceGroupRepository.GetByRegion(v.ctx, v.checked.currency)
		if err == nil {
			v.checked.priceGroup = pricegroup
		}
		return nil
	}

	if orderType == pkg.OrderType_simple {
		return orderErrorCurrencyIsRequired
	}

	v.checked.isCurrencyPredefined = false

	countryCode := v.getCountry()
	if countryCode == "" {
		v.checked.currency = v.checked.merchant.GetProcessingDefaultCurrency()
		pricegroup, err := v.priceGroupRepository.GetByRegion(v.ctx, v.checked.currency)
		if err == nil {
			v.checked.priceGroup = pricegroup
		}
		return nil
	}

	country, err := v.country.GetByIsoCodeA2(v.ctx, countryCode)
	if err != nil {
		v.checked.currency = v.checked.merchant.GetProcessingDefaultCurrency()
		pricegroup, err := v.priceGroupRepository.GetByRegion(v.ctx, v.checked.currency)
		if err == nil {
			v.checked.priceGroup = pricegroup
		}
		return nil
	}

	pricegroup, err := v.priceGroupRepository.GetById(v.ctx, country.PriceGroupId)
	if err != nil {
		v.checked.currency = v.checked.merchant.GetProcessingDefaultCurrency()
		pricegroup, err := v.priceGroupRepository.GetByRegion(v.ctx, v.checked.currency)
		if err == nil {
			v.checked.priceGroup = pricegroup
		}
		return nil
	}

	v.checked.currency = pricegroup.Currency
	v.checked.priceGroup = pricegroup

	return nil
}

func (v *OrderCreateRequestProcessor) processAmount() {
	v.checked.amount = v.request.Amount
}

func (v *OrderCreateRequestProcessor) processMetadata() {
	v.checked.metadata = v.request.Metadata
}

func (v *OrderCreateRequestProcessor) processPrivateMetadata() {
	v.checked.privateMetadata = v.request.PrivateMetadata
}

func (v *OrderCreateRequestProcessor) getCountry() string {
	if v.checked.user == nil {
		return ""
	}
	return v.checked.user.GetCountry()
}

func (v *OrderCreateRequestProcessor) processPayerIp(ctx context.Context) error {
	address, err := v.getAddressByIp(ctx, v.checked.user.Ip)

	if err != nil {
		return err
	}

	// fully replace address, to avoid inconsistency
	v.checked.user.Address = address

	return nil
}

func (v *OrderCreateRequestProcessor) processPaylinkKeyProducts() error {
	amount, priceGroup, items, _, err := v.processKeyProducts(
		v.ctx,
		v.checked.project.Id,
		v.request.Products,
		v.checked.priceGroup,
		DefaultLanguage,
		v.request.PlatformId,
	)

	if err != nil {
		return err
	}

	if len(v.request.PlatformId) == 0 && len(items) > 0 {
		v.request.PlatformId = items[0].PlatformId
	}

	v.checked.priceGroup = priceGroup
	v.checked.products = v.request.Products
	v.checked.currency = priceGroup.Currency
	v.checked.amount = amount
	v.checked.items = items

	return nil
}

func (v *OrderCreateRequestProcessor) processPaylinkProducts(_ context.Context) error {
	amount, priceGroup, items, isBuyForVirtual, err := v.processProducts(
		v.ctx,
		v.checked.project.Id,
		v.request.Products,
		v.checked.priceGroup,
		DefaultLanguage,
	)

	v.checked.isBuyForVirtualCurrency = isBuyForVirtual

	if err != nil {
		return err
	}

	v.checked.priceGroup = priceGroup
	v.checked.products = v.request.Products
	v.checked.currency = priceGroup.Currency
	v.checked.amount = amount
	v.checked.items = items

	return nil
}

func (v *OrderCreateRequestProcessor) processPaymentMethod(pm *billingpb.PaymentMethod) error {
	if pm.IsActive == false {
		return orderErrorPaymentMethodInactive
	}

	if _, err := v.paymentSystemRepository.GetById(v.ctx, pm.PaymentSystemId); err != nil {
		return orderErrorPaymentSystemInactive
	}

	_, err := v.Service.getPaymentSettings(pm, v.checked.currency, v.checked.mccCode, v.checked.operatingCompanyId, "", v.checked.project.IsProduction())

	if err != nil {
		return err
	}

	v.checked.paymentMethod = pm

	return nil
}

func (v *OrderCreateRequestProcessor) processLimitAmounts() (err error) {
	amount := v.checked.amount

	pmls, err := v.paymentMinLimitSystemRepository.GetByCurrency(v.ctx, v.checked.currency)
	if err != nil {
		return errorPaymentMinLimitSystemNotFound
	}

	if amount < pmls.Amount {
		return orderErrorAmountLowerThanMinLimitSystem
	}

	if v.checked.project.LimitsCurrency != "" && v.checked.project.LimitsCurrency != v.checked.currency {
		if !helper.Contains(v.supportedCurrencies, v.checked.project.LimitsCurrency) {
			return orderErrorCurrencyNotFound
		}
		req := &currenciespb.ExchangeCurrencyCurrentForMerchantRequest{
			From:              v.checked.currency,
			To:                v.checked.project.LimitsCurrency,
			MerchantId:        v.checked.merchant.Id,
			RateType:          currenciespb.RateTypeOxr,
			ExchangeDirection: currenciespb.ExchangeDirectionSell,
			Amount:            amount,
		}

		rsp, err := v.curService.ExchangeCurrencyCurrentForMerchant(v.ctx, req)

		if err != nil {
			zap.S().Error(
				pkg.ErrorGrpcServiceCallFailed,
				zap.Error(err),
				zap.String(errorFieldService, "CurrencyRatesService"),
				zap.String(errorFieldMethod, "ExchangeCurrencyCurrentForMerchant"),
			)

			return orderErrorConvertionCurrency
		}

		amount = rsp.ExchangedAmount
	}

	if amount < v.checked.project.MinPaymentAmount {
		return orderErrorAmountLowerThanMinAllowed
	}

	if v.checked.project.MaxPaymentAmount > 0 && amount > v.checked.project.MaxPaymentAmount {
		return orderErrorAmountGreaterThanMaxAllowed
	}

	if v.checked.paymentMethod != nil {
		if v.request.Amount < v.checked.paymentMethod.MinPaymentAmount {
			return orderErrorAmountLowerThanMinAllowedPaymentMethod
		}

		if v.checked.paymentMethod.MaxPaymentAmount > 0 && v.request.Amount > v.checked.paymentMethod.MaxPaymentAmount {
			return orderErrorAmountGreaterThanMaxAllowedPaymentMethod
		}
	}

	return
}

func (v *OrderCreateRequestProcessor) processRecurringSettings() (err error) {
	if v.request.RecurringPeriod == "" {
		return nil
	}

	if v.checked.paymentMethod != nil && !v.checked.paymentMethod.RecurringAllowed {
		return orderErrorRecurringNotAllowed
	}

	currentTime := time.Now().UTC()
	dateEnd := currentTime.AddDate(1, 0, 0)

	if v.request.RecurringDateEnd != "" {
		inputDateEnd, err := time.Parse(billingpb.FilterDateFormat, v.request.RecurringDateEnd)

		if err != nil {
			return orderErrorRecurringDateEndInvalid
		}

		dateEnd = inputDateEnd.UTC()
	}

	dateEnd = time.Date(dateEnd.Year(), dateEnd.Month(), dateEnd.Day(), 23, 59, 59, 0, dateEnd.Location())
	currentTime = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
	delta := dateEnd.Sub(currentTime).Hours()

	var interval float64

	switch v.request.RecurringPeriod {
	case recurringpb.RecurringPeriodDay:
		interval = delta / 24
		break
	case recurringpb.RecurringPeriodWeek:
		interval = delta / 24 / 7
		break
	case recurringpb.RecurringPeriodMonth:
		interval = (delta / 24 / 365) * 12
		break
	case recurringpb.RecurringPeriodYear:
		interval = delta / 24 / 365
		break
	default:
		return orderErrorRecurringInvalidPeriod
	}

	interval = math.Floor(interval)
	totalDays := math.Floor(delta / 24)

	if interval < 1 || totalDays > 365 ||
		(v.request.RecurringPeriod == recurringpb.RecurringPeriodDay && interval > 365) ||
		(v.request.RecurringPeriod == recurringpb.RecurringPeriodDay && interval < 7) ||
		(v.request.RecurringPeriod == recurringpb.RecurringPeriodWeek && interval > 52) ||
		(v.request.RecurringPeriod == recurringpb.RecurringPeriodMonth && interval > 12) ||
		(v.request.RecurringPeriod == recurringpb.RecurringPeriodYear && interval > 1) {
		return orderErrorRecurringDateEndOutOfRange
	}

	v.checked.recurringPeriod = v.request.RecurringPeriod
	v.checked.recurringInterval = int32(1)
	v.checked.recurringDateEnd = dateEnd.Format(billingpb.FilterDateFormat)

	return
}

func (v *OrderCreateRequestProcessor) processSignature() error {
	var hashString string

	if v.request.IsJson == false {
		var keys []string
		var elements []string

		for k := range v.request.RawParams {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, k := range keys {
			value := k + "=" + v.request.RawParams[k]
			elements = append(elements, value)
		}

		hashString = strings.Join(elements, "") + v.checked.project.SecretKey
	} else {
		hashString = v.request.RawBody + v.checked.project.SecretKey
	}

	h := sha512.New()
	h.Write([]byte(hashString))

	if hex.EncodeToString(h.Sum(nil)) != v.request.Signature {
		return orderErrorSignatureInvalid
	}

	return nil
}

// Calculate VAT for order
func (v *OrderCreateRequestProcessor) processOrderVat(order *billingpb.Order) error {
	order.Tax = &billingpb.OrderTax{
		Amount:   0,
		Rate:     0,
		Type:     taxTypeVat,
		Currency: order.Currency,
	}
	order.TotalPaymentAmount = order.OrderAmount
	order.ChargeAmount = order.TotalPaymentAmount
	order.ChargeCurrency = order.Currency

	countryCode := order.GetCountry()

	if countryCode == "" {
		return nil
	}

	if countryCode == CountryCodeUSA {
		order.Tax.Type = taxTypeSalesTax
	}

	if order.VatPayer == billingpb.VatPayerNobody {
		return nil
	}

	if countryCode != "" {
		country, err := v.country.GetByIsoCodeA2(v.ctx, countryCode)
		if err != nil {
			return err
		}
		if country.VatEnabled == false {
			return nil
		}
	}

	req := &taxpb.GeoIdentity{
		Country: countryCode,
	}

	if countryCode == CountryCodeUSA {
		req.Zip = order.GetPostalCode()
	}

	rsp, err := v.tax.GetRate(v.ctx, req)

	if err != nil {
		v.logError("Tax service return error", []interface{}{"error", err.Error(), "request", req})
		return err
	}

	order.Tax.Rate = rsp.Rate

	switch order.VatPayer {

	case billingpb.VatPayerBuyer:
		order.Tax.Amount = v.FormatAmount(order.OrderAmount*order.Tax.Rate, order.Currency)
		order.TotalPaymentAmount = v.FormatAmount(order.OrderAmount+order.Tax.Amount, order.Currency)
		order.ChargeAmount = order.TotalPaymentAmount
		break

	case billingpb.VatPayerSeller:
		order.Tax.Amount = v.FormatAmount(tools.GetPercentPartFromAmount(order.TotalPaymentAmount, order.Tax.Rate), order.Currency)
		break

	default:
		return orderErrorVatPayerUnknown
	}

	return nil
}

func (v *OrderCreateRequestProcessor) processCustomerToken(ctx context.Context) error {
	token, err := v.getTokenBy(v.request.Token)

	if err != nil {
		return err
	}

	customer, err := v.getCustomerById(ctx, token.CustomerId)

	if err != nil {
		return err
	}

	v.request.Type = token.Settings.Type
	v.request.PlatformId = token.Settings.PlatformId

	v.request.ProjectId = token.Settings.ProjectId
	v.request.Description = token.Settings.Description
	v.request.Amount = token.Settings.Amount
	v.request.Currency = token.Settings.Currency
	v.request.Products = token.Settings.ProductsIds
	v.request.Metadata = token.Settings.Metadata
	v.request.PaymentMethod = token.Settings.PaymentMethod

	if token.Settings.ReturnUrl != nil {
		v.request.UrlSuccess = token.Settings.ReturnUrl.Success
		v.request.UrlFail = token.Settings.ReturnUrl.Fail
	}

	v.checked.user = &billingpb.OrderUser{
		ExternalId: token.User.Id,
		Address:    token.User.Address,
		Metadata:   token.User.Metadata,
	}

	if token.User.Name != nil {
		v.checked.user.Name = token.User.Name.Value
	}

	if token.User.Email != nil {
		v.checked.user.Email = token.User.Email.Value
		v.checked.user.EmailVerified = token.User.Email.Verified
	}

	if token.User.Phone != nil {
		v.checked.user.Phone = token.User.Phone.Value
		v.checked.user.PhoneVerified = token.User.Phone.Verified
	}

	if token.User.Ip != nil {
		v.checked.user.Ip = token.User.Ip.Value
	}

	if token.User.Locale != nil {
		v.checked.user.Locale = token.User.Locale.Value
	}

	v.request.ButtonCaption = token.Settings.ButtonCaption

	v.checked.user.Id = customer.Id
	v.checked.user.Uuid = customer.Uuid
	v.checked.user.Object = pkg.ObjectTypeUser
	v.checked.user.TechEmail = customer.TechEmail

	if token.Settings.RecurringPeriod != "" {
		v.request.RecurringPeriod = token.Settings.RecurringPeriod
		v.request.RecurringDateEnd = token.Settings.RecurringDateEnd

		if err = v.processRecurringSettings(); err != nil {
			return err
		}
	}

	return nil
}

func (v *OrderCreateRequestProcessor) processUserData() (err error) {
	customer := new(billingpb.Customer)
	tokenReq := v.transformOrderUser2TokenRequest(v.request.User)

	if v.request.Token == "" {
		customer, _ = v.findCustomer(v.ctx, tokenReq, v.checked.project)
	}

	if customer != nil {
		customer, err = v.updateCustomer(v.ctx, tokenReq, v.checked.project, customer)
	} else {
		customer, err = v.createCustomer(v.ctx, tokenReq, v.checked.project)
	}

	if err != nil {
		return err
	}

	v.checked.user = v.request.User
	v.checked.user.Id = customer.Id
	v.checked.user.Uuid = customer.Uuid
	v.checked.user.Object = pkg.ObjectTypeUser
	v.checked.user.TechEmail = customer.TechEmail

	return
}

// GetById payment methods of project for rendering in payment form
func (v *PaymentFormProcessor) processRenderFormPaymentMethods(
	ctx context.Context,
) ([]*billingpb.PaymentFormPaymentMethod, error) {
	var projectPms []*billingpb.PaymentFormPaymentMethod

	paymentMethods, err := v.service.paymentMethodRepository.ListByOrder(
		ctx,
		v.order,
	)

	if err != nil {
		zap.S().Errorw("ListByOrder failed", "error", err, "order_id", v.order.Id, "order_uuid", v.order.Uuid)
		return nil, orderErrorUnknown
	}

	for _, pm := range paymentMethods {
		if pm.IsActive == false {
			continue
		}

		ps, err := v.service.paymentSystemRepository.GetById(ctx, pm.PaymentSystemId)

		if err != nil {
			zap.S().Errorw("GetById failed", "error", err, "order_id", v.order.Id, "order_uuid", v.order.Uuid)
			continue
		}

		if ps.IsActive == false {
			continue
		}

		if v.order.OrderAmount < pm.MinPaymentAmount ||
			(pm.MaxPaymentAmount > 0 && v.order.OrderAmount > pm.MaxPaymentAmount) {
			continue
		}
		_, err = v.service.getPaymentSettings(pm, v.order.Currency, v.order.MccCode, v.order.OperatingCompanyId, "", v.order.IsProduction)

		if err != nil {
			zap.S().Errorw("GetPaymentSettings failed", "error", err, "order_id", v.order.Id, "order_uuid", v.order.Uuid)
			continue
		}

		formPm := &billingpb.PaymentFormPaymentMethod{
			Id:            pm.Id,
			Name:          pm.Name,
			Type:          pm.Type,
			Group:         pm.Group,
			AccountRegexp: pm.AccountRegexp,
		}

		err = v.processPaymentMethodsData(ctx, formPm)

		if err != nil {
			zap.S().Errorw(
				"Process payment Method data failed",
				"error", err,
				"order_id", v.order.Id,
			)
			continue
		}

		projectPms = append(projectPms, formPm)
	}

	if len(projectPms) <= 0 {
		zap.S().Errorw("Not found any active payment methods", "order_id", v.order.Id, "order_uuid", v.order.Uuid)
		return projectPms, orderErrorPaymentMethodNotAllowed
	}

	return projectPms, nil
}

func (v *PaymentFormProcessor) processPaymentMethodsData(ctx context.Context, pm *billingpb.PaymentFormPaymentMethod) error {
	pm.HasSavedCards = false

	if pm.IsBankCard() == true {
		req := &recurringpb.SavedCardRequest{Token: v.order.User.Id}
		rsp, err := v.service.rep.FindSavedCards(ctx, req)

		if err != nil {
			zap.S().Errorw(
				"Get saved cards from repository failed",
				"error", err,
				"token", v.order.User.Id,
				"project_id", v.order.Project.Id,
				"order_id", v.order.Id,
			)
		} else {
			pm.HasSavedCards = len(rsp.SavedCards) > 0
			pm.SavedCards = []*billingpb.SavedCard{}

			for _, v := range rsp.SavedCards {
				d := &billingpb.SavedCard{
					Id:         v.Id,
					Pan:        v.MaskedPan,
					CardHolder: v.CardHolder,
					Expire:     &billingpb.CardExpire{Month: v.Expire.Month, Year: v.Expire.Year},
				}

				pm.SavedCards = append(pm.SavedCards, d)
			}

		}
	}

	return nil
}

func (v *PaymentCreateProcessor) reserveKeysForOrder(ctx context.Context, order *billingpb.Order) error {
	if len(order.Keys) == 0 {
		zap.S().Infow("[ProcessOrderKeyProducts] reserving keys", "order_id", order.Id)
		keys := make([]string, len(order.Products))
		for i, productId := range order.Products {
			reserveRes := &billingpb.PlatformKeyReserveResponse{}
			reserveReq := &billingpb.PlatformKeyReserveRequest{
				PlatformId:   order.PlatformId,
				MerchantId:   order.Project.MerchantId,
				OrderId:      order.Id,
				KeyProductId: productId,
				Ttl:          oneDayTtl,
			}

			err := v.service.ReserveKeyForOrder(ctx, reserveReq, reserveRes)
			if err != nil {
				zap.L().Error(
					pkg.ErrorGrpcServiceCallFailed,
					zap.Error(err),
					zap.String(errorFieldService, "KeyService"),
					zap.String(errorFieldMethod, "ReserveKeyForOrder"),
				)
				return err
			}

			if reserveRes.Status != billingpb.ResponseStatusOk {
				zap.S().Errorw("[ProcessOrderKeyProducts] can't reserve key. Cancelling reserved before", "message", reserveRes.Message, "order_id", order.Id)

				// we should cancel reservation for keys reserved before
				for _, keyToCancel := range keys {
					if len(keyToCancel) > 0 {
						cancelRes := &billingpb.EmptyResponseWithStatus{}
						err := v.service.CancelRedeemKeyForOrder(ctx, &billingpb.KeyForOrderRequest{KeyId: keyToCancel}, cancelRes)
						if err != nil {
							zap.L().Error(
								pkg.ErrorGrpcServiceCallFailed,
								zap.Error(err),
								zap.String(errorFieldService, "KeyService"),
								zap.String(errorFieldMethod, "CancelRedeemKeyForOrder"),
							)
						} else if cancelRes.Status != billingpb.ResponseStatusOk {
							zap.S().Errorw("[ProcessOrderKeyProducts] error during cancelling reservation", "message", cancelRes.Message, "order_id", order.Id)
						} else {
							zap.S().Infow("[ProcessOrderKeyProducts] successful canceled reservation", "order_id", order.Id, "key_id", keyToCancel)
						}
					}
				}

				return reserveRes.Message
			}
			zap.S().Infow("[ProcessOrderKeyProducts] reserved for product", "product ", productId, "reserveRes ", reserveRes, "order_id", order.Id)
			keys[i] = reserveRes.KeyId
		}

		order.Keys = keys
	}

	return nil
}

// Validate data received from payment form and write validated data to order
func (v *PaymentCreateProcessor) processPaymentFormData(ctx context.Context) error {
	if _, ok := v.data[billingpb.PaymentCreateFieldOrderId]; !ok ||
		v.data[billingpb.PaymentCreateFieldOrderId] == "" {
		return orderErrorCreatePaymentRequiredFieldIdNotFound
	}

	if _, ok := v.data[billingpb.PaymentCreateFieldPaymentMethodId]; !ok ||
		v.data[billingpb.PaymentCreateFieldPaymentMethodId] == "" {
		return orderErrorCreatePaymentRequiredFieldPaymentMethodNotFound
	}

	if _, ok := v.data[billingpb.PaymentCreateFieldEmail]; !ok ||
		v.data[billingpb.PaymentCreateFieldEmail] == "" {
		return orderErrorCreatePaymentRequiredFieldEmailNotFound
	}

	order, err := v.service.getOrderByUuidToForm(ctx, v.data[billingpb.PaymentCreateFieldOrderId])

	if err != nil {
		return err
	}

	if order.PrivateStatus != recurringpb.OrderStatusNew && order.PrivateStatus != recurringpb.OrderStatusPaymentSystemComplete {
		return orderErrorAlreadyProcessed
	}

	if order.UserAddressDataRequired == true {
		country, ok := v.data[billingpb.PaymentCreateFieldUserCountry]

		if !ok || country == "" {
			return orderErrorCreatePaymentRequiredFieldUserCountryNotFound
		}

		if country == CountryCodeUSA {
			zip, ok := v.data[billingpb.PaymentCreateFieldUserZip]

			if !ok || zip == "" {
				return orderErrorCreatePaymentRequiredFieldUserZipNotFound
			}

			zipData, err := v.service.zipCodeRepository.GetByZipAndCountry(ctx, zip, country)

			if err == nil && zipData != nil {
				v.data[billingpb.PaymentCreateFieldUserCity] = zipData.City
				v.data[billingpb.PaymentCreateFieldUserState] = zipData.State.Code
			}
		}
	}

	merchant, err := v.service.merchantRepository.GetById(ctx, order.GetMerchantId())
	if err != nil {
		return merchantErrorNotFound
	}

	if order.MccCode == "" {
		order.MccCode = merchant.MccCode
		order.IsHighRisk = merchant.IsHighRisk()
	}

	order.OperatingCompanyId, err = v.service.getOrderOperatingCompanyId(ctx, order.GetCountry(), merchant)
	if err != nil {
		return err
	}

	processor := &OrderCreateRequestProcessor{
		Service: v.service,
		request: &billingpb.OrderCreateRequest{
			ProjectId: order.Project.Id,
			Amount:    order.OrderAmount,
		},
		checked: &orderCreateRequestProcessorChecked{
			currency:           order.Currency,
			amount:             order.OrderAmount,
			mccCode:            order.MccCode,
			operatingCompanyId: order.OperatingCompanyId,
		},
		ctx: ctx,
	}

	if err := processor.processProject(); err != nil {
		return err
	}

	pm, err := v.service.paymentMethodRepository.GetById(ctx, v.data[billingpb.PaymentCreateFieldPaymentMethodId])
	if err != nil {
		return orderErrorPaymentMethodNotFound
	}

	if pm.IsActive == false {
		return orderErrorPaymentMethodInactive
	}

	ps, err := v.service.paymentSystemRepository.GetById(ctx, pm.PaymentSystemId)
	if err != nil {
		return orderErrorPaymentSystemInactive
	}

	if err := processor.processLimitAmounts(); err != nil {
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			return e
		}
		return err
	}

	if order.User.Ip != v.ip {
		order.User.Ip = v.ip
	}

	updCustomerReq := &billingpb.TokenRequest{User: &billingpb.TokenUser{}}

	if val, ok := v.data[billingpb.PaymentCreateFieldEmail]; ok {
		order.User.Email = val
		updCustomerReq.User.Email = &billingpb.TokenUserEmailValue{Value: val}
	}

	order.PaymentRequisites = make(map[string]string)

	if order.UserAddressDataRequired == true {
		if order.BillingAddress == nil {
			order.BillingAddress = &billingpb.OrderBillingAddress{}
		}

		if order.BillingAddress.Country != v.data[billingpb.PaymentCreateFieldUserCountry] {
			order.BillingAddress.Country = v.data[billingpb.PaymentCreateFieldUserCountry]
		}

		if order.BillingAddress.Country == CountryCodeUSA {
			if order.BillingAddress.City != v.data[billingpb.PaymentCreateFieldUserCity] {
				order.BillingAddress.City = v.data[billingpb.PaymentCreateFieldUserCity]
			}

			if order.BillingAddress.PostalCode != v.data[billingpb.PaymentCreateFieldUserZip] {
				order.BillingAddress.PostalCode = v.data[billingpb.PaymentCreateFieldUserZip]
			}

			if order.BillingAddress.State != v.data[billingpb.PaymentCreateFieldUserState] {
				order.BillingAddress.State = v.data[billingpb.PaymentCreateFieldUserState]
			}
		}

		err = processor.processOrderVat(order)
		if err != nil {
			zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error(), "method", "processOrderVat")
			return err
		}
		updCustomerReq.User.Address = order.BillingAddress
	}

	restricted, err := v.service.applyCountryRestriction(ctx, order, order.GetCountry())
	if err != nil {
		zap.L().Error(
			"v.service.applyCountryRestriction Method failed",
			zap.Error(err),
			zap.Any("order", order),
		)
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			return e
		}
		return orderErrorUnknown
	}
	if restricted {
		return orderCountryPaymentRestrictedError
	}

	var customer *billingpb.Customer

	if helper.IsIdentified(order.User.Id) == true {
		customer, _ = v.service.customerRepository.GetById(ctx, order.User.Id)

		if customer != nil && customer.Email != order.User.Email {
			customer = nil
		}
	}

	if customer != nil {
		customer, err := v.service.updateCustomerFromRequest(ctx, order, updCustomerReq, v.ip, v.acceptLanguage, v.userAgent)

		if err != nil {
			v.service.logError("Update customer data by request failed", []interface{}{"error", err.Error(), "data", updCustomerReq})
		} else {
			if order.User.Locale == "" && customer.Locale != "" &&
				customer.Locale != order.User.Locale {
				order.User.Locale = customer.Locale
			}
		}
	} else {
		customer, err := v.service.createCustomerFromRequest(ctx, order, updCustomerReq, v.ip, v.acceptLanguage, v.userAgent)
		if err != nil {
			v.service.logError("Create customer data by request failed", []interface{}{"error", err.Error(), "data", updCustomerReq})
			return err
		}

		order.User.Id = customer.Id
		order.User.Uuid = customer.Uuid
	}

	delete(v.data, billingpb.PaymentCreateFieldOrderId)
	delete(v.data, billingpb.PaymentCreateFieldPaymentMethodId)
	delete(v.data, billingpb.PaymentCreateFieldEmail)

	if pm.IsBankCard() == true {
		if id, ok := v.data[billingpb.PaymentCreateFieldStoredCardId]; ok {
			storedCard, err := v.service.rep.FindSavedCardById(ctx, &recurringpb.FindByStringValue{Value: id})

			if err != nil {
				v.service.logError("Get data about stored card failed", []interface{}{"err", err.Error(), "id", id})
			}

			if storedCard == nil {
				v.service.logError("Get data about stored card failed", []interface{}{"id", id})
				return orderGetSavedCardError
			}

			if storedCard.Token != order.User.Id {
				v.service.logError("Alarm: user try use not own bank card for payment", []interface{}{"user_id", order.User.Id, "card_id", id})
				return orderErrorRecurringCardNotOwnToUser
			}

			order.PaymentRequisites[billingpb.PaymentCreateFieldPan] = storedCard.MaskedPan
			order.PaymentRequisites[billingpb.PaymentCreateFieldMonth] = storedCard.Expire.Month
			order.PaymentRequisites[billingpb.PaymentCreateFieldYear] = storedCard.Expire.Year
			order.PaymentRequisites[billingpb.PaymentCreateFieldHolder] = storedCard.CardHolder
			order.PaymentRequisites[billingpb.PaymentCreateFieldRecurringId] = storedCard.RecurringId
		} else {
			validator := &bankCardValidator{
				Pan:    v.data[billingpb.PaymentCreateFieldPan],
				Cvv:    v.data[billingpb.PaymentCreateFieldCvv],
				Month:  v.data[billingpb.PaymentCreateFieldMonth],
				Year:   v.data[billingpb.PaymentCreateFieldYear],
				Holder: v.data[billingpb.PaymentCreateFieldHolder],
			}

			if err := validator.Validate(); err != nil {
				return err
			}

			order.PaymentRequisites[billingpb.PaymentCreateFieldPan] = stringTools.MaskBankCardNumber(v.data[billingpb.PaymentCreateFieldPan])
			order.PaymentRequisites[billingpb.PaymentCreateFieldMonth] = v.data[billingpb.PaymentCreateFieldMonth]

			if len(v.data[billingpb.PaymentCreateFieldYear]) < 3 {
				v.data[billingpb.PaymentCreateFieldYear] = strconv.Itoa(time.Now().UTC().Year())[:2] + v.data[billingpb.PaymentCreateFieldYear]
			}

			order.PaymentRequisites[billingpb.PaymentCreateFieldYear] = v.data[billingpb.PaymentCreateFieldYear]
		}

		bin := v.service.getBinData(ctx, order.PaymentRequisites[billingpb.PaymentCreateFieldPan])

		if bin != nil {
			order.PaymentRequisites[billingpb.PaymentCreateBankCardFieldBrand] = bin.CardBrand
			order.PaymentRequisites[billingpb.PaymentCreateBankCardFieldType] = bin.CardType
			order.PaymentRequisites[billingpb.PaymentCreateBankCardFieldCategory] = bin.CardCategory
			order.PaymentRequisites[billingpb.PaymentCreateBankCardFieldIssuerName] = bin.BankName
			order.PaymentRequisites[billingpb.PaymentCreateBankCardFieldIssuerCountry] = bin.BankCountryName
			order.PaymentRequisites[billingpb.PaymentCreateBankCardFieldIssuerCountryIsoCode] = bin.BankCountryIsoCode
		}
	} else {
		account := ""

		if acc, ok := v.data[billingpb.PaymentCreateFieldEWallet]; ok {
			account = acc
		}

		if acc, ok := v.data[billingpb.PaymentCreateFieldCrypto]; ok {
			account = acc
		}

		if account == "" {
			return payment_system.PaymentSystemErrorEWalletIdentifierIsInvalid
		}

		order.PaymentRequisites = v.data
	}

	if order.PaymentMethod == nil {
		order.PaymentMethod = &billingpb.PaymentMethodOrder{
			Id:              pm.Id,
			Name:            pm.Name,
			PaymentSystemId: ps.Id,
			Group:           pm.Group,
			ExternalId:      pm.ExternalId,
			Handler:         ps.Handler,
			RefundAllowed:   pm.RefundAllowed,
		}
	}

	methodName, err := order.GetCostPaymentMethodName()

	if err == nil {
		order.PaymentMethod.Params, err = v.service.getPaymentSettings(
			pm,
			processor.checked.currency,
			processor.checked.mccCode,
			processor.checked.operatingCompanyId,
			methodName,
			order.IsProduction,
		)

		if err != nil {
			return err
		}
	}

	v.checked.project = processor.checked.project
	v.checked.paymentMethod = pm
	v.checked.order = order

	if v.checked.project.CallbackProtocol == billingpb.ProjectCallbackProtocolDefault &&
		v.checked.project.WebhookMode == pkg.ProjectWebhookPreApproval {
		err = v.service.webhookCheckUser(order)

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) GetOrderKeyProducts(ctx context.Context, projectId string, productIds []string) ([]*billingpb.KeyProduct, error) {
	if len(productIds) == 0 {
		return nil, orderErrorProductsEmpty
	}

	result := billingpb.ListKeyProductsResponse{}

	err := s.GetKeyProductsForOrder(ctx, &billingpb.GetKeyProductsForOrderRequest{
		ProjectId: projectId,
		Ids:       productIds,
	}, &result)

	if err != nil {
		zap.L().Error(
			"v.GetKeyProductsForOrder Method failed",
			zap.Error(err),
		)
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			return nil, e
		}
		return nil, orderErrorUnknown
	}

	if result.Count != int64(len(productIds)) {
		return nil, orderErrorProductsInvalid
	}

	return result.Products, nil
}

func (s *Service) GetOrderKeyProductsAmount(products []*billingpb.KeyProduct, group *billingpb.PriceGroup, platformId string) (float64, error) {
	if len(products) == 0 {
		return 0, orderErrorProductsEmpty
	}

	sum := float64(0)

	for _, p := range products {
		amount, err := p.GetPriceInCurrencyAndPlatform(group, platformId)

		if err != nil {
			return 0, orderErrorNoProductsCommonCurrency
		}

		sum += amount
	}

	return sum, nil
}

func (s *Service) GetOrderProducts(ctx context.Context, projectId string, productIds []string) ([]*billingpb.Product, error) {
	if len(productIds) == 0 {
		return nil, orderErrorProductsEmpty
	}

	result := billingpb.ListProductsResponse{}

	err := s.GetProductsForOrder(ctx, &billingpb.GetProductsForOrderRequest{
		ProjectId: projectId,
		Ids:       productIds,
	}, &result)

	if err != nil {
		zap.L().Error(
			"v.GetProductsForOrder Method failed",
			zap.Error(err),
		)
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			return nil, e
		}
		return nil, orderErrorUnknown
	}

	if result.Total != int64(len(productIds)) {
		return nil, orderErrorProductsInvalid
	}

	return result.Products, nil
}

func (s *Service) GetOrderProductsAmount(products []*billingpb.Product, group *billingpb.PriceGroup) (float64, error) {
	if len(products) == 0 {
		return 0, orderErrorProductsEmpty
	}

	sum := float64(0)

	for _, p := range products {
		amount, err := p.GetPriceInCurrency(group)

		if err != nil {
			return 0, err
		}

		sum += amount
	}

	totalAmount := tools.FormatAmount(sum)

	return totalAmount, nil
}

func (s *Service) GetOrderProductsItems(products []*billingpb.Product, language string, group *billingpb.PriceGroup) ([]*billingpb.OrderItem, error) {
	var result []*billingpb.OrderItem

	if len(products) == 0 {
		return nil, orderErrorProductsEmpty
	}

	isDefaultLanguage := language == DefaultLanguage

	for _, p := range products {
		var (
			amount      float64
			name        string
			description string
			err         error
		)

		amount, err = p.GetPriceInCurrency(group)
		if err != nil {
			return nil, orderErrorProductsPrice
		}

		name, err = p.GetLocalizedName(language)
		if err != nil {
			if isDefaultLanguage {
				return nil, orderErrorNoNameInRequiredLanguage
			}
			name, err = p.GetLocalizedName(DefaultLanguage)
			if err != nil {
				return nil, orderErrorNoNameInDefaultLanguage
			}
		}

		description, err = p.GetLocalizedDescription(language)
		if err != nil {
			if isDefaultLanguage {
				return nil, orderErrorNoDescriptionInRequiredLanguage
			}
			description, err = p.GetLocalizedDescription(DefaultLanguage)
			if err != nil {
				return nil, orderErrorNoDescriptionInDefaultLanguage
			}
		}

		item := &billingpb.OrderItem{
			Id:          p.Id,
			Object:      p.Object,
			Sku:         p.Sku,
			Name:        name,
			Description: description,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
			Images:      p.Images,
			Url:         p.Url,
			Metadata:    p.Metadata,
			Amount:      amount,
			Currency:    group.Currency,
		}
		result = append(result, item)
	}

	return result, nil
}

func (s *Service) GetOrderKeyProductsItems(products []*billingpb.KeyProduct, language string, group *billingpb.PriceGroup, platformId string) ([]*billingpb.OrderItem, error) {
	var result []*billingpb.OrderItem

	if len(products) == 0 {
		return nil, orderErrorProductsEmpty
	}

	isDefaultLanguage := language == DefaultLanguage

	for _, p := range products {
		var (
			amount      float64
			name        string
			description string
			err         error
		)

		amount, err = p.GetPriceInCurrencyAndPlatform(group, platformId)
		if err != nil {
			return nil, orderErrorProductsPrice
		}

		name, err = p.GetLocalizedName(language)
		if err != nil {
			if isDefaultLanguage {
				return nil, orderErrorNoNameInRequiredLanguage
			}
			name, err = p.GetLocalizedName(DefaultLanguage)
			if err != nil {
				return nil, orderErrorNoNameInDefaultLanguage
			}
		}

		description, err = p.GetLocalizedDescription(language)
		if err != nil {
			if isDefaultLanguage {
				return nil, orderErrorNoDescriptionInRequiredLanguage
			}
			description, err = p.GetLocalizedDescription(DefaultLanguage)
			if err != nil {
				return nil, orderErrorNoDescriptionInDefaultLanguage
			}
		}

		item := &billingpb.OrderItem{
			Id:          p.Id,
			Object:      p.Object,
			Sku:         p.Sku,
			Name:        name,
			Description: description,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
			Images:      []string{getImageByLanguage(DefaultLanguage, p.Cover)},
			Url:         p.Url,
			Metadata:    p.Metadata,
			Amount:      amount,
			Currency:    group.Currency,
			PlatformId:  platformId,
		}
		result = append(result, item)
	}

	return result, nil
}

func (s *Service) filterPlatforms(orderProducts []*billingpb.KeyProduct) []string {
	// filter available platformIds for all products in request
	var platformIds []string
	for i, product := range orderProducts {
		var platformsToCheck []string
		for _, pl := range product.Platforms {
			platformsToCheck = append(platformsToCheck, pl.Id)
		}

		if i > 0 {
			platformIds = intersect(platformIds, platformsToCheck)
		} else {
			platformIds = platformsToCheck
		}
	}

	return platformIds
}

func (s *Service) ProcessOrderVirtualCurrency(ctx context.Context, order *billingpb.Order) error {
	var (
		country    string
		currency   string
		priceGroup *billingpb.PriceGroup
	)

	merchant, _ := s.merchantRepository.GetById(ctx, order.Project.MerchantId)
	defaultCurrency := merchant.GetProcessingDefaultCurrency()

	if defaultCurrency == "" {
		zap.S().Infow("merchant payout currency not found", "order.Uuid", order.Uuid)
		return orderErrorNoProductsCommonCurrency
	}

	defaultPriceGroup, err := s.priceGroupRepository.GetByRegion(ctx, defaultCurrency)
	if err != nil {
		zap.S().Errorw("Price group not found", "currency", currency)
		return orderErrorUnknown
	}

	currency = defaultCurrency
	priceGroup = defaultPriceGroup

	country = order.GetCountry()

	if country != "" {
		countryData, err := s.country.GetByIsoCodeA2(ctx, country)
		if err != nil {
			zap.S().Errorw("Country not found", "country", country)
			return orderErrorUnknown
		}

		priceGroup, err = s.priceGroupRepository.GetById(ctx, countryData.PriceGroupId)
		if err != nil {
			zap.S().Errorw("Price group not found", "countryData", countryData)
			return orderErrorUnknown
		}

		currency = priceGroup.Currency
	}

	zap.S().Infow("try to use detected currency for order amount", "currency", currency, "order.Uuid", order.Uuid)

	project, err := s.project.GetById(ctx, order.GetProjectId())

	if err != nil || project == nil || project.VirtualCurrency == nil {
		return orderErrorVirtualCurrencyNotFilled
	}

	amount, err := s.GetAmountForVirtualCurrency(order.VirtualCurrencyAmount, priceGroup, project.VirtualCurrency.Prices)
	if err != nil {
		if priceGroup.Id == defaultPriceGroup.Id {
			return err
		}

		// try to get order Amount in default currency, if it differs from requested one
		amount, err = s.GetAmountForVirtualCurrency(order.VirtualCurrencyAmount, defaultPriceGroup, project.VirtualCurrency.Prices)
		if err != nil {
			return err
		}
	}

	amount = s.FormatAmount(amount, currency)

	order.Currency = currency
	order.OrderAmount = amount
	order.TotalPaymentAmount = amount
	order.ChargeAmount = amount
	order.ChargeCurrency = currency

	return nil
}

func (s *Service) GetAmountForVirtualCurrency(virtualAmount float64, group *billingpb.PriceGroup, prices []*billingpb.ProductPrice) (float64, error) {
	for _, price := range prices {
		if price.Currency == group.Currency {
			return virtualAmount * price.Amount, nil
		}
	}

	return 0, virtualCurrencyPayoutCurrencyMissed
}

func (s *Service) ProcessOrderKeyProducts(ctx context.Context, order *billingpb.Order) ([]*billingpb.Platform, error) {
	if order.ProductType != pkg.OrderType_key {
		return nil, nil
	}

	priceGroup, err := s.getOrderPriceGroup(ctx, order)
	if err != nil {
		zap.L().Error(
			"ProcessOrderKeyProducts getOrderPriceGroup failed",
			zap.Error(err),
			zap.String("order.Uuid", order.Uuid),
		)
		return nil, err
	}

	locale := DefaultLanguage
	if order.User != nil && order.User.Locale != "" {
		locale = order.User.Locale
	}

	amount, priceGroup, items, platforms, err := s.processKeyProducts(
		ctx,
		order.Project.Id,
		order.Products,
		priceGroup,
		locale,
		order.PlatformId,
	)

	if err != nil {
		return nil, err
	}

	if len(order.PlatformId) == 0 && len(items) > 0 {
		order.PlatformId = items[0].PlatformId
	}

	order.Currency = priceGroup.Currency
	order.OrderAmount = amount
	order.TotalPaymentAmount = amount

	order.ChargeAmount = order.TotalPaymentAmount
	order.ChargeCurrency = order.Currency

	order.Items = items

	return platforms, nil
}

func (s *Service) ProcessOrderProducts(ctx context.Context, order *billingpb.Order) error {

	if order.ProductType != pkg.OrderType_product {
		return nil
	}
	priceGroup, err := s.getOrderPriceGroup(ctx, order)
	if err != nil {
		zap.L().Error(
			"ProcessOrderProducts getOrderPriceGroup failed",
			zap.Error(err),
			zap.String("order.Uuid", order.Uuid),
		)
		return err
	}

	locale := DefaultLanguage
	if order.User != nil && order.User.Locale != "" {
		locale = order.User.Locale
	}

	amount, priceGroup, items, _, err := s.processProducts(
		ctx,
		order.Project.Id,
		order.Products,
		priceGroup,
		locale,
	)

	if err != nil {
		return err
	}

	order.Currency = priceGroup.Currency

	order.OrderAmount = amount
	order.TotalPaymentAmount = amount

	order.ChargeAmount = order.TotalPaymentAmount
	order.ChargeCurrency = order.Currency

	order.Items = items

	return nil
}

func (s *Service) processAmountForFiatCurrency(
	_ context.Context,
	_ *billingpb.Project,
	orderProducts []*billingpb.Product,
	priceGroup *billingpb.PriceGroup,
	defaultPriceGroup *billingpb.PriceGroup,
) (float64, *billingpb.PriceGroup, error) {
	// try to get order Amount in requested currency
	amount, err := s.GetOrderProductsAmount(orderProducts, priceGroup)
	if err != nil {
		if err != billingpb.ProductNoPriceInCurrencyError {
			return 0, nil, err
		}

		if priceGroup.Id == defaultPriceGroup.Id {
			return 0, nil, err
		}

		// try to get order Amount in fallback currency
		amount, err = s.GetOrderProductsAmount(orderProducts, defaultPriceGroup)
		if err != nil {
			return 0, nil, err
		}
		return amount, defaultPriceGroup, nil
	}

	return amount, priceGroup, nil
}

func (s *Service) processAmountForVirtualCurrency(
	_ context.Context,
	project *billingpb.Project,
	orderProducts []*billingpb.Product,
	priceGroup *billingpb.PriceGroup,
	defaultPriceGroup *billingpb.PriceGroup,
) (float64, *billingpb.PriceGroup, error) {

	if project.VirtualCurrency == nil || len(project.VirtualCurrency.Prices) == 0 {
		return 0, nil, orderErrorVirtualCurrencyNotFilled
	}

	var amount float64

	usedPriceGroup := priceGroup

	virtualAmount, err := s.GetOrderProductsAmount(orderProducts, &billingpb.PriceGroup{Currency: billingpb.VirtualCurrencyPriceGroup})
	if err != nil {
		zap.L().Error(pkg.MethodFinishedWithError, zap.Error(err))
		return 0, nil, err
	}

	amount, err = s.GetAmountForVirtualCurrency(virtualAmount, usedPriceGroup, project.VirtualCurrency.Prices)
	if err != nil {
		zap.L().Error(pkg.MethodFinishedWithError, zap.Error(err))
		if priceGroup.Id == defaultPriceGroup.Id {
			return 0, nil, err
		}

		// try to get order Amount in fallback currency
		usedPriceGroup = defaultPriceGroup
		amount, err = s.GetAmountForVirtualCurrency(virtualAmount, usedPriceGroup, project.VirtualCurrency.Prices)
		if err != nil {
			zap.L().Error(pkg.MethodFinishedWithError, zap.Error(err))
			return 0, nil, err
		}
	}

	return amount, usedPriceGroup, nil
}

func (s *Service) notifyPaylinkError(ctx context.Context, paylinkId string, err error, req interface{}, order interface{}) {
	msg := map[string]interface{}{
		"event":     "error",
		"paylinkId": paylinkId,
		"message":   "Invalid paylink",
		"error":     err.Error(),
		"request":   req,
		"order":     order,
	}
	_ = s.centrifugoDashboard.Publish(ctx, centrifugoChannel, msg)
}

func (v *PaymentCreateProcessor) GetMerchantId() string {
	return v.checked.project.MerchantId
}

func (s *Service) processCustomerData(
	ctx context.Context,
	customerId string,
	order *billingpb.Order,
	req *billingpb.PaymentFormJsonDataRequest,
	browserCustomer *BrowserCookieCustomer,
	locale string,
) (*billingpb.Customer, error) {
	customer, err := s.getCustomerById(ctx, customerId)

	if err != nil {
		return nil, err
	}

	tokenReq := &billingpb.TokenRequest{
		User: &billingpb.TokenUser{
			Ip:             &billingpb.TokenUserIpValue{Value: req.Ip},
			Locale:         &billingpb.TokenUserLocaleValue{Value: locale},
			AcceptLanguage: req.Locale,
			UserAgent:      req.UserAgent,
		},
	}
	project := &billingpb.Project{
		Id:         order.Project.Id,
		MerchantId: order.Project.MerchantId,
	}

	browserCustomer.CustomerId = customer.Id
	_, err = s.updateCustomer(ctx, tokenReq, project, customer)

	return customer, err
}

func (s *Service) IsOrderCanBePaying(
	ctx context.Context,
	req *billingpb.IsOrderCanBePayingRequest,
	rsp *billingpb.IsOrderCanBePayingResponse,
) error {
	order, err := s.getOrderByUuidToForm(ctx, req.OrderId)
	rsp.Status = billingpb.ResponseStatusBadData

	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Message = e
			return nil
		}
		return err
	}

	if order != nil && order.GetProjectId() != req.ProjectId {
		rsp.Message = orderErrorOrderCreatedAnotherProject
		return nil
	}

	rsp.Status = billingpb.ResponseStatusOk
	rsp.Item = order

	return nil
}

func (s *Service) fillPaymentDataCard(order *billingpb.Order) error {
	first6 := ""
	last4 := ""
	pan, ok := order.PaymentMethodTxnParams[billingpb.PaymentCreateFieldPan]
	if !ok || pan == "" {
		pan, ok = order.PaymentRequisites["pan"]
		if !ok {
			pan = ""
		}
	}
	order.PaymentMethodPayerAccount = pan
	if len(pan) >= 6 {
		first6 = pan[0:6]
		last4 = pan[len(pan)-4:]
	}

	cardBrand, ok := order.PaymentRequisites["card_brand"]
	if !ok {
		cardBrand = ""
	}

	month, ok := order.PaymentRequisites["month"]
	if !ok {
		month = ""
	}

	year, ok := order.PaymentRequisites["year"]
	if !ok {
		year = ""
	}

	order.PaymentMethod.Card = &billingpb.PaymentMethodCard{
		Masked:      pan,
		First6:      first6,
		Last4:       last4,
		ExpiryMonth: month,
		ExpiryYear:  year,
		Brand:       cardBrand,
		Secure3D:    order.PaymentMethodTxnParams[billingpb.TxnParamsFieldBankCardIs3DS] == "1",
	}
	b, err := json.Marshal(order.PaymentMethod.Card)
	if err != nil {
		return err
	}
	fp, err := bcrypt.GenerateFromPassword([]byte(string(b)), bcrypt.MinCost)
	if err == nil {
		order.PaymentMethod.Card.Fingerprint = string(fp)
	}
	return nil
}

func (s *Service) fillPaymentDataEwallet(order *billingpb.Order) error {
	account := order.PaymentMethodTxnParams[billingpb.PaymentCreateFieldEWallet]
	order.PaymentMethodPayerAccount = account
	order.PaymentMethod.Wallet = &billingpb.PaymentMethodWallet{
		Brand:   order.PaymentMethod.Name,
		Account: account,
	}
	return nil
}

func (s *Service) fillPaymentDataCrypto(order *billingpb.Order) error {
	address := order.PaymentMethodTxnParams[billingpb.PaymentCreateFieldCrypto]
	order.PaymentMethodPayerAccount = address
	order.PaymentMethod.CryptoCurrency = &billingpb.PaymentMethodCrypto{
		Brand:   order.PaymentMethod.Name,
		Address: address,
	}
	return nil
}

func (s *Service) SetUserNotifySales(
	ctx context.Context,
	req *billingpb.SetUserNotifyRequest,
	_ *billingpb.EmptyResponse,
) error {

	order, err := s.getOrderByUuid(ctx, req.OrderUuid)

	if err != nil {
		s.logError(orderErrorNotFound.Message, []interface{}{"error", err.Error(), "request", req})
		return orderErrorNotFound
	}

	if req.EnableNotification && req.Email == "" {
		return orderErrorEmailRequired
	}

	order.NotifySale = req.EnableNotification
	order.NotifySaleEmail = req.Email
	err = s.updateOrder(ctx, order)
	if err != nil {
		return err
	}

	if !req.EnableNotification {
		return nil
	}

	data := &billingpb.NotifyUserSales{
		Email:   req.Email,
		OrderId: order.Id,
		Date:    time.Now().Format(time.RFC3339),
	}
	if order.User != nil {
		data.UserId = order.User.Id
	}

	if err = s.notifySalesRepository.Insert(ctx, data); err != nil {
		return err
	}

	if helper.IsIdentified(order.User.Id) == true {
		customer, err := s.getCustomerById(ctx, order.User.Id)
		if err != nil {
			return err
		}
		project, err := s.project.GetById(ctx, order.Project.Id)
		if err != nil {
			return err
		}

		customer.NotifySale = req.EnableNotification
		customer.NotifySaleEmail = req.Email

		tokenReq := s.transformOrderUser2TokenRequest(order.User)
		_, err = s.updateCustomer(ctx, tokenReq, project, customer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) SetUserNotifyNewRegion(
	ctx context.Context,
	req *billingpb.SetUserNotifyRequest,
	_ *billingpb.EmptyResponse,
) error {
	order, err := s.getOrderByUuid(ctx, req.OrderUuid)

	if err != nil {
		s.logError(orderErrorNotFound.Message, []interface{}{"error", err.Error(), "request", req})
		return orderErrorNotFound
	}

	if order.CountryRestriction.PaymentsAllowed {
		s.logError(orderErrorNotRestricted.Message, []interface{}{"request", req})
		return orderErrorNotRestricted
	}

	if req.EnableNotification && req.Email == "" {
		return orderErrorEmailRequired
	}

	if order.User == nil {
		order.User = &billingpb.OrderUser{}
	}
	order.User.NotifyNewRegion = req.EnableNotification
	order.User.NotifyNewRegionEmail = req.Email
	err = s.updateOrder(ctx, order)
	if err != nil {
		return err
	}

	if !(req.EnableNotification && order.CountryRestriction != nil) {
		return nil
	}

	data := &billingpb.NotifyUserNewRegion{
		Email:            req.Email,
		OrderId:          order.Id,
		UserId:           order.User.Id,
		Date:             time.Now().Format(time.RFC3339),
		CountryIsoCodeA2: order.CountryRestriction.IsoCodeA2,
	}

	if err = s.notifyRegionRepository.Insert(ctx, data); err != nil {
		return err
	}

	if helper.IsIdentified(order.User.Id) == true {
		customer, err := s.getCustomerById(ctx, order.User.Id)
		if err != nil {
			return err
		}
		project, err := s.project.GetById(ctx, order.Project.Id)
		if err != nil {
			return err
		}

		customer.NotifyNewRegion = req.EnableNotification
		customer.NotifyNewRegionEmail = req.Email

		tokenReq := s.transformOrderUser2TokenRequest(order.User)
		_, err = s.updateCustomer(ctx, tokenReq, project, customer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) applyCountryRestriction(
	ctx context.Context,
	order *billingpb.Order,
	countryCode string,
) (restricted bool, err error) {
	restricted = false
	if countryCode == "" {
		order.UserAddressDataRequired = true
		order.CountryRestriction = &billingpb.CountryRestriction{
			PaymentsAllowed: false,
			ChangeAllowed:   true,
		}
		return
	}

	country, err := s.country.GetByIsoCodeA2(ctx, countryCode)
	if err != nil {
		return
	}

	merchant, err := s.merchantRepository.GetById(ctx, order.GetMerchantId())
	if err != nil {
		return
	}

	paymentsAllowed, changeAllowed := country.GetPaymentRestrictions(merchant.IsHighRisk())

	order.CountryRestriction = &billingpb.CountryRestriction{
		IsoCodeA2:       countryCode,
		PaymentsAllowed: paymentsAllowed,
		ChangeAllowed:   changeAllowed,
	}
	if paymentsAllowed {
		return
	}
	if changeAllowed {
		order.UserAddressDataRequired = true
		return
	}
	order.PrivateStatus = recurringpb.OrderStatusPaymentSystemDeclined
	restricted = true
	err = s.updateOrder(ctx, order)
	if err != nil && err.Error() == orderErrorNotFound.Error() {
		err = nil
	}
	return
}

func (s *Service) PaymentFormPlatformChanged(ctx context.Context, req *billingpb.PaymentFormUserChangePlatformRequest, rsp *billingpb.PaymentFormDataChangeResponse) error {
	order, err := s.getOrderByUuidToForm(ctx, req.OrderId)

	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	rsp.Status = billingpb.ResponseStatusOk

	order.PlatformId = req.Platform

	if order.ProductType == pkg.OrderType_product {
		err = s.ProcessOrderProducts(ctx, order)
	} else if order.ProductType == pkg.OrderType_key {
		_, err = s.ProcessOrderKeyProducts(ctx, order)
	}

	if err != nil {
		if pid := order.PrivateMetadata["PaylinkId"]; pid != "" {
			s.notifyPaylinkError(ctx, pid, err, req, order)
		}
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	processor := &OrderCreateRequestProcessor{Service: s, ctx: ctx}
	err = processor.processOrderVat(order)
	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error(), "method", "processOrderVat")
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusSystemError
			rsp.Message = e
			return nil
		}
		return err
	}

	err = s.setOrderChargeAmountAndCurrency(ctx, order)
	if err != nil {
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	err = s.updateOrder(ctx, order)

	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err.Error())
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusSystemError
			rsp.Message = e
			return nil
		}
		return err
	}

	rsp.Item = order.GetPaymentFormDataChangeResult()

	return nil
}

func (s *Service) OrderReceipt(
	ctx context.Context,
	req *billingpb.OrderReceiptRequest,
	rsp *billingpb.OrderReceiptResponse,
) error {
	order, err := s.orderRepository.GetByUuid(ctx, req.OrderId)

	if err != nil {
		zap.L().Error(pkg.MethodFinishedWithError, zap.Error(err))

		rsp.Status = billingpb.ResponseStatusBadData
		rsp.Message = err.(*billingpb.ResponseErrorMessage)

		return nil
	}

	if order.ReceiptId != req.ReceiptId {
		zap.L().Error(
			orderErrorReceiptNotEquals.Message,
			zap.String("Requested receipt", req.ReceiptId),
			zap.String("Order receipt", order.ReceiptId),
		)

		rsp.Status = billingpb.ResponseStatusBadData
		rsp.Message = orderErrorReceiptNotEquals

		return nil
	}

	receipt, err := s.getOrderReceiptObject(ctx, order)
	if err != nil {
		zap.L().Error("get order receipt object failed", zap.Error(err))
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	rsp.Status = billingpb.ResponseStatusOk
	rsp.Receipt = receipt

	return nil
}

func (s *Service) getOrderReceiptObject(ctx context.Context, order *billingpb.Order) (*billingpb.OrderReceipt, error) {
	merchant, err := s.merchantRepository.GetById(ctx, order.GetMerchantId())

	if err != nil {
		zap.L().Error(orderErrorMerchantForOrderNotFound.Message, zap.Error(err))
		return nil, orderErrorMerchantForOrderNotFound
	}

	totalPrice, err := s.formatter.FormatCurrency(DefaultLanguage, order.OrderAmount, order.Currency)
	if err != nil {
		zap.L().Error(
			orderErrorDuringFormattingCurrency.Message,
			zap.Float64("price", order.OrderAmount),
			zap.String("locale", DefaultLanguage),
			zap.String("currency", order.Currency),
		)
		return nil, orderErrorDuringFormattingCurrency
	}

	totalAmount, err := s.formatter.FormatCurrency(DefaultLanguage, order.TotalPaymentAmount, order.Currency)
	if err != nil {
		zap.L().Error(
			orderErrorDuringFormattingCurrency.Message,
			zap.Float64("price", order.TotalPaymentAmount),
			zap.String("locale", DefaultLanguage),
			zap.String("currency", order.Currency),
		)
		return nil, orderErrorDuringFormattingCurrency
	}

	vatInOrderCurrency, err := s.formatter.FormatCurrency(DefaultLanguage, order.Tax.Amount, order.Tax.Currency)
	if err != nil {
		zap.L().Error(
			orderErrorDuringFormattingCurrency.Message,
			zap.Float64("price", order.Tax.Amount),
			zap.String("locale", DefaultLanguage),
			zap.String("currency", order.Tax.Currency),
		)
		return nil, orderErrorDuringFormattingCurrency
	}

	vatInChargeCurrency, err := s.formatter.FormatCurrency(DefaultLanguage, order.GetTaxAmountInChargeCurrency(), order.ChargeCurrency)
	if err != nil {
		zap.L().Error(
			orderErrorDuringFormattingCurrency.Message,
			zap.Float64("price", order.GetTaxAmountInChargeCurrency()),
			zap.String("locale", DefaultLanguage),
			zap.String("currency", order.ChargeCurrency),
		)
		return nil, orderErrorDuringFormattingCurrency
	}

	totalCharge, err := s.formatter.FormatCurrency(DefaultLanguage, order.ChargeAmount, order.ChargeCurrency)
	if err != nil {
		zap.L().Error(
			orderErrorDuringFormattingCurrency.Message,
			zap.Float64("price", order.ChargeAmount),
			zap.String("locale", DefaultLanguage),
			zap.String("currency", order.ChargeCurrency),
		)
		return nil, orderErrorDuringFormattingCurrency
	}

	date, err := s.formatter.FormatDateTime(DefaultLanguage, time.Unix(order.CreatedAt.Seconds, 0))
	if err != nil {
		zap.L().Error(
			orderErrorDuringFormattingDate.Message,
			zap.Any("date", order.CreatedAt),
			zap.String("locale", DefaultLanguage),
		)
		return nil, orderErrorDuringFormattingDate
	}

	items := make([]*billingpb.OrderReceiptItem, len(order.Items))

	currency := order.Currency
	if order.IsBuyForVirtualCurrency {
		project, err := s.project.GetById(ctx, order.GetProjectId())

		if err != nil {
			zap.L().Error(
				projectErrorUnknown.Message,
				zap.Error(err),
				zap.String("order.uuid", order.Uuid),
			)
			return nil, projectErrorUnknown
		}

		var ok = false
		currency, ok = project.VirtualCurrency.Name[DefaultLanguage]

		if !ok {
			zap.L().Error(
				projectErrorVirtualCurrencyNameDefaultLangRequired.Message,
				zap.Error(err),
				zap.String("order.uuid", order.Uuid),
			)
			return nil, projectErrorVirtualCurrencyNameDefaultLangRequired
		}
	}

	for i, item := range order.Items {
		price, err := s.formatter.FormatCurrency(DefaultLanguage, item.Amount, currency)

		// Virtual currency always returns error but formatting with Name
		if err != nil && order.IsBuyForVirtualCurrency == false {
			zap.L().Error(
				orderErrorDuringFormattingCurrency.Message,
				zap.Float64("price", item.Amount),
				zap.String("locale", DefaultLanguage),
				zap.String("currency", item.Currency),
			)
			return nil, orderErrorDuringFormattingCurrency
		}

		items[i] = &billingpb.OrderReceiptItem{Name: item.Name, Price: price}
	}

	var platformName = ""

	if platform, ok := availablePlatforms[order.PlatformId]; ok {
		platformName = platform.Name
	}

	oc, err := s.operatingCompanyRepository.GetById(ctx, order.OperatingCompanyId)

	if err != nil {
		zap.L().Error(pkg.MethodFinishedWithError, zap.Error(err))
		return nil, err
	}

	receipt := &billingpb.OrderReceipt{
		TotalPrice:                 totalPrice,
		TransactionId:              order.Uuid,
		TransactionDate:            date,
		ProjectName:                order.Project.Name[DefaultLanguage],
		MerchantName:               merchant.GetCompanyName(),
		Items:                      items,
		OrderType:                  order.Type,
		PlatformName:               platformName,
		PaymentPartner:             oc.Name,
		VatPayer:                   order.VatPayer,
		VatInOrderCurrency:         vatInOrderCurrency,
		VatInChargeCurrency:        vatInChargeCurrency,
		TotalAmount:                totalAmount,
		TotalCharge:                totalCharge,
		ReceiptId:                  order.ReceiptId,
		Url:                        order.ReceiptUrl,
		VatRate:                    fmt.Sprintf("%.2f", order.Tax.Rate*100) + "%",
		CustomerEmail:              order.User.Email,
		CustomerUuid:               order.User.Uuid,
		SubscriptionViewUrl:        "",
		SubscriptionsManagementUrl: "",
	}

	subscription, err := s.rep.FindSubscriptions(ctx, &recurringpb.FindSubscriptionsRequest{
		MerchantId: order.Project.MerchantId,
		CustomerId: order.User.Id,
	})

	if err != nil {
		zap.L().Error(pkg.MethodFinishedWithError, zap.Error(err))
		return nil, err
	}

	if order.Recurring {
		receipt.SubscriptionViewUrl = fmt.Sprintf("%s/subscriptions/%s", s.cfg.CheckoutUrl, order.RecurringId)

		if order.RecurringSettings != nil {
			receipt.RecurringPeriod = order.RecurringSettings.Period
			receipt.RecurringInterval = fmt.Sprintf("%d", order.RecurringSettings.Interval)
			receipt.RecurringDateEnd = order.RecurringSettings.DateEnd
		}
	}

	if len(subscription.List) > 0 {
		receipt.SubscriptionsManagementUrl = fmt.Sprintf("%s/subscriptions", s.cfg.CheckoutUrl)
	}

	return receipt, nil
}

func (v *OrderCreateRequestProcessor) UserCountryExists() bool {
	return v.checked != nil && v.checked.user != nil && v.checked.user.Address != nil &&
		v.checked.user.Address.Country != ""
}

func intersect(a []string, b []string) []string {
	set := make([]string, 0)
	hash := make(map[string]bool)

	for _, v := range a {
		hash[v] = true
	}

	for _, v := range b {
		if _, found := hash[v]; found {
			set = append(set, v)
		}
	}

	return set
}

func (s *Service) hasPaymentCosts(ctx context.Context, order *billingpb.Order) bool {
	country, err := s.country.GetByIsoCodeA2(ctx, order.GetCountry())

	if err != nil {
		return false
	}

	methodName, err := order.GetCostPaymentMethodName()

	if err != nil {
		return false
	}

	_, err = s.paymentChannelCostSystemRepository.Find(
		ctx,
		methodName,
		country.PayerTariffRegion,
		country.IsoCodeA2,
		order.MccCode,
		order.OperatingCompanyId,
	)

	if err != nil {
		return false
	}

	data := &billingpb.PaymentChannelCostMerchantRequest{
		MerchantId:     order.GetMerchantId(),
		Name:           methodName,
		PayoutCurrency: order.GetMerchantRoyaltyCurrency(),
		Amount:         order.ChargeAmount,
		Region:         country.PayerTariffRegion,
		Country:        country.IsoCodeA2,
		MccCode:        order.MccCode,
	}
	_, err = s.getPaymentChannelCostMerchant(ctx, data)

	if err != nil {
		zap.L().Info("debug_1", zap.String("method", "PaymentChannelCostMerchantRequest"))
	}

	return err == nil
}

func (s *Service) paymentSystemPaymentCallbackComplete(ctx context.Context, order *billingpb.Order) error {
	ch := s.cfg.GetCentrifugoOrderChannel(order.Uuid)
	message := map[string]string{
		billingpb.PaymentCreateFieldOrderId: order.Uuid,
		"status":                            paymentSystemPaymentProcessingSuccessStatus,
	}

	return s.centrifugoPaymentForm.Publish(ctx, ch, message)
}

func (v *OrderCreateRequestProcessor) processVirtualCurrency(_ context.Context) error {
	amount := v.request.Amount
	virtualCurrency := v.checked.project.VirtualCurrency

	if virtualCurrency == nil || len(virtualCurrency.Prices) <= 0 {
		return orderErrorVirtualCurrencyNotFilled
	}

	_, frac := math.Modf(amount)

	if virtualCurrency.SellCountType == pkg.ProjectSellCountTypeIntegral && frac > 0 {
		return orderErrorVirtualCurrencyFracNotSupported
	}

	if amount < virtualCurrency.MinPurchaseValue ||
		(virtualCurrency.MaxPurchaseValue > 0 && amount > virtualCurrency.MaxPurchaseValue) {
		return orderErrorVirtualCurrencyLimits
	}

	v.checked.virtualAmount = amount
	return nil
}

func (s *Service) OrderReCreateProcess(
	ctx context.Context,
	req *billingpb.OrderReCreateProcessRequest,
	res *billingpb.OrderCreateProcessResponse,
) error {
	res.Status = billingpb.ResponseStatusOk

	order, err := s.orderRepository.GetByUuid(ctx, req.OrderId)
	if err != nil {
		zap.S().Errorw(pkg.ErrorGrpcServiceCallFailed, "err", err.Error(), "data", req)
		res.Status = billingpb.ResponseStatusNotFound
		res.Message = orderErrorUnknown
		return nil
	}

	if !order.CanBeRecreated() {
		res.Status = billingpb.ResponseStatusBadData
		res.Message = orderErrorWrongPrivateStatus
		return nil
	}

	newOrder := new(billingpb.Order)
	err = copier.Copy(&newOrder, &order)

	if err != nil {
		zap.S().Error(
			"Copy order to new structure order by refund failed",
			zap.Error(err),
			zap.Any("order", order),
		)

		res.Status = billingpb.ResponseStatusSystemError
		res.Message = orderErrorUnknown

		return nil
	}

	newOrder.PrivateStatus = recurringpb.OrderStatusNew
	newOrder.Status = recurringpb.OrderPublicStatusCreated
	newOrder.Id = primitive.NewObjectID().Hex()
	newOrder.Uuid = uuid.New().String()
	newOrder.ReceiptId = uuid.New().String()
	newOrder.CreatedAt = ptypes.TimestampNow()
	newOrder.UpdatedAt = ptypes.TimestampNow()
	newOrder.Canceled = false
	newOrder.CanceledAt = nil
	newOrder.ReceiptUrl = ""
	newOrder.PaymentMethod = nil

	newOrder.User = &billingpb.OrderUser{
		Id:            order.User.Id,
		Uuid:          order.User.Uuid,
		Phone:         order.User.Phone,
		PhoneVerified: order.User.PhoneVerified,
		Metadata:      order.Metadata,
		Object:        order.User.Object,
		Name:          order.User.Name,
		ExternalId:    order.User.ExternalId,
	}

	if err = s.orderRepository.Insert(ctx, newOrder); err != nil {
		res.Status = billingpb.ResponseStatusBadData
		res.Message = orderErrorCanNotCreate
		return nil
	}

	res.Item = newOrder

	return nil
}

func (s *Service) getAddressByIp(ctx context.Context, ip string) (order *billingpb.OrderBillingAddress, err error) {
	rsp, err := s.geo.GetIpData(ctx, &geoip.GeoIpDataRequest{IP: ip})
	if err != nil {
		zap.L().Error(
			"GetIpData failed",
			zap.Error(err),
			zap.String("ip", ip),
		)

		return nil, orderErrorPayerRegionUnknown
	}

	address := &billingpb.OrderBillingAddress{
		Country: rsp.Country.IsoCode,
		City:    rsp.City.Names["en"],
	}

	if rsp.Postal != nil {
		address.PostalCode = rsp.Postal.Code
	}

	if len(rsp.Subdivisions) > 0 {
		address.State = rsp.Subdivisions[0].IsoCode
	}

	return address, nil
}

func (s *Service) getOrderPriceGroup(ctx context.Context, order *billingpb.Order) (priceGroup *billingpb.PriceGroup, err error) {
	if order.IsCurrencyPredefined {
		priceGroup, err = s.priceGroupRepository.GetByRegion(ctx, order.Currency)
		return
	}

	merchant, err := s.merchantRepository.GetById(ctx, order.GetMerchantId())
	if err != nil {
		return
	}

	defaultPriceGroup, err := s.priceGroupRepository.GetByRegion(ctx, merchant.GetProcessingDefaultCurrency())

	countryCode := order.GetCountry()
	if countryCode == "" {
		return defaultPriceGroup, nil
	}

	country, err := s.country.GetByIsoCodeA2(ctx, countryCode)
	if err != nil {
		return defaultPriceGroup, nil
	}

	priceGroup, err = s.priceGroupRepository.GetById(ctx, country.PriceGroupId)
	return
}

func (s *Service) setOrderChargeAmountAndCurrency(ctx context.Context, order *billingpb.Order) (err error) {
	order.ChargeAmount = order.TotalPaymentAmount
	order.ChargeCurrency = order.Currency

	if order.PaymentRequisites == nil {
		return nil
	}

	if order.PaymentMethod == nil {
		return nil
	}

	binCountryCode, ok := order.PaymentRequisites[billingpb.PaymentCreateBankCardFieldIssuerCountryIsoCode]
	if !ok || binCountryCode == "" {
		return nil
	}

	binCardBrand, ok := order.PaymentRequisites[billingpb.PaymentCreateBankCardFieldBrand]
	if !ok || binCardBrand == "" {
		return nil
	}

	if order.PaymentIpCountry != "" {
		order.IsIpCountryMismatchBin = order.PaymentIpCountry != binCountryCode
	}

	binCountry, err := s.country.GetByIsoCodeA2(ctx, binCountryCode)
	if err != nil {
		return err
	}
	if binCountry.Currency == order.Currency {
		return nil
	}

	sCurr, err := s.curService.GetPriceCurrencies(ctx, &currenciespb.EmptyRequest{})
	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(errorFieldService, "CurrencyRatesService"),
			zap.String(errorFieldMethod, "GetPriceCurrencies"),
			zap.Any(errorFieldEntrySource, order.Id),
		)
		return err
	}
	if !helper.Contains(sCurr.Currencies, binCountry.Currency) {
		return nil
	}

	// check that we have terminal in payment method for bin country currency
	pm, err := s.paymentMethodRepository.GetById(ctx, order.PaymentMethod.Id)
	if err != nil {
		return nil
	}

	_, err = s.getPaymentSettings(pm, binCountry.Currency, order.MccCode, order.OperatingCompanyId, binCardBrand, order.IsProduction)
	if err != nil {
		return nil
	}

	reqCur := &currenciespb.ExchangeCurrencyCurrentCommonRequest{
		From:              order.Currency,
		To:                binCountry.Currency,
		RateType:          currenciespb.RateTypePaysuper,
		Amount:            order.TotalPaymentAmount,
		ExchangeDirection: currenciespb.ExchangeDirectionSell,
	}

	rspCur, err := s.curService.ExchangeCurrencyCurrentCommon(ctx, reqCur)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(errorFieldService, "CurrencyRatesService"),
			zap.String(errorFieldMethod, "ExchangeCurrencyCurrentCommon"),
			zap.Any(errorFieldRequest, reqCur),
			zap.Any(errorFieldEntrySource, order.Id),
		)

		return orderErrorConvertionCurrency
	}

	order.ChargeCurrency = binCountry.Currency
	order.ChargeAmount = s.FormatAmount(rspCur.ExchangedAmount, binCountry.Currency)

	return nil
}

func (s *Service) checkVirtualCurrencyProduct(products []*billingpb.Product) bool {
	if len(products) == 0 {
		return false
	}

	for _, product := range products {
		if len(product.Prices) != 1 {
			return false
		}
		if product.Prices[0].IsVirtualCurrency == false {
			return false
		}
	}

	return true
}

func (s *Service) processProducts(
	ctx context.Context,
	projectId string,
	productIds []string,
	priceGroup *billingpb.PriceGroup,
	locale string,
) (amount float64, usedPriceGroup *billingpb.PriceGroup, items []*billingpb.OrderItem, isBuyForVirtualCurrency bool, err error) {
	project, err := s.project.GetById(ctx, projectId)
	if err != nil {
		return
	}
	if project.IsDeleted() == true {
		err = orderErrorProjectInactive
		return
	}

	orderProducts, err := s.GetOrderProducts(ctx, project.Id, productIds)
	if err != nil {
		return
	}

	merchant, err := s.merchantRepository.GetById(ctx, project.MerchantId)
	if err != nil {
		return
	}

	defaultPriceGroup, err := s.priceGroupRepository.GetByRegion(ctx, merchant.GetProcessingDefaultCurrency())
	if err != nil {
		return
	}

	if priceGroup == nil {
		priceGroup = defaultPriceGroup
	}

	isBuyForVirtualCurrency = s.checkVirtualCurrencyProduct(orderProducts)

	if isBuyForVirtualCurrency {
		amount, usedPriceGroup, err = s.processAmountForVirtualCurrency(ctx, project, orderProducts, priceGroup, defaultPriceGroup)
	} else {
		amount, usedPriceGroup, err = s.processAmountForFiatCurrency(ctx, project, orderProducts, priceGroup, defaultPriceGroup)
	}

	if err != nil {
		return
	}

	amount = s.FormatAmount(amount, usedPriceGroup.Currency)

	if isBuyForVirtualCurrency {
		items, err = s.GetOrderProductsItems(orderProducts, locale, &billingpb.PriceGroup{Currency: billingpb.VirtualCurrencyPriceGroup})
	} else {
		items, err = s.GetOrderProductsItems(orderProducts, locale, usedPriceGroup)
	}

	return
}

func (s *Service) processKeyProducts(
	ctx context.Context,
	projectId string,
	productIds []string,
	priceGroup *billingpb.PriceGroup,
	locale string,
	platformId string,
) (amount float64, usedPriceGroup *billingpb.PriceGroup, items []*billingpb.OrderItem, platforms []*billingpb.Platform, err error) {

	project, err := s.project.GetById(ctx, projectId)
	if err != nil {
		return
	}
	if project.IsDeleted() == true {
		err = orderErrorProjectInactive
		return
	}

	orderProducts, err := s.GetOrderKeyProducts(ctx, project.Id, productIds)
	if err != nil {
		return
	}

	platformIds := s.filterPlatforms(orderProducts)
	if len(platformIds) == 0 {
		zap.L().Error("No available platformIds")
		err = orderErrorNoPlatforms
		return
	}

	platforms = make([]*billingpb.Platform, len(platformIds))
	for i, v := range platformIds {
		platforms[i] = availablePlatforms[v]
	}
	sort.Slice(platforms, func(i, j int) bool {
		return platforms[i].Order < platforms[j].Order
	})

	if platformId == "" {
		platformId = platforms[0].Id
	}

	merchant, err := s.merchantRepository.GetById(ctx, project.MerchantId)
	if err != nil {
		return
	}

	defaultPriceGroup, err := s.priceGroupRepository.GetByRegion(ctx, merchant.GetProcessingDefaultCurrency())
	if err != nil {
		return
	}

	if priceGroup == nil {
		priceGroup = defaultPriceGroup
	}

	usedPriceGroup = priceGroup

	amount, err = s.GetOrderKeyProductsAmount(orderProducts, priceGroup, platformId)
	if err != nil {
		if err != orderErrorNoProductsCommonCurrency {
			return
		} else {

			if priceGroup.Id == defaultPriceGroup.Id {
				return
			}

			usedPriceGroup = defaultPriceGroup

			// try to get order Amount in fallback currency
			amount, err = s.GetOrderKeyProductsAmount(orderProducts, defaultPriceGroup, platformId)
			if err != nil {
				return
			}
		}
	}

	amount = s.FormatAmount(amount, usedPriceGroup.Currency)

	items, err = s.GetOrderKeyProductsItems(orderProducts, locale, usedPriceGroup, platformId)

	return
}

func (s *Service) addRecurringSubscription(
	ctx context.Context, order *billingpb.Order, h payment_system.PaymentSystemInterface, data map[string]string,
) (*recurringpb.Subscription, string, error) {
	maskedPan := ""

	if order.PaymentMethod.ExternalId == recurringpb.PaymentSystemGroupAliasBankCard {
		maskedPan = stringTools.MaskBankCardNumber(data[billingpb.PaymentCreateFieldPan])
	}

	expireAt, _ := time.Parse(billingpb.FilterDateFormat, order.RecurringSettings.DateEnd)
	expireAt = time.Date(expireAt.Year(), expireAt.Month(), expireAt.Day(), 23, 59, 59, 0, expireAt.Location())
	tsExpireAt, _ := ptypes.TimestampProto(expireAt)

	subscription := &recurringpb.Subscription{
		OrderId:      order.Id,
		MerchantId:   order.Project.MerchantId,
		ProjectId:    order.Project.Id,
		CustomerId:   order.User.Id,
		CustomerUuid: order.User.Uuid,
		MaskedPan:    maskedPan,
		CustomerInfo: &recurringpb.CustomerInfo{
			ExternalId: order.User.ExternalId,
			Email:      order.User.Email,
			Phone:      order.User.Phone,
		},
		IsActive:    false,
		Period:      order.RecurringSettings.Period,
		ExpireAt:    tsExpireAt,
		Amount:      order.ChargeAmount,
		Currency:    order.ChargeCurrency,
		ProjectName: order.Project.Name,
	}

	subscription.ItemType = order.ProductType

	for _, item := range order.Items {
		subscription.ItemList = append(subscription.ItemList, item.Id)
	}

	res, err := s.rep.AddSubscription(ctx, subscription)

	if err != nil || res.Status != billingpb.ResponseStatusOk {
		zap.L().Error(
			"Unable to add recurring subscription",
			zap.Error(err),
			zap.Any("subscription", subscription),
		)
		return nil, "", orderErrorRecurringUnableToAdd
	}

	subscription.Id = res.SubscriptionId

	url, err := h.CreateRecurringSubscription(
		order,
		subscription,
		s.cfg.GetRedirectUrlSuccess(nil),
		s.cfg.GetRedirectUrlFail(nil),
		data,
	)

	if err != nil {
		zap.L().Error(
			"h.CreateRecurringSubscription Method failed",
			zap.Error(err),
			zap.Any("order", order),
		)
		return nil, "", err
	}

	resUpdate, err := s.rep.UpdateSubscription(ctx, subscription)

	if err != nil || resUpdate.Status != billingpb.ResponseStatusOk {
		zap.L().Error(
			"Unable to update recurring subscription",
			zap.Error(err),
			zap.Any("subscription", subscription),
		)
		return nil, "", orderErrorRecurringUnableToUpdate
	}

	return subscription, url, nil
}

// Set caption for redirect button in payment form
func (m *orderCreateRequestProcessorChecked) setRedirectButtonCaption(caption string) {
	m.project.RedirectSettings.ButtonCaption = caption
}

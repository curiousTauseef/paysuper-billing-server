package service

import (
	"context"
	"errors"
	"github.com/paysuper/paysuper-billing-server/internal/config"
	"github.com/paysuper/paysuper-billing-server/internal/database"
	"github.com/paysuper/paysuper-billing-server/internal/mocks"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	casbinMocks "github.com/paysuper/paysuper-proto/go/casbinpb/mocks"
	"github.com/paysuper/paysuper-proto/go/recurringpb"
	recurringMocks "github.com/paysuper/paysuper-proto/go/recurringpb/mocks"
	reportingMocks "github.com/paysuper/paysuper-proto/go/reporterpb/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
	"testing"
)

type RecurringTestSuite struct {
	suite.Suite
	service *Service
}

func Test_Recurring(t *testing.T) {
	suite.Run(t, new(RecurringTestSuite))
}

func (suite *RecurringTestSuite) SetupTest() {
	cfg, err := config.NewConfig()
	assert.NoError(suite.T(), err, "Config load failed")

	db, err := mongodb.NewDatabase()
	assert.NoError(suite.T(), err, "Database connection failed")

	redisdb := mocks.NewTestRedis()
	cache, err := database.NewCacheRedis(redisdb, "cache")

	if err != nil {
		suite.FailNow("Cache redis initialize failed", "%v", err)
	}

	casbin := &casbinMocks.CasbinService{}

	suite.service = NewBillingService(
		db,
		cfg,
		mocks.NewGeoIpServiceTestOk(),
		mocks.NewRepositoryServiceOk(),
		&mocks.TaxServiceOkMock{},
		mocks.NewBrokerMockOk(),
		nil,
		cache,
		mocks.NewCurrencyServiceMockOk(),
		mocks.NewDocumentSignerMockOk(),
		&reportingMocks.ReporterService{},
		mocks.NewFormatterOK(),
		mocks.NewBrokerMockOk(),
		casbin,
		mocks.NewNotifierOk(),
		mocks.NewBrokerMockOk(),
	)

	if err := suite.service.Init(); err != nil {
		suite.FailNow("Billing service initialization failed", "%v", err)
	}
}

func (suite *RecurringTestSuite) TearDownTest() {
	err := suite.service.db.Drop()

	if err != nil {
		suite.FailNow("Database deletion failed", "%v", err)
	}

	err = suite.service.db.Close()

	if err != nil {
		suite.FailNow("Database close failed", "%v", err)
	}
}

func (suite *RecurringTestSuite) TestRecurring_DeleteSavedCard_Ok() {
	customer := &BrowserCookieCustomer{
		VirtualCustomerId: primitive.NewObjectID().Hex(),
		Ip:                "127.0.0.1",
		AcceptLanguage:    "fr-CA",
		UserAgent:         "windows",
		SessionCount:      0,
	}
	cookie, err := suite.service.generateBrowserCookie(customer)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), cookie)

	req := &billingpb.DeleteSavedCardRequest{
		Id:     primitive.NewObjectID().Hex(),
		Cookie: cookie,
	}
	rsp := &billingpb.EmptyResponseWithStatus{}
	err = suite.service.DeleteSavedCard(context.TODO(), req, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusOk, rsp.Status)
	assert.Empty(suite.T(), rsp.Message)
}

func (suite *RecurringTestSuite) TestRecurring_DeleteSavedCard_IncorrectCookie_Error() {
	req := &billingpb.DeleteSavedCardRequest{
		Id:     primitive.NewObjectID().Hex(),
		Cookie: primitive.NewObjectID().Hex(),
	}
	rsp := &billingpb.EmptyResponseWithStatus{}
	err := suite.service.DeleteSavedCard(context.TODO(), req, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusBadData, rsp.Status)
	assert.Equal(suite.T(), recurringErrorIncorrectCookie, rsp.Message)
}

func (suite *RecurringTestSuite) TestRecurring_DeleteSavedCard_DontHaveCustomerId_Error() {
	customer := &BrowserCookieCustomer{
		Ip:             "127.0.0.1",
		AcceptLanguage: "fr-CA",
		UserAgent:      "windows",
		SessionCount:   0,
	}
	cookie, err := suite.service.generateBrowserCookie(customer)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), cookie)

	req := &billingpb.DeleteSavedCardRequest{
		Id:     primitive.NewObjectID().Hex(),
		Cookie: cookie,
	}
	rsp := &billingpb.EmptyResponseWithStatus{}
	err = suite.service.DeleteSavedCard(context.TODO(), req, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusNotFound, rsp.Status)
	assert.Equal(suite.T(), recurringCustomerNotFound, rsp.Message)
}

func (suite *RecurringTestSuite) TestRecurring_DeleteSavedCard_RealCustomer_Ok() {
	project := &billingpb.Project{
		Id:         primitive.NewObjectID().Hex(),
		MerchantId: primitive.NewObjectID().Hex(),
	}
	req0 := &billingpb.TokenRequest{
		User: &billingpb.TokenUser{
			Id: primitive.NewObjectID().Hex(),
			Locale: &billingpb.TokenUserLocaleValue{
				Value: "en",
			},
		},
		Settings: &billingpb.TokenSettings{
			ProjectId: project.Id,
			Amount:    100,
			Currency:  "USD",
			Type:      pkg.OrderType_simple,
		},
	}
	customer, err := suite.service.createCustomer(context.TODO(), req0, project)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), customer)

	browserCustomer := &BrowserCookieCustomer{
		CustomerId:     customer.Id,
		Ip:             "127.0.0.1",
		AcceptLanguage: "fr-CA",
		UserAgent:      "windows",
		SessionCount:   0,
	}
	cookie, err := suite.service.generateBrowserCookie(browserCustomer)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), cookie)

	req := &billingpb.DeleteSavedCardRequest{
		Id:     primitive.NewObjectID().Hex(),
		Cookie: cookie,
	}
	rsp := &billingpb.EmptyResponseWithStatus{}
	err = suite.service.DeleteSavedCard(context.TODO(), req, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusOk, rsp.Status)
	assert.Empty(suite.T(), rsp.Message)
}

func (suite *RecurringTestSuite) TestRecurring_DeleteSavedCard_RealCustomerNotFound_Error() {
	browserCustomer := &BrowserCookieCustomer{
		CustomerId:     primitive.NewObjectID().Hex(),
		Ip:             "127.0.0.1",
		AcceptLanguage: "fr-CA",
		UserAgent:      "windows",
		SessionCount:   0,
	}
	cookie, err := suite.service.generateBrowserCookie(browserCustomer)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), cookie)

	req := &billingpb.DeleteSavedCardRequest{
		Id:     primitive.NewObjectID().Hex(),
		Cookie: cookie,
	}
	rsp := &billingpb.EmptyResponseWithStatus{}
	err = suite.service.DeleteSavedCard(context.TODO(), req, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusNotFound, rsp.Status)
	assert.Equal(suite.T(), recurringCustomerNotFound, rsp.Message)
}

func (suite *RecurringTestSuite) TestRecurring_DeleteSavedCard_RecurringServiceSystem_Error() {
	browserCustomer := &BrowserCookieCustomer{
		VirtualCustomerId: primitive.NewObjectID().Hex(),
		Ip:                "127.0.0.1",
		AcceptLanguage:    "fr-CA",
		UserAgent:         "windows",
		SessionCount:      0,
	}
	cookie, err := suite.service.generateBrowserCookie(browserCustomer)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), cookie)

	suite.service.rep = mocks.NewRepositoryServiceError()

	req := &billingpb.DeleteSavedCardRequest{
		Id:     primitive.NewObjectID().Hex(),
		Cookie: cookie,
	}
	rsp := &billingpb.EmptyResponseWithStatus{}
	err = suite.service.DeleteSavedCard(context.TODO(), req, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusSystemError, rsp.Status)
	assert.Equal(suite.T(), recurringErrorUnknown, rsp.Message)
}

func (suite *RecurringTestSuite) TestRecurring_DeleteSavedCard_RecurringServiceResult_Error() {
	browserCustomer := &BrowserCookieCustomer{
		VirtualCustomerId: primitive.NewObjectID().Hex(),
		Ip:                "127.0.0.1",
		AcceptLanguage:    "fr-CA",
		UserAgent:         "windows",
		SessionCount:      0,
	}
	cookie, err := suite.service.generateBrowserCookie(browserCustomer)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), cookie)

	suite.service.rep = mocks.NewRepositoryServiceEmpty()

	req := &billingpb.DeleteSavedCardRequest{
		Id:     primitive.NewObjectID().Hex(),
		Cookie: cookie,
	}
	rsp := &billingpb.EmptyResponseWithStatus{}
	err = suite.service.DeleteSavedCard(context.TODO(), req, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusBadData, rsp.Status)
	assert.Equal(suite.T(), recurringSavedCardNotFount, rsp.Message)
}

func (suite *RecurringTestSuite) TestRecurring_DeleteSavedCard_RecurringServiceResultSystemError_Error() {
	browserCustomer := &BrowserCookieCustomer{
		VirtualCustomerId: "ffffffffffffffffffffffff",
		Ip:                "127.0.0.1",
		AcceptLanguage:    "fr-CA",
		UserAgent:         "windows",
		SessionCount:      0,
	}
	cookie, err := suite.service.generateBrowserCookie(browserCustomer)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), cookie)

	suite.service.rep = mocks.NewRepositoryServiceEmpty()

	req := &billingpb.DeleteSavedCardRequest{
		Id:     primitive.NewObjectID().Hex(),
		Cookie: cookie,
	}
	rsp := &billingpb.EmptyResponseWithStatus{}
	err = suite.service.DeleteSavedCard(context.TODO(), req, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusSystemError, rsp.Status)
	assert.Equal(suite.T(), recurringErrorUnknown, rsp.Message)
}

func (suite *RecurringTestSuite) TestOrder_DeleteRecurringSubscription_WithoutCookie_Ok() {
	var (
		customerId = "customer_id"
		orderId    = "order_id"
		psId       = "payment_system_id"
		psHandler  = "payment_system_handler"
	)

	subscription := &recurringpb.Subscription{
		OrderId:    "order_id",
		CustomerId: customerId,
	}
	order := &billingpb.Order{
		PaymentMethod: &billingpb.PaymentMethodOrder{
			PaymentSystemId: psId,
		},
	}

	recurring := &recurringMocks.RepositoryService{}
	recurring.On("GetSubscription", mock.Anything, mock.Anything).Return(&recurringpb.GetSubscriptionResponse{
		Status:       billingpb.ResponseStatusOk,
		Subscription: subscription,
	}, nil)
	recurring.On("DeleteSubscription", mock.Anything, subscription).Return(&recurringpb.DeleteSubscriptionResponse{
		Status: billingpb.ResponseStatusOk,
	}, nil)
	suite.service.rep = recurring

	orderRepository := &mocks.OrderRepositoryInterface{}
	orderRepository.On("GetById", mock.Anything, orderId).Return(order, nil)
	suite.service.orderRepository = orderRepository

	psRepository := &mocks.PaymentSystemRepositoryInterface{}
	psRepository.On("GetById", mock.Anything, psId).Return(&billingpb.PaymentSystem{
		Handler: psHandler,
	}, nil)
	suite.service.paymentSystemRepository = psRepository

	paymentSystem := &mocks.PaymentSystemInterface{}
	paymentSystem.On("DeleteRecurringSubscription", order, subscription).Return(nil)

	gatewayManagerMock := &mocks.PaymentSystemManagerInterface{}
	gatewayManagerMock.On("GetGateway", psHandler).Return(paymentSystem, nil)
	suite.service.paymentSystemGateway = gatewayManagerMock

	rsp := &billingpb.EmptyResponseWithStatus{}
	err := suite.service.DeleteRecurringSubscription(context.Background(), &billingpb.DeleteRecurringSubscriptionRequest{
		SubscriptionId: order.RecurringId,
		CustomerId:     customerId,
	}, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusOk, rsp.Status)
}

func (suite *RecurringTestSuite) TestOrder_DeleteRecurringSubscription_WithCookie_Ok() {
	var (
		customerId = "customer_id"
		orderId    = "order_id"
		psId       = "payment_system_id"
		psHandler  = "payment_system_handler"
	)

	browserCustomer := &BrowserCookieCustomer{
		CustomerId:     customerId,
		Ip:             "127.0.0.1",
		AcceptLanguage: "fr-CA",
		UserAgent:      "windows",
		SessionCount:   0,
	}
	cookie, err := suite.service.generateBrowserCookie(browserCustomer)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), cookie)

	subscription := &recurringpb.Subscription{
		OrderId:    "order_id",
		CustomerId: customerId,
	}
	order := &billingpb.Order{
		PaymentMethod: &billingpb.PaymentMethodOrder{
			PaymentSystemId: psId,
		},
	}

	recurring := &recurringMocks.RepositoryService{}
	recurring.On("GetSubscription", mock.Anything, mock.Anything).Return(&recurringpb.GetSubscriptionResponse{
		Status:       billingpb.ResponseStatusOk,
		Subscription: subscription,
	}, nil)
	recurring.On("DeleteSubscription", mock.Anything, subscription).Return(&recurringpb.DeleteSubscriptionResponse{
		Status: billingpb.ResponseStatusOk,
	}, nil)
	suite.service.rep = recurring

	orderRepository := &mocks.OrderRepositoryInterface{}
	orderRepository.On("GetById", mock.Anything, orderId).Return(order, nil)
	suite.service.orderRepository = orderRepository

	psRepository := &mocks.PaymentSystemRepositoryInterface{}
	psRepository.On("GetById", mock.Anything, psId).Return(&billingpb.PaymentSystem{
		Handler: psHandler,
	}, nil)
	suite.service.paymentSystemRepository = psRepository

	paymentSystem := &mocks.PaymentSystemInterface{}
	paymentSystem.On("DeleteRecurringSubscription", order, subscription).Return(nil)

	gatewayManagerMock := &mocks.PaymentSystemManagerInterface{}
	gatewayManagerMock.On("GetGateway", psHandler).Return(paymentSystem, nil)
	suite.service.paymentSystemGateway = gatewayManagerMock

	rsp := &billingpb.EmptyResponseWithStatus{}
	err = suite.service.DeleteRecurringSubscription(context.Background(), &billingpb.DeleteRecurringSubscriptionRequest{
		SubscriptionId: order.RecurringId,
		Cookie:         cookie,
	}, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusOk, rsp.Status)
}

func (suite *RecurringTestSuite) TestOrder_DeleteRecurringSubscription_BadCookie_Error() {
	rsp := &billingpb.EmptyResponseWithStatus{}
	err := suite.service.DeleteRecurringSubscription(context.Background(), &billingpb.DeleteRecurringSubscriptionRequest{
		SubscriptionId: "recurring_id",
		Cookie:         "cookie",
	}, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusForbidden, rsp.Status)
	assert.Equal(suite.T(), recurringCustomerNotFound, rsp.Message)
}

func (suite *RecurringTestSuite) TestOrder_DeleteRecurringSubscription_NoCustomerOnCookie_Error() {
	browserCustomer := &BrowserCookieCustomer{
		VirtualCustomerId: "customerId",
		Ip:                "127.0.0.1",
		AcceptLanguage:    "fr-CA",
		UserAgent:         "windows",
		SessionCount:      0,
	}
	cookie, err := suite.service.generateBrowserCookie(browserCustomer)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), cookie)

	rsp := &billingpb.EmptyResponseWithStatus{}
	err = suite.service.DeleteRecurringSubscription(context.Background(), &billingpb.DeleteRecurringSubscriptionRequest{
		SubscriptionId: "recurring_id",
		Cookie:         cookie,
	}, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusBadData, rsp.Status)
	assert.Equal(suite.T(), recurringCustomerNotFound, rsp.Message)
}

func (suite *RecurringTestSuite) TestOrder_DeleteRecurringSubscription_SubscriptionNotFound_Error() {
	recurring := &recurringMocks.RepositoryService{}
	recurring.On("GetSubscription", mock.Anything, mock.Anything).Return(&recurringpb.GetSubscriptionResponse{
		Status: billingpb.ResponseStatusNotFound,
	}, nil)
	suite.service.rep = recurring

	rsp := &billingpb.EmptyResponseWithStatus{}
	err := suite.service.DeleteRecurringSubscription(context.Background(), &billingpb.DeleteRecurringSubscriptionRequest{
		SubscriptionId: "recurring_id",
		CustomerId:     "customer_id",
	}, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusNotFound, rsp.Status)
	assert.Equal(suite.T(), orderErrorRecurringSubscriptionNotFound, rsp.Message)
}

func (suite *RecurringTestSuite) TestOrder_DeleteRecurringSubscription_AccessDeny_Error() {
	recurring := &recurringMocks.RepositoryService{}
	recurring.On("GetSubscription", mock.Anything, mock.Anything).Return(&recurringpb.GetSubscriptionResponse{
		Status:       billingpb.ResponseStatusOk,
		Subscription: &recurringpb.Subscription{CustomerId: "customer_id2"},
	}, nil)
	suite.service.rep = recurring

	rsp := &billingpb.EmptyResponseWithStatus{}
	err := suite.service.DeleteRecurringSubscription(context.Background(), &billingpb.DeleteRecurringSubscriptionRequest{
		SubscriptionId: "recurring_id",
		CustomerId:     "customer_id",
	}, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusForbidden, rsp.Status)
	assert.Equal(suite.T(), recurringCustomerNotFound, rsp.Message)
}

func (suite *RecurringTestSuite) TestOrder_DeleteRecurringSubscription_OrderNotFound_Error() {
	var (
		customerId = "customer_id"
		orderId    = "order_id"
	)

	subscription := &recurringpb.Subscription{
		OrderId:    "order_id",
		CustomerId: customerId,
	}

	recurring := &recurringMocks.RepositoryService{}
	recurring.On("GetSubscription", mock.Anything, mock.Anything).Return(&recurringpb.GetSubscriptionResponse{
		Status:       billingpb.ResponseStatusOk,
		Subscription: subscription,
	}, nil)
	suite.service.rep = recurring

	orderRepository := &mocks.OrderRepositoryInterface{}
	orderRepository.On("GetById", mock.Anything, orderId).Return(nil, errors.New("notfound"))
	suite.service.orderRepository = orderRepository

	rsp := &billingpb.EmptyResponseWithStatus{}
	err := suite.service.DeleteRecurringSubscription(context.Background(), &billingpb.DeleteRecurringSubscriptionRequest{
		SubscriptionId: "recurring_id",
		CustomerId:     customerId,
	}, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusNotFound, rsp.Status)
	assert.Equal(suite.T(), orderErrorNotFound, rsp.Message)
}

func (suite *RecurringTestSuite) TestOrder_DeleteRecurringSubscription_PaymentSystemNotFound_Error() {
	var (
		customerId = "customer_id"
		orderId    = "order_id"
		psId       = "payment_system_id"
	)

	subscription := &recurringpb.Subscription{
		OrderId:    "order_id",
		CustomerId: customerId,
	}
	order := &billingpb.Order{
		PaymentMethod: &billingpb.PaymentMethodOrder{
			PaymentSystemId: psId,
		},
	}

	recurring := &recurringMocks.RepositoryService{}
	recurring.On("GetSubscription", mock.Anything, mock.Anything).Return(&recurringpb.GetSubscriptionResponse{
		Status:       billingpb.ResponseStatusOk,
		Subscription: subscription,
	}, nil)
	suite.service.rep = recurring

	orderRepository := &mocks.OrderRepositoryInterface{}
	orderRepository.On("GetById", mock.Anything, orderId).Return(order, nil)
	suite.service.orderRepository = orderRepository

	psRepository := &mocks.PaymentSystemRepositoryInterface{}
	psRepository.On("GetById", mock.Anything, psId).Return(nil, errors.New("notfound"))
	suite.service.paymentSystemRepository = psRepository

	rsp := &billingpb.EmptyResponseWithStatus{}
	err := suite.service.DeleteRecurringSubscription(context.Background(), &billingpb.DeleteRecurringSubscriptionRequest{
		SubscriptionId: "recurring_id",
		CustomerId:     customerId,
	}, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusNotFound, rsp.Status)
	assert.Equal(suite.T(), orderErrorPaymentSystemInactive, rsp.Message)
}

func (suite *RecurringTestSuite) TestOrder_DeleteRecurringSubscription_PaymentSystemGatewayNotFound_Error() {
	var (
		customerId = "customer_id"
		orderId    = "order_id"
		psId       = "payment_system_id"
		psHandler  = "payment_system_handler"
	)

	subscription := &recurringpb.Subscription{
		OrderId:    "order_id",
		CustomerId: customerId,
	}
	order := &billingpb.Order{
		PaymentMethod: &billingpb.PaymentMethodOrder{
			PaymentSystemId: psId,
		},
	}

	recurring := &recurringMocks.RepositoryService{}
	recurring.On("GetSubscription", mock.Anything, mock.Anything).Return(&recurringpb.GetSubscriptionResponse{
		Status:       billingpb.ResponseStatusOk,
		Subscription: subscription,
	}, nil)
	suite.service.rep = recurring

	orderRepository := &mocks.OrderRepositoryInterface{}
	orderRepository.On("GetById", mock.Anything, orderId).Return(order, nil)
	suite.service.orderRepository = orderRepository

	psRepository := &mocks.PaymentSystemRepositoryInterface{}
	psRepository.On("GetById", mock.Anything, psId).Return(&billingpb.PaymentSystem{
		Handler: psHandler,
	}, nil)
	suite.service.paymentSystemRepository = psRepository

	paymentSystem := &mocks.PaymentSystemInterface{}
	paymentSystem.On("DeleteRecurringSubscription", order, subscription).Return(nil)

	gatewayManagerMock := &mocks.PaymentSystemManagerInterface{}
	gatewayManagerMock.On("GetGateway", psHandler).Return(nil, errors.New("notfound"))
	suite.service.paymentSystemGateway = gatewayManagerMock

	rsp := &billingpb.EmptyResponseWithStatus{}
	err := suite.service.DeleteRecurringSubscription(context.Background(), &billingpb.DeleteRecurringSubscriptionRequest{
		SubscriptionId: "recurring_id",
		CustomerId:     customerId,
	}, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusSystemError, rsp.Status)
	assert.Equal(suite.T(), orderErrorPaymentSystemInactive, rsp.Message)
}

func (suite *RecurringTestSuite) TestOrder_DeleteRecurringSubscription_DeleteSubscriptionOnPaymentSystem_Error() {
	var (
		customerId = "customer_id"
		orderId    = "order_id"
		psId       = "payment_system_id"
		psHandler  = "payment_system_handler"
	)

	subscription := &recurringpb.Subscription{
		OrderId:    "order_id",
		CustomerId: customerId,
	}
	order := &billingpb.Order{
		PaymentMethod: &billingpb.PaymentMethodOrder{
			PaymentSystemId: psId,
		},
	}

	recurring := &recurringMocks.RepositoryService{}
	recurring.On("GetSubscription", mock.Anything, mock.Anything).Return(&recurringpb.GetSubscriptionResponse{
		Status:       billingpb.ResponseStatusOk,
		Subscription: subscription,
	}, nil)
	suite.service.rep = recurring

	orderRepository := &mocks.OrderRepositoryInterface{}
	orderRepository.On("GetById", mock.Anything, orderId).Return(order, nil)
	suite.service.orderRepository = orderRepository

	psRepository := &mocks.PaymentSystemRepositoryInterface{}
	psRepository.On("GetById", mock.Anything, psId).Return(&billingpb.PaymentSystem{
		Handler: psHandler,
	}, nil)
	suite.service.paymentSystemRepository = psRepository

	paymentSystem := &mocks.PaymentSystemInterface{}
	paymentSystem.On("DeleteRecurringSubscription", order, subscription).Return(errors.New("error"))

	gatewayManagerMock := &mocks.PaymentSystemManagerInterface{}
	gatewayManagerMock.On("GetGateway", psHandler).Return(paymentSystem, nil)
	suite.service.paymentSystemGateway = gatewayManagerMock

	rsp := &billingpb.EmptyResponseWithStatus{}
	err := suite.service.DeleteRecurringSubscription(context.Background(), &billingpb.DeleteRecurringSubscriptionRequest{
		SubscriptionId: order.RecurringId,
		CustomerId:     customerId,
	}, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusSystemError, rsp.Status)
	assert.Equal(suite.T(), recurringErrorDeleteSubscription, rsp.Message)
}

func (suite *RecurringTestSuite) TestOrder_DeleteRecurringSubscription_DeleteFromRepository_Error() {
	var (
		customerId = "customer_id"
		orderId    = "order_id"
		psId       = "payment_system_id"
		psHandler  = "payment_system_handler"
	)

	subscription := &recurringpb.Subscription{
		OrderId:    "order_id",
		CustomerId: customerId,
	}
	order := &billingpb.Order{
		PaymentMethod: &billingpb.PaymentMethodOrder{
			PaymentSystemId: psId,
		},
	}

	recurring := &recurringMocks.RepositoryService{}
	recurring.On("GetSubscription", mock.Anything, mock.Anything).Return(&recurringpb.GetSubscriptionResponse{
		Status:       billingpb.ResponseStatusOk,
		Subscription: subscription,
	}, nil)
	recurring.On("DeleteSubscription", mock.Anything, subscription).Return(&recurringpb.DeleteSubscriptionResponse{
		Status: billingpb.ResponseStatusSystemError,
	}, nil)
	suite.service.rep = recurring

	orderRepository := &mocks.OrderRepositoryInterface{}
	orderRepository.On("GetById", mock.Anything, orderId).Return(order, nil)
	suite.service.orderRepository = orderRepository

	psRepository := &mocks.PaymentSystemRepositoryInterface{}
	psRepository.On("GetById", mock.Anything, psId).Return(&billingpb.PaymentSystem{
		Handler: psHandler,
	}, nil)
	suite.service.paymentSystemRepository = psRepository

	paymentSystem := &mocks.PaymentSystemInterface{}
	paymentSystem.On("DeleteRecurringSubscription", order, subscription).Return(nil)

	gatewayManagerMock := &mocks.PaymentSystemManagerInterface{}
	gatewayManagerMock.On("GetGateway", psHandler).Return(paymentSystem, nil)
	suite.service.paymentSystemGateway = gatewayManagerMock

	rsp := &billingpb.EmptyResponseWithStatus{}
	err := suite.service.DeleteRecurringSubscription(context.Background(), &billingpb.DeleteRecurringSubscriptionRequest{
		SubscriptionId: order.RecurringId,
		CustomerId:     customerId,
	}, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusSystemError, rsp.Status)
	assert.Equal(suite.T(), recurringErrorDeleteSubscription, rsp.Message)
}

func (suite *RecurringTestSuite) TestOrder_DeleteRecurringSubscription_SkipWithEmptyOrder_Ok() {
	var (
		customerId  = "customer_id"
		recurringId = "recurring_id"
	)

	subscription := &recurringpb.Subscription{
		CustomerId: customerId,
		OrderId:    primitive.NilObjectID.Hex(),
	}

	recurring := &recurringMocks.RepositoryService{}
	recurring.On("GetSubscription", mock.Anything, mock.Anything).Return(&recurringpb.GetSubscriptionResponse{
		Status:       billingpb.ResponseStatusOk,
		Subscription: subscription,
	}, nil)
	recurring.On("DeleteSubscription", mock.Anything, subscription).Return(&recurringpb.DeleteSubscriptionResponse{
		Status: billingpb.ResponseStatusOk,
	}, nil)
	suite.service.rep = recurring

	rsp := &billingpb.EmptyResponseWithStatus{}
	err := suite.service.DeleteRecurringSubscription(context.Background(), &billingpb.DeleteRecurringSubscriptionRequest{
		SubscriptionId: recurringId,
		CustomerId:     customerId,
	}, rsp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), billingpb.ResponseStatusOk, rsp.Status)
}

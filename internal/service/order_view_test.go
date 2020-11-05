package service

import (
	"context"
	"github.com/go-redis/redis"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang/protobuf/ptypes"
	"github.com/jinzhu/now"
	"github.com/paysuper/paysuper-billing-server/internal/config"
	"github.com/paysuper/paysuper-billing-server/internal/database"
	"github.com/paysuper/paysuper-billing-server/internal/mocks"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	casbinMocks "github.com/paysuper/paysuper-proto/go/casbinpb/mocks"
	reportingMocks "github.com/paysuper/paysuper-proto/go/reporterpb/mocks"
	tools "github.com/paysuper/paysuper-tools/number"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	rabbitmq "gopkg.in/ProtocolONE/rabbitmq.v1/pkg"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
	"testing"
	"time"
)

type OrderViewTestSuite struct {
	suite.Suite
	service *Service
	log     *zap.Logger
	cache   database.CacheInterface

	merchant            *billingpb.Merchant
	projectFixedAmount  *billingpb.Project
	projectWithProducts *billingpb.Project
	paymentMethod       *billingpb.PaymentMethod
	paymentSystem       *billingpb.PaymentSystem
	paylink1            *billingpb.Paylink
	customer            *billingpb.Customer
	cookie              string
}

func Test_OrderView(t *testing.T) {
	suite.Run(t, new(OrderViewTestSuite))
}

func (suite *OrderViewTestSuite) SetupTest() {
	cfg, err := config.NewConfig()
	if err != nil {
		suite.FailNow("Config load failed", "%v", err)
	}

	cfg.CardPayApiUrl = "https://sandbox.cardpay.com"
	cfg.OrderViewUpdateBatchSize = 20

	m, err := migrate.New(
		"file://../../migrations/tests",
		cfg.MongoDsn)
	assert.NoError(suite.T(), err, "Migrate init failed")

	err = m.Up()
	if err != nil && err.Error() != "no change" {
		suite.FailNow("Migrations failed", "%v", err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 50*time.Second)
	opts := []mongodb.Option{mongodb.Context(ctx)}
	db, err := mongodb.NewDatabase(opts...)
	if err != nil {
		suite.FailNow("Database connection failed", "%v", err)
	}

	suite.log, err = zap.NewProduction()

	if err != nil {
		suite.FailNow("Logger initialization failed", "%v", err)
	}

	broker, err := rabbitmq.NewBroker(cfg.BrokerAddress)

	if err != nil {
		suite.FailNow("Creating RabbitMQ publisher failed", "%v", err)
	}

	redisClient := database.NewRedis(
		&redis.Options{
			Addr:     cfg.RedisHost,
			Password: cfg.RedisPassword,
		},
	)

	redisdb := mocks.NewTestRedis()
	suite.cache, err = database.NewCacheRedis(redisdb, "cache")
	suite.service = NewBillingService(
		db,
		cfg,
		mocks.NewGeoIpServiceTestOk(),
		mocks.NewRepositoryServiceOk(),
		mocks.NewTaxServiceOkMock(),
		broker,
		redisClient,
		suite.cache,
		mocks.NewCurrencyServiceMockOk(),
		mocks.NewDocumentSignerMockOk(),
		&reportingMocks.ReporterService{},
		mocks.NewFormatterOK(),
		mocks.NewBrokerMockOk(),
		&casbinMocks.CasbinService{},
		mocks.NewNotifierOk(),
		mocks.NewBrokerMockOk(),
	)

	if err := suite.service.Init(); err != nil {
		suite.FailNow("Billing service initialization failed", "%v", err)
	}

	suite.merchant, suite.projectFixedAmount, suite.paymentMethod, suite.paymentSystem, suite.customer = HelperCreateEntitiesForTests(suite.Suite, suite.service)

	suite.projectWithProducts = &billingpb.Project{
		Id:                       primitive.NewObjectID().Hex(),
		CallbackCurrency:         "RUB",
		CallbackProtocol:         "default",
		LimitsCurrency:           "USD",
		MaxPaymentAmount:         15000,
		MinPaymentAmount:         1,
		Name:                     map[string]string{"en": "test project 1"},
		IsProductsCheckout:       true,
		AllowDynamicRedirectUrls: true,
		SecretKey:                "test project 1 secret key",
		Status:                   billingpb.ProjectStatusDraft,
		MerchantId:               suite.merchant.Id,
		VatPayer:                 billingpb.VatPayerBuyer,
		WebhookTesting: &billingpb.WebHookTesting{
			Products: &billingpb.ProductsTesting{
				NonExistingUser:  true,
				ExistingUser:     true,
				CorrectPayment:   true,
				IncorrectPayment: true,
			},
			VirtualCurrency: &billingpb.VirtualCurrencyTesting{
				NonExistingUser:  true,
				ExistingUser:     true,
				CorrectPayment:   true,
				IncorrectPayment: true,
			},
			Keys: &billingpb.KeysTesting{IsPassed: true},
		},
	}

	if err := suite.service.project.Insert(context.TODO(), suite.projectWithProducts); err != nil {
		suite.FailNow("Insert project test data failed", "%v", err)
	}

	products := CreateProductsForProject(suite.Suite, suite.service, suite.projectWithProducts, 1)
	assert.Len(suite.T(), products, 1)

	paylinkBod, _ := ptypes.TimestampProto(now.BeginningOfDay())
	paylinkExpiresAt, _ := ptypes.TimestampProto(time.Now().Add(1 * time.Hour))

	suite.paylink1 = &billingpb.Paylink{
		Id:                   primitive.NewObjectID().Hex(),
		Object:               "paylink",
		Products:             []string{products[0].Id},
		ExpiresAt:            paylinkExpiresAt,
		CreatedAt:            paylinkBod,
		UpdatedAt:            paylinkBod,
		MerchantId:           suite.projectWithProducts.MerchantId,
		ProjectId:            suite.projectWithProducts.Id,
		Name:                 "Willy Wonka Strikes Back",
		IsExpired:            false,
		Visits:               0,
		NoExpiryDate:         false,
		ProductsType:         "product",
		Deleted:              false,
		TotalTransactions:    0,
		SalesCount:           0,
		ReturnsCount:         0,
		Conversion:           0,
		GrossSalesAmount:     0,
		GrossReturnsAmount:   0,
		GrossTotalAmount:     0,
		TransactionsCurrency: "",
	}

	err = suite.service.paylinkRepository.Insert(context.TODO(), suite.paylink1)
	assert.NoError(suite.T(), err)

	browserCustomer := &BrowserCookieCustomer{
		CustomerId: suite.customer.Id,
		Ip:         "127.0.0.1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	suite.cookie, err = suite.service.generateBrowserCookie(browserCustomer)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), suite.cookie)
}

func (suite *OrderViewTestSuite) TearDownTest() {
	err := suite.service.db.Drop()

	if err != nil {
		suite.FailNow("Database deletion failed", "%v", err)
	}

	err = suite.service.db.Close()

	if err != nil {
		suite.FailNow("Database close failed", "%v", err)
	}
}

func (suite *OrderViewTestSuite) Test_OrderView_updateOrderView() {
	amounts := []float64{100, 10}
	currencies := []string{"RUB", "USD"}
	countries := []string{"RU", "FI"}
	var orders []*billingpb.Order

	count := 0
	for count < suite.service.cfg.OrderViewUpdateBatchSize+10 {
		order := HelperCreateAndPayOrder(
			suite.Suite,
			suite.service,
			amounts[count%2],
			currencies[count%2],
			countries[count%2],
			suite.projectFixedAmount,
			suite.paymentMethod,
			suite.cookie,
		)
		assert.NotNil(suite.T(), order)
		orders = append(orders, order)

		count++
	}
	err := suite.service.updateOrderView(context.TODO(), []string{})
	assert.NoError(suite.T(), err)
}

func (suite *OrderViewTestSuite) Test_OrderView_GetOrderFromViewPublic_Ok() {
	order := HelperCreateAndPayOrder(
		suite.Suite,
		suite.service,
		100,
		"USD",
		"RU",
		suite.projectFixedAmount,
		suite.paymentMethod,
		suite.cookie,
	)

	assert.False(suite.T(), suite.projectFixedAmount.IsProduction())
	orderPublic, err := suite.service.orderViewRepository.GetPublicOrderBy(context.TODO(), order.Id, "", "")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), orderPublic)
	assert.False(suite.T(), orderPublic.IsProduction)
}

func (suite *OrderViewTestSuite) Test_OrderView_GetOrderFromViewPublic_ProductionProject_Ok() {
	productionProject := &billingpb.Project{
		Id:                       primitive.NewObjectID().Hex(),
		CallbackCurrency:         "RUB",
		CallbackProtocol:         "default",
		LimitsCurrency:           "USD",
		MaxPaymentAmount:         15000,
		MinPaymentAmount:         1,
		Name:                     map[string]string{"en": "test prod project 1"},
		IsProductsCheckout:       false,
		AllowDynamicRedirectUrls: true,
		SecretKey:                "test prod project 1 secret key",
		Status:                   billingpb.ProjectStatusInProduction,
		MerchantId:               suite.merchant.Id,
		VatPayer:                 billingpb.VatPayerBuyer,
	}

	if err := suite.service.project.Insert(context.TODO(), productionProject); err != nil {
		suite.FailNow("Insert project test data failed", "%v", err)
	}

	order := HelperCreateAndPayOrder(
		suite.Suite,
		suite.service,
		100,
		"USD",
		"RU",
		productionProject,
		suite.paymentMethod,
		suite.cookie,
	)

	assert.True(suite.T(), productionProject.IsProduction())
	orderPublic, err := suite.service.orderViewRepository.GetPublicOrderBy(context.TODO(), order.Id, "", "")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), orderPublic)
	assert.True(suite.T(), orderPublic.IsProduction)
}

func (suite *OrderViewTestSuite) Test_OrderView_GetOrderFromViewPrivate_Ok() {
	order := HelperCreateAndPayOrder(
		suite.Suite,
		suite.service,
		100,
		"USD",
		"RU",
		suite.projectFixedAmount,
		suite.paymentMethod,
		suite.cookie,
	)

	op, err := suite.service.orderViewRepository.GetPrivateOrderBy(context.TODO(), order.Id, "", "")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), op)
	assert.False(suite.T(), op.IsProduction)
	assert.NotNil(suite.T(), op.MerchantInfo)
	assert.NotEmpty(suite.T(), op.MerchantInfo.CompanyName)
	assert.NotEmpty(suite.T(), op.MerchantInfo.AgreementNumber)
	assert.NotNil(suite.T(), op.OrderChargeBeforeVat)
	assert.NotNil(suite.T(), op.TaxRate)
}

func (suite *OrderViewTestSuite) Test_OrderView_CountTransactions_Ok() {
	_ = HelperCreateAndPayOrder(
		suite.Suite,
		suite.service,
		100,
		"RUB",
		"RU",
		suite.projectFixedAmount,
		suite.paymentMethod,
		suite.cookie,
	)
	_ = HelperCreateAndPayOrder(
		suite.Suite,
		suite.service,
		200,
		"USD",
		"FI",
		suite.projectFixedAmount,
		suite.paymentMethod,
		suite.cookie,
	)

	count, err := suite.service.orderViewRepository.CountTransactions(context.TODO(), bson.M{})
	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), count, 2)

	count, err = suite.service.orderViewRepository.CountTransactions(context.TODO(), bson.M{"country_code": "FI"})
	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), count, 1)
}

func (suite *OrderViewTestSuite) Test_OrderView_GetTransactionsPublic_Ok() {
	transactions, err := suite.service.orderViewRepository.GetTransactionsPublic(context.TODO(), bson.M{}, 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), transactions, 0)

	_ = HelperCreateAndPayOrder(
		suite.Suite,
		suite.service,
		100,
		"USD",
		"RU",
		suite.projectFixedAmount,
		suite.paymentMethod,
		suite.cookie,
	)

	transactions, err = suite.service.orderViewRepository.GetTransactionsPublic(context.TODO(), bson.M{}, 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), transactions, 1)
	assert.IsType(suite.T(), &billingpb.OrderViewPublic{}, transactions[0])
}

func (suite *OrderViewTestSuite) Test_OrderView_GetTransactionsPrivate_Ok() {
	transactions, err := suite.service.orderViewRepository.GetTransactionsPrivate(context.TODO(), bson.M{}, 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), transactions, 0)

	_ = HelperCreateAndPayOrder(
		suite.Suite,
		suite.service,
		100,
		"USD",
		"RU",
		suite.projectFixedAmount,
		suite.paymentMethod,
		suite.cookie,
	)

	transactions, err = suite.service.orderViewRepository.GetTransactionsPrivate(context.TODO(), bson.M{}, 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), transactions, 1)
	assert.IsType(suite.T(), &billingpb.OrderViewPrivate{}, transactions[0])
}

func (suite *OrderViewTestSuite) Test_OrderView_GetRoyaltySummary_Ok_NoTransactions() {
	to := time.Now().Add(time.Duration(5) * time.Hour)
	from := to.Add(-time.Duration(10) * time.Hour)

	summaryItems, summaryTotal, ordersIds, err := suite.service.orderViewRepository.GetRoyaltySummary(context.TODO(), suite.merchant.Id, suite.merchant.GetPayoutCurrency(), from, to, true)
	assert.NoError(suite.T(), err)

	assert.Len(suite.T(), summaryItems, 0)

	assert.NotNil(suite.T(), summaryTotal)
	assert.Equal(suite.T(), summaryTotal.Product, "")
	assert.Equal(suite.T(), summaryTotal.Region, "")
	assert.Equal(suite.T(), summaryTotal.TotalTransactions, int32(0))
	assert.Equal(suite.T(), summaryTotal.SalesCount, int32(0))
	assert.Equal(suite.T(), summaryTotal.ReturnsCount, int32(0))
	assert.Equal(suite.T(), tools.FormatAmount(summaryTotal.GrossSalesAmount), float64(0))
	assert.Equal(suite.T(), tools.FormatAmount(summaryTotal.GrossReturnsAmount), float64(0))
	assert.Equal(suite.T(), tools.FormatAmount(summaryTotal.GrossTotalAmount), float64(0))
	assert.Equal(suite.T(), tools.FormatAmount(summaryTotal.TotalFees), float64(0))
	assert.Equal(suite.T(), tools.FormatAmount(summaryTotal.TotalVat), float64(0))
	assert.Equal(suite.T(), tools.FormatAmount(summaryTotal.PayoutAmount), float64(0))
	assert.Empty(suite.T(), ordersIds)

	controlTotal := summaryTotal.GrossSalesAmount - summaryTotal.TotalFees - summaryTotal.TotalVat
	assert.Equal(suite.T(), summaryTotal.PayoutAmount, controlTotal)
}

func (suite *OrderViewTestSuite) Test_OrderView_GetRoyaltySummary_Ok_OnlySales() {
	countries := []string{"RU", "FI"}
	var orders []*billingpb.Order
	numberOfOrders := 3

	suite.projectFixedAmount.Status = billingpb.ProjectStatusInProduction
	if err := suite.service.project.Update(context.TODO(), suite.projectFixedAmount); err != nil {
		suite.FailNow("Update project test data failed", "%v", err)
	}

	count := 0
	for count < numberOfOrders {
		order := HelperCreateAndPayOrder(
			suite.Suite,
			suite.service,
			10,
			"USD",
			countries[count%2],
			suite.projectFixedAmount,
			suite.paymentMethod,
			suite.cookie,
		)
		assert.NotNil(suite.T(), order)
		orders = append(orders, order)

		count++
	}

	to := time.Now().Add(time.Duration(5) * time.Hour)
	from := to.Add(-time.Duration(10) * time.Hour)

	summaryItems, summaryTotal, orderIds, err := suite.service.orderViewRepository.GetRoyaltySummary(context.TODO(), suite.merchant.Id, suite.merchant.GetPayoutCurrency(), from, to, true)
	assert.NoError(suite.T(), err)

	assert.Len(suite.T(), summaryItems, 2)

	assert.Equal(suite.T(), summaryItems[0].Product, suite.projectFixedAmount.Name["en"])
	assert.Equal(suite.T(), summaryItems[0].Region, "FI")
	assert.EqualValues(suite.T(), summaryItems[0].TotalTransactions, 1)
	assert.EqualValues(suite.T(), summaryItems[0].SalesCount, 1)
	assert.EqualValues(suite.T(), summaryItems[0].ReturnsCount, 0)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[0].GrossSalesAmount), 12)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[0].GrossReturnsAmount), 0)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[0].GrossTotalAmount), 12)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[0].TotalFees), 0.65)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[0].TotalVat), 2)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[0].PayoutAmount), 9.34)

	controlTotal0 := summaryItems[0].GrossSalesAmount - summaryItems[0].TotalFees - summaryItems[0].TotalVat
	assert.Equal(suite.T(), tools.FormatAmount(summaryItems[0].PayoutAmount), tools.FormatAmount(controlTotal0))

	assert.Equal(suite.T(), summaryItems[1].Product, suite.projectFixedAmount.Name["en"])
	assert.Equal(suite.T(), summaryItems[1].Region, "RU")
	assert.EqualValues(suite.T(), summaryItems[1].TotalTransactions, 2)
	assert.EqualValues(suite.T(), summaryItems[1].SalesCount, 2)
	assert.EqualValues(suite.T(), summaryItems[1].ReturnsCount, 0)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[1].GrossSalesAmount), 24)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[1].GrossReturnsAmount), 0)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[1].GrossTotalAmount), 24)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[1].TotalFees), 1.31)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[1].TotalVat), 4.09)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[1].PayoutAmount), 18.59)

	controlTotal1 := summaryItems[1].GrossSalesAmount - summaryItems[1].TotalFees - summaryItems[1].TotalVat
	assert.Equal(suite.T(), summaryItems[1].PayoutAmount, controlTotal1)

	assert.NotNil(suite.T(), summaryTotal)
	assert.Equal(suite.T(), summaryTotal.Product, "")
	assert.Equal(suite.T(), summaryTotal.Region, "")
	assert.EqualValues(suite.T(), summaryTotal.TotalTransactions, 3)
	assert.EqualValues(suite.T(), summaryTotal.SalesCount, 3)
	assert.EqualValues(suite.T(), summaryTotal.ReturnsCount, 0)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryTotal.GrossSalesAmount), 36)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryTotal.GrossReturnsAmount), 0)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryTotal.GrossTotalAmount), 36)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryTotal.TotalFees), 1.96)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryTotal.TotalVat), 6.09)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryTotal.PayoutAmount), 27.93)
	assert.Len(suite.T(), orderIds, numberOfOrders)

	controlTotal := summaryTotal.GrossSalesAmount - summaryTotal.TotalFees - summaryTotal.TotalVat
	assert.Equal(suite.T(), tools.FormatAmount(summaryTotal.PayoutAmount), tools.FormatAmount(controlTotal))
}

func (suite *OrderViewTestSuite) Test_OrderView_GetRoyaltySummary_HasExistsReportId() {
	countries := []string{"RU", "FI"}
	var orders []*billingpb.Order
	numberOfOrders := 3

	suite.projectFixedAmount.Status = billingpb.ProjectStatusInProduction
	if err := suite.service.project.Update(context.TODO(), suite.projectFixedAmount); err != nil {
		suite.FailNow("Update project test data failed", "%v", err)
	}

	count := 0
	for count < numberOfOrders {
		order := HelperCreateAndPayOrder(
			suite.Suite,
			suite.service,
			10,
			"USD",
			countries[count%2],
			suite.projectFixedAmount,
			suite.paymentMethod,
			suite.cookie,
		)
		assert.NotNil(suite.T(), order)
		orders = append(orders, order)

		count++
	}

	to := time.Now().Add(time.Duration(5) * time.Hour)
	from := to.Add(-time.Duration(10) * time.Hour)

	//goland:noinspection GoNilness
	id, _ := primitive.ObjectIDFromHex(orders[0].Id)
	err := suite.service.orderRepository.IncludeOrdersToRoyaltyReport(ctx, "report_id", []primitive.ObjectID{id})
	assert.NoError(suite.T(), err)

	//goland:noinspection GoNilness
	err = suite.service.updateOrderView(ctx, []string{orders[0].Id})
	assert.NoError(suite.T(), err)

	_, _, orderIds, err := suite.service.orderViewRepository.GetRoyaltySummary(context.TODO(), suite.merchant.Id, suite.merchant.GetPayoutCurrency(), from, to, true)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), orderIds, numberOfOrders-1)

	_, _, orderIds, err = suite.service.orderViewRepository.GetRoyaltySummary(context.TODO(), suite.merchant.Id, suite.merchant.GetPayoutCurrency(), from, to, false)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), orderIds, numberOfOrders)
}

func (suite *OrderViewTestSuite) Test_OrderView_GetRoyaltySummary_Ok_SalesAndRefunds() {
	countries := []string{"RU", "FI"}
	var orders []*billingpb.Order
	numberOfOrders := 3

	suite.projectFixedAmount.Status = billingpb.ProjectStatusInProduction
	if err := suite.service.project.Update(context.TODO(), suite.projectFixedAmount); err != nil {
		suite.FailNow("Update project test data failed", "%v", err)
	}

	count := 0
	for count < numberOfOrders {
		order := HelperCreateAndPayOrder(
			suite.Suite,
			suite.service,
			10,
			"USD",
			countries[count%2],
			suite.projectFixedAmount,
			suite.paymentMethod,
			suite.cookie,
		)
		assert.NotNil(suite.T(), order)
		orders = append(orders, order)

		count++
	}

	suite.paymentSystem.Handler = "mock_ok"
	err := suite.service.paymentSystemRepository.Update(context.TODO(), suite.paymentSystem)
	assert.NoError(suite.T(), err)

	for _, order := range orders {
		refund := HelperMakeRefund(suite.Suite, suite.service, order, order.ChargeAmount, false)
		assert.NotNil(suite.T(), refund)
	}

	to := time.Now().Add(time.Duration(5) * time.Hour)
	from := to.Add(-time.Duration(10) * time.Hour)

	summaryItems, summaryTotal, orderIds, err := suite.service.orderViewRepository.GetRoyaltySummary(context.TODO(), suite.merchant.Id, suite.merchant.GetPayoutCurrency(), from, to, true)
	assert.NoError(suite.T(), err)

	assert.Len(suite.T(), summaryItems, 2)

	assert.Equal(suite.T(), summaryItems[0].Product, suite.projectFixedAmount.Name["en"])
	assert.Equal(suite.T(), summaryItems[0].Region, "FI")
	assert.EqualValues(suite.T(), summaryItems[0].TotalTransactions, 2)
	assert.EqualValues(suite.T(), summaryItems[0].SalesCount, 1)
	assert.EqualValues(suite.T(), summaryItems[0].ReturnsCount, 1)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[0].GrossSalesAmount), 12)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[0].GrossReturnsAmount), 12)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[0].GrossTotalAmount), 0)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[0].TotalFees), 0.65)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[0].TotalVat), 0)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[0].PayoutAmount), -0.65)

	controlTotal0 := summaryItems[0].GrossTotalAmount - summaryItems[0].TotalFees - summaryItems[0].TotalVat
	assert.Equal(suite.T(), tools.FormatAmount(summaryItems[0].PayoutAmount), tools.FormatAmount(controlTotal0))

	assert.Equal(suite.T(), summaryItems[1].Product, suite.projectFixedAmount.Name["en"])
	assert.Equal(suite.T(), summaryItems[1].Region, "RU")
	assert.EqualValues(suite.T(), summaryItems[1].TotalTransactions, 4)
	assert.EqualValues(suite.T(), summaryItems[1].SalesCount, 2)
	assert.EqualValues(suite.T(), summaryItems[1].ReturnsCount, 2)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[1].GrossSalesAmount), 24)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[1].GrossReturnsAmount), 24)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[1].GrossTotalAmount), 0)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[1].TotalFees), 1.31)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[1].TotalVat), 0)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryItems[1].PayoutAmount), -1.31)

	controlTotal1 := summaryItems[1].GrossTotalAmount - summaryItems[1].TotalFees - summaryItems[1].TotalVat
	assert.Equal(suite.T(), tools.FormatAmount(summaryItems[1].PayoutAmount), tools.FormatAmount(controlTotal1))

	assert.NotNil(suite.T(), summaryTotal)
	assert.Equal(suite.T(), summaryTotal.Product, "")
	assert.Equal(suite.T(), summaryTotal.Region, "")
	assert.EqualValues(suite.T(), summaryTotal.TotalTransactions, 6)
	assert.EqualValues(suite.T(), summaryTotal.SalesCount, 3)
	assert.EqualValues(suite.T(), summaryTotal.ReturnsCount, 3)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryTotal.GrossSalesAmount), 36)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryTotal.GrossReturnsAmount), 36)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryTotal.GrossTotalAmount), 0)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryTotal.TotalFees), 1.96)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryTotal.TotalVat), 0)
	assert.EqualValues(suite.T(), tools.FormatAmount(summaryTotal.PayoutAmount), -1.96)
	assert.Len(suite.T(), orderIds, int(summaryTotal.TotalTransactions))

	controlTotal := summaryTotal.GrossTotalAmount - summaryTotal.TotalFees - summaryTotal.TotalVat
	assert.Equal(suite.T(), tools.FormatAmount(summaryTotal.PayoutAmount), tools.FormatAmount(controlTotal))
}

func (suite *OrderViewTestSuite) Test_OrderView_PaylinkStat() {
	countries := []string{"RU", "FI"}
	referrers := []string{"http://steam.com", "http://games.mail.ru/someurl"}
	utmSources := []string{"yandex", "google"}
	utmMedias := []string{"cpc", ""}
	utmCampaign := []string{"45249779", "dfsdf"}
	yesterday := time.Now().Add(-24 * time.Hour).Unix()
	tomorrow := time.Now().Add(24 * time.Hour).Unix()
	maxVisits := 3
	maxOrders := 4
	maxRefunds := 1
	var orders []*billingpb.Order

	n, err := suite.service.paylinkVisitsRepository.CountPaylinkVisits(context.TODO(), suite.paylink1.Id, yesterday, tomorrow)
	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), n, 0)

	visitsReq := &billingpb.PaylinkRequestById{
		Id: suite.paylink1.Id,
	}

	count := 0
	for count < maxVisits {
		err = suite.service.IncrPaylinkVisits(context.TODO(), visitsReq, &billingpb.EmptyResponse{})
		assert.NoError(suite.T(), err)
		count++
	}

	count = 0
	for count < maxOrders {
		err = suite.service.IncrPaylinkVisits(context.TODO(), visitsReq, &billingpb.EmptyResponse{})
		assert.NoError(suite.T(), err)

		order := HelperCreateAndPayPaylinkOrder(
			suite.Suite,
			suite.service,
			suite.paylink1.Id,
			countries[count%2],
			suite.cookie,
			suite.paymentMethod,
			&billingpb.OrderIssuer{
				Url:         referrers[count%2],
				UtmSource:   utmSources[count%2],
				UtmMedium:   utmMedias[count%2],
				UtmCampaign: utmCampaign[count%2],
			},
		)
		assert.NotNil(suite.T(), order)
		orders = append(orders, order)
		count++
	}

	count = 0
	for count < maxRefunds {
		refund := HelperMakeRefund(suite.Suite, suite.service, orders[count], orders[count].ChargeAmount, false)
		assert.NotNil(suite.T(), refund)
		count++
	}

	n, err = suite.service.paylinkVisitsRepository.CountPaylinkVisits(context.TODO(), suite.paylink1.Id, yesterday, tomorrow)
	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), n, maxVisits+maxOrders)

	req := &billingpb.PaylinkRequest{
		Id:         suite.paylink1.Id,
		MerchantId: suite.paylink1.MerchantId,
	}

	res := &billingpb.GetPaylinkResponse{}
	err = suite.service.GetPaylink(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.Status, billingpb.ResponseStatusOk)
	assert.Equal(suite.T(), res.Item.Visits, int32(maxVisits+maxOrders))
	assert.Equal(suite.T(), res.Item.TotalTransactions, int32(maxOrders+maxRefunds))
	assert.EqualValues(suite.T(), res.Item.SalesCount, maxOrders)
	assert.EqualValues(suite.T(), res.Item.ReturnsCount, maxRefunds)
	assert.Equal(suite.T(), res.Item.Conversion, tools.ToPrecise(float64(maxOrders)/float64(maxVisits+maxOrders)))
	assert.Equal(suite.T(), res.Item.TransactionsCurrency, suite.merchant.Banking.Currency)
	assert.EqualValues(suite.T(), res.Item.GrossSalesAmount, 177.506506)
	assert.EqualValues(suite.T(), res.Item.GrossReturnsAmount, 45.397897)
	assert.EqualValues(suite.T(), res.Item.GrossTotalAmount, 132.108609)

	// stat by country
	stat, err := suite.service.orderViewRepository.GetPaylinkStatByCountry(context.TODO(), suite.paylink1.Id, suite.paylink1.MerchantId, yesterday, tomorrow)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stat.Total)
	assert.Len(suite.T(), stat.Top, len(countries))

	assert.Equal(suite.T(), stat.Top[0].CountryCode, "FI")
	assert.EqualValues(suite.T(), stat.Top[0].TotalTransactions, 2)
	assert.EqualValues(suite.T(), stat.Top[0].SalesCount, 2)
	assert.EqualValues(suite.T(), stat.Top[0].ReturnsCount, 0)
	assert.Equal(suite.T(), stat.Top[0].TransactionsCurrency, suite.merchant.Banking.Currency)
	assert.EqualValues(suite.T(), stat.Top[0].GrossSalesAmount, 88.526744)
	assert.EqualValues(suite.T(), stat.Top[0].GrossReturnsAmount, 0)
	assert.EqualValues(suite.T(), stat.Top[0].GrossTotalAmount, 88.526744)

	assert.Equal(suite.T(), stat.Top[1].CountryCode, "RU")
	assert.EqualValues(suite.T(), stat.Top[1].TotalTransactions, 3)
	assert.EqualValues(suite.T(), stat.Top[1].SalesCount, 2)
	assert.EqualValues(suite.T(), stat.Top[1].ReturnsCount, 1)
	assert.Equal(suite.T(), stat.Top[1].TransactionsCurrency, suite.merchant.Banking.Currency)
	assert.EqualValues(suite.T(), stat.Top[1].GrossSalesAmount, 88.979762)
	assert.EqualValues(suite.T(), stat.Top[1].GrossReturnsAmount, 45.397897)
	assert.EqualValues(suite.T(), stat.Top[1].GrossTotalAmount, 43.581865)

	assert.EqualValues(suite.T(), stat.Total.TotalTransactions, maxOrders+maxRefunds)
	assert.EqualValues(suite.T(), stat.Total.SalesCount, maxOrders)
	assert.EqualValues(suite.T(), stat.Total.ReturnsCount, maxRefunds)
	assert.Equal(suite.T(), stat.Total.TransactionsCurrency, suite.merchant.Banking.Currency)
	assert.EqualValues(suite.T(), stat.Total.GrossSalesAmount, 177.506506)
	assert.EqualValues(suite.T(), stat.Total.GrossReturnsAmount, 45.397897)
	assert.EqualValues(suite.T(), stat.Total.GrossTotalAmount, 132.108609)

	// stat by referrer

	stat, err = suite.service.orderViewRepository.GetPaylinkStatByReferrer(context.TODO(), suite.paylink1.Id, suite.paylink1.MerchantId, yesterday, tomorrow)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stat.Total)
	assert.Len(suite.T(), stat.Top, len(referrers))

	assert.Equal(suite.T(), stat.Top[0].ReferrerHost, "games.mail.ru")
	assert.EqualValues(suite.T(), stat.Top[0].TotalTransactions, 2)
	assert.EqualValues(suite.T(), stat.Top[0].SalesCount, 2)
	assert.EqualValues(suite.T(), stat.Top[0].ReturnsCount, 0)
	assert.Equal(suite.T(), stat.Top[0].TransactionsCurrency, suite.merchant.Banking.Currency)
	assert.EqualValues(suite.T(), stat.Top[0].GrossSalesAmount, 88.526744)
	assert.EqualValues(suite.T(), stat.Top[0].GrossReturnsAmount, 0)
	assert.EqualValues(suite.T(), stat.Top[0].GrossTotalAmount, 88.526744)

	assert.Equal(suite.T(), stat.Top[1].ReferrerHost, "steam.com")
	assert.EqualValues(suite.T(), stat.Top[1].TotalTransactions, 3)
	assert.EqualValues(suite.T(), stat.Top[1].SalesCount, 2)
	assert.EqualValues(suite.T(), stat.Top[1].ReturnsCount, 1)
	assert.Equal(suite.T(), stat.Top[1].TransactionsCurrency, suite.merchant.Banking.Currency)
	assert.EqualValues(suite.T(), stat.Top[1].GrossSalesAmount, 88.979762)
	assert.EqualValues(suite.T(), stat.Top[1].GrossReturnsAmount, 45.397897)
	assert.EqualValues(suite.T(), stat.Top[1].GrossTotalAmount, 43.581865)

	assert.EqualValues(suite.T(), stat.Total.TotalTransactions, maxOrders+maxRefunds)
	assert.EqualValues(suite.T(), stat.Total.SalesCount, maxOrders)
	assert.EqualValues(suite.T(), stat.Total.ReturnsCount, maxRefunds)
	assert.Equal(suite.T(), stat.Total.TransactionsCurrency, suite.merchant.Banking.Currency)
	assert.EqualValues(suite.T(), stat.Total.GrossSalesAmount, 177.506506)
	assert.EqualValues(suite.T(), stat.Total.GrossReturnsAmount, 45.397897)
	assert.EqualValues(suite.T(), stat.Total.GrossTotalAmount, 132.108609)

	// stat by date

	stat, err = suite.service.orderViewRepository.GetPaylinkStatByDate(context.TODO(), suite.paylink1.Id, suite.paylink1.MerchantId, yesterday, tomorrow)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stat.Total)
	assert.Len(suite.T(), stat.Top, 1)

	assert.Equal(suite.T(), stat.Top[0].Date, time.Now().Format("2006-01-02"))
	assert.Equal(suite.T(), stat.Top[0].TotalTransactions, int32(maxOrders+maxRefunds))
	assert.EqualValues(suite.T(), stat.Top[0].SalesCount, maxOrders)
	assert.EqualValues(suite.T(), stat.Top[0].ReturnsCount, maxRefunds)
	assert.Equal(suite.T(), stat.Top[0].TransactionsCurrency, suite.merchant.Banking.Currency)
	assert.EqualValues(suite.T(), stat.Top[0].GrossSalesAmount, 177.506506)
	assert.EqualValues(suite.T(), stat.Top[0].GrossReturnsAmount, 45.397897)
	assert.EqualValues(suite.T(), stat.Top[0].GrossTotalAmount, 132.108609)

	assert.Equal(suite.T(), stat.Total.TotalTransactions, int32(maxOrders+maxRefunds))
	assert.EqualValues(suite.T(), stat.Total.SalesCount, maxOrders)
	assert.EqualValues(suite.T(), stat.Total.ReturnsCount, maxRefunds)
	assert.Equal(suite.T(), stat.Total.TransactionsCurrency, suite.merchant.Banking.Currency)
	assert.EqualValues(suite.T(), stat.Total.GrossSalesAmount, 177.506506)
	assert.EqualValues(suite.T(), stat.Total.GrossReturnsAmount, 45.397897)
	assert.EqualValues(suite.T(), stat.Total.GrossTotalAmount, 132.108609)

	// stat by utm
	stat, err = suite.service.orderViewRepository.GetPaylinkStatByUtm(context.TODO(), suite.paylink1.Id, suite.paylink1.MerchantId, yesterday, tomorrow)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stat.Total)
	assert.Len(suite.T(), stat.Top, 2)

	assert.NotNil(suite.T(), stat.Top[0].Utm)
	assert.Equal(suite.T(), stat.Top[0].Utm.UtmSource, "google")
	assert.Equal(suite.T(), stat.Top[0].Utm.UtmMedium, "")
	assert.Equal(suite.T(), stat.Top[0].Utm.UtmCampaign, "dfsdf")
	assert.EqualValues(suite.T(), stat.Top[0].TotalTransactions, 2)
	assert.EqualValues(suite.T(), stat.Top[0].SalesCount, 2)
	assert.EqualValues(suite.T(), stat.Top[0].ReturnsCount, 0)
	assert.Equal(suite.T(), stat.Top[0].TransactionsCurrency, suite.merchant.Banking.Currency)
	assert.EqualValues(suite.T(), stat.Top[0].GrossSalesAmount, 88.526744)
	assert.EqualValues(suite.T(), stat.Top[0].GrossReturnsAmount, 0)
	assert.EqualValues(suite.T(), stat.Top[0].GrossTotalAmount, 88.526744)

	assert.NotNil(suite.T(), stat.Top[1].Utm)
	assert.Equal(suite.T(), stat.Top[1].Utm.UtmSource, "yandex")
	assert.Equal(suite.T(), stat.Top[1].Utm.UtmMedium, "cpc")
	assert.Equal(suite.T(), stat.Top[1].Utm.UtmCampaign, "45249779")
	assert.EqualValues(suite.T(), stat.Top[1].TotalTransactions, 3)
	assert.EqualValues(suite.T(), stat.Top[1].SalesCount, 2)
	assert.EqualValues(suite.T(), stat.Top[1].ReturnsCount, 1)
	assert.Equal(suite.T(), stat.Top[1].TransactionsCurrency, suite.merchant.Banking.Currency)
	assert.EqualValues(suite.T(), stat.Top[1].GrossSalesAmount, 88.979762)
	assert.EqualValues(suite.T(), stat.Top[1].GrossReturnsAmount, 45.397897)
	assert.EqualValues(suite.T(), stat.Top[1].GrossTotalAmount, 43.581865)

	assert.Equal(suite.T(), stat.Total.TotalTransactions, int32(maxOrders+maxRefunds))
	assert.EqualValues(suite.T(), stat.Total.SalesCount, maxOrders)
	assert.EqualValues(suite.T(), stat.Total.ReturnsCount, maxRefunds)
	assert.Equal(suite.T(), stat.Total.TransactionsCurrency, suite.merchant.Banking.Currency)
	assert.EqualValues(suite.T(), stat.Total.GrossSalesAmount, 177.506506)
	assert.EqualValues(suite.T(), stat.Total.GrossReturnsAmount, 45.397897)
	assert.EqualValues(suite.T(), stat.Total.GrossTotalAmount, 132.108609)
}

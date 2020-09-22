package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/ProtocolONE/geoip-service/pkg/proto"
	"github.com/go-redis/redis"
	"github.com/paysuper/paysuper-billing-server/internal/config"
	"github.com/paysuper/paysuper-billing-server/internal/database"
	"github.com/paysuper/paysuper-billing-server/internal/helper"
	"github.com/paysuper/paysuper-billing-server/internal/repository"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-i18n"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/casbinpb"
	"github.com/paysuper/paysuper-proto/go/currenciespb"
	"github.com/paysuper/paysuper-proto/go/document_signerpb"
	"github.com/paysuper/paysuper-proto/go/notifierpb"
	"github.com/paysuper/paysuper-proto/go/recurringpb"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
	"github.com/paysuper/paysuper-proto/go/taxpb"
	httpTools "github.com/paysuper/paysuper-tools/http"
	tools "github.com/paysuper/paysuper-tools/number"
	"go.uber.org/zap"
	"gopkg.in/ProtocolONE/rabbitmq.v1/pkg"
	"gopkg.in/gomail.v2"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
	"strings"
	"sync"
)

const (
	errorNotFound = "%s not found"

	errorBbNotFoundMessage = "not found"

	HeaderContentType   = "Content-Type"
	HeaderAuthorization = "Authorization"
	HeaderContentLength = "Content-Length"

	MIMEApplicationForm = "application/x-www-form-urlencoded"
	MIMEApplicationJSON = "application/json"

	CountryCodeUSA = "US"

	DefaultLanguage = "en"

	centrifugoChannel = "paysuper-billing-server"
)

type Service struct {
	db                                     mongodb.SourceInterface
	mx                                     sync.Mutex
	cfg                                    *config.Config
	ctx                                    context.Context
	geo                                    proto.GeoIpService
	rep                                    recurringpb.RepositoryService
	tax                                    taxpb.TaxService
	broker                                 rabbitmq.BrokerInterface
	redis                                  redis.Cmdable
	cacher                                 database.CacheInterface
	curService                             currenciespb.CurrencyRatesService
	smtpCl                                 gomail.SendCloser
	supportedCurrencies                    []string
	currenciesPrecision                    map[string]int32
	documentSigner                         document_signerpb.DocumentSignerService
	centrifugoPaymentForm                  CentrifugoInterface
	centrifugoDashboard                    CentrifugoInterface
	formatter                              paysuper_i18n.Formatter
	reporterService                        reporterpb.ReporterService
	postmarkBroker                         rabbitmq.BrokerInterface
	casbinService                          casbinpb.CasbinService
	notifier                               notifierpb.NotifierService
	paymentSystemGateway                   *Gateway
	country                                repository.CountryRepositoryInterface
	refundRepository                       repository.RefundRepositoryInterface
	orderRepository                        repository.OrderRepositoryInterface
	userRoleRepository                     repository.UserRoleRepositoryInterface
	zipCodeRepository                      repository.ZipCodeRepositoryInterface
	userProfileRepository                  repository.UserProfileRepositoryInterface
	turnoverRepository                     repository.TurnoverRepositoryInterface
	priceGroupRepository                   repository.PriceGroupRepositoryInterface
	merchantRepository                     repository.MerchantRepositoryInterface
	merchantBalanceRepository              repository.MerchantBalanceRepositoryInterface
	moneyBackCostMerchantRepository        repository.MoneyBackCostMerchantRepositoryInterface
	moneyBackCostSystemRepository          repository.MoneyBackCostSystemRepositoryInterface
	project                                repository.ProjectRepositoryInterface
	priceTableRepository                   repository.PriceTableRepositoryInterface
	notificationRepository                 repository.NotificationRepositoryInterface
	operatingCompanyRepository             repository.OperatingCompanyRepositoryInterface
	bankBinRepository                      repository.BankBinRepositoryInterface
	notifySalesRepository                  repository.NotifySalesRepositoryInterface
	notifyRegionRepository                 repository.NotifyRegionRepositoryInterface
	paymentSystemRepository                repository.PaymentSystemRepositoryInterface
	paymentMethodRepository                repository.PaymentMethodRepositoryInterface
	paymentChannelCostSystemRepository     repository.PaymentChannelCostSystemRepositoryInterface
	paymentChannelCostMerchantRepository   repository.PaymentChannelCostMerchantRepositoryInterface
	paymentMinLimitSystemRepository        repository.PaymentMinLimitSystemRepositoryInterface
	keyRepository                          repository.KeyRepositoryInterface
	keyProductRepository                   repository.KeyProductRepositoryInterface
	productRepository                      repository.ProductRepositoryInterface
	paylinkRepository                      repository.PaylinkRepositoryInterface
	paylinkVisitsRepository                repository.PaylinkVisitRepositoryInterface
	royaltyReportRepository                repository.RoyaltyReportRepositoryInterface
	vatReportRepository                    repository.VatReportRepositoryInterface
	payoutRepository                       repository.PayoutRepositoryInterface
	customerRepository                     repository.CustomerRepositoryInterface
	accountingRepository                   repository.AccountingEntryRepositoryInterface
	merchantTariffsSettingsRepository      repository.MerchantTariffsSettingsInterface
	merchantPaymentTariffsRepository       repository.MerchantPaymentTariffsInterface
	orderViewRepository                    repository.OrderViewRepositoryInterface
	merchantPaymentMethodHistoryRepository repository.MerchantPaymentMethodHistoryRepositoryInterface
	feedbackRepository                     repository.FeedbackRepositoryInterface
	dashboardRepository                    repository.DashboardRepositoryInterface
	validateUserBroker                     rabbitmq.BrokerInterface
	autoincrementRepository                repository.AutoincrementRepositoryInterface
	moneyRegistry                          map[string]*helper.Money
	moneyRegistryMx                        sync.Mutex
}

func newBillingServerResponseError(status int32, message *billingpb.ResponseErrorMessage) *billingpb.ResponseError {
	return &billingpb.ResponseError{
		Status:  status,
		Message: message,
	}
}

func newBillingServerErrorMsg(code, msg string, details ...string) *billingpb.ResponseErrorMessage {
	var det string
	if len(details) > 0 && details[0] != "" {
		det = details[0]
	} else {
		det = ""
	}
	return &billingpb.ResponseErrorMessage{Code: code, Message: msg, Details: det}
}

func NewBillingService(
	db mongodb.SourceInterface,
	cfg *config.Config,
	geo proto.GeoIpService,
	rep recurringpb.RepositoryService,
	tax taxpb.TaxService,
	broker rabbitmq.BrokerInterface,
	redis redis.Cmdable,
	cache database.CacheInterface,
	curService currenciespb.CurrencyRatesService,
	documentSigner document_signerpb.DocumentSignerService,
	reporterService reporterpb.ReporterService,
	formatter paysuper_i18n.Formatter,
	postmarkBroker rabbitmq.BrokerInterface,
	casbinService casbinpb.CasbinService,
	notifier notifierpb.NotifierService,
	validateUserBroker rabbitmq.BrokerInterface,
) *Service {
	return &Service{
		db:                 db,
		cfg:                cfg,
		geo:                geo,
		rep:                rep,
		tax:                tax,
		broker:             broker,
		redis:              redis,
		cacher:             cache,
		curService:         curService,
		documentSigner:     documentSigner,
		reporterService:    reporterService,
		formatter:          formatter,
		postmarkBroker:     postmarkBroker,
		casbinService:      casbinService,
		notifier:           notifier,
		validateUserBroker: validateUserBroker,
		moneyRegistry:      make(map[string]*helper.Money),
	}
}

func (s *Service) Init() (err error) {
	s.centrifugoPaymentForm = newCentrifugo(s.cfg.CentrifugoPaymentForm, httpTools.NewLoggedHttpClient(zap.S()))
	s.centrifugoDashboard = newCentrifugo(s.cfg.CentrifugoDashboard, httpTools.NewLoggedHttpClient(zap.S()))
	s.paymentSystemGateway = s.newPaymentSystemGateway()

	s.refundRepository = repository.NewRefundRepository(s.db)
	s.orderRepository = repository.NewOrderRepository(s.db)
	s.country = repository.NewCountryRepository(s.db, s.cacher)
	s.userRoleRepository = repository.NewUserRoleRepository(s.db, s.cacher)
	s.zipCodeRepository = repository.NewZipCodeRepository(s.db, s.cacher)
	s.userProfileRepository = repository.NewUserProfileRepository(s.db)
	s.turnoverRepository = repository.NewTurnoverRepository(s.db, s.cacher)
	s.priceGroupRepository = repository.NewPriceGroupRepository(s.db, s.cacher)
	s.merchantRepository = repository.NewMerchantRepository(s.db, s.cacher)
	s.merchantBalanceRepository = repository.NewMerchantBalanceRepository(s.db, s.cacher)
	s.moneyBackCostMerchantRepository = repository.NewMoneyBackCostMerchantRepository(s.db, s.cacher)
	s.moneyBackCostSystemRepository = repository.NewMoneyBackCostSystemRepository(s.db, s.cacher)
	s.project = repository.NewProjectRepository(s.db, s.cacher)
	s.priceTableRepository = repository.NewPriceTableRepository(s.db)
	s.notificationRepository = repository.NewNotificationRepository(s.db)
	s.operatingCompanyRepository = repository.NewOperatingCompanyRepository(s.db, s.cacher)
	s.bankBinRepository = repository.NewBankBinRepository(s.db)
	s.notifySalesRepository = repository.NewNotifySalesRepository(s.db)
	s.notifyRegionRepository = repository.NewNotifyRegionRepository(s.db)
	s.paymentSystemRepository = repository.NewPaymentSystemRepository(s.db, s.cacher)
	s.paymentMethodRepository = repository.NewPaymentMethodRepository(s.db, s.cacher)
	s.paymentChannelCostSystemRepository = repository.NewPaymentChannelCostSystemRepository(s.db, s.cacher)
	s.paymentChannelCostMerchantRepository = repository.NewPaymentChannelCostMerchantRepository(s.db, s.cacher)
	s.paymentMinLimitSystemRepository = repository.NewPaymentMinLimitSystemRepository(s.db, s.cacher)
	s.keyRepository = repository.NewKeyRepository(s.db)
	s.keyProductRepository = repository.NewKeyProductRepository(s.db)
	s.productRepository = repository.NewProductRepository(s.db, s.cacher)
	s.paylinkRepository = repository.NewPaylinkRepository(s.db, s.cacher)
	s.paylinkVisitsRepository = repository.NewPaylinkVisitRepository(s.db)
	s.royaltyReportRepository = repository.NewRoyaltyReportRepository(s.db, s.cacher)
	s.vatReportRepository = repository.NewVatReportRepository(s.db)
	s.payoutRepository = repository.NewPayoutRepository(s.db, s.cacher)
	s.customerRepository = repository.NewCustomerRepository(s.db)
	s.accountingRepository = repository.NewAccountingEntryRepository(s.db)
	s.merchantTariffsSettingsRepository = repository.NewMerchantTariffsSettingsRepository(s.db, s.cacher)
	s.merchantPaymentTariffsRepository = repository.NewMerchantPaymentTariffsRepository(s.db, s.cacher)
	s.orderViewRepository = repository.NewOrderViewRepository(s.db)
	s.merchantPaymentMethodHistoryRepository = repository.NewMerchantPaymentMethodHistoryRepository(s.db)
	s.feedbackRepository = repository.NewFeedbackRepository(s.db)
	s.dashboardRepository = repository.NewDashboardRepository(s.db, s.cacher)
	s.autoincrementRepository = repository.NewAutoincrementRepository(s.db)

	sCurr, err := s.curService.GetSupportedCurrencies(context.TODO(), &currenciespb.EmptyRequest{})
	if err != nil {
		zap.S().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(errorFieldService, "CurrencyRatesService"),
			zap.String(errorFieldMethod, "GetSupportedCurrencies"),
		)

		return err
	}

	s.supportedCurrencies = sCurr.Currencies

	cp, err := s.curService.GetCurrenciesPrecision(context.TODO(), &currenciespb.EmptyRequest{})
	if err != nil {
		zap.S().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(errorFieldService, "CurrencyRatesService"),
			zap.String(errorFieldMethod, "GetCurrenciesPrecision"),
		)

		return err
	}

	s.currenciesPrecision = cp.Values

	return
}

func (s *Service) getCurrencyPrecision(currency string) int32 {
	p, ok := s.currenciesPrecision[currency]
	if !ok {
		return 2
	}
	return p
}

func (s *Service) FormatAmount(amount float64, currency string) float64 {
	p := s.getCurrencyPrecision(currency)
	return tools.ToFixed(amount, int(p))
}

func (s *Service) logError(msg string, data []interface{}) {
	zap.S().Errorw(msg, data...)
}

func (s *Service) UpdateOrder(ctx context.Context, req *billingpb.Order, _ *billingpb.EmptyResponse) error {
	err := s.updateOrder(ctx, req)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) IsDbNotFoundError(err error) bool {
	return err.Error() == errorBbNotFoundMessage
}

func (s *Service) getCountryFromAcceptLanguage(acceptLanguage string) (string, string) {
	it := strings.Split(acceptLanguage, ",")

	if strings.Index(it[0], "-") == -1 {
		return "", ""
	}

	it1 := strings.Split(it[0], "-")
	return it[0], strings.ToUpper(it1[1])
}

func (s *Service) getDefaultPaymentMethodCommissions() *billingpb.MerchantPaymentMethodCommissions {
	return &billingpb.MerchantPaymentMethodCommissions{
		Fee: pkg.DefaultPaymentMethodFee,
		PerTransaction: &billingpb.MerchantPaymentMethodPerTransactionCommission{
			Fee:      pkg.DefaultPaymentMethodPerTransactionFee,
			Currency: pkg.DefaultPaymentMethodCurrency,
		},
	}
}

func (s *Service) CheckProjectRequestSignature(
	ctx context.Context,
	req *billingpb.CheckProjectRequestSignatureRequest,
	rsp *billingpb.CheckProjectRequestSignatureResponse,
) error {
	p := &OrderCreateRequestProcessor{
		Service: s,
		request: &billingpb.OrderCreateRequest{ProjectId: req.ProjectId},
		checked: &orderCreateRequestProcessorChecked{},
		ctx:     ctx,
	}

	err := p.processProject()

	if err != nil {
		zap.S().Errorw(pkg.MethodFinishedWithError, "err", err)
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			rsp.Status = billingpb.ResponseStatusBadData
			rsp.Message = e
			return nil
		}
		return err
	}

	if p.checked.project.SecretKey != req.Signature {
		rsp.Status = billingpb.ResponseStatusBadData
		rsp.Message = orderErrorSignatureInvalid

		return nil
	}

	rsp.Status = billingpb.ResponseStatusOk

	return nil
}

func (s *Service) getMerchantCentrifugoChannel(merchantId string) string {
	return fmt.Sprintf(s.cfg.CentrifugoMerchantChannel, merchantId)
}

func (s *Service) reporterServiceCreateFile(ctx context.Context, req *reporterpb.ReportFile) error {
	rsp, err := s.reporterService.CreateFile(ctx, req)

	if err != nil || (rsp != nil && rsp.Status != billingpb.ResponseStatusOk) {
		if err == nil {
			err = errors.New(rsp.Message.Message)
		}

		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(errorFieldService, reporterpb.ServiceName),
			zap.String(errorFieldMethod, "CreateFile"),
			zap.Any(errorFieldRequest, req),
		)
	}

	return err
}

func (s *Service) round(merchantId, fieldKey string, val float64) (float64, error) {
	key := merchantId + "_" + fieldKey
	money, ok := s.moneyRegistry[key]

	if !ok {
		s.moneyRegistryMx.Lock()
		money = helper.NewMoney()
		s.moneyRegistry[key] = money
		s.moneyRegistryMx.Unlock()
	}

	rounded, err := money.Round(val)

	if err != nil {
		zap.L().Error(
			billingpb.ErrorUnableRound,
			zap.Error(err),
			zap.String(billingpb.ErrorFieldKey, key),
			zap.Float64(billingpb.ErrorFieldValue, val),
		)
	}

	return rounded, err
}

package internal

import (
	"context"
	"errors"
	"github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/handlers"
	"github.com/ProtocolONE/geoip-service/pkg"
	"github.com/ProtocolONE/geoip-service/pkg/proto"
	"github.com/go-redis/redis"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang/protobuf/ptypes"
	"github.com/micro/cli"
	"github.com/micro/go-micro"
	goConfig "github.com/micro/go-micro/config"
	"github.com/micro/go-micro/config/source"
	goConfigCli "github.com/micro/go-micro/config/source/cli"
	"github.com/micro/go-plugins/client/selector/static"
	metrics "github.com/micro/go-plugins/wrapper/monitoring/prometheus"
	"github.com/paysuper/paysuper-billing-server/internal/config"
	"github.com/paysuper/paysuper-billing-server/internal/database"
	"github.com/paysuper/paysuper-billing-server/internal/service"
	"github.com/paysuper/paysuper-billing-server/pkg"
	paysuperI18n "github.com/paysuper/paysuper-i18n"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/casbinpb"
	"github.com/paysuper/paysuper-proto/go/currenciespb"
	"github.com/paysuper/paysuper-proto/go/document_signerpb"
	"github.com/paysuper/paysuper-proto/go/notifierpb"
	"github.com/paysuper/paysuper-proto/go/postmarkpb"
	"github.com/paysuper/paysuper-proto/go/recurringpb"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
	"github.com/paysuper/paysuper-proto/go/taxpb"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gopkg.in/ProtocolONE/rabbitmq.v1/pkg"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Application struct {
	cfg          *config.Config
	database     mongodb.SourceInterface
	redis        *redis.Client
	redisCluster *redis.ClusterClient
	service      micro.Service
	httpServer   *http.Server
	router       *http.ServeMux
	logger       *zap.Logger
	svc          *service.Service
	CliArgs      goConfig.Config
}

func NewApplication() *Application {
	return &Application{}
}

func (app *Application) Init() {
	app.initLogger()

	cfg, err := config.NewConfig()

	if err != nil {
		app.logger.Fatal("Config load failed", zap.Error(err))
	}

	app.cfg = cfg

	app.logger.Info("db migrations started")

	migrations, err := migrate.New(pkg.MigrationSource, app.cfg.MongoDsn)

	if err != nil {
		app.logger.Fatal("Migrations initialization failed", zap.Error(err))
	}

	migrations.LockTimeout = time.Duration(cfg.MigrationsLockTimeout) * time.Second

	err = migrations.Up()

	if err != nil && err != migrate.ErrNoChange && err != migrate.ErrNilVersion {
		app.logger.Fatal("Migrations processing failed", zap.Error(err))
	}

	app.logger.Info("db migrations applied")

	db, err := mongodb.NewDatabase()
	if err != nil {
		app.logger.Fatal("Database connection failed", zap.Error(err))
	}

	app.database = db

	app.redis = database.NewRedis(
		&redis.Options{
			Addr:     cfg.RedisHost,
			Password: cfg.RedisPassword,
		},
	)

	if err != nil {
		app.logger.Fatal("Connection to Redis failed", zap.Error(err))
	}

	broker, err := rabbitmq.NewBroker(app.cfg.BrokerAddress)

	if err != nil {
		app.logger.Fatal("Creating RabbitMQ publisher failed", zap.Error(err))
	}

	postmarkBroker, err := rabbitmq.NewBroker(app.cfg.BrokerAddress)

	if err != nil {
		app.logger.Fatal("Creating postmark broker failed", zap.Error(err))
	}

	postmarkBroker.SetExchangeName(postmarkpb.PostmarkSenderTopicName)

	validateUserBroker, err := rabbitmq.NewBroker(app.cfg.BrokerAddress)

	if err != nil {
		app.logger.Fatal("Creating validate user broker failed", zap.Error(err))
	}

	validateUserBroker.SetExchangeName(notifierpb.PayOneTopicNameValidateUser)

	options := []micro.Option{
		micro.Name(billingpb.ServiceName),
		micro.WrapHandler(metrics.NewHandlerWrapper()),
		micro.AfterStop(func() error {
			app.logger.Info("Micro service stopped")
			app.Stop()
			return nil
		}),
		micro.Flags(
			cli.StringFlag{
				Name:  "task",
				Value: "",
				Usage: "running task",
			},
			cli.StringFlag{
				Name:  "date",
				Value: "",
				Usage: "task context date, i.e. 2006-01-02T15:04:05Z07:00",
			},
			cli.StringFlag{
				Name:  "orderid",
				Value: "",
				Usage: "selected order id",
			},
			cli.StringFlag{
				Name:  "force",
				Value: "",
				Usage: "force rebuild accounting entries for order",
			},
		),
	}

	if os.Getenv("MICRO_SELECTOR") == "static" {
		log.Println("Use micro selector `static`")
		options = append(options, micro.Selector(static.NewSelector()))
	}

	app.logger.Info("Initialize micro service")

	app.service = micro.NewService(options...)

	var clisrc source.Source

	app.service.Init(
		micro.Action(func(c *cli.Context) {
			clisrc = goConfigCli.NewSource(
				goConfigCli.Context(c),
			)
		}),
	)

	app.CliArgs = goConfig.NewConfig()
	err = app.CliArgs.Load(clisrc)
	if err != nil {
		app.logger.Fatal("Cli args load failed", zap.Error(err))
	}

	geoService := proto.NewGeoIpService(geoip.ServiceName, app.service.Client())
	repService := recurringpb.NewRepositoryService(recurringpb.PayOneRepositoryServiceName, app.service.Client())
	taxService := taxpb.NewTaxService(taxpb.ServiceName, app.service.Client())
	curService := currenciespb.NewCurrencyRatesService(currenciespb.ServiceName, app.service.Client())
	documentSignerService := document_signerpb.NewDocumentSignerService(document_signerpb.ServiceName, app.service.Client())
	reporter := reporterpb.NewReporterService(reporterpb.ServiceName, app.service.Client())
	casbin := casbinpb.NewCasbinService(casbinpb.ServiceName, app.service.Client())
	webHookNotifier := notifierpb.NewNotifierService(notifierpb.ServiceName, app.service.Client())

	formatter, err := paysuperI18n.NewFormatter([]string{"i18n/rules"}, []string{"i18n/messages"})

	if err != nil {
		app.logger.Fatal("Create il8n formatter failed", zap.Error(err))
	}

	app.redisCluster = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        cfg.CacheRedis.Address,
		Password:     cfg.CacheRedis.Password,
		MaxRetries:   cfg.CacheRedis.MaxRetries,
		MaxRedirects: cfg.CacheRedis.MaxRedirects,
		PoolSize:     cfg.CacheRedis.PoolSize,
	})
	cache, err := database.NewCacheRedis(app.redisCluster, cfg.CacheRedis.Version)

	if err != nil {
		app.logger.Error("Unable to initialize cache for the application", zap.Error(err))
	} else {
		go func() {
			if err = cache.CleanOldestVersion(); err != nil {
				app.logger.Error("Unable to clean oldest versions of cache", zap.Error(err))
			}
		}()
	}

	app.svc = service.NewBillingService(
		app.database,
		app.cfg,
		geoService,
		repService,
		taxService,
		broker,
		app.redis,
		cache,
		curService,
		documentSignerService,
		reporter,
		formatter,
		postmarkBroker,
		casbin,
		webHookNotifier,
		validateUserBroker,
	)

	if err := app.svc.Init(); err != nil {
		app.logger.Fatal("Create service instance failed", zap.Error(err))
	}

	err = billingpb.RegisterBillingServiceHandler(app.service.Server(), app.svc)

	if err != nil {
		app.logger.Fatal("Service init failed", zap.Error(err))
	}

	app.router = http.NewServeMux()
	app.initHealth()
	app.initMetrics()
}

func (app *Application) initLogger() {
	var err error

	logger, err := zap.NewProduction()

	if err != nil {
		log.Fatalf("Application logger initialization failed with error: %s\n", err)
	}
	app.logger = logger.Named(pkg.LoggerName)
	zap.ReplaceGlobals(app.logger)
}

func (app *Application) initHealth() {
	h := health.New()
	err := h.AddChecks([]*health.Config{
		{
			Name:     "health-check",
			Checker:  app,
			Interval: time.Duration(1) * time.Second,
			Fatal:    true,
		},
	})

	if err != nil {
		app.logger.Fatal("Health check register failed", zap.Error(err))
	}

	if err = h.Start(); err != nil {
		app.logger.Fatal("Health check start failed", zap.Error(err))
	}

	app.logger.Info("Health check listener started", zap.String("port", app.cfg.MetricsPort))

	app.router.HandleFunc("/health", handlers.NewJSONHandlerFunc(h, nil))
}

func (app *Application) initMetrics() {
	app.router.Handle("/metrics", promhttp.Handler())
}

func (app *Application) Run() {
	app.httpServer = &http.Server{
		Addr:              ":" + app.cfg.MetricsPort,
		Handler:           app.router,
		ReadTimeout:       time.Duration(app.cfg.MetricsReadTimeout) * time.Second,
		ReadHeaderTimeout: time.Duration(app.cfg.MetricsReadHeaderTimeout) * time.Second,
	}

	go func() {
		if err := app.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.logger.Fatal("Http server starting failed", zap.Error(err))
		}
	}()

	if err := app.service.Run(); err != nil {
		app.logger.Fatal("Micro service starting failed", zap.Error(err))
	}
}

func (app *Application) Status() (interface{}, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err := app.database.Ping(ctx)

	if err != nil {
		return "fail", errors.New("mongodb connection lost: " + err.Error())
	}

	err = app.redis.Ping().Err()

	if err != nil {
		return "fail", errors.New("redis connection lost: " + err.Error())
	}

	err = app.redisCluster.Ping().Err()

	if err != nil {
		return "fail", errors.New("redis cluster connection lost: " + err.Error())
	}

	return "ok", nil
}

func (app *Application) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if app.httpServer != nil {
		if err := app.httpServer.Shutdown(ctx); err != nil {
			app.logger.Error("Http server shutdown failed", zap.Error(err))
		}
		app.logger.Info("Http server stopped")
	}

	_ = app.database.Close()
	app.logger.Info("Database connection closed")

	if err := app.redis.Close(); err != nil {
		zap.S().Error("Redis connection close failed", zap.Error(err))
	} else {
		zap.S().Info("Redis connection closed")
	}

	if err := app.logger.Sync(); err != nil {
		app.logger.Error("Logger sync failed", zap.Error(err))
	} else {
		app.logger.Info("Logger synced")
	}
}

func (app *Application) TaskProcessVatReports(date string) error {
	zap.S().Info("Start to processing vat reports")
	req := &billingpb.ProcessVatReportsRequest{
		Date: ptypes.TimestampNow(),
	}
	if date != "" {
		date, err := time.Parse("2006-01-02", date)
		if err != nil {
			return err
		}
		if date.After(time.Now()) {
			return errors.New(pkg.ErrorVatReportDateCantBeInFuture)
		}
		req.Date, err = ptypes.TimestampProto(date)
		if err != nil {
			return err
		}
	}
	return app.svc.ProcessVatReports(context.TODO(), req, &billingpb.EmptyResponse{})
}

func (app *Application) TaskCreateRoyaltyReport() error {
	return app.svc.CreateRoyaltyReport(context.TODO(), &billingpb.CreateRoyaltyReportRequest{}, &billingpb.CreateRoyaltyReportRequest{})
}

func (app *Application) TaskAutoAcceptRoyaltyReports() error {
	return app.svc.AutoAcceptRoyaltyReports(context.TODO(), &billingpb.EmptyRequest{}, &billingpb.EmptyResponse{})
}

func (app *Application) TaskAutoCreatePayouts() error {
	return app.svc.AutoCreatePayoutDocuments(context.TODO(), &billingpb.EmptyRequest{}, &billingpb.EmptyResponse{})
}

func (app *Application) TaskRebuildOrderView() error {
	return app.svc.RebuildOrderView(context.TODO())
}

func (app *Application) TaskRebuildAccountingEntries(orderId string, force bool) error {
	return app.svc.RebuildAccountingEntries(context.TODO(), orderId, force)
}

func (app *Application) TaskFixReportDates() error {
	return app.svc.TaskFixReportDates(context.TODO())
}

func (app *Application) TaskMerchantsMigrate() error {
	return app.svc.MerchantsMigrate(context.TODO())
}

func (app *Application) TaskFixTaxes() error {
	return app.svc.FixTaxes(context.TODO())
}

func (app *Application) TaskRebuildPayouts() error {
	return app.svc.TaskRebuildPayoutsRoyalties()
}

func (app *Application) MigrateCustomers() error {
	return app.svc.MigrateCustomers(context.Background())
}

func (app *Application) UpdateFirstPayments() error {
	return app.svc.UpdateFirstPayments(context.Background())
}

func (app *Application) TaskCreatePayout() error {
	rsp := &billingpb.CreatePayoutDocumentResponse{}
	err := app.svc.CreatePayoutDocument(
		context.Background(),
		&billingpb.CreatePayoutDocumentRequest{
			MerchantId: "5dbac7bb120a810001a8fe80",
			Ip:         "127.0.0.1",
		},
		rsp,
	)

	if err != nil {
		return err
	}

	if rsp.Status != billingpb.ResponseStatusOk {
		return rsp.Message
	}

	return nil
}

func (app *Application) KeyDaemonStart() {
	zap.L().Info("Key daemon started", zap.Int64("RestartInterval", app.cfg.KeyDaemonRestartInterval))

	go func() {
		interval := time.Duration(app.cfg.KeyDaemonRestartInterval) * time.Second
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

		for {
			zap.S().Debug("Key daemon working")

			select {
			case <-shutdown:
				zap.S().Info("Key daemon stopping")
				return
			default:
				count, err := app.svc.KeyDaemonProcess(context.TODO())
				if err != nil {
					zap.L().Error("Key daemon process failed", zap.Error(err))
				}

				zap.S().Debugf("Key daemon job finished", "count", count)
				time.Sleep(interval)
			}
		}
	}()
}

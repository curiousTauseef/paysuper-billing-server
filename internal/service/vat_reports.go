package service

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jinzhu/now"
	"github.com/paysuper/paysuper-billing-server/internal/helper"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/currenciespb"
	"github.com/paysuper/paysuper-proto/go/postmarkpb"
	"github.com/paysuper/paysuper-proto/go/taxpb"
	tools "github.com/paysuper/paysuper-tools/number"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"go.uber.org/zap"
	"time"
)

const (
	collectionVatReports = "vat_reports"

	VatPeriodEvery1Month = 1
	VatPeriodEvery2Month = 2
	VatPeriodEvery3Month = 3

	errorMsgVatReportTaxServiceGetRateFailed   = "tax service get rate error"
	errorMsgVatReportTurnoverNotFound          = "turnover not found"
	errorMsgVatReportRatesPolicyNotImplemented = "selected currency rates policy not implemented yet"
	errorMsgVatReportCantGetTimeForDate        = "cannot get vat report time for date"
)

var (
	errorVatReportNotEnabledForCountry          = newBillingServerErrorMsg("vr000001", "vat not enabled for country")
	errorVatReportPeriodNotConfiguredForCountry = newBillingServerErrorMsg("vr000002", "vat period not configured for country")
	errorVatReportCurrencyExchangeFailed        = newBillingServerErrorMsg("vr000003", "currency exchange failed")
	errorVatReportStatusChangeNotAllowed        = newBillingServerErrorMsg("vr000004", "vat report status change not allowed")
	errorVatReportStatusChangeFailed            = newBillingServerErrorMsg("vr000005", "vat report status change failed")
	errorVatReportQueryError                    = newBillingServerErrorMsg("vr000006", "vat report db query error")
	errorVatReportNotFound                      = newBillingServerErrorMsg("vr000007", "vat report not found")
	errorVatReportInternal                      = newBillingServerErrorMsg("vr000008", "vat report internal error")
	errorVatReportStatusIsTheSame               = newBillingServerErrorMsg("vr000009", "vat report status is the same")

	VatReportOnStatusNotifyToCentrifugo = []string{
		pkg.VatReportStatusNeedToPay,
		pkg.VatReportStatusPaid,
		pkg.VatReportStatusOverdue,
		pkg.VatReportStatusCanceled,
	}

	VatReportOnStatusNotifyToEmail = []string{
		pkg.VatReportStatusNeedToPay,
		pkg.VatReportStatusOverdue,
		pkg.VatReportStatusCanceled,
	}

	VatReportStatusAllowManualChangeFrom = []string{
		pkg.VatReportStatusNeedToPay,
		pkg.VatReportStatusOverdue,
	}

	VatReportStatusAllowManualChangeTo = []string{
		pkg.VatReportStatusPaid,
		pkg.VatReportStatusCanceled,
	}

	AccountingEntriesLocalAmountsUpdate = []string{
		pkg.AccountingEntryTypeRealGrossRevenue,
		pkg.AccountingEntryTypeRealTaxFee,
		pkg.AccountingEntryTypeCentralBankTaxFee,
		pkg.AccountingEntryTypeRealRefund,
		pkg.AccountingEntryTypeRealRefundTaxFee,
	}
)

type vatReportProcessor struct {
	*Service
	date               time.Time
	ts                 *timestamp.Timestamp
	countries          []*billingpb.Country
	orderViewUpdateIds map[string]bool
}

func NewVatReportProcessor(s *Service, ctx context.Context, date *timestamp.Timestamp) (*vatReportProcessor, error) {
	ts, err := ptypes.Timestamp(date)
	if err != nil {
		return nil, err
	}
	eod := now.New(ts).EndOfDay()
	eodTimestamp, err := ptypes.TimestampProto(eod)
	if err != nil {
		return nil, err
	}
	countries, err := s.country.FindByVatEnabled(ctx)
	if err != nil {
		return nil, err
	}

	processor := &vatReportProcessor{
		Service:            s,
		date:               eod,
		ts:                 eodTimestamp,
		countries:          countries.Countries,
		orderViewUpdateIds: make(map[string]bool),
	}

	return processor, nil
}

func (s *Service) GetVatReportsDashboard(
	ctx context.Context,
	req *billingpb.EmptyRequest,
	res *billingpb.VatReportsResponse,
) error {
	res.Status = billingpb.ResponseStatusOk

	reports, err := s.vatReportRepository.GetByStatus(
		ctx,
		[]string{pkg.VatReportStatusThreshold, pkg.VatReportStatusNeedToPay, pkg.VatReportStatusOverdue},
	)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			res.Status = billingpb.ResponseStatusNotFound
			res.Message = errorVatReportNotFound
			return nil
		}

		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorVatReportQueryError
		return nil
	}

	res.Data = &billingpb.VatReportsPaginate{
		Count: int32(len(reports)),
		Items: reports,
	}

	return nil
}

func (s *Service) GetVatReportsForCountry(
	ctx context.Context,
	req *billingpb.VatReportsRequest,
	res *billingpb.VatReportsResponse,
) error {
	res.Status = billingpb.ResponseStatusOk

	reports, err := s.vatReportRepository.GetByCountry(ctx, req.Country, req.Sort, req.Offset, req.Limit)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			res.Status = billingpb.ResponseStatusNotFound
			res.Message = errorVatReportNotFound
			return nil
		}

		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorVatReportQueryError
		return nil
	}

	res.Data = &billingpb.VatReportsPaginate{
		Count: int32(len(reports)),
		Items: reports,
	}

	return nil
}

func (s *Service) GetVatReport(
	ctx context.Context,
	req *billingpb.VatReportRequest,
	res *billingpb.VatReportResponse,
) error {
	res.Status = billingpb.ResponseStatusOk

	vr, err := s.vatReportRepository.GetById(ctx, req.Id)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			res.Status = billingpb.ResponseStatusNotFound
			res.Message = errorVatReportNotFound
			return nil
		}

		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorVatReportQueryError
		return nil
	}

	res.Vat = vr

	return nil
}

func (s *Service) GetVatReportTransactions(
	ctx context.Context,
	req *billingpb.VatTransactionsRequest,
	res *billingpb.PrivateTransactionsResponse,
) error {
	res.Status = billingpb.ResponseStatusOk

	vr, err := s.vatReportRepository.GetById(ctx, req.VatReportId)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			res.Status = billingpb.ResponseStatusNotFound
			res.Message = errorVatReportNotFound
			return nil
		}

		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorVatReportQueryError
		return nil
	}

	from, err := ptypes.Timestamp(vr.DateFrom)
	if err != nil {
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorVatReportInternal
		return nil
	}
	to, err := ptypes.Timestamp(vr.DateTo)
	if err != nil {
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorVatReportInternal
		return nil
	}

	match := bson.M{
		"pm_order_close_date": bson.M{
			"$gte": now.New(from).BeginningOfDay(),
			"$lte": now.New(to).EndOfDay(),
		},
		"country_code":         vr.Country,
		"operating_company_id": vr.OperatingCompanyId,
		"is_production":        true,
	}

	n, err := s.orderViewRepository.CountTransactions(ctx, match)

	if err != nil {
		return err
	}

	vts, err := s.orderViewRepository.GetTransactionsPrivate(ctx, match, req.Limit, req.Offset)

	if err != nil {
		return err
	}

	res.Data = &billingpb.PrivateTransactionsPaginate{
		Count: int32(n),
		Items: vts,
	}

	return nil
}

func (s *Service) ProcessVatReports(
	ctx context.Context,
	req *billingpb.ProcessVatReportsRequest,
	res *billingpb.EmptyResponse,
) error {

	handler, err := NewVatReportProcessor(s, ctx, req.Date)
	if err != nil {
		return err
	}

	zap.S().Info("process accounting entries")
	err = handler.ProcessAccountingEntries(ctx)
	if err != nil {
		return err
	}

	zap.S().Info("updating order view")
	err = handler.UpdateOrderView(ctx)
	if err != nil {
		return err
	}

	zap.S().Info("calc annual turnovers")
	err = s.CalcAnnualTurnovers(ctx, &billingpb.EmptyRequest{}, &billingpb.EmptyResponse{})
	if err != nil {
		return err
	}

	zap.S().Info("processing vat reports")
	err = handler.ProcessVatReports(ctx)
	if err != nil {
		return err
	}

	zap.S().Info("updating vat reports status")
	err = handler.ProcessVatReportsStatus(ctx)
	if err != nil {
		return err
	}

	zap.S().Info("processing vat reports finished successfully")

	return nil
}

func (s *Service) UpdateVatReportStatus(
	ctx context.Context,
	req *billingpb.UpdateVatReportStatusRequest,
	res *billingpb.ResponseError,
) error {
	res.Status = billingpb.ResponseStatusOk

	vr, err := s.vatReportRepository.GetById(ctx, req.Id)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			res.Status = billingpb.ResponseStatusNotFound
			res.Message = errorVatReportNotFound
			return nil
		}

		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorVatReportQueryError
		return nil
	}

	if vr.Status == req.Status {
		res.Status = billingpb.ResponseStatusNotModified
		res.Message = errorVatReportStatusIsTheSame
		return nil
	}

	if !helper.Contains(VatReportStatusAllowManualChangeFrom, vr.Status) {
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorVatReportStatusChangeNotAllowed
		return nil
	}

	if !helper.Contains(VatReportStatusAllowManualChangeTo, req.Status) {
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorVatReportStatusChangeNotAllowed
		return nil
	}

	vr.Status = req.Status
	if vr.Status == pkg.VatReportStatusPaid {
		vr.PaidAt = ptypes.TimestampNow()
	}

	err = s.updateVatReport(ctx, vr)
	if err != nil {
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorVatReportStatusChangeFailed
		return nil
	}

	return nil
}

func (s *Service) updateVatReport(ctx context.Context, vr *billingpb.VatReport) error {
	if err := s.vatReportRepository.Update(ctx, vr); err != nil {
		return err
	}

	if helper.Contains(VatReportOnStatusNotifyToCentrifugo, vr.Status) {
		if err := s.centrifugoDashboard.Publish(ctx, s.cfg.CentrifugoFinancierChannel, vr); err != nil {
			return err
		}
	}

	if helper.Contains(VatReportOnStatusNotifyToEmail, vr.Status) {
		payload := &postmarkpb.Payload{
			TemplateAlias: s.cfg.EmailTemplates.VatReportChanged,
			TemplateModel: map[string]string{
				"country": vr.Country,
				"status":  vr.Status,
			},
			To: s.cfg.EmailNotificationFinancierRecipient,
		}

		err := s.postmarkBroker.Publish(postmarkpb.PostmarkSenderTopicName, payload, amqp.Table{})

		if err != nil {
			zap.L().Error(
				"Publication message about vat report status change to queue failed",
				zap.Error(err),
				zap.Any("report", vr),
			)
		}
	}
	return nil
}

func (s *Service) getVatReportTime(VatPeriodMonth int32, date time.Time) (from, to time.Time, err error) {
	var nowTime *now.Now
	if date.IsZero() {
		nowTime = now.New(time.Now())
	} else {
		nowTime = now.New(date)
	}

	switch VatPeriodMonth {
	case VatPeriodEvery1Month:
		from = nowTime.BeginningOfMonth()
		to = nowTime.EndOfMonth()
		return
	case VatPeriodEvery2Month:
		from = nowTime.BeginningOfMonth()
		to = nowTime.EndOfMonth()
		if from.Month()%2 == 0 {
			from = now.New(from.AddDate(0, 0, -1)).BeginningOfMonth()
		} else {
			to = now.New(to.AddDate(0, 0, 1)).EndOfMonth()
		}
		return
	case VatPeriodEvery3Month:
		from = nowTime.BeginningOfQuarter()
		to = nowTime.EndOfQuarter()
		return
	}

	err = errorVatReportPeriodNotConfiguredForCountry
	return
}

func (s *Service) getVatReportTimeForDate(VatPeriodMonth int32, date time.Time) (from, to time.Time, err error) {
	return s.getVatReportTime(VatPeriodMonth, date)
}

func (s *Service) getLastVatReportTime(VatPeriodMonth int32) (from, to time.Time, err error) {
	return s.getVatReportTime(VatPeriodMonth, time.Time{})
}

func (h *vatReportProcessor) ProcessVatReportsStatus(ctx context.Context) error {
	reports, err := h.vatReportRepository.GetByStatus(ctx, []string{pkg.VatReportStatusThreshold, pkg.VatReportStatusNeedToPay})

	if err != nil {
		return err
	}

	currentUnixTime := time.Now().Unix()

	for _, report := range reports {
		country := h.getCountry(report.Country)
		if country == nil || country.VatEnabled == false {
			continue
		}
		currentFrom, _, err := h.Service.getLastVatReportTime(country.VatPeriodMonth)
		if err != nil {
			return err
		}
		reportDateFrom, err := ptypes.Timestamp(report.DateFrom)
		if err != nil {
			return err
		}

		if reportDateFrom.Unix() >= currentFrom.Unix() {
			continue
		}

		if report.Status == pkg.VatReportStatusNeedToPay {
			reportDeadline, err := ptypes.Timestamp(report.PayUntilDate)
			if err != nil {
				return err
			}
			if currentUnixTime >= reportDeadline.Unix() {
				report.Status = pkg.VatReportStatusOverdue
				err = h.Service.updateVatReport(ctx, report)
				if err != nil {
					return err
				}
			}
			continue
		}

		noThreshold := country.VatThreshold.Year == 0 && country.VatThreshold.World == 0

		thresholdExceeded := (country.VatThreshold.Year > 0 && report.CountryAnnualTurnover >= country.VatThreshold.Year) ||
			(country.VatThreshold.World > 0 && report.WorldAnnualTurnover >= country.VatThreshold.World)

		amountsGtZero := report.VatAmount > 0 || report.CorrectionAmount > 0 || report.DeductionAmount > 0

		if (noThreshold || thresholdExceeded) && amountsGtZero {
			report.Status = pkg.VatReportStatusNeedToPay
		} else {
			report.Status = pkg.VatReportStatusExpired
		}
		err = h.Service.updateVatReport(ctx, report)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *vatReportProcessor) getCountry(countryCode string) *billingpb.Country {
	for _, c := range h.countries {
		if c.IsoCodeA2 == countryCode {
			return c
		}
	}
	return nil
}

func (h *vatReportProcessor) ProcessVatReports(ctx context.Context) error {
	operatingCompanies, err := h.Service.operatingCompanyRepository.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, oc := range operatingCompanies {
		for _, c := range h.countries {
			err := h.processVatReportForPeriod(ctx, c, oc.Id)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *vatReportProcessor) ProcessAccountingEntries(ctx context.Context) error {
	for _, c := range h.countries {
		err := h.processAccountingEntriesForPeriod(ctx, c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *vatReportProcessor) UpdateOrderView(ctx context.Context) error {
	if len(h.orderViewUpdateIds) == 0 {
		return nil
	}

	ids := make([]string, 0, len(h.orderViewUpdateIds))
	for k := range h.orderViewUpdateIds {
		ids = append(ids, k)
	}

	err := h.Service.updateOrderView(ctx, ids)
	if err != nil {
		return err
	}

	return nil
}

func (h *vatReportProcessor) processVatReportForPeriod(ctx context.Context, country *billingpb.Country, operatingCompanyId string) error {

	from, to, err := h.Service.getVatReportTimeForDate(country.VatPeriodMonth, h.date)
	if err != nil {
		zap.S().Warnw("generating vat report failed", "country", country.IsoCodeA2, "err", err.Error())
		return err
	}

	zap.S().Infow("generating vat report",
		"country", country.IsoCodeA2,
		"from", from.Format(time.RFC3339),
		"to", to.Format(time.RFC3339),
	)

	req := &taxpb.GeoIdentity{
		Country: country.IsoCodeA2,
	}

	rsp, err := h.Service.tax.GetRate(ctx, req)
	if err != nil {
		zap.L().Error(errorMsgVatReportTaxServiceGetRateFailed, zap.Error(err))
		return err
	}

	rate := rsp.Rate

	report := &billingpb.VatReport{
		Id:                 primitive.NewObjectID().Hex(),
		Country:            country.IsoCodeA2,
		VatRate:            rate,
		Currency:           country.Currency,
		Status:             pkg.VatReportStatusThreshold,
		CorrectionAmount:   0,
		CreatedAt:          ptypes.TimestampNow(),
		UpdatedAt:          ptypes.TimestampNow(),
		OperatingCompanyId: operatingCompanyId,
	}

	report.DateFrom, err = ptypes.TimestampProto(from)
	if err != nil {
		return err
	}
	report.DateTo, err = ptypes.TimestampProto(to)
	if err != nil {
		return err
	}
	report.PayUntilDate, err = ptypes.TimestampProto(to.AddDate(0, 0, int(country.VatDeadlineDays)))
	if err != nil {
		return err
	}

	countryTurnover, err := h.Service.turnoverRepository.Get(ctx, operatingCompanyId, country.IsoCodeA2, from.Year())

	if err != nil {
		zap.S().Warn(
			errorMsgVatReportTurnoverNotFound,
			zap.String("country", country.IsoCodeA2),
			zap.Any("year", from.Year()),
			zap.Error(err),
		)
		return nil
	}
	report.CountryAnnualTurnover = tools.FormatAmount(countryTurnover.Amount)

	worldTurnover, err := h.Service.turnoverRepository.Get(ctx, operatingCompanyId, "", from.Year())

	if err != nil {
		return err
	}
	report.WorldAnnualTurnover = worldTurnover.Amount

	targetCurrency := ""
	if country.VatEnabled {
		targetCurrency = country.VatCurrency
	}
	if targetCurrency == "" {
		targetCurrency = country.Currency
	}
	if worldTurnover.Currency != targetCurrency {
		report.WorldAnnualTurnover, err = h.exchangeAmount(
			ctx,
			worldTurnover.Currency,
			targetCurrency,
			worldTurnover.Amount,
			country.VatCurrencyRatesSource,
		)

		if err != nil {
			return err
		}
	}

	report.WorldAnnualTurnover = tools.FormatAmount(report.WorldAnnualTurnover)

	isLastDayOfPeriod := h.date.Unix() == to.Unix()
	isCurrencyRatesPolicyOnDay := country.VatCurrencyRatesPolicy == pkg.VatCurrencyRatesPolicyOnDay
	report.AmountsApproximate = !(isCurrencyRatesPolicyOnDay || (!isCurrencyRatesPolicyOnDay && isLastDayOfPeriod))

	res, err := h.orderViewRepository.GetVatSummary(h.ctx, operatingCompanyId, country.IsoCodeA2, false, from, to)

	if err != nil {
		return err
	}

	if len(res) == 1 {
		report.TransactionsCount = res[0].Count
		report.GrossRevenue = tools.FormatAmount(res[0].PaymentGrossRevenueLocal - res[0].PaymentRefundGrossRevenueLocal)
		report.VatAmount = tools.FormatAmount(res[0].PaymentTaxFeeLocal - res[0].PaymentRefundTaxFeeLocal)
		report.FeesAmount = res[0].PaymentFeesTotal + res[0].PaymentRefundFeesTotal
	}

	res, err = h.orderViewRepository.GetVatSummary(h.ctx, operatingCompanyId, country.IsoCodeA2, true, from, to)

	if err != nil {
		return err
	}

	if len(res) == 1 {
		report.TransactionsCount += res[0].Count
		report.DeductionAmount = tools.FormatAmount(res[0].PaymentRefundTaxFeeLocal)
		report.FeesAmount += res[0].PaymentFeesTotal + res[0].PaymentRefundFeesTotal
	}

	report.FeesAmount = tools.FormatAmount(report.FeesAmount)

	vr, err := h.vatReportRepository.GetByCountryPeriod(ctx, report.Country, from, to)

	if err == mongo.ErrNoDocuments {
		return h.Service.vatReportRepository.Insert(ctx, report)
	}

	if err != nil {
		return err
	}

	report.Id = vr.Id
	report.CreatedAt = vr.CreatedAt
	return h.Service.updateVatReport(ctx, report)

}

func (h *vatReportProcessor) processAccountingEntriesForPeriod(ctx context.Context, country *billingpb.Country) error {
	if !country.VatEnabled {
		return errorVatReportNotEnabledForCountry
	}

	if country.VatCurrencyRatesPolicy == pkg.VatCurrencyRatesPolicyOnDay {
		return nil
	}

	if country.VatCurrencyRatesPolicy == pkg.VatCurrencyRatesPolicyAvgMonth {
		zap.S().Warnf(
			errorMsgVatReportRatesPolicyNotImplemented,
			"country", country.IsoCodeA2,
			"policy", pkg.VatCurrencyRatesPolicyAvgMonth,
		)
		return nil
	}

	from, to, err := h.Service.getVatReportTimeForDate(country.VatPeriodMonth, h.date)
	if err != nil {
		zap.L().Error(
			errorMsgVatReportCantGetTimeForDate,
			zap.Error(err),
			zap.String("country", country.IsoCodeA2),
			zap.Time("date", h.date),
		)
		return nil
	}

	aes, err := h.accountingRepository.FindByTypeCountryDates(
		h.ctx,
		country.IsoCodeA2,
		AccountingEntriesLocalAmountsUpdate,
		now.New(from).BeginningOfDay(),
		now.New(to).EndOfDay(),
	)

	if err != nil || len(aes) == 0 {
		return nil
	}

	var aesRealTaxFee = make(map[string]*billingpb.AccountingEntry)
	for _, ae := range aes {
		if ae.Type != pkg.AccountingEntryTypeRealTaxFee {
			continue
		}
		aesRealTaxFee[ae.Source.Id] = ae
	}

	var list []*billingpb.AccountingEntry

	for _, ae := range aes {
		if ae.Type == pkg.AccountingEntryTypeRealTaxFee {
			continue
		}
		amount := ae.LocalAmount
		if ae.LocalCurrency != ae.OriginalCurrency {
			amount, err = h.exchangeAmount(
				ctx,
				ae.OriginalCurrency,
				ae.LocalCurrency,
				ae.OriginalAmount,
				country.VatCurrencyRatesSource,
			)

			if err != nil {
				return err
			}
		}
		if ae.Type == pkg.AccountingEntryTypeCentralBankTaxFee {
			realTaxFee, ok := aesRealTaxFee[ae.Source.Id]
			if ok {
				ae.LocalAmount = ae.LocalAmount - realTaxFee.LocalAmount
			}
		}
		if amount == ae.LocalAmount {
			continue
		}

		list = append(list, ae)

		h.orderViewUpdateIds[ae.Source.Id] = true
	}

	if len(list) == 0 {
		return nil
	}

	if err = h.accountingRepository.BulkWrite(h.ctx, list); err != nil {
		return err
	}

	return nil
}

func (h *vatReportProcessor) exchangeAmount(
	ctx context.Context,
	from, to string,
	amount float64,
	source string,
) (float64, error) {
	req := &currenciespb.ExchangeCurrencyByDateCommonRequest{
		From:              from,
		To:                to,
		RateType:          currenciespb.RateTypeCentralbanks,
		ExchangeDirection: currenciespb.ExchangeDirectionBuy,
		Source:            source,
		Amount:            amount,
		Datetime:          h.ts,
	}

	rsp, err := h.Service.curService.ExchangeCurrencyByDateCommon(ctx, req)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(errorFieldService, "CurrencyRatesService"),
			zap.String(errorFieldMethod, "ExchangeCurrencyCurrentCommon"),
			zap.Any(errorFieldRequest, req),
		)

		return 0, errorVatReportCurrencyExchangeFailed
	}
	return rsp.ExchangedAmount, nil
}

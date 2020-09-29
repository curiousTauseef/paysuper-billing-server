package service

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	"github.com/jinzhu/now"
	"github.com/paysuper/paysuper-billing-server/internal/helper"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/errors"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/currenciespb"
	tools "github.com/paysuper/paysuper-tools/number"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"time"
)

const (
	errorCannotCalculateTurnoverCountry = "can not calculate turnover for country"
	errorCannotCalculateTurnoverWorld   = "can not calculate turnover for world"
)

var (
	errorTurnoversCurrencyRatesPolicyNotSupported = errors.NewBillingServerErrorMsg("to000001", "vat currency rates policy not supported")
	errorTurnoversExchangeFailed                  = errors.NewBillingServerErrorMsg("to000002", "currency exchange failed")
)

func (s *Service) CalcAnnualTurnovers(ctx context.Context, req *billingpb.EmptyRequest, res *billingpb.EmptyResponse) error {
	operatingCompanies, err := s.operatingCompanyRepository.GetAll(ctx)
	if err != nil {
		return err
	}

	countries, err := s.country.FindByVatEnabled(ctx)
	if err != nil {
		return err
	}
	for _, operatingCompany := range operatingCompanies {
		for _, country := range countries.Countries {
			err = s.calcAnnualTurnover(ctx, country.IsoCodeA2, operatingCompany.Id)
			if err != nil {
				if err != errorTurnoversCurrencyRatesPolicyNotSupported {
					zap.L().Error(errorCannotCalculateTurnoverCountry,
						zap.String("country", country.IsoCodeA2),
						zap.Error(err))
					return err
				}
				zap.L().Warn(errorCannotCalculateTurnoverCountry,
					zap.String("country", country.IsoCodeA2),
					zap.Error(err))
			}
		}

		err = s.calcAnnualTurnover(ctx, "", operatingCompany.Id)
		if err != nil {
			zap.L().Error(errorCannotCalculateTurnoverWorld, zap.Error(err))
			return err
		}
	}

	return nil
}

func (s *Service) calcAnnualTurnover(ctx context.Context, countryCode, operatingCompanyId string) error {

	var (
		targetCurrency = "EUR"
		ratesType      = currenciespb.RateTypeOxr
		ratesSource    = ""
		currencyPolicy = pkg.VatCurrencyRatesPolicyOnDay
		year           = now.BeginningOfYear()
		from           = now.BeginningOfYear()
		to             = now.EndOfDay()
		amount         = float64(0)
		VatPeriodMonth = int32(0)
		err            error
	)

	if countryCode != "" {
		country, err := s.country.GetByIsoCodeA2(ctx, countryCode)
		if err != nil {
			return errorCountryNotFound
		}
		if country.VatEnabled {
			targetCurrency = country.VatCurrency
		}
		if targetCurrency == "" {
			targetCurrency = country.Currency
		}
		VatPeriodMonth = country.VatPeriodMonth
		currencyPolicy = country.VatCurrencyRatesPolicy
		ratesType = currenciespb.RateTypeCentralbanks
		ratesSource = country.VatCurrencyRatesSource
	}

	switch currencyPolicy {
	case pkg.VatCurrencyRatesPolicyOnDay:
		amount, err = s.getTurnover(ctx, from, to, countryCode, targetCurrency, currencyPolicy, ratesType, ratesSource, operatingCompanyId)
		break
	case pkg.VatCurrencyRatesPolicyLastDay:
		from, to, err = s.getLastVatReportTime(VatPeriodMonth)
		if err != nil {
			return err
		}

		count := 0
		for from.Unix() >= year.Unix() {
			diff := to.Sub(time.Now())
			if diff.Seconds() > 0 {
				to = now.EndOfDay()
			}
			amnt, err := s.getTurnover(ctx, from, to, countryCode, targetCurrency, currencyPolicy, ratesType, ratesSource, operatingCompanyId)
			if err != nil {
				return err
			}
			amount += amnt
			count++
			from, to, err = s.getVatReportTimeForDate(VatPeriodMonth, from.AddDate(0, 0, -1))
			if err != nil {
				return err
			}
		}
		break
	default:
		err = errorTurnoversCurrencyRatesPolicyNotSupported
		return err
	}

	at := &billingpb.AnnualTurnover{
		Year:               int32(year.Year()),
		Country:            countryCode,
		Amount:             tools.FormatAmount(amount),
		Currency:           targetCurrency,
		OperatingCompanyId: operatingCompanyId,
	}

	err = s.turnoverRepository.Upsert(ctx, at)

	if err != nil {
		return err
	}
	return nil

}

func (s *Service) getTurnover(
	ctx context.Context,
	from, to time.Time,
	countryCode, targetCurrency, currencyPolicy, ratesType, ratesSource, operatingCompanyId string,
) (amount float64, err error) {
	policy := []string{pkg.VatCurrencyRatesPolicyOnDay, pkg.VatCurrencyRatesPolicyLastDay}

	if !helper.Contains(policy, currencyPolicy) {
		err = errorTurnoversCurrencyRatesPolicyNotSupported
		return
	}

	res, err := s.orderViewRepository.GetTurnoverSummary(ctx, operatingCompanyId, countryCode, currencyPolicy, from, to)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = nil
			return
		}
		return
	}

	toTimestamp, err := ptypes.TimestampProto(to)
	if err != nil {
		err = errorTurnoversExchangeFailed
		return
	}

	for _, v := range res {
		if v.Id == targetCurrency {
			amount += v.Amount
			continue
		}

		req := &currenciespb.ExchangeCurrencyByDateCommonRequest{
			From:              v.Id,
			To:                targetCurrency,
			RateType:          ratesType,
			ExchangeDirection: currenciespb.ExchangeDirectionBuy,
			Source:            ratesSource,
			Amount:            v.Amount,
			Datetime:          toTimestamp,
		}

		rsp, err := s.curService.ExchangeCurrencyByDateCommon(ctx, req)

		if err != nil {
			zap.L().Error(
				pkg.ErrorGrpcServiceCallFailed,
				zap.Error(err),
				zap.String(errorFieldService, "CurrencyRatesService"),
				zap.String(errorFieldMethod, "ExchangeCurrencyCurrentCommon"),
				zap.Any(errorFieldRequest, req),
			)

			return 0, errorTurnoversExchangeFailed
		} else {
			amount += rsp.ExchangedAmount
		}
	}

	return
}

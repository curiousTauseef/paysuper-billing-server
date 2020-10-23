package mocks

import (
	"context"
	"errors"
	"github.com/golang/protobuf/ptypes"
	"github.com/micro/go-micro/client"
	"github.com/paysuper/paysuper-proto/go/currenciespb"
	tools "github.com/paysuper/paysuper-tools/number"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	eurPriceInRub      = float64(72)
	eurPriceInRubCb    = float64(72.5)
	eurPriceInRubStock = float64(70.5)

	usdPriceInRub      = float64(65)
	usdPriceInRubCb    = float64(65.5)
	usdPriceInRubStock = float64(64)

	merchantMarkupValuePercent = float64(2)
	commonMarkupValuePercent   = float64(0)
)

var (
	MerchantIdMock = primitive.NewObjectID().Hex()
)

type CurrencyServiceMockOk struct{}
type CurrencyServiceMockError struct{}

func NewCurrencyServiceMockOk() currenciespb.CurrencyRatesService {
	return &CurrencyServiceMockOk{}
}

func NewCurrencyServiceMockError() currenciespb.CurrencyRatesService {
	return &CurrencyServiceMockError{}
}

func (s *CurrencyServiceMockOk) GetCurrenciesPrecision(ctx context.Context, in *currenciespb.EmptyRequest, opts ...client.CallOption) (*currenciespb.CurrenciesPrecisionResponse, error) {
	return &currenciespb.CurrenciesPrecisionResponse{
		Values: map[string]int32{
			"AED": 2,
			"ALL": 2,
			"AMD": 2,
			"ARS": 2,
			"AUD": 2,
			"BGN": 2,
			"BHD": 3,
			"BRL": 2,
			"BYN": 2,
			"CAD": 2,
			"CHF": 2,
			"CLP": 0,
			"CNY": 2,
			"COP": 2,
			"CZK": 2,
			"DKK": 2,
			"EGP": 2,
			"EUR": 2,
			"GBP": 2,
			"GHS": 2,
			"HKD": 2,
			"HRK": 2,
			"HUF": 2,
			"IDR": 2,
			"ILS": 2,
			"INR": 2,
			"ISK": 0,
			"JPY": 0,
			"KES": 2,
			"KRW": 0,
			"KWD": 3,
			"KZT": 2,
			"MXN": 2,
			"MYR": 2,
			"NOK": 2,
			"NZD": 2,
			"PEN": 2,
			"PHP": 2,
			"PLN": 2,
			"QAR": 2,
			"RON": 2,
			"RSD": 2,
			"RUB": 2,
			"SAR": 2,
			"SEK": 2,
			"SGD": 2,
			"THB": 2,
			"TRY": 2,
			"TWD": 2,
			"TZS": 2,
			"UAH": 2,
			"USD": 2,
			"UYU": 2,
			"VND": 0,
			"ZAR": 2,
		},
	}, nil
}

func (s *CurrencyServiceMockOk) GetRateCurrentCommon(
	ctx context.Context,
	in *currenciespb.GetRateCurrentCommonRequest,
	opts ...client.CallOption,
) (*currenciespb.RateData, error) {
	return &currenciespb.RateData{}, nil
}

func (s *CurrencyServiceMockOk) GetRateByDateCommon(
	ctx context.Context,
	in *currenciespb.GetRateByDateCommonRequest,
	opts ...client.CallOption,
) (*currenciespb.RateData, error) {
	return &currenciespb.RateData{
		Id:        primitive.NewObjectID().Hex(),
		CreatedAt: ptypes.TimestampNow(),
		Pair:      in.From + in.To,
		Rate:      3,
		Source:    primitive.NewObjectID().Hex(),
		Volume:    3,
	}, nil
}

func (s *CurrencyServiceMockOk) GetRateCurrentForMerchant(
	ctx context.Context,
	in *currenciespb.GetRateCurrentForMerchantRequest,
	opts ...client.CallOption,
) (*currenciespb.RateData, error) {
	return &currenciespb.RateData{
		Id:        primitive.NewObjectID().Hex(),
		CreatedAt: ptypes.TimestampNow(),
		Pair:      in.From + in.To,
		Rate:      3,
		Source:    primitive.NewObjectID().Hex(),
		Volume:    3,
	}, nil
}

func (s *CurrencyServiceMockOk) GetRateByDateForMerchant(
	ctx context.Context,
	in *currenciespb.GetRateByDateForMerchantRequest,
	opts ...client.CallOption,
) (*currenciespb.RateData, error) {
	return &currenciespb.RateData{}, nil
}

func (s *CurrencyServiceMockOk) ExchangeCurrencyCurrentCommon(
	ctx context.Context,
	in *currenciespb.ExchangeCurrencyCurrentCommonRequest,
	opts ...client.CallOption,
) (*currenciespb.ExchangeCurrencyResponse, error) {

	divider := float64(1)

	if in.RateType == currenciespb.RateTypePaysuper || in.RateType == currenciespb.RateTypeStock {

		switch in.ExchangeDirection {

		case currenciespb.ExchangeDirectionSell:
			divider = 1 - (commonMarkupValuePercent / 100)

		case currenciespb.ExchangeDirectionBuy:
			divider = 1 + (commonMarkupValuePercent / 100)
		}
	}

	newIn := &currenciespb.ExchangeCurrencyCurrentForMerchantRequest{
		From:              in.From,
		To:                in.To,
		RateType:          in.RateType,
		Source:            in.Source,
		Amount:            in.Amount,
		MerchantId:        "",
		ExchangeDirection: in.ExchangeDirection,
	}

	return s.exchangeCurrency(divider, newIn)
}

func (s *CurrencyServiceMockOk) ExchangeCurrencyCurrentForMerchant(
	ctx context.Context,
	in *currenciespb.ExchangeCurrencyCurrentForMerchantRequest,
	opts ...client.CallOption,
) (*currenciespb.ExchangeCurrencyResponse, error) {

	divider := float64(1)

	switch in.ExchangeDirection {

	case currenciespb.ExchangeDirectionSell:
		divider = 1 - (merchantMarkupValuePercent / 100)

	case currenciespb.ExchangeDirectionBuy:
		divider = 1 + (merchantMarkupValuePercent / 100)
	}

	return s.exchangeCurrency(divider, in)
}

func (s *CurrencyServiceMockOk) exchangeCurrency(
	divider float64,
	in *currenciespb.ExchangeCurrencyCurrentForMerchantRequest,
) (*currenciespb.ExchangeCurrencyResponse, error) {

	if in.From == "RUB" && in.To == "EUR" {
		rate := tools.ToPrecise((1 / eurPriceInRub) / divider)
		a := in.Amount * rate

		return &currenciespb.ExchangeCurrencyResponse{
			ExchangedAmount: tools.ToPrecise(a),
			ExchangeRate:    rate,
			Correction:      0,
			OriginalRate:    rate,
		}, nil
	}
	if in.From == "EUR" && in.To == "RUB" {
		rate := tools.ToPrecise(eurPriceInRub / divider)
		a := in.Amount * rate

		return &currenciespb.ExchangeCurrencyResponse{
			ExchangedAmount: tools.ToPrecise(a),
			ExchangeRate:    rate,
			Correction:      0,
			OriginalRate:    rate,
		}, nil
	}

	if in.From == "USD" && in.To == "EUR" {
		if in.RateType == currenciespb.RateTypeStock {
			rate := tools.ToPrecise((usdPriceInRubStock / eurPriceInRubStock) / divider)
			a := in.Amount * rate

			return &currenciespb.ExchangeCurrencyResponse{
				ExchangedAmount: tools.ToPrecise(a),
				ExchangeRate:    rate,
				Correction:      0,
				OriginalRate:    rate,
			}, nil
		}

		if in.RateType == currenciespb.RateTypeCentralbanks {
			rate := tools.ToPrecise((usdPriceInRubCb / eurPriceInRubCb) / divider)
			a := in.Amount * rate

			return &currenciespb.ExchangeCurrencyResponse{
				ExchangedAmount: tools.ToPrecise(a),
				ExchangeRate:    rate,
				Correction:      0,
				OriginalRate:    rate,
			}, nil
		}

		rate := tools.ToPrecise((usdPriceInRub / eurPriceInRub) / divider)
		a := in.Amount * rate
		return &currenciespb.ExchangeCurrencyResponse{
			ExchangedAmount: tools.ToPrecise(a),
			ExchangeRate:    rate,
			Correction:      0,
			OriginalRate:    rate,
		}, nil
	}

	if in.From == "EUR" && in.To == "USD" {
		if in.RateType == currenciespb.RateTypeStock {
			rate := tools.ToPrecise((eurPriceInRubStock / usdPriceInRubStock) / divider)
			a := in.Amount * rate

			return &currenciespb.ExchangeCurrencyResponse{
				ExchangedAmount: tools.ToPrecise(a),
				ExchangeRate:    rate,
				Correction:      0,
				OriginalRate:    rate,
			}, nil
		}

		if in.RateType == currenciespb.RateTypeCentralbanks {
			rate := tools.ToPrecise((eurPriceInRubCb / usdPriceInRubCb) / divider)
			a := in.Amount * rate

			return &currenciespb.ExchangeCurrencyResponse{
				ExchangedAmount: tools.ToPrecise(a),
				ExchangeRate:    rate,
				Correction:      0,
				OriginalRate:    rate,
			}, nil
		}

		rate := tools.ToPrecise((eurPriceInRub / usdPriceInRub) / divider)
		a := in.Amount * rate
		return &currenciespb.ExchangeCurrencyResponse{
			ExchangedAmount: tools.ToPrecise(a),
			ExchangeRate:    rate,
			Correction:      0,
			OriginalRate:    rate,
		}, nil
	}

	if in.From == "USD" && in.To == "RUB" {
		if in.RateType == currenciespb.RateTypeStock {
			rate := tools.ToPrecise(usdPriceInRubStock / divider)
			a := in.Amount * rate

			return &currenciespb.ExchangeCurrencyResponse{
				ExchangedAmount: tools.ToPrecise(a),
				ExchangeRate:    rate,
				Correction:      0,
				OriginalRate:    rate,
			}, nil
		}

		if in.RateType == currenciespb.RateTypeCentralbanks {
			rate := tools.ToPrecise(usdPriceInRubCb / divider)
			a := in.Amount * rate

			return &currenciespb.ExchangeCurrencyResponse{
				ExchangedAmount: tools.ToPrecise(a),
				ExchangeRate:    rate,
				Correction:      0,
				OriginalRate:    rate,
			}, nil
		}

		rate := tools.ToPrecise(usdPriceInRub / divider)
		a := in.Amount * rate
		return &currenciespb.ExchangeCurrencyResponse{
			ExchangedAmount: tools.ToPrecise(a),
			ExchangeRate:    rate,
			Correction:      0,
			OriginalRate:    rate,
		}, nil
	}
	if in.From == "RUB" && in.To == "USD" {
		if in.RateType == currenciespb.RateTypeStock {
			rate := tools.ToPrecise((1 / usdPriceInRubStock) / divider)
			a := in.Amount * rate

			return &currenciespb.ExchangeCurrencyResponse{
				ExchangedAmount: tools.ToPrecise(a),
				ExchangeRate:    rate,
				Correction:      0,
				OriginalRate:    rate,
			}, nil
		}

		if in.RateType == currenciespb.RateTypeCentralbanks {
			rate := tools.ToPrecise((1 / usdPriceInRubCb) / divider)
			a := in.Amount * rate

			return &currenciespb.ExchangeCurrencyResponse{
				ExchangedAmount: tools.ToPrecise(a),
				ExchangeRate:    rate,
				Correction:      0,
				OriginalRate:    rate,
			}, nil
		}

		rate := tools.ToPrecise((1 / usdPriceInRub) / divider)
		a := in.Amount * rate
		return &currenciespb.ExchangeCurrencyResponse{
			ExchangedAmount: tools.ToPrecise(a),
			ExchangeRate:    rate,
			Correction:      0,
			OriginalRate:    rate,
		}, nil
	}

	return &currenciespb.ExchangeCurrencyResponse{
		ExchangedAmount: 10,
		ExchangeRate:    0.25,
		Correction:      2,
		OriginalRate:    0.5,
	}, nil
}

func (s *CurrencyServiceMockOk) ExchangeCurrencyByDateCommon(
	ctx context.Context,
	in *currenciespb.ExchangeCurrencyByDateCommonRequest,
	opts ...client.CallOption,
) (*currenciespb.ExchangeCurrencyResponse, error) {
	if in.From == "TRY" && in.To == "EUR" {
		return &currenciespb.ExchangeCurrencyResponse{
			ExchangedAmount: in.Amount * 6,
		}, nil
	}
	if in.From == "TRY" && in.To == "RUB" {
		return &currenciespb.ExchangeCurrencyResponse{
			ExchangedAmount: in.Amount * 10,
		}, nil
	}

	divider := float64(1)

	if in.RateType == currenciespb.RateTypePaysuper || in.RateType == currenciespb.RateTypeStock {

		switch in.ExchangeDirection {

		case currenciespb.ExchangeDirectionSell:
			divider = 1 - (commonMarkupValuePercent / 100)

		case currenciespb.ExchangeDirectionBuy:
			divider = 1 + (commonMarkupValuePercent / 100)
		}
	}

	newIn := &currenciespb.ExchangeCurrencyCurrentForMerchantRequest{
		From:              in.From,
		To:                in.To,
		RateType:          in.RateType,
		Source:            in.Source,
		Amount:            in.Amount,
		MerchantId:        "",
		ExchangeDirection: in.ExchangeDirection,
	}

	return s.exchangeCurrency(divider, newIn)
}

func (s *CurrencyServiceMockOk) ExchangeCurrencyByDateForMerchant(
	ctx context.Context,
	in *currenciespb.ExchangeCurrencyByDateForMerchantRequest,
	opts ...client.CallOption,
) (*currenciespb.ExchangeCurrencyResponse, error) {

	divider := float64(1)

	switch in.ExchangeDirection {

	case currenciespb.ExchangeDirectionSell:
		divider = 1 - (merchantMarkupValuePercent / 100)

	case currenciespb.ExchangeDirectionBuy:
		divider = 1 + (merchantMarkupValuePercent / 100)
	}

	req := &currenciespb.ExchangeCurrencyCurrentForMerchantRequest{
		From:              in.From,
		To:                in.To,
		RateType:          in.RateType,
		Source:            in.Source,
		Amount:            in.Amount,
		MerchantId:        in.MerchantId,
		ExchangeDirection: in.ExchangeDirection,
	}

	return s.exchangeCurrency(divider, req)
}

func (s *CurrencyServiceMockOk) GetCommonRateCorrectionRule(
	ctx context.Context,
	in *currenciespb.CommonCorrectionRuleRequest,
	opts ...client.CallOption,
) (*currenciespb.CorrectionRule, error) {
	return &currenciespb.CorrectionRule{}, nil
}

func (s *CurrencyServiceMockOk) GetMerchantRateCorrectionRule(
	ctx context.Context,
	in *currenciespb.MerchantCorrectionRuleRequest,
	opts ...client.CallOption,
) (*currenciespb.CorrectionRule, error) {
	return &currenciespb.CorrectionRule{}, nil
}

func (s *CurrencyServiceMockOk) AddCommonRateCorrectionRule(
	ctx context.Context,
	in *currenciespb.CommonCorrectionRule,
	opts ...client.CallOption,
) (*currenciespb.EmptyResponse, error) {
	return &currenciespb.EmptyResponse{}, nil
}

func (s *CurrencyServiceMockOk) AddMerchantRateCorrectionRule(
	ctx context.Context,
	in *currenciespb.CorrectionRule,
	opts ...client.CallOption,
) (*currenciespb.EmptyResponse, error) {
	return &currenciespb.EmptyResponse{}, nil
}

func (s *CurrencyServiceMockOk) SetPaysuperCorrectionCorridor(
	ctx context.Context,
	in *currenciespb.CorrectionCorridor,
	opts ...client.CallOption,
) (*currenciespb.EmptyResponse, error) {
	return &currenciespb.EmptyResponse{}, nil
}

func (s *CurrencyServiceMockOk) GetSupportedCurrencies(
	ctx context.Context,
	in *currenciespb.EmptyRequest,
	opts ...client.CallOption,
) (*currenciespb.CurrenciesList, error) {
	return &currenciespb.CurrenciesList{
		Currencies: []string{"USD", "EUR", "RUB", "GBP"},
	}, nil
}

func (s *CurrencyServiceMockOk) GetSettlementCurrencies(
	ctx context.Context,
	in *currenciespb.EmptyRequest,
	opts ...client.CallOption,
) (*currenciespb.CurrenciesList, error) {
	return &currenciespb.CurrenciesList{Currencies: []string{"USD", "EUR"}}, nil
}

func (s *CurrencyServiceMockOk) GetPriceCurrencies(
	ctx context.Context,
	in *currenciespb.EmptyRequest,
	opts ...client.CallOption,
) (*currenciespb.CurrenciesList, error) {
	return &currenciespb.CurrenciesList{
		Currencies: []string{"AED", "ARS", "AUD", "BHD", "BRL", "CAD", "CHF", "CLP", "CNY", "COP", "CRC", "CZK", "DKK",
			"EGP", "EUR", "GBP", "HKD", "HRK", "HUF", "IDR", "ILS", "INR", "JPY", "KRW", "KZT", "MXN",
			"MYR", "NOK", "NZD", "PEN", "PHP", "PLN", "QAR", "RON", "RSD", "RUB", "SAR", "SEK", "SGD",
			"THB", "TRY", "TWD", "USD", "VND", "ZAR"},
	}, nil
}

func (s *CurrencyServiceMockOk) GetVatCurrencies(
	ctx context.Context,
	in *currenciespb.EmptyRequest,
	opts ...client.CallOption,
) (*currenciespb.CurrenciesList, error) {
	return &currenciespb.CurrenciesList{}, nil
}

func (s *CurrencyServiceMockOk) GetAccountingCurrencies(
	ctx context.Context,
	in *currenciespb.EmptyRequest,
	opts ...client.CallOption,
) (*currenciespb.CurrenciesList, error) {
	return &currenciespb.CurrenciesList{}, nil
}

func (s *CurrencyServiceMockError) GetCurrenciesPrecision(ctx context.Context, in *currenciespb.EmptyRequest, opts ...client.CallOption) (*currenciespb.CurrenciesPrecisionResponse, error) {
	panic("implement me")
}

func (s *CurrencyServiceMockError) GetRateCurrentCommon(
	ctx context.Context,
	in *currenciespb.GetRateCurrentCommonRequest,
	opts ...client.CallOption,
) (*currenciespb.RateData, error) {
	return nil, errors.New(SomeError)
}

func (s *CurrencyServiceMockError) GetRateByDateCommon(
	ctx context.Context,
	in *currenciespb.GetRateByDateCommonRequest,
	opts ...client.CallOption,
) (*currenciespb.RateData, error) {
	return nil, errors.New(SomeError)
}

func (s *CurrencyServiceMockError) GetRateCurrentForMerchant(
	ctx context.Context,
	in *currenciespb.GetRateCurrentForMerchantRequest,
	opts ...client.CallOption,
) (*currenciespb.RateData, error) {
	if in.MerchantId == MerchantIdMock {
		return &currenciespb.RateData{Rate: 10}, nil
	}

	return nil, errors.New(SomeError)
}

func (s *CurrencyServiceMockError) GetRateByDateForMerchant(
	ctx context.Context,
	in *currenciespb.GetRateByDateForMerchantRequest,
	opts ...client.CallOption,
) (*currenciespb.RateData, error) {
	return nil, errors.New(SomeError)
}

func (s *CurrencyServiceMockError) ExchangeCurrencyCurrentCommon(
	ctx context.Context,
	in *currenciespb.ExchangeCurrencyCurrentCommonRequest,
	opts ...client.CallOption,
) (*currenciespb.ExchangeCurrencyResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *CurrencyServiceMockError) ExchangeCurrencyCurrentForMerchant(
	ctx context.Context,
	in *currenciespb.ExchangeCurrencyCurrentForMerchantRequest,
	opts ...client.CallOption,
) (*currenciespb.ExchangeCurrencyResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *CurrencyServiceMockError) ExchangeCurrencyByDateCommon(
	ctx context.Context,
	in *currenciespb.ExchangeCurrencyByDateCommonRequest,
	opts ...client.CallOption,
) (*currenciespb.ExchangeCurrencyResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *CurrencyServiceMockError) ExchangeCurrencyByDateForMerchant(
	ctx context.Context,
	in *currenciespb.ExchangeCurrencyByDateForMerchantRequest,
	opts ...client.CallOption,
) (*currenciespb.ExchangeCurrencyResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *CurrencyServiceMockError) GetCommonRateCorrectionRule(
	ctx context.Context,
	in *currenciespb.CommonCorrectionRuleRequest,
	opts ...client.CallOption,
) (*currenciespb.CorrectionRule, error) {
	return nil, errors.New(SomeError)
}

func (s *CurrencyServiceMockError) GetMerchantRateCorrectionRule(
	ctx context.Context,
	in *currenciespb.MerchantCorrectionRuleRequest,
	opts ...client.CallOption,
) (*currenciespb.CorrectionRule, error) {
	return nil, errors.New(SomeError)
}

func (s *CurrencyServiceMockError) AddCommonRateCorrectionRule(
	ctx context.Context,
	in *currenciespb.CommonCorrectionRule,
	opts ...client.CallOption,
) (*currenciespb.EmptyResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *CurrencyServiceMockError) AddMerchantRateCorrectionRule(
	ctx context.Context,
	in *currenciespb.CorrectionRule,
	opts ...client.CallOption,
) (*currenciespb.EmptyResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *CurrencyServiceMockError) SetPaysuperCorrectionCorridor(
	ctx context.Context,
	in *currenciespb.CorrectionCorridor,
	opts ...client.CallOption,
) (*currenciespb.EmptyResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *CurrencyServiceMockError) GetSupportedCurrencies(
	ctx context.Context,
	in *currenciespb.EmptyRequest,
	opts ...client.CallOption,
) (*currenciespb.CurrenciesList, error) {
	return nil, errors.New(SomeError)
}

func (s *CurrencyServiceMockError) GetSettlementCurrencies(
	ctx context.Context,
	in *currenciespb.EmptyRequest,
	opts ...client.CallOption,
) (*currenciespb.CurrenciesList, error) {
	return nil, errors.New(SomeError)
}

func (s *CurrencyServiceMockError) GetPriceCurrencies(
	ctx context.Context,
	in *currenciespb.EmptyRequest,
	opts ...client.CallOption,
) (*currenciespb.CurrenciesList, error) {
	return nil, errors.New(SomeError)
}

func (s *CurrencyServiceMockError) GetVatCurrencies(
	ctx context.Context,
	in *currenciespb.EmptyRequest,
	opts ...client.CallOption,
) (*currenciespb.CurrenciesList, error) {
	return nil, errors.New(SomeError)
}

func (s *CurrencyServiceMockError) GetAccountingCurrencies(
	ctx context.Context,
	in *currenciespb.EmptyRequest,
	opts ...client.CallOption,
) (*currenciespb.CurrenciesList, error) {
	return nil, errors.New(SomeError)
}

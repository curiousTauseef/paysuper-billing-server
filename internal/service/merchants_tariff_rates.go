package service

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

func (s *Service) getMerchantTariffRates(
	ctx context.Context,
	in *billingpb.GetMerchantTariffRatesRequest,
) (*billingpb.GetMerchantTariffRatesResponseItems, error) {
	mccCode, err := getMccByOperationsType(in.MerchantOperationsType)

	if err != nil {
		return nil, err
	}

	payment, err := s.merchantPaymentTariffsRepository.Find(ctx, in.HomeRegion, in.PayerRegion, mccCode, in.MinAmount, in.MaxAmount)

	if err != nil {
		return nil, merchantErrorUnknown
	}

	if len(payment) <= 0 {
		return nil, merchantTariffsNotFound
	}

	settings, err := s.merchantTariffsSettingsRepository.GetByMccCode(ctx, mccCode)

	if err != nil {
		return nil, merchantTariffsNotFound
	}

	tariffs := &billingpb.GetMerchantTariffRatesResponseItems{
		Payment:       payment,
		Refund:        settings.Refund,
		Chargeback:    settings.Chargeback,
		Payout:        settings.Payout,
		MinimalPayout: settings.MinimalPayout,
		MccCode:       settings.MccCode,
	}

	return tariffs, nil
}

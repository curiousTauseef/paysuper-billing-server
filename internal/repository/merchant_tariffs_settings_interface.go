package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// MerchantTariffsSettingsInterface is abstraction layer for working with merchant tariffs settings and representation in database.
type MerchantTariffsSettingsInterface interface {
	// Insert add the merchant tariffs settings to the collection.
	Insert(ctx context.Context, merchant *billingpb.MerchantTariffRatesSettings) error

	// GetByMccCode returns the merchant tariffs settings by mcc code.
	GetByMccCode(ctx context.Context, code string) (*billingpb.MerchantTariffRatesSettings, error)
}

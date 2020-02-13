package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// MerchantPaymentTariffsInterface is abstraction layer for working with merchant payment tariffs and representation in database.
type MerchantPaymentTariffsInterface interface {
	// MultipleInsert adds the multiple merchant payment tariffs to the collection.
	MultipleInsert(context.Context, []*billingpb.MerchantTariffRatesPayment) error

	// Find return merchant payment tariffs by home region, payer region, mcc code, min amount and max amount.
	Find(context.Context, string, string, string, float64, float64) ([]*billingpb.MerchantTariffRatesPayment, error)
}

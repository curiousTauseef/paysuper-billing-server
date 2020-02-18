package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// MerchantPaymentMethodHistoryRepositoryInterface is abstraction layer for working with merchant payment method history and representation in database.
type MerchantPaymentMethodHistoryRepositoryInterface interface {
	// Insert adds the merchant payment method history to the collection.
	Insert(context.Context, *billingpb.MerchantPaymentMethodHistory) error
}

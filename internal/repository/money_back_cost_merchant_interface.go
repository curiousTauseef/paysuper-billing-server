package repository

import (
	"context"
	internalPkg "github.com/paysuper/paysuper-billing-server/internal/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// MoneyBackCostMerchantRepositoryInterface is abstraction layer for working with cost of merchant in money back case
// and representation in database.
type MoneyBackCostMerchantRepositoryInterface interface {
	// Insert adds the cost of merchant in money back to the collection.
	Insert(context.Context, *billingpb.MoneyBackCostMerchant) error

	// MultipleInsert adds the multiple cost of merchant in money back to the collection.
	MultipleInsert(context.Context, []*billingpb.MoneyBackCostMerchant) error

	// Update updates the cost of merchant in money back in the collection.
	Update(context.Context, *billingpb.MoneyBackCostMerchant) error

	// Find returns list of cost of merchant in money back by merchantId, name, payout currency, undo reason, region,
	// country, mcc code and payment stage.
	Find(context.Context, string, string, string, string, string, string, string, int32) ([]*internalPkg.MoneyBackCostMerchantSet, error)

	// GetById returns the cost of merchant in money back by unique identity.
	GetById(context.Context, string) (*billingpb.MoneyBackCostMerchant, error)

	// Delete deleting the cost of merchant in money back.
	Delete(context.Context, *billingpb.MoneyBackCostMerchant) error

	// GetAllForMerchant returns the list of merchant cost in money back by merchant id.
	GetAllForMerchant(context.Context, string) (*billingpb.MoneyBackCostMerchantList, error)

	// Update all merchant's tariffs
	DeleteAndInsertMany(ctx context.Context, merchantId string, tariffs []*billingpb.MoneyBackCostMerchant) error
}

package repository

import (
	"context"
	internalPkg "github.com/paysuper/paysuper-billing-server/internal/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// MoneyBackCostSystemRepositoryInterface is abstraction layer for working with cost of system in money back case
// and representation in database.
type MoneyBackCostSystemRepositoryInterface interface {
	// Insert adds the cost of system in money back to the collection.
	Insert(context.Context, *billingpb.MoneyBackCostSystem) error

	// MultipleInsert adds the multiple cost of system in money back to the collection.
	MultipleInsert(context.Context, []*billingpb.MoneyBackCostSystem) error

	// Update updates the cost of system in money back in the collection.
	Update(context.Context, *billingpb.MoneyBackCostSystem) error

	// Find returns list of cost of system in money back by name, payout currency, undo reason, region, country,
	// mcc code, operating company id and payment stage.
	Find(context.Context, string, string, string, string, string, string, string, int32) ([]*internalPkg.MoneyBackCostSystemSet, error)

	// GetById returns the cost of system in money back by unique identity.
	GetById(context.Context, string) (*billingpb.MoneyBackCostSystem, error)

	// Delete deleting the cost of system in money back.
	Delete(context.Context, *billingpb.MoneyBackCostSystem) error

	// GetAllForMerchant returns the list of system cost in money back.
	GetAll(context.Context) (*billingpb.MoneyBackCostSystemList, error)
}

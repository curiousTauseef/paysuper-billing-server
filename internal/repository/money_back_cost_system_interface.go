package repository

import (
	"context"
	internalPkg "github.com/paysuper/paysuper-billing-server/internal/pkg"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

// MoneyBackCostSystemRepositoryInterface is abstraction layer for working with cost of system in money back case
// and representation in database.
type MoneyBackCostSystemRepositoryInterface interface {
	Insert(context.Context, *billingpb.MoneyBackCostSystem) error
	MultipleInsert(context.Context, []*billingpb.MoneyBackCostSystem) error
	Update(context.Context, *billingpb.MoneyBackCostSystem) error
	Find(context.Context, string, string, string, string, string, string, string, int32) ([]*internalPkg.MoneyBackCostSystemSet, error)
	GetById(context.Context, string) (*billingpb.MoneyBackCostSystem, error)
	Delete(context.Context, *billingpb.MoneyBackCostSystem) error
	GetAll(context.Context) (*billingpb.MoneyBackCostSystemList, error)
}

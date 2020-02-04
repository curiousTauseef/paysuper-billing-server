package repository

import (
	"context"
	intPkg "github.com/paysuper/paysuper-billing-server/internal/pkg"
)

const (
	collectionBinData = "bank_bin"
)

// BankBinRepositoryInterface is abstraction layer for working with bank bin and representation in database.
type BankBinRepositoryInterface interface {
	// Insert adds the price table to the collection.
	Insert(context.Context, *intPkg.BinData) error

	// MultipleInsert adds the multiple price groups to the collection.
	MultipleInsert(context.Context, []*intPkg.BinData) error

	// GetByBin returns the bank bin by bin number.
	GetByBin(context.Context, int32) (*intPkg.BinData, error)
}

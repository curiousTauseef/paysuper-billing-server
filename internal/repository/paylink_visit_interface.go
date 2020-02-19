package repository

import (
	"context"
)

// PaylinkVisitRepositoryInterface is abstraction layer for working with paylink visit and representation in database.
type PaylinkVisitRepositoryInterface interface {
	// CountPaylinkVisits counts and returns visits by identifier of paylink between days.
	CountPaylinkVisits(context.Context, string, int64, int64) (int64, error)

	// IncrVisits increment visit by paylink identifier.
	IncrVisits(ctx context.Context, id string) error
}

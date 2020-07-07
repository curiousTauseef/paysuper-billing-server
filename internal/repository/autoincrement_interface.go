package repository

import "context"

type AutoincrementRepositoryInterface interface {
	GatPayoutAutoincrementId(ctx context.Context) (int64, error)
}

package service

import (
	"context"
	"go.uber.org/zap"
)

// emulate update batching, because aggregarion pipeline, ended with $merge,
// does not return any documents in result,
// so, this query cannot be iterated with driver's BatchSize() and Next() methods
func (s *Service) updateOrderView(ctx context.Context, ids []string) error {
	var err error

	batchSize := s.cfg.OrderViewUpdateBatchSize
	count := len(ids)

	if count == 0 {
		ids, err = s.accountingRepository.GetDistinctBySourceId(ctx)

		if err != nil {
			return err
		}

		count = len(ids)
	}

	if count > 0 && count <= batchSize {
		return s.orderRepository.UpdateOrderView(ctx, ids)
	}

	var batches [][]string

	for batchSize < len(ids) {
		ids, batches = ids[batchSize:], append(batches, ids[0:batchSize:batchSize])
	}
	batches = append(batches, ids)
	for _, batchIds := range batches {
		err = s.orderRepository.UpdateOrderView(ctx, batchIds)

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) RebuildOrderView(ctx context.Context) error {
	zap.L().Info("start rebuilding order view")

	err := s.updateOrderView(ctx, []string{})

	if err != nil {
		zap.L().Error("rebuilding order view failed with error", zap.Error(err))
		return err
	}

	zap.L().Info("rebuilding order view finished successfully")

	return nil
}

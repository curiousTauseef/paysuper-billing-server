package service

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.uber.org/zap"
)

func (s *Service) GetRecommendedPriceTable(
	ctx context.Context,
	req *billingpb.RecommendedPriceTableRequest,
	res *billingpb.RecommendedPriceTableResponse,
) error {
	table, err := s.priceTableRepository.GetByRegion(ctx, req.Currency)

	if err != nil {
		zap.L().Error("Price table not found", zap.Any("req", req))
		return nil
	}

	res.Ranges = table.Ranges

	return nil
}

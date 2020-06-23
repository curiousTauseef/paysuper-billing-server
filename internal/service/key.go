package service

import (
	"bufio"
	"bytes"
	"context"
	"github.com/paysuper/paysuper-billing-server/pkg/errors"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

func (s *Service) UploadKeysFile(
	ctx context.Context,
	req *billingpb.PlatformKeysFileRequest,
	res *billingpb.PlatformKeysFileResponse,
) error {
	scanner := bufio.NewScanner(bytes.NewReader(req.File))
	count, err := s.keyRepository.CountKeysByProductPlatform(ctx, req.KeyProductId, req.PlatformId)

	if err != nil {
		zap.S().Errorf(errors.KeyErrorNotFound.Message, "err", err.Error(), "keyProductId", req.KeyProductId, "platformId", req.PlatformId)
		res.Status = billingpb.ResponseStatusNotFound
		res.Message = errors.KeyErrorNotFound
		return nil
	}

	res.TotalCount = int32(count)

	// Process key by line
	for scanner.Scan() {
		key := &billingpb.Key{
			Id:           primitive.NewObjectID().Hex(),
			Code:         scanner.Text(),
			KeyProductId: req.KeyProductId,
			PlatformId:   req.PlatformId,
		}

		if err := s.keyRepository.Insert(ctx, key); err != nil {
			zap.S().Errorf(errors.KeyErrorFailedToInsert.Message, "err", err, "key", key)
			continue
		}

		res.TotalCount++
		res.KeysProcessed++
	}

	// tell about errors
	if err = scanner.Err(); err != nil {
		zap.S().Errorf(errors.KeyErrorFileProcess.Message, "err", err.Error())
		res.Message = errors.KeyErrorFileProcess
		res.Status = billingpb.ResponseStatusBadData
		return nil
	}

	res.Status = billingpb.ResponseStatusOk

	return nil
}

func (s *Service) GetAvailableKeysCount(
	ctx context.Context,
	req *billingpb.GetPlatformKeyCountRequest,
	res *billingpb.GetPlatformKeyCountResponse,
) error {
	keyProduct, err := s.keyProductRepository.GetById(ctx, req.KeyProductId)

	if err != nil {
		zap.S().Errorf(keyProductNotFound.Message, "err", err.Error(), "keyProductId", req.KeyProductId, "platformId", req.PlatformId)
		res.Status = billingpb.ResponseStatusNotFound
		res.Message = keyProductNotFound
		return nil
	}

	if keyProduct.MerchantId != req.MerchantId {
		zap.S().Error(keyProductMerchantMismatch.Message, "keyProductId", req.KeyProductId)
		res.Status = billingpb.ResponseStatusNotFound
		res.Message = keyProductMerchantMismatch
		return nil
	}

	count, err := s.keyRepository.CountKeysByProductPlatform(ctx, req.KeyProductId, req.PlatformId)

	if err != nil {
		zap.S().Errorf(errors.KeyErrorNotFound.Message, "err", err.Error(), "keyProductId", req.KeyProductId, "platformId", req.PlatformId)
		res.Status = billingpb.ResponseStatusNotFound
		res.Message = errors.KeyErrorNotFound
		return nil
	}

	res.Count = int32(count)
	res.Status = billingpb.ResponseStatusOk

	return nil
}

func (s *Service) GetKeyByID(
	ctx context.Context,
	req *billingpb.KeyForOrderRequest,
	res *billingpb.GetKeyForOrderRequestResponse,
) error {
	key, err := s.keyRepository.GetById(ctx, req.KeyId)

	if err != nil {
		zap.S().Errorf(errors.KeyErrorNotFound.Message, "err", err.Error(), "keyId", req.KeyId)
		res.Status = billingpb.ResponseStatusNotFound
		res.Message = errors.KeyErrorNotFound
		return nil
	}

	res.Key = key

	return nil
}

func (s *Service) ReserveKeyForOrder(
	ctx context.Context,
	req *billingpb.PlatformKeyReserveRequest,
	res *billingpb.PlatformKeyReserveResponse,
) error {
	zap.S().Infow("[ReserveKeyForOrder] called", "order_id", req.OrderId, "platform_id", req.PlatformId, "KeyProductId", req.KeyProductId)
	key, err := s.keyRepository.ReserveKey(ctx, req.KeyProductId, req.PlatformId, req.OrderId, req.Ttl)

	if err != nil {
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errors.KeyErrorReserve
		return nil
	}

	zap.S().Infow("[ReserveKeyForOrder] reserved key", "req.order_id", req.OrderId, "key.order_id", key.OrderId, "key.id", key.Id, "key.RedeemedAt", key.RedeemedAt, "key.KeyProductId", key.KeyProductId)

	res.KeyId = key.Id
	res.Status = billingpb.ResponseStatusOk

	return nil
}

func (s *Service) FinishRedeemKeyForOrder(
	ctx context.Context,
	req *billingpb.KeyForOrderRequest,
	res *billingpb.GetKeyForOrderRequestResponse,
) error {
	key, err := s.keyRepository.FinishRedeemById(ctx, req.KeyId)

	if err != nil {
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errors.KeyErrorFinish
		return nil
	}

	res.Key = key
	res.Status = billingpb.ResponseStatusOk

	return nil
}

func (s *Service) CancelRedeemKeyForOrder(
	ctx context.Context,
	req *billingpb.KeyForOrderRequest,
	res *billingpb.EmptyResponseWithStatus,
) error {
	_, err := s.keyRepository.CancelById(ctx, req.KeyId)

	if err != nil {
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errors.KeyErrorCanceled
		return nil
	}

	res.Status = billingpb.ResponseStatusOk

	return nil
}

func (s *Service) KeyDaemonProcess(ctx context.Context) (int, error) {
	counter := 0
	keys, err := s.keyRepository.FindUnfinished(ctx)

	if err != nil {
		return counter, err
	}

	for _, key := range keys {
		_, err = s.keyRepository.CancelById(ctx, key.Id)

		if err != nil {
			continue
		}

		counter++
	}

	return counter, nil
}

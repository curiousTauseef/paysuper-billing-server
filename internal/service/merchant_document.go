package service

import (
	"context"
	"github.com/paysuper/paysuper-billing-server/pkg/errors"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/postmarkpb"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

var (
	errorMerchantDocumentNotFound     = errors.NewBillingServerErrorMsg("md000001", "unable to get merchant document list")
	errorMerchantDocumentAccessDenied = errors.NewBillingServerErrorMsg("md000002", "access denied")
	errorMerchantDocumentUnableInsert = errors.NewBillingServerErrorMsg("md000003", "unable to add merchant document")
)

func (s *Service) AddMerchantDocument(
	ctx context.Context,
	req *billingpb.MerchantDocument,
	res *billingpb.AddMerchantDocumentResponse,
) error {
	req.Id = primitive.NewObjectID().Hex()
	err := s.merchantDocumentRepository.Insert(ctx, req)

	if err != nil {
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorMerchantDocumentUnableInsert
		return nil
	}

	role, err := s.userRoleRepository.GetMerchantOwner(ctx, req.MerchantId)
	if err != nil {
		zap.L().Error("can't get merchant owner", zap.Error(err), zap.String("merchant_id", req.MerchantId))
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorMerchantDocumentUnableInsert
		return nil
	}

	var payload *postmarkpb.Payload

	if role.UserId == req.UserId {
		payload = &postmarkpb.Payload{
			TemplateAlias: s.cfg.EmailTemplates.MerchantDocumentUploaded,
			TemplateModel: map[string]string{
				"merchant_id": req.MerchantId,
				"document_id": req.Id,
				"user_id":     req.UserId,
				"name":        req.OriginalName,
				"file_path":   req.FilePath,
			},
			To: s.cfg.EmailAdminDocumentViewer,
		}
	} else {
		payload = &postmarkpb.Payload{
			TemplateAlias: s.cfg.EmailTemplates.AdminDocumentUploaded,
			TemplateModel: map[string]string{
			},
			To: role.Email,
		}
	}

	err = s.postmarkBroker.Publish(postmarkpb.PostmarkSenderTopicName, payload, amqp.Table{})
	if err != nil {
		zap.L().Error("can't send email", zap.Error(err), zap.Any("payload", payload))
	}

	res.Item = req
	res.Status = billingpb.ResponseStatusOk

	return nil
}

func (s *Service) GetMerchantDocument(
	ctx context.Context,
	req *billingpb.GetMerchantDocumentRequest,
	res *billingpb.GetMerchantDocumentResponse,
) error {
	item, err := s.merchantDocumentRepository.GetById(ctx, req.Id)

	if err != nil {
		res.Status = billingpb.ResponseStatusNotFound
		res.Message = errorMerchantDocumentNotFound
		return nil
	}

	if item.MerchantId != req.MerchantId {
		res.Status = billingpb.ResponseStatusForbidden
		res.Message = errorMerchantDocumentAccessDenied
		return nil
	}

	res.Item = item
	res.Status = billingpb.ResponseStatusOk

	return nil
}

func (s *Service) GetMerchantDocuments(
	ctx context.Context,
	req *billingpb.GetMerchantDocumentsRequest,
	res *billingpb.GetMerchantDocumentsResponse,
) error {
	var err error

	res.List, err = s.merchantDocumentRepository.GetByMerchantId(ctx, req.MerchantId, req.Offset, req.Limit)

	if err != nil {
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorMerchantDocumentNotFound
		return nil
	}

	res.Count, err = s.merchantDocumentRepository.CountByMerchantId(ctx, req.MerchantId)

	if err != nil {
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorMerchantDocumentNotFound
		return nil
	}

	res.Status = billingpb.ResponseStatusOk

	return nil
}

package service

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-billing-server/pkg/errors"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	errorOperatingCompanyCountryAlreadyExists = errors.NewBillingServerErrorMsg("oc000001", "operating company for one of passed country already exists")
	errorOperatingCompanyCountryUnknown       = errors.NewBillingServerErrorMsg("oc000002", "operating company country unknown")
	errorOperatingCompanyNotFound             = errors.NewBillingServerErrorMsg("oc000003", "operating company not found")
)

func (s *Service) GetOperatingCompaniesList(
	ctx context.Context,
	req *billingpb.EmptyRequest,
	res *billingpb.GetOperatingCompaniesListResponse,
) (err error) {
	res.Items, err = s.operatingCompanyRepository.GetAll(ctx)
	if err != nil {
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			res.Status = billingpb.ResponseStatusBadData
			res.Message = e
			return nil
		}
		return
	}

	res.Status = billingpb.ResponseStatusOk
	return
}

func (s *Service) AddOperatingCompany(
	ctx context.Context,
	req *billingpb.OperatingCompany,
	res *billingpb.EmptyResponseWithStatus,
) (err error) {
	oc := &billingpb.OperatingCompany{
		Id:               primitive.NewObjectID().Hex(),
		PaymentCountries: []string{},
		CreatedAt:        ptypes.TimestampNow(),
	}

	if req.Id != "" {
		oc, err = s.operatingCompanyRepository.GetById(ctx, req.Id)
		if err != nil {
			res.Status = billingpb.ResponseStatusBadData
			res.Message = errorOperatingCompanyNotFound
			return nil
		}
	}

	if req.PaymentCountries == nil || len(req.PaymentCountries) == 0 {
		ocCheck, err := s.operatingCompanyRepository.GetByPaymentCountry(ctx, "")
		if err != nil && err != mongo.ErrNoDocuments {
			if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
				res.Status = billingpb.ResponseStatusBadData
				res.Message = e
				return nil
			}
			return err
		}
		if ocCheck != nil && ocCheck.Id != oc.Id {
			res.Status = billingpb.ResponseStatusBadData
			res.Message = errorOperatingCompanyCountryAlreadyExists
			return nil
		}
		oc.PaymentCountries = []string{}

	} else {
		for _, countryCode := range req.PaymentCountries {
			ocCheck, err := s.operatingCompanyRepository.GetByPaymentCountry(ctx, countryCode)
			if err != nil && err != mongo.ErrNoDocuments {
				if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
					res.Status = billingpb.ResponseStatusBadData
					res.Message = e
					return nil
				}
				return err
			}
			if ocCheck != nil && ocCheck.Id != oc.Id {
				res.Status = billingpb.ResponseStatusBadData
				res.Message = errorOperatingCompanyCountryAlreadyExists
				return nil
			}

			_, err = s.country.GetByIsoCodeA2(ctx, countryCode)
			if err != nil {
				res.Status = billingpb.ResponseStatusBadData
				res.Message = errorOperatingCompanyCountryUnknown
				return nil
			}
		}
		oc.PaymentCountries = req.PaymentCountries
	}

	oc.UpdatedAt = ptypes.TimestampNow()
	oc.Name = req.Name
	oc.Country = req.Country
	oc.RegistrationNumber = req.RegistrationNumber
	oc.RegistrationDate = req.RegistrationDate
	oc.VatNumber = req.VatNumber
	oc.Address = req.Address
	oc.SignatoryName = req.SignatoryName
	oc.SignatoryPosition = req.SignatoryPosition
	oc.BankingDetails = req.BankingDetails
	oc.VatAddress = req.VatAddress
	oc.Email = req.Email

	err = s.operatingCompanyRepository.Upsert(ctx, oc)
	if err != nil {
		if e, ok := err.(*billingpb.ResponseErrorMessage); ok {
			res.Status = billingpb.ResponseStatusBadData
			res.Message = e
			return nil
		}
		return
	}

	res.Status = billingpb.ResponseStatusOk
	return
}

func (s *Service) GetOperatingCompany(
	ctx context.Context,
	req *billingpb.GetOperatingCompanyRequest,
	res *billingpb.GetOperatingCompanyResponse,
) (err error) {
	oc, err := s.operatingCompanyRepository.GetById(ctx, req.Id)

	if err != nil {
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorOperatingCompanyNotFound
		return nil
	}

	res.Status = billingpb.ResponseStatusOk
	res.Company = oc

	return
}

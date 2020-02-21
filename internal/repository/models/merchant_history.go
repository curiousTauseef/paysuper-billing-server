package models

import (
	"errors"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type merchantPaymentMethodHistoryMapper struct {

}

func (m *merchantPaymentMethodHistoryMapper) MapObjectToMgo(obj interface{}) (interface{}, error) {
	p := obj.(*billingpb.MerchantPaymentMethodHistory)
	st := &MgoMerchantPaymentMethodHistory{}

	if len(p.Id) <= 0 {
		st.Id = primitive.NewObjectID()
	} else {
		oid, err := primitive.ObjectIDFromHex(p.Id)

		if err != nil {
			return nil, errors.New(billingpb.ErrorInvalidObjectId)
		}

		st.Id = oid
	}

	if len(p.MerchantId) <= 0 {
		return nil, errors.New(billingpb.ErrorInvalidObjectId)
	} else {
		oid, err := primitive.ObjectIDFromHex(p.MerchantId)

		if err != nil {
			return nil, errors.New(billingpb.ErrorInvalidObjectId)
		}

		st.MerchantId = oid
	}

	if len(p.UserId) <= 0 {
		return nil, errors.New(billingpb.ErrorInvalidObjectId)
	} else {
		oid, err := primitive.ObjectIDFromHex(p.UserId)

		if err != nil {
			return nil, errors.New(billingpb.ErrorInvalidObjectId)
		}

		st.UserId = oid
	}

	if p.CreatedAt != nil {
		t, err := ptypes.Timestamp(p.CreatedAt)

		if err != nil {
			return nil, err
		}

		st.CreatedAt = t
	} else {
		st.CreatedAt = time.Now()
	}

	oid, err := primitive.ObjectIDFromHex(p.PaymentMethod.PaymentMethod.Id)

	if err != nil {
		return nil, err
	}

	st.PaymentMethod = &MgoMerchantPaymentMethod{
		PaymentMethod: &MgoMerchantPaymentMethodIdentification{
			Id:   oid,
			Name: p.PaymentMethod.PaymentMethod.Name,
		},
		Commission:  p.PaymentMethod.Commission,
		Integration: p.PaymentMethod.Integration,
		IsActive:    p.PaymentMethod.IsActive,
	}

	return st, nil
}

func (*merchantPaymentMethodHistoryMapper) MapMgoToObject(obj interface{}) (interface{}, error) {
	var err error
	decoded := obj.(*MgoMerchantPaymentMethodHistory)
	m := &billingpb.MerchantPaymentMethodHistory{}

	m.Id = decoded.Id.Hex()
	m.MerchantId = decoded.MerchantId.Hex()
	m.UserId = decoded.UserId.Hex()
	m.CreatedAt, err = ptypes.TimestampProto(decoded.CreatedAt)
	if err != nil {
		return nil, err
	}

	m.PaymentMethod = &billingpb.MerchantPaymentMethod{
		PaymentMethod: &billingpb.MerchantPaymentMethodIdentification{
			Id:   decoded.PaymentMethod.PaymentMethod.Id.Hex(),
			Name: decoded.PaymentMethod.PaymentMethod.Name,
		},
		Commission:  decoded.PaymentMethod.Commission,
		Integration: decoded.PaymentMethod.Integration,
		IsActive:    decoded.PaymentMethod.IsActive,
	}

	return m, nil
}

func NewMerchantPaymentMethodHistoryMapper() Mapper {
	return &merchantPaymentMethodHistoryMapper{}
}

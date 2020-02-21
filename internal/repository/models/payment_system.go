package models

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type paymentSystemMapper struct{}

func NewPaymentSystemMapper() Mapper {
	return &paymentSystemMapper{}
}

type MgoPaymentSystem struct {
	Id                 primitive.ObjectID `bson:"_id" faker:"objectId"`
	Name               string             `bson:"name"`
	Handler            string             `bson:"handler"`
	Country            string             `bson:"country"`
	AccountingCurrency string             `bson:"accounting_currency"`
	AccountingPeriod   string             `bson:"accounting_period"`
	IsActive           bool               `bson:"is_active"`
	CreatedAt          time.Time          `bson:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at"`
}

func (m *paymentSystemMapper) MapObjectToMgo(obj interface{}) (interface{}, error) {
	in := obj.(*billingpb.PaymentSystem)

	out := &MgoPaymentSystem{
		Name:               in.Name,
		Country:            in.Country,
		AccountingCurrency: in.AccountingCurrency,
		AccountingPeriod:   in.AccountingPeriod,
		IsActive:           in.IsActive,
		Handler:            in.Handler,
	}

	if len(in.Id) <= 0 {
		out.Id = primitive.NewObjectID()
	} else {
		oid, err := primitive.ObjectIDFromHex(in.Id)

		if err != nil {
			return nil, err
		}

		out.Id = oid
	}

	if in.CreatedAt != nil {
		t, err := ptypes.Timestamp(in.CreatedAt)

		if err != nil {
			return nil, err
		}

		out.CreatedAt = t
	} else {
		out.CreatedAt = time.Now()
	}

	if in.UpdatedAt != nil {
		t, err := ptypes.Timestamp(in.UpdatedAt)

		if err != nil {
			return nil, err
		}

		out.UpdatedAt = t
	} else {
		out.UpdatedAt = time.Now()
	}

	return out, nil
}

func (m *paymentSystemMapper) MapMgoToObject(obj interface{}) (interface{}, error) {
	var err error
	in := obj.(*MgoPaymentSystem)

	out := &billingpb.PaymentSystem{
		Id:                 in.Id.Hex(),
		Name:               in.Name,
		Country:            in.Country,
		AccountingCurrency: in.AccountingCurrency,
		AccountingPeriod:   in.AccountingPeriod,
		IsActive:           in.IsActive,
		Handler:            in.Handler,
	}

	out.CreatedAt, err = ptypes.TimestampProto(in.CreatedAt)

	if err != nil {
		return nil, err
	}

	out.UpdatedAt, err = ptypes.TimestampProto(in.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return out, nil
}

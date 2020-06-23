package models

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type paymentChannelCostSystemMapper struct{}

func NewPaymentChannelCostSystemMapper() Mapper {
	return &paymentChannelCostSystemMapper{}
}

type MgoPaymentChannelCostSystem struct {
	Id                 primitive.ObjectID `bson:"_id" faker:"objectId"`
	Name               string             `bson:"name"`
	Region             string             `bson:"region"`
	Country            string             `bson:"country"`
	Percent            float64            `bson:"percent"`
	FixAmount          float64            `bson:"fix_amount"`
	FixAmountCurrency  string             `bson:"fix_amount_currency"`
	CreatedAt          time.Time          `bson:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at"`
	IsActive           bool               `bson:"is_active"`
	MccCode            string             `bson:"mcc_code"`
	OperatingCompanyId string             `bson:"operating_company_id" faker:"objectId"`
}

func (m *paymentChannelCostSystemMapper) MapObjectToMgo(obj interface{}) (interface{}, error) {
	in := obj.(*billingpb.PaymentChannelCostSystem)

	out := &MgoPaymentChannelCostSystem{
		Name:               in.Name,
		Region:             in.Region,
		Country:            in.Country,
		Percent:            in.Percent,
		FixAmount:          in.FixAmount,
		FixAmountCurrency:  in.FixAmountCurrency,
		IsActive:           in.IsActive,
		MccCode:            in.MccCode,
		OperatingCompanyId: in.OperatingCompanyId,
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

func (m *paymentChannelCostSystemMapper) MapMgoToObject(obj interface{}) (interface{}, error) {
	var err error
	in := obj.(*MgoPaymentChannelCostSystem)

	out := &billingpb.PaymentChannelCostSystem{
		Id:                 in.Id.Hex(),
		Name:               in.Name,
		Region:             in.Region,
		Country:            in.Country,
		Percent:            in.Percent,
		FixAmount:          in.FixAmount,
		FixAmountCurrency:  in.FixAmountCurrency,
		IsActive:           in.IsActive,
		MccCode:            in.MccCode,
		OperatingCompanyId: in.OperatingCompanyId,
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

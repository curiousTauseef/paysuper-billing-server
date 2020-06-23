package models

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type priceGroupMapper struct{}

func NewPriceGroupMapper() Mapper {
	return &priceGroupMapper{}
}

type MgoPriceGroup struct {
	Id            primitive.ObjectID `bson:"_id" faker:"objectId"`
	Currency      string             `bson:"currency"`
	Region        string             `bson:"region"`
	InflationRate float64            `bson:"inflation_rate"`
	Fraction      float64            `bson:"fraction"`
	IsActive      bool               `bson:"is_active"`
	CreatedAt     time.Time          `bson:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at"`
}

func (m *priceGroupMapper) MapObjectToMgo(obj interface{}) (interface{}, error) {
	in := obj.(*billingpb.PriceGroup)

	out := &MgoPriceGroup{
		Region:        in.Region,
		Currency:      in.Currency,
		InflationRate: in.InflationRate,
		Fraction:      in.Fraction,
		IsActive:      in.IsActive,
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

func (m *priceGroupMapper) MapMgoToObject(obj interface{}) (interface{}, error) {
	var err error
	in := obj.(*MgoPriceGroup)

	out := &billingpb.PriceGroup{
		Id:            in.Id.Hex(),
		Region:        in.Region,
		Currency:      in.Currency,
		InflationRate: in.InflationRate,
		Fraction:      in.Fraction,
		IsActive:      in.IsActive,
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

package models

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"time"
)

type zipCodeMapper struct{}

func NewZipCodeMapper() Mapper {
	return &zipCodeMapper{}
}

type MgoZipCode struct {
	Zip       string                  `bson:"zip"`
	Country   string                  `bson:"country"`
	City      string                  `bson:"city"`
	State     *billingpb.ZipCodeState `bson:"state"`
	CreatedAt time.Time               `bson:"created_at"`
}

func (m *zipCodeMapper) MapObjectToMgo(obj interface{}) (interface{}, error) {
	in := obj.(*billingpb.ZipCode)

	out := &MgoZipCode{
		Zip:     in.Zip,
		Country: in.Country,
		City:    in.City,
		State:   in.State,
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

	return out, nil
}

func (m *zipCodeMapper) MapMgoToObject(obj interface{}) (interface{}, error) {
	var err error
	in := obj.(*MgoZipCode)

	out := &billingpb.ZipCode{
		Zip:     in.Zip,
		Country: in.Country,
		City:    in.City,
		State:   in.State,
	}

	out.CreatedAt, err = ptypes.TimestampProto(in.CreatedAt)

	if err != nil {
		return nil, err
	}

	return out, nil
}

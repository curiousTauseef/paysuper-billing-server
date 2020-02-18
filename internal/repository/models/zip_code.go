package models

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"time"
)

type MgoZipCode struct {
	Zip       string                  `bson:"zip"`
	Country   string                  `bson:"country"`
	City      string                  `bson:"city"`
	State     *billingpb.ZipCodeState `bson:"state"`
	CreatedAt time.Time               `bson:"created_at"`
}

func MapZipCodeToMgo(m *billingpb.ZipCode) (*MgoZipCode, error) {
	obj := &MgoZipCode{
		Zip:     m.Zip,
		Country: m.Country,
		City:    m.City,
		State:   m.State,
	}

	if m.CreatedAt != nil {
		t, err := ptypes.Timestamp(m.CreatedAt)

		if err != nil {
			return nil, err
		}

		obj.CreatedAt = t
	} else {
		obj.CreatedAt = time.Now()
	}

	return obj, nil
}

func MapMgoToZipCode(decoded *MgoZipCode) (*billingpb.ZipCode, error) {
	var err error

	obj := &billingpb.ZipCode{
		Zip:     decoded.Zip,
		Country: decoded.Country,
		City:    decoded.City,
		State:   decoded.State,
	}

	obj.CreatedAt, err = ptypes.TimestampProto(decoded.CreatedAt)

	if err != nil {
		return nil, err
	}

	return obj, nil
}

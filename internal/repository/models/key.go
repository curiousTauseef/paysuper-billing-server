package models

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type MgoKey struct {
	Id           primitive.ObjectID  `bson:"_id" faker:"objectId"`
	Code         string              `bson:"code"`
	KeyProductId primitive.ObjectID  `bson:"key_product_id" faker:"objectId"`
	PlatformId   string              `bson:"platform_id"`
	OrderId      *primitive.ObjectID `bson:"order_id" faker:"objectIdPointer"`
	CreatedAt    time.Time           `bson:"created_at"`
	ReservedTo   time.Time           `bson:"reserved_to"`
	RedeemedAt   time.Time           `bson:"redeemed_at"`
}

type keyMapper struct {
}

func (k *keyMapper) MapObjectToMgo(obj interface{}) (interface{}, error) {
	m := obj.(*billingpb.Key)
	oid, err := primitive.ObjectIDFromHex(m.Id)

	if err != nil {
		return nil, err
	}

	keyProductOid, err := primitive.ObjectIDFromHex(m.KeyProductId)

	if err != nil {
		return nil, err
	}

	st := &MgoKey{
		Id:           oid,
		PlatformId:   m.PlatformId,
		KeyProductId: keyProductOid,
		Code:         m.Code,
	}

	if m.OrderId != "" {
		oid, err := primitive.ObjectIDFromHex(m.OrderId)

		if err != nil {
			return nil, err
		}

		orderId := oid
		st.OrderId = &orderId
	}

	if m.RedeemedAt != nil {
		if st.RedeemedAt, err = ptypes.Timestamp(m.RedeemedAt); err != nil {
			return nil, err
		}
	} else {
		st.RedeemedAt = time.Time{}
	}

	if m.ReservedTo != nil {
		if st.ReservedTo, err = ptypes.Timestamp(m.ReservedTo); err != nil {
			return nil, err
		}
	} else {
		st.ReservedTo = time.Time{}
	}

	if m.CreatedAt != nil {
		if st.CreatedAt, err = ptypes.Timestamp(m.CreatedAt); err != nil {
			return nil, err
		}
	} else {
		st.CreatedAt = time.Now()
	}

	return st, nil
}

func (*keyMapper) MapMgoToObject(obj interface{}) (interface{}, error) {
	var err error
	decoded := obj.(*MgoKey)
	k := &billingpb.Key{}
	k.Id = decoded.Id.Hex()
	k.Code = decoded.Code
	k.KeyProductId = decoded.KeyProductId.Hex()
	k.PlatformId = decoded.PlatformId

	if decoded.OrderId != nil {
		k.OrderId = decoded.OrderId.Hex()
	}

	if k.CreatedAt, err = ptypes.TimestampProto(decoded.CreatedAt); err != nil {
		return nil, err
	}

	if k.RedeemedAt, err = ptypes.TimestampProto(decoded.RedeemedAt); err != nil {
		return nil, err
	}

	if k.ReservedTo, err = ptypes.TimestampProto(decoded.ReservedTo); err != nil {
		return nil, err
	}

	return k, nil
}

func NewKeyMapper() Mapper {
	return &keyMapper{}
}

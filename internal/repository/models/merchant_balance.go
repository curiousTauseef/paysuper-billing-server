package models

import (
	"errors"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type merchantBalanceMapper struct {
}

type MgoMerchantBalance struct {
	Id             primitive.ObjectID `bson:"_id" faker:"objectId"`
	MerchantId     primitive.ObjectID `bson:"merchant_id" faker:"objectId"`
	Currency       string             `bson:"currency"`
	Debit          float64            `bson:"debit"`
	Credit         float64            `bson:"credit"`
	RollingReserve float64            `bson:"rolling_reserve"`
	Total          float64            `bson:"total"`
	CreatedAt      time.Time          `bson:"created_at"`
}

func (h *merchantBalanceMapper) MapObjectToMgo(obj interface{}) (interface{}, error) {
	m := obj.(*billingpb.MerchantBalance)
	st := &MgoMerchantBalance{
		Currency:       m.Currency,
		Debit:          m.Debit,
		Credit:         m.Credit,
		RollingReserve: m.RollingReserve,
		Total:          m.Total,
	}
	if len(m.Id) <= 0 {
		st.Id = primitive.NewObjectID()
	} else {
		oid, err := primitive.ObjectIDFromHex(m.Id)

		if err != nil {
			return nil, errors.New(billingpb.ErrorInvalidObjectId)
		}

		st.Id = oid
	}

	merchantOid, err := primitive.ObjectIDFromHex(m.MerchantId)

	if err != nil {
		return nil, errors.New(billingpb.ErrorInvalidObjectId)
	}

	st.MerchantId = merchantOid

	if m.CreatedAt != nil {
		t, err := ptypes.Timestamp(m.CreatedAt)

		if err != nil {
			return nil, err
		}

		st.CreatedAt = t
	} else {
		st.CreatedAt = time.Now()
	}
	return st, nil
}

func (h *merchantBalanceMapper) MapMgoToObject(obj interface{}) (interface{}, error) {
	var err error
	decoded := obj.(*MgoMerchantBalance)
	m := &billingpb.MerchantBalance{}

	m.Id = decoded.Id.Hex()
	m.MerchantId = decoded.MerchantId.Hex()
	m.Currency = decoded.Currency
	m.Debit = decoded.Debit
	m.Credit = decoded.Credit
	m.RollingReserve = decoded.RollingReserve
	m.Total = decoded.Total

	m.CreatedAt, err = ptypes.TimestampProto(decoded.CreatedAt)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func NewMerchantBalanceMapper() Mapper {
	return &merchantBalanceMapper{}
}

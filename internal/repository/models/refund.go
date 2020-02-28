package models

import (
	"errors"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type MgoRefundOrder struct {
	Id   primitive.ObjectID `bson:"id"`
	Uuid string             `bson:"uuid"`
}

type MgoRefund struct {
	Id             primitive.ObjectID `bson:"_id"`
	OriginalOrder  *MgoRefundOrder    `bson:"original_order"`
	ExternalId     string             `bson:"external_id"`
	Amount         float64            `bson:"amount"`
	CreatorId      primitive.ObjectID `bson:"creator_id"`
	Currency       string             `bson:"currency"`
	Status         int32              `bson:"status"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
	PayerData      *billingpb.RefundPayerData   `bson:"payer_data"`
	SalesTax       float32            `bson:"sales_tax"`
	IsChargeback   bool               `bson:"is_chargeback"`
	CreatedOrderId primitive.ObjectID `bson:"created_order_id,omitempty"`
	Reason         string             `bson:"reason"`
}


type refundMapper struct {
}

func (r *refundMapper) MapObjectToMgo(obj interface{}) (interface{}, error) {
	m := obj.(*billingpb.Refund)
	oid, err := primitive.ObjectIDFromHex(m.OriginalOrder.Id)

	if err != nil {
		return nil, err
	}

	creatorOid, err := primitive.ObjectIDFromHex(m.CreatorId)

	if err != nil {
		return nil, err
	}

	st := &MgoRefund{
		OriginalOrder: &MgoRefundOrder{
			Id:   oid,
			Uuid: m.OriginalOrder.Uuid,
		},
		ExternalId:   m.ExternalId,
		Amount:       m.Amount,
		CreatorId:    creatorOid,
		Currency:     m.Currency,
		Status:       m.Status,
		PayerData:    m.PayerData,
		SalesTax:     m.SalesTax,
		IsChargeback: m.IsChargeback,
		Reason:       m.Reason,
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

	if m.CreatedOrderId != "" {
		oid, err := primitive.ObjectIDFromHex(m.CreatedOrderId)

		if err != nil {
			return nil, errors.New(billingpb.ErrorInvalidObjectId)
		}

		st.CreatedOrderId = oid
	}

	if m.CreatedAt != nil {
		t, err := ptypes.Timestamp(m.CreatedAt)

		if err != nil {
			return nil, err
		}

		st.CreatedAt = t
	} else {
		st.CreatedAt = time.Now()
	}

	if m.UpdatedAt != nil {
		t, err := ptypes.Timestamp(m.UpdatedAt)

		if err != nil {
			return nil, err
		}

		st.UpdatedAt = t
	} else {
		st.UpdatedAt = time.Now()
	}

	return st, nil
}

func (r *refundMapper) MapMgoToObject(obj interface{}) (interface{}, error) {
	var err error
	m := &billingpb.Refund{}
	decoded := obj.(*MgoRefund)

	m.Id = decoded.Id.Hex()
	m.OriginalOrder = &billingpb.RefundOrder{
		Id:   decoded.OriginalOrder.Id.Hex(),
		Uuid: decoded.OriginalOrder.Uuid,
	}
	m.ExternalId = decoded.ExternalId
	m.Amount = decoded.Amount
	m.CreatorId = decoded.CreatorId.Hex()
	m.Currency = decoded.Currency
	m.Status = decoded.Status
	m.PayerData = decoded.PayerData
	m.SalesTax = decoded.SalesTax
	m.IsChargeback = decoded.IsChargeback
	m.Reason = decoded.Reason

	if !decoded.CreatedOrderId.IsZero() {
		m.CreatedOrderId = decoded.CreatedOrderId.Hex()
	}

	m.CreatedAt, err = ptypes.TimestampProto(decoded.CreatedAt)

	if err != nil {
		return nil, err
	}

	m.UpdatedAt, err = ptypes.TimestampProto(decoded.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return m, nil
}

func NewRefundMapper() Mapper {
	return &refundMapper{}
}
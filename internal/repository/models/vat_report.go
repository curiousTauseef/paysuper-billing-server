package models

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type vatReportMapper struct{}

func NewVatReportMapper() Mapper {
	return &vatReportMapper{}
}

type MgoVatReport struct {
	Id                    primitive.ObjectID `bson:"_id" faker:"objectId"`
	Country               string             `bson:"country"`
	VatRate               float64            `bson:"vat_rate"`
	Currency              string             `bson:"currency"`
	TransactionsCount     int32              `bson:"transactions_count"`
	GrossRevenue          float64            `bson:"gross_revenue"`
	VatAmount             float64            `bson:"vat_amount"`
	FeesAmount            float64            `bson:"fees_amount"`
	DeductionAmount       float64            `bson:"deduction_amount"`
	CorrectionAmount      float64            `bson:"correction_amount"`
	Status                string             `bson:"status"`
	CountryAnnualTurnover float64            `bson:"country_annual_turnover"`
	WorldAnnualTurnover   float64            `bson:"world_annual_turnover"`
	AmountsApproximate    bool               `bson:"amounts_approximate"`
	OperatingCompanyId    string             `bson:"operating_company_id" faker:"objectId"`
	DateFrom              time.Time          `bson:"date_from"`
	DateTo                time.Time          `bson:"date_to"`
	CreatedAt             time.Time          `bson:"created_at"`
	UpdatedAt             time.Time          `bson:"updated_at"`
	PayUntilDate          time.Time          `bson:"pay_until_date"`
	PaidAt                time.Time          `bson:"paid_at"`
}

func (m *vatReportMapper) MapObjectToMgo(obj interface{}) (interface{}, error) {
	in := obj.(*billingpb.VatReport)

	out := &MgoVatReport{
		Country:               in.Country,
		VatRate:               in.VatRate,
		Currency:              in.Currency,
		TransactionsCount:     in.TransactionsCount,
		GrossRevenue:          in.GrossRevenue,
		VatAmount:             in.VatAmount,
		FeesAmount:            in.FeesAmount,
		DeductionAmount:       in.DeductionAmount,
		CorrectionAmount:      in.CorrectionAmount,
		Status:                in.Status,
		CountryAnnualTurnover: in.CountryAnnualTurnover,
		WorldAnnualTurnover:   in.WorldAnnualTurnover,
		AmountsApproximate:    in.AmountsApproximate,
		OperatingCompanyId:    in.OperatingCompanyId,
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

	t, err := ptypes.Timestamp(in.DateFrom)
	if err != nil {
		return nil, err
	}
	out.DateFrom = t

	t, err = ptypes.Timestamp(in.DateTo)
	if err != nil {
		return nil, err
	}
	out.DateTo = t

	t, err = ptypes.Timestamp(in.PayUntilDate)
	if err != nil {
		return nil, err
	}
	out.PayUntilDate = t

	if in.PaidAt != nil {
		t, err = ptypes.Timestamp(in.PaidAt)
		if err != nil {
			return nil, err
		}
		out.PaidAt = t
	}

	return out, nil
}

func (m *vatReportMapper) MapMgoToObject(obj interface{}) (interface{}, error) {
	var err error
	in := obj.(*MgoVatReport)

	out := &billingpb.VatReport{
		Id:                    in.Id.Hex(),
		Country:               in.Country,
		VatRate:               in.VatRate,
		Currency:              in.Currency,
		TransactionsCount:     in.TransactionsCount,
		GrossRevenue:          in.GrossRevenue,
		VatAmount:             in.VatAmount,
		FeesAmount:            in.FeesAmount,
		DeductionAmount:       in.DeductionAmount,
		CorrectionAmount:      in.CorrectionAmount,
		CountryAnnualTurnover: in.CountryAnnualTurnover,
		WorldAnnualTurnover:   in.WorldAnnualTurnover,
		AmountsApproximate:    in.AmountsApproximate,
		Status:                in.Status,
		OperatingCompanyId:    in.OperatingCompanyId,
	}

	out.CreatedAt, err = ptypes.TimestampProto(in.CreatedAt)
	if err != nil {
		return nil, err
	}

	out.UpdatedAt, err = ptypes.TimestampProto(in.UpdatedAt)
	if err != nil {
		return nil, err
	}

	out.DateFrom, err = ptypes.TimestampProto(in.DateFrom)
	if err != nil {
		return nil, err
	}

	out.DateTo, err = ptypes.TimestampProto(in.DateTo)
	if err != nil {
		return nil, err
	}

	out.PayUntilDate, err = ptypes.TimestampProto(in.PayUntilDate)
	if err != nil {
		return nil, err
	}

	out.PaidAt, err = ptypes.TimestampProto(in.PaidAt)
	if err != nil {
		return nil, err
	}

	return out, nil
}

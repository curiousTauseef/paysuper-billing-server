package models

import (
	"errors"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type MgoMerchant struct {
	Id                                            primitive.ObjectID                   `bson:"_id"`
	User                                          *MgoMerchantUser                     `bson:"user"`
	Company                                       *billingpb.MerchantCompanyInfo       `bson:"company"`
	Contacts                                      *billingpb.MerchantContact           `bson:"contacts"`
	Banking                                       *billingpb.MerchantBanking           `bson:"banking"`
	Status                                        int32                                `bson:"status"`
	CreatedAt                                     time.Time                            `bson:"created_at"`
	UpdatedAt                                     time.Time                            `bson:"updated_at"`
	FirstPaymentAt                                time.Time                            `bson:"first_payment_at"`
	IsVatEnabled                                  bool                                 `bson:"is_vat_enabled"`
	IsCommissionToUserEnabled                     bool                                 `bson:"is_commission_to_user_enabled"`
	HasMerchantSignature                          bool                                 `bson:"has_merchant_signature"`
	HasPspSignature                               bool                                 `bson:"has_psp_signature"`
	LastPayout                                    *MgoMerchantLastPayout               `bson:"last_payout"`
	IsSigned                                      bool                                 `bson:"is_signed"`
	PaymentMethods                                map[string]*MgoMerchantPaymentMethod `bson:"payment_methods"`
	AgreementType                                 int32                                `bson:"agreement_type"`
	AgreementSentViaMail                          bool                                 `bson:"agreement_sent_via_mail"`
	MailTrackingLink                              string                               `bson:"mail_tracking_link"`
	S3AgreementName                               string                               `bson:"s3_agreement_name"`
	PayoutCostAmount                              float64                              `bson:"payout_cost_amount"`
	PayoutCostCurrency                            string                               `bson:"payout_cost_currency"`
	MinPayoutAmount                               float64                              `bson:"min_payout_amount"`
	RollingReserveThreshold                       float64                              `bson:"rolling_reserve_amount"`
	RollingReserveDays                            int32                                `bson:"rolling_reserve_days"`
	RollingReserveChargebackTransactionsThreshold float64                              `bson:"rolling_reserve_chargeback_transactions_threshold"`
	ItemMinCostAmount                             float64                              `bson:"item_min_cost_amount"`
	ItemMinCostCurrency                           string                               `bson:"item_min_cost_currency"`
	Tariff                                        *billingpb.MerchantTariff            `bson:"tariff"`
	AgreementSignatureData                        *MgoMerchantAgreementSignatureData   `bson:"agreement_signature_data"`
	Steps                                         *billingpb.MerchantCompletedSteps    `bson:"steps"`
	AgreementTemplate                             string                               `bson:"agreement_template"`
	ReceivedDate                                  time.Time                            `bson:"received_date"`
	StatusLastUpdatedAt                           time.Time                            `bson:"status_last_updated_at"`
	AgreementNumber                               string                               `bson:"agreement_number"`
	MinimalPayoutLimit                            float32                              `bson:"minimal_payout_limit"`
	ManualPayoutsEnabled                          bool                                 `bson:"manual_payouts_enabled"`
	MccCode                                       string                               `bson:"mcc_code"`
	OperatingCompanyId                            string                               `bson:"operating_company_id"`
	MerchantOperationsType                        string                               `bson:"merchant_operations_type"`
	DontChargeVat                                 bool                                 `bson:"dont_charge_vat"`
}

type MgoMerchantPaymentMethodIdentification struct {
	Id   primitive.ObjectID `bson:"id"`
	Name string             `bson:"name"`
}

type MgoMerchantPaymentMethod struct {
	PaymentMethod *MgoMerchantPaymentMethodIdentification     `bson:"payment_method"`
	Commission    *billingpb.MerchantPaymentMethodCommissions `bson:"commission"`
	Integration   *billingpb.MerchantPaymentMethodIntegration `bson:"integration"`
	IsActive      bool                                        `bson:"is_active"`
}

type MgoMerchantAgreementSignatureDataSignUrl struct {
	SignUrl   string    `bson:"sign_url"`
	ExpiresAt time.Time `bson:"expires_at"`
}

type MgoMerchantAgreementSignatureData struct {
	DetailsUrl          string                                    `bson:"details_url"`
	FilesUrl            string                                    `bson:"files_url"`
	SignatureRequestId  string                                    `bson:"signature_request_id"`
	MerchantSignatureId string                                    `bson:"merchant_signature_id"`
	PsSignatureId       string                                    `bson:"ps_signature_id"`
	MerchantSignUrl     *MgoMerchantAgreementSignatureDataSignUrl `bson:"merchant_sign_url"`
	PsSignUrl           *MgoMerchantAgreementSignatureDataSignUrl `bson:"ps_sign_url"`
}

type MgoMerchantUser struct {
	Id               string    `bson:"id"`
	Email            string    `bson:"email"`
	FirstName        string    `bson:"first_name"`
	LastName         string    `bson:"last_name"`
	ProfileId        string    `bson:"profile_id"`
	RegistrationDate time.Time `bson:"registration_date"`
}

type MgoMerchantLastPayout struct {
	Date   time.Time `bson:"date"`
	Amount float64   `bson:"amount"`
}

type MgoMerchantCommon struct {
	Id      primitive.ObjectID             `bson:"_id"`
	Company *billingpb.MerchantCompanyInfo `bson:"company"`
	Banking *billingpb.MerchantBanking     `bson:"banking"`
	Status  int32                          `bson:"status"`
}

type MgoMerchantPaymentMethodHistory struct {
	Id            primitive.ObjectID        `bson:"_id"`
	MerchantId    primitive.ObjectID        `bson:"merchant_id"`
	PaymentMethod *MgoMerchantPaymentMethod `bson:"payment_method"`
	CreatedAt     time.Time                 `bson:"created_at" json:"created_at"`
	UserId        primitive.ObjectID        `bson:"user_id"`
}

type merchantMapper struct {
}

type merchantCommonMapper struct {
}

func (*merchantMapper) MapObjectToMgo(obj interface{}) (interface{}, error) {
	m := obj.(*billingpb.Merchant)

	st := &MgoMerchant{
		Company:                   m.Company,
		Contacts:                  m.Contacts,
		Banking:                   m.Banking,
		Status:                    m.Status,
		IsVatEnabled:              m.IsVatEnabled,
		IsCommissionToUserEnabled: m.IsCommissionToUserEnabled,
		HasMerchantSignature:      m.HasMerchantSignature,
		HasPspSignature:           m.HasPspSignature,
		IsSigned:                  m.IsSigned,
		AgreementType:             m.AgreementType,
		AgreementSentViaMail:      m.AgreementSentViaMail,
		MailTrackingLink:          m.MailTrackingLink,
		S3AgreementName:           m.S3AgreementName,
		PayoutCostAmount:          m.PayoutCostAmount,
		PayoutCostCurrency:        m.PayoutCostCurrency,
		MinPayoutAmount:           m.MinPayoutAmount,
		RollingReserveThreshold:   m.RollingReserveThreshold,
		RollingReserveDays:        m.RollingReserveDays,
		RollingReserveChargebackTransactionsThreshold: m.RollingReserveChargebackTransactionsThreshold,
		ItemMinCostAmount:      m.ItemMinCostAmount,
		ItemMinCostCurrency:    m.ItemMinCostCurrency,
		Tariff:                 m.Tariff,
		Steps:                  m.Steps,
		AgreementTemplate:      m.AgreementTemplate,
		AgreementNumber:        m.AgreementNumber,
		MinimalPayoutLimit:     m.MinimalPayoutLimit,
		ManualPayoutsEnabled:   m.ManualPayoutsEnabled,
		MccCode:                m.MccCode,
		OperatingCompanyId:     m.OperatingCompanyId,
		MerchantOperationsType: m.MerchantOperationsType,
		DontChargeVat:          m.DontChargeVat,
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

	if m.User != nil {
		st.User = &MgoMerchantUser{
			Id:        m.User.Id,
			Email:     m.User.Email,
			FirstName: m.User.FirstName,
			LastName:  m.User.LastName,
			ProfileId: m.User.ProfileId,
		}

		if m.User.RegistrationDate != nil {
			t, err := ptypes.Timestamp(m.User.RegistrationDate)

			if err != nil {
				return nil, err
			}

			st.User.RegistrationDate = t
		}
	}

	if m.ReceivedDate != nil {
		t, err := ptypes.Timestamp(m.ReceivedDate)

		if err != nil {
			return nil, err
		}

		st.ReceivedDate = t
	}

	if m.StatusLastUpdatedAt != nil {
		t, err := ptypes.Timestamp(m.StatusLastUpdatedAt)

		if err != nil {
			return nil, err
		}

		st.StatusLastUpdatedAt = t
	}

	if m.FirstPaymentAt != nil {
		t, err := ptypes.Timestamp(m.FirstPaymentAt)

		if err != nil {
			return nil, err
		}

		st.FirstPaymentAt = t
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

	if m.LastPayout != nil {
		st.LastPayout = &MgoMerchantLastPayout{
			Amount: m.LastPayout.Amount,
		}

		t, err := ptypes.Timestamp(m.LastPayout.Date)

		if err != nil {
			return nil, err
		}

		st.LastPayout.Date = t
	}

	if len(m.PaymentMethods) > 0 {
		st.PaymentMethods = make(map[string]*MgoMerchantPaymentMethod, len(m.PaymentMethods))

		for k, v := range m.PaymentMethods {
			oid, err := primitive.ObjectIDFromHex(v.PaymentMethod.Id)

			if err != nil {
				return nil, err
			}

			st.PaymentMethods[k] = &MgoMerchantPaymentMethod{
				PaymentMethod: &MgoMerchantPaymentMethodIdentification{
					Id:   oid,
					Name: v.PaymentMethod.Name,
				},
				Commission:  v.Commission,
				Integration: v.Integration,
				IsActive:    v.IsActive,
			}
		}
	}

	if m.AgreementSignatureData != nil {
		st.AgreementSignatureData = &MgoMerchantAgreementSignatureData{
			DetailsUrl:          m.AgreementSignatureData.DetailsUrl,
			FilesUrl:            m.AgreementSignatureData.FilesUrl,
			SignatureRequestId:  m.AgreementSignatureData.SignatureRequestId,
			MerchantSignatureId: m.AgreementSignatureData.MerchantSignatureId,
			PsSignatureId:       m.AgreementSignatureData.PsSignatureId,
		}

		if m.AgreementSignatureData.MerchantSignUrl != nil {
			st.AgreementSignatureData.MerchantSignUrl = &MgoMerchantAgreementSignatureDataSignUrl{
				SignUrl: m.AgreementSignatureData.MerchantSignUrl.SignUrl,
			}

			t, err := ptypes.Timestamp(m.AgreementSignatureData.MerchantSignUrl.ExpiresAt)

			if err != nil {
				return nil, err
			}

			st.AgreementSignatureData.MerchantSignUrl.ExpiresAt = t
		}

		if m.AgreementSignatureData.PsSignUrl != nil {
			st.AgreementSignatureData.PsSignUrl = &MgoMerchantAgreementSignatureDataSignUrl{
				SignUrl: m.AgreementSignatureData.PsSignUrl.SignUrl,
			}

			t, err := ptypes.Timestamp(m.AgreementSignatureData.PsSignUrl.ExpiresAt)

			if err != nil {
				return nil, err
			}

			st.AgreementSignatureData.PsSignUrl.ExpiresAt = t
		}
	}

	return st, nil
}

func (*merchantMapper) MapMgoToObject(obj interface{}) (interface{}, error) {
	var err error
	decoded := obj.(*MgoMerchant)
	m := &billingpb.Merchant{}

	m.Id = decoded.Id.Hex()
	m.Company = decoded.Company
	m.Contacts = decoded.Contacts
	m.Banking = decoded.Banking
	m.Status = decoded.Status
	m.IsVatEnabled = decoded.IsVatEnabled
	m.IsCommissionToUserEnabled = decoded.IsCommissionToUserEnabled
	m.HasMerchantSignature = decoded.HasMerchantSignature
	m.HasPspSignature = decoded.HasPspSignature
	m.IsSigned = decoded.IsSigned
	m.AgreementType = decoded.AgreementType
	m.AgreementSentViaMail = decoded.AgreementSentViaMail
	m.MailTrackingLink = decoded.MailTrackingLink
	m.S3AgreementName = decoded.S3AgreementName
	m.PayoutCostAmount = decoded.PayoutCostAmount
	m.PayoutCostCurrency = decoded.PayoutCostCurrency
	m.MinPayoutAmount = decoded.MinPayoutAmount
	m.RollingReserveThreshold = decoded.RollingReserveThreshold
	m.RollingReserveDays = decoded.RollingReserveDays
	m.RollingReserveChargebackTransactionsThreshold = decoded.RollingReserveChargebackTransactionsThreshold
	m.ItemMinCostAmount = decoded.ItemMinCostAmount
	m.ItemMinCostCurrency = decoded.ItemMinCostCurrency
	m.Tariff = decoded.Tariff
	m.Steps = decoded.Steps
	m.AgreementTemplate = decoded.AgreementTemplate
	m.AgreementNumber = decoded.AgreementNumber
	m.MinimalPayoutLimit = decoded.MinimalPayoutLimit
	m.ManualPayoutsEnabled = decoded.ManualPayoutsEnabled
	m.MccCode = decoded.MccCode
	m.OperatingCompanyId = decoded.OperatingCompanyId
	m.MerchantOperationsType = decoded.MerchantOperationsType
	m.DontChargeVat = decoded.DontChargeVat

	if decoded.User != nil {
		m.User = &billingpb.MerchantUser{
			Id:        decoded.User.Id,
			Email:     decoded.User.Email,
			FirstName: decoded.User.FirstName,
			LastName:  decoded.User.LastName,
			ProfileId: decoded.User.ProfileId,
		}

		if !decoded.User.RegistrationDate.IsZero() {
			m.User.RegistrationDate, err = ptypes.TimestampProto(decoded.User.RegistrationDate)

			if err != nil {
				return nil, err
			}
		}
	}

	if !decoded.ReceivedDate.IsZero() {
		m.ReceivedDate, err = ptypes.TimestampProto(decoded.ReceivedDate)

		if err != nil {
			return nil, err
		}
	}

	if !decoded.StatusLastUpdatedAt.IsZero() {
		m.StatusLastUpdatedAt, err = ptypes.TimestampProto(decoded.StatusLastUpdatedAt)

		if err != nil {
			return nil, err
		}
	}

	m.FirstPaymentAt, err = ptypes.TimestampProto(decoded.FirstPaymentAt)

	if err != nil {
		return nil, err
	}

	m.CreatedAt, err = ptypes.TimestampProto(decoded.CreatedAt)

	if err != nil {
		return nil, err
	}

	m.UpdatedAt, err = ptypes.TimestampProto(decoded.UpdatedAt)

	if err != nil {
		return nil, err
	}

	if decoded.LastPayout != nil {
		m.LastPayout = &billingpb.MerchantLastPayout{
			Amount: decoded.LastPayout.Amount,
		}

		m.LastPayout.Date, err = ptypes.TimestampProto(decoded.LastPayout.Date)

		if err != nil {
			return nil, err
		}
	}

	if len(decoded.PaymentMethods) > 0 {
		m.PaymentMethods = make(map[string]*billingpb.MerchantPaymentMethod, len(decoded.PaymentMethods))

		for k, v := range decoded.PaymentMethods {
			m.PaymentMethods[k] = &billingpb.MerchantPaymentMethod{
				PaymentMethod: &billingpb.MerchantPaymentMethodIdentification{},
				Commission:    v.Commission,
				Integration:   v.Integration,
				IsActive:      v.IsActive,
			}

			if v.PaymentMethod != nil {
				m.PaymentMethods[k].PaymentMethod.Id = v.PaymentMethod.Id.Hex()
				m.PaymentMethods[k].PaymentMethod.Name = v.PaymentMethod.Name
			}
		}
	}

	if decoded.AgreementSignatureData != nil {
		m.AgreementSignatureData = &billingpb.MerchantAgreementSignatureData{
			DetailsUrl:          decoded.AgreementSignatureData.DetailsUrl,
			FilesUrl:            decoded.AgreementSignatureData.FilesUrl,
			SignatureRequestId:  decoded.AgreementSignatureData.SignatureRequestId,
			MerchantSignatureId: decoded.AgreementSignatureData.MerchantSignatureId,
			PsSignatureId:       decoded.AgreementSignatureData.PsSignatureId,
		}

		if decoded.AgreementSignatureData.MerchantSignUrl != nil {
			m.AgreementSignatureData.MerchantSignUrl = &billingpb.MerchantAgreementSignatureDataSignUrl{
				SignUrl: decoded.AgreementSignatureData.MerchantSignUrl.SignUrl,
			}

			t, err := ptypes.TimestampProto(decoded.AgreementSignatureData.MerchantSignUrl.ExpiresAt)

			if err != nil {
				return nil, err
			}

			m.AgreementSignatureData.MerchantSignUrl.ExpiresAt = t
		}

		if decoded.AgreementSignatureData.PsSignUrl != nil {
			m.AgreementSignatureData.PsSignUrl = &billingpb.MerchantAgreementSignatureDataSignUrl{
				SignUrl: decoded.AgreementSignatureData.PsSignUrl.SignUrl,
			}

			t, err := ptypes.TimestampProto(decoded.AgreementSignatureData.PsSignUrl.ExpiresAt)

			if err != nil {
				return nil, err
			}

			m.AgreementSignatureData.PsSignUrl.ExpiresAt = t
		}
	}

	return m, nil
}

func NewMerchantMapper() Mapper {
	return &merchantMapper{}
}

func (*merchantCommonMapper) MapObjectToMgo(obj interface{}) (interface{}, error) {
	panic("implement me")
}

func (*merchantCommonMapper) MapMgoToObject(obj interface{}) (interface{}, error) {
	decoded := obj.(*MgoMerchantCommon)
	m := &billingpb.MerchantCommon{}

	m.Id = decoded.Id.Hex()
	m.Status = decoded.Status

	if decoded.Company != nil {
		m.Name = decoded.Company.Name
	}

	if decoded.Banking != nil {
		m.Currency = decoded.Banking.Currency
	}

	return m, nil
}

func NewCommonMerchantMapper() Mapper {
	return &merchantCommonMapper{}
}

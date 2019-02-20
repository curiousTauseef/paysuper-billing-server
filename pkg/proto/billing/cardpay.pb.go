// Code generated by protoc-gen-go. DO NOT EDIT.
// source: billing/cardpay.proto

package billing // import "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type CardPayAddress struct {
	Country              string   `protobuf:"bytes,1,opt,name=country,proto3" json:"country,omitempty"`
	City                 string   `protobuf:"bytes,2,opt,name=city,proto3" json:"city,omitempty"`
	Phone                string   `protobuf:"bytes,3,opt,name=phone,proto3" json:"phone,omitempty"`
	State                string   `protobuf:"bytes,4,opt,name=state,proto3" json:"state,omitempty"`
	Street               string   `protobuf:"bytes,5,opt,name=street,proto3" json:"street,omitempty"`
	Zip                  string   `protobuf:"bytes,6,opt,name=zip,proto3" json:"zip,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-" structure:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-" structure:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-" structure:"-"`
}

func (m *CardPayAddress) Reset()         { *m = CardPayAddress{} }
func (m *CardPayAddress) String() string { return proto.CompactTextString(m) }
func (*CardPayAddress) ProtoMessage()    {}
func (*CardPayAddress) Descriptor() ([]byte, []int) {
	return fileDescriptor_cardpay_b3de4ad0513c2c93, []int{0}
}
func (m *CardPayAddress) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CardPayAddress.Unmarshal(m, b)
}
func (m *CardPayAddress) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CardPayAddress.Marshal(b, m, deterministic)
}
func (dst *CardPayAddress) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CardPayAddress.Merge(dst, src)
}
func (m *CardPayAddress) XXX_Size() int {
	return xxx_messageInfo_CardPayAddress.Size(m)
}
func (m *CardPayAddress) XXX_DiscardUnknown() {
	xxx_messageInfo_CardPayAddress.DiscardUnknown(m)
}

var xxx_messageInfo_CardPayAddress proto.InternalMessageInfo

func (m *CardPayAddress) GetCountry() string {
	if m != nil {
		return m.Country
	}
	return ""
}

func (m *CardPayAddress) GetCity() string {
	if m != nil {
		return m.City
	}
	return ""
}

func (m *CardPayAddress) GetPhone() string {
	if m != nil {
		return m.Phone
	}
	return ""
}

func (m *CardPayAddress) GetState() string {
	if m != nil {
		return m.State
	}
	return ""
}

func (m *CardPayAddress) GetStreet() string {
	if m != nil {
		return m.Street
	}
	return ""
}

func (m *CardPayAddress) GetZip() string {
	if m != nil {
		return m.Zip
	}
	return ""
}

type CardPayItem struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Description          string   `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	Count                int32    `protobuf:"varint,3,opt,name=count,proto3" json:"count,omitempty"`
	Price                float64  `protobuf:"fixed64,4,opt,name=price,proto3" json:"price,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-" structure:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-" structure:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-" structure:"-"`
}

func (m *CardPayItem) Reset()         { *m = CardPayItem{} }
func (m *CardPayItem) String() string { return proto.CompactTextString(m) }
func (*CardPayItem) ProtoMessage()    {}
func (*CardPayItem) Descriptor() ([]byte, []int) {
	return fileDescriptor_cardpay_b3de4ad0513c2c93, []int{1}
}
func (m *CardPayItem) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CardPayItem.Unmarshal(m, b)
}
func (m *CardPayItem) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CardPayItem.Marshal(b, m, deterministic)
}
func (dst *CardPayItem) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CardPayItem.Merge(dst, src)
}
func (m *CardPayItem) XXX_Size() int {
	return xxx_messageInfo_CardPayItem.Size(m)
}
func (m *CardPayItem) XXX_DiscardUnknown() {
	xxx_messageInfo_CardPayItem.DiscardUnknown(m)
}

var xxx_messageInfo_CardPayItem proto.InternalMessageInfo

func (m *CardPayItem) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *CardPayItem) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *CardPayItem) GetCount() int32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *CardPayItem) GetPrice() float64 {
	if m != nil {
		return m.Price
	}
	return 0
}

type CardPayMerchantOrder struct {
	Id                   string          `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Description          string          `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	Items                []*CardPayItem  `protobuf:"bytes,3,rep,name=items,proto3" json:"items,omitempty"`
	ShippingAddress      *CardPayAddress `protobuf:"bytes,4,opt,name=shipping_address,json=shippingAddress,proto3" json:"shipping_address,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-" bson:"-" structure:"-"`
	XXX_unrecognized     []byte          `json:"-" bson:"-" structure:"-"`
	XXX_sizecache        int32           `json:"-" bson:"-" structure:"-"`
}

func (m *CardPayMerchantOrder) Reset()         { *m = CardPayMerchantOrder{} }
func (m *CardPayMerchantOrder) String() string { return proto.CompactTextString(m) }
func (*CardPayMerchantOrder) ProtoMessage()    {}
func (*CardPayMerchantOrder) Descriptor() ([]byte, []int) {
	return fileDescriptor_cardpay_b3de4ad0513c2c93, []int{2}
}
func (m *CardPayMerchantOrder) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CardPayMerchantOrder.Unmarshal(m, b)
}
func (m *CardPayMerchantOrder) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CardPayMerchantOrder.Marshal(b, m, deterministic)
}
func (dst *CardPayMerchantOrder) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CardPayMerchantOrder.Merge(dst, src)
}
func (m *CardPayMerchantOrder) XXX_Size() int {
	return xxx_messageInfo_CardPayMerchantOrder.Size(m)
}
func (m *CardPayMerchantOrder) XXX_DiscardUnknown() {
	xxx_messageInfo_CardPayMerchantOrder.DiscardUnknown(m)
}

var xxx_messageInfo_CardPayMerchantOrder proto.InternalMessageInfo

func (m *CardPayMerchantOrder) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *CardPayMerchantOrder) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *CardPayMerchantOrder) GetItems() []*CardPayItem {
	if m != nil {
		return m.Items
	}
	return nil
}

func (m *CardPayMerchantOrder) GetShippingAddress() *CardPayAddress {
	if m != nil {
		return m.ShippingAddress
	}
	return nil
}

type CallbackCardPayBankCardAccount struct {
	Holder               string   `protobuf:"bytes,1,opt,name=holder,proto3" json:"holder,omitempty"`
	IssuingCountryCode   string   `protobuf:"bytes,2,opt,name=issuing_country_code,json=issuingCountryCode,proto3" json:"issuing_country_code,omitempty"`
	MaskedPan            string   `protobuf:"bytes,3,opt,name=masked_pan,json=maskedPan,proto3" json:"masked_pan,omitempty"`
	Token                string   `protobuf:"bytes,4,opt,name=token,proto3" json:"token,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-" structure:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-" structure:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-" structure:"-"`
}

func (m *CallbackCardPayBankCardAccount) Reset()         { *m = CallbackCardPayBankCardAccount{} }
func (m *CallbackCardPayBankCardAccount) String() string { return proto.CompactTextString(m) }
func (*CallbackCardPayBankCardAccount) ProtoMessage()    {}
func (*CallbackCardPayBankCardAccount) Descriptor() ([]byte, []int) {
	return fileDescriptor_cardpay_b3de4ad0513c2c93, []int{3}
}
func (m *CallbackCardPayBankCardAccount) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CallbackCardPayBankCardAccount.Unmarshal(m, b)
}
func (m *CallbackCardPayBankCardAccount) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CallbackCardPayBankCardAccount.Marshal(b, m, deterministic)
}
func (dst *CallbackCardPayBankCardAccount) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CallbackCardPayBankCardAccount.Merge(dst, src)
}
func (m *CallbackCardPayBankCardAccount) XXX_Size() int {
	return xxx_messageInfo_CallbackCardPayBankCardAccount.Size(m)
}
func (m *CallbackCardPayBankCardAccount) XXX_DiscardUnknown() {
	xxx_messageInfo_CallbackCardPayBankCardAccount.DiscardUnknown(m)
}

var xxx_messageInfo_CallbackCardPayBankCardAccount proto.InternalMessageInfo

func (m *CallbackCardPayBankCardAccount) GetHolder() string {
	if m != nil {
		return m.Holder
	}
	return ""
}

func (m *CallbackCardPayBankCardAccount) GetIssuingCountryCode() string {
	if m != nil {
		return m.IssuingCountryCode
	}
	return ""
}

func (m *CallbackCardPayBankCardAccount) GetMaskedPan() string {
	if m != nil {
		return m.MaskedPan
	}
	return ""
}

func (m *CallbackCardPayBankCardAccount) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

type CallbackCardPayCryptoCurrencyAccount struct {
	CryptoAddress        string   `protobuf:"bytes,1,opt,name=crypto_address,json=cryptoAddress,proto3" json:"crypto_address,omitempty"`
	CryptoTransactionId  string   `protobuf:"bytes,2,opt,name=crypto_transaction_id,json=cryptoTransactionId,proto3" json:"crypto_transaction_id,omitempty"`
	PrcAmount            string   `protobuf:"bytes,3,opt,name=prc_amount,json=prcAmount,proto3" json:"prc_amount,omitempty"`
	PrcCurrency          string   `protobuf:"bytes,4,opt,name=prc_currency,json=prcCurrency,proto3" json:"prc_currency,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-" structure:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-" structure:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-" structure:"-"`
}

func (m *CallbackCardPayCryptoCurrencyAccount) Reset()         { *m = CallbackCardPayCryptoCurrencyAccount{} }
func (m *CallbackCardPayCryptoCurrencyAccount) String() string { return proto.CompactTextString(m) }
func (*CallbackCardPayCryptoCurrencyAccount) ProtoMessage()    {}
func (*CallbackCardPayCryptoCurrencyAccount) Descriptor() ([]byte, []int) {
	return fileDescriptor_cardpay_b3de4ad0513c2c93, []int{4}
}
func (m *CallbackCardPayCryptoCurrencyAccount) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CallbackCardPayCryptoCurrencyAccount.Unmarshal(m, b)
}
func (m *CallbackCardPayCryptoCurrencyAccount) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CallbackCardPayCryptoCurrencyAccount.Marshal(b, m, deterministic)
}
func (dst *CallbackCardPayCryptoCurrencyAccount) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CallbackCardPayCryptoCurrencyAccount.Merge(dst, src)
}
func (m *CallbackCardPayCryptoCurrencyAccount) XXX_Size() int {
	return xxx_messageInfo_CallbackCardPayCryptoCurrencyAccount.Size(m)
}
func (m *CallbackCardPayCryptoCurrencyAccount) XXX_DiscardUnknown() {
	xxx_messageInfo_CallbackCardPayCryptoCurrencyAccount.DiscardUnknown(m)
}

var xxx_messageInfo_CallbackCardPayCryptoCurrencyAccount proto.InternalMessageInfo

func (m *CallbackCardPayCryptoCurrencyAccount) GetCryptoAddress() string {
	if m != nil {
		return m.CryptoAddress
	}
	return ""
}

func (m *CallbackCardPayCryptoCurrencyAccount) GetCryptoTransactionId() string {
	if m != nil {
		return m.CryptoTransactionId
	}
	return ""
}

func (m *CallbackCardPayCryptoCurrencyAccount) GetPrcAmount() string {
	if m != nil {
		return m.PrcAmount
	}
	return ""
}

func (m *CallbackCardPayCryptoCurrencyAccount) GetPrcCurrency() string {
	if m != nil {
		return m.PrcCurrency
	}
	return ""
}

type CardPayCustomer struct {
	Email                string   `protobuf:"bytes,1,opt,name=email,proto3" json:"email,omitempty"`
	Ip                   string   `protobuf:"bytes,2,opt,name=ip,proto3" json:"ip,omitempty"`
	Id                   string   `protobuf:"bytes,3,opt,name=id,proto3" json:"id,omitempty"`
	Locale               string   `protobuf:"bytes,4,opt,name=locale,proto3" json:"locale,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-" structure:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-" structure:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-" structure:"-"`
}

func (m *CardPayCustomer) Reset()         { *m = CardPayCustomer{} }
func (m *CardPayCustomer) String() string { return proto.CompactTextString(m) }
func (*CardPayCustomer) ProtoMessage()    {}
func (*CardPayCustomer) Descriptor() ([]byte, []int) {
	return fileDescriptor_cardpay_b3de4ad0513c2c93, []int{5}
}
func (m *CardPayCustomer) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CardPayCustomer.Unmarshal(m, b)
}
func (m *CardPayCustomer) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CardPayCustomer.Marshal(b, m, deterministic)
}
func (dst *CardPayCustomer) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CardPayCustomer.Merge(dst, src)
}
func (m *CardPayCustomer) XXX_Size() int {
	return xxx_messageInfo_CardPayCustomer.Size(m)
}
func (m *CardPayCustomer) XXX_DiscardUnknown() {
	xxx_messageInfo_CardPayCustomer.DiscardUnknown(m)
}

var xxx_messageInfo_CardPayCustomer proto.InternalMessageInfo

func (m *CardPayCustomer) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *CardPayCustomer) GetIp() string {
	if m != nil {
		return m.Ip
	}
	return ""
}

func (m *CardPayCustomer) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *CardPayCustomer) GetLocale() string {
	if m != nil {
		return m.Locale
	}
	return ""
}

type CardPayEWalletAccount struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-" structure:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-" structure:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-" structure:"-"`
}

func (m *CardPayEWalletAccount) Reset()         { *m = CardPayEWalletAccount{} }
func (m *CardPayEWalletAccount) String() string { return proto.CompactTextString(m) }
func (*CardPayEWalletAccount) ProtoMessage()    {}
func (*CardPayEWalletAccount) Descriptor() ([]byte, []int) {
	return fileDescriptor_cardpay_b3de4ad0513c2c93, []int{6}
}
func (m *CardPayEWalletAccount) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CardPayEWalletAccount.Unmarshal(m, b)
}
func (m *CardPayEWalletAccount) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CardPayEWalletAccount.Marshal(b, m, deterministic)
}
func (dst *CardPayEWalletAccount) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CardPayEWalletAccount.Merge(dst, src)
}
func (m *CardPayEWalletAccount) XXX_Size() int {
	return xxx_messageInfo_CardPayEWalletAccount.Size(m)
}
func (m *CardPayEWalletAccount) XXX_DiscardUnknown() {
	xxx_messageInfo_CardPayEWalletAccount.DiscardUnknown(m)
}

var xxx_messageInfo_CardPayEWalletAccount proto.InternalMessageInfo

func (m *CardPayEWalletAccount) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type CallbackCardPayPaymentData struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Amount               float64  `protobuf:"fixed64,2,opt,name=amount,proto3" json:"amount,omitempty"`
	AuthCode             string   `protobuf:"bytes,3,opt,name=auth_code,json=authCode,proto3" json:"auth_code,omitempty"`
	Created              string   `protobuf:"bytes,4,opt,name=created,proto3" json:"created,omitempty"`
	Currency             string   `protobuf:"bytes,5,opt,name=currency,proto3" json:"currency,omitempty"`
	DeclineCode          string   `protobuf:"bytes,6,opt,name=decline_code,json=declineCode,proto3" json:"decline_code,omitempty"`
	DeclineReason        string   `protobuf:"bytes,7,opt,name=decline_reason,json=declineReason,proto3" json:"decline_reason,omitempty"`
	Description          string   `protobuf:"bytes,8,opt,name=description,proto3" json:"description,omitempty"`
	Is_3D                bool     `protobuf:"varint,9,opt,name=is_3d,json=is3d,proto3" json:"is_3d,omitempty"`
	Note                 string   `protobuf:"bytes,10,opt,name=note,proto3" json:"note,omitempty"`
	Rrn                  string   `protobuf:"bytes,11,opt,name=rrn,proto3" json:"rrn,omitempty"`
	Status               string   `protobuf:"bytes,12,opt,name=status,proto3" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-" structure:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-" structure:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-" structure:"-"`
}

func (m *CallbackCardPayPaymentData) Reset()         { *m = CallbackCardPayPaymentData{} }
func (m *CallbackCardPayPaymentData) String() string { return proto.CompactTextString(m) }
func (*CallbackCardPayPaymentData) ProtoMessage()    {}
func (*CallbackCardPayPaymentData) Descriptor() ([]byte, []int) {
	return fileDescriptor_cardpay_b3de4ad0513c2c93, []int{7}
}
func (m *CallbackCardPayPaymentData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CallbackCardPayPaymentData.Unmarshal(m, b)
}
func (m *CallbackCardPayPaymentData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CallbackCardPayPaymentData.Marshal(b, m, deterministic)
}
func (dst *CallbackCardPayPaymentData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CallbackCardPayPaymentData.Merge(dst, src)
}
func (m *CallbackCardPayPaymentData) XXX_Size() int {
	return xxx_messageInfo_CallbackCardPayPaymentData.Size(m)
}
func (m *CallbackCardPayPaymentData) XXX_DiscardUnknown() {
	xxx_messageInfo_CallbackCardPayPaymentData.DiscardUnknown(m)
}

var xxx_messageInfo_CallbackCardPayPaymentData proto.InternalMessageInfo

func (m *CallbackCardPayPaymentData) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *CallbackCardPayPaymentData) GetAmount() float64 {
	if m != nil {
		return m.Amount
	}
	return 0
}

func (m *CallbackCardPayPaymentData) GetAuthCode() string {
	if m != nil {
		return m.AuthCode
	}
	return ""
}

func (m *CallbackCardPayPaymentData) GetCreated() string {
	if m != nil {
		return m.Created
	}
	return ""
}

func (m *CallbackCardPayPaymentData) GetCurrency() string {
	if m != nil {
		return m.Currency
	}
	return ""
}

func (m *CallbackCardPayPaymentData) GetDeclineCode() string {
	if m != nil {
		return m.DeclineCode
	}
	return ""
}

func (m *CallbackCardPayPaymentData) GetDeclineReason() string {
	if m != nil {
		return m.DeclineReason
	}
	return ""
}

func (m *CallbackCardPayPaymentData) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *CallbackCardPayPaymentData) GetIs_3D() bool {
	if m != nil {
		return m.Is_3D
	}
	return false
}

func (m *CallbackCardPayPaymentData) GetNote() string {
	if m != nil {
		return m.Note
	}
	return ""
}

func (m *CallbackCardPayPaymentData) GetRrn() string {
	if m != nil {
		return m.Rrn
	}
	return ""
}

func (m *CallbackCardPayPaymentData) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

type CardPayCallbackRecurringDataFilling struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-" structure:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-" structure:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-" structure:"-"`
}

func (m *CardPayCallbackRecurringDataFilling) Reset()         { *m = CardPayCallbackRecurringDataFilling{} }
func (m *CardPayCallbackRecurringDataFilling) String() string { return proto.CompactTextString(m) }
func (*CardPayCallbackRecurringDataFilling) ProtoMessage()    {}
func (*CardPayCallbackRecurringDataFilling) Descriptor() ([]byte, []int) {
	return fileDescriptor_cardpay_b3de4ad0513c2c93, []int{8}
}
func (m *CardPayCallbackRecurringDataFilling) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CardPayCallbackRecurringDataFilling.Unmarshal(m, b)
}
func (m *CardPayCallbackRecurringDataFilling) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CardPayCallbackRecurringDataFilling.Marshal(b, m, deterministic)
}
func (dst *CardPayCallbackRecurringDataFilling) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CardPayCallbackRecurringDataFilling.Merge(dst, src)
}
func (m *CardPayCallbackRecurringDataFilling) XXX_Size() int {
	return xxx_messageInfo_CardPayCallbackRecurringDataFilling.Size(m)
}
func (m *CardPayCallbackRecurringDataFilling) XXX_DiscardUnknown() {
	xxx_messageInfo_CardPayCallbackRecurringDataFilling.DiscardUnknown(m)
}

var xxx_messageInfo_CardPayCallbackRecurringDataFilling proto.InternalMessageInfo

func (m *CardPayCallbackRecurringDataFilling) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type CardPayCallbackRecurringData struct {
	Id                   string                               `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Amount               float64                              `protobuf:"fixed64,2,opt,name=amount,proto3" json:"amount,omitempty"`
	AuthCode             string                               `protobuf:"bytes,3,opt,name=auth_code,json=authCode,proto3" json:"auth_code,omitempty"`
	Created              string                               `protobuf:"bytes,4,opt,name=created,proto3" json:"created,omitempty"`
	Currency             string                               `protobuf:"bytes,5,opt,name=currency,proto3" json:"currency,omitempty"`
	DeclineCode          string                               `protobuf:"bytes,6,opt,name=decline_code,json=declineCode,proto3" json:"decline_code,omitempty"`
	DeclineReason        string                               `protobuf:"bytes,7,opt,name=decline_reason,json=declineReason,proto3" json:"decline_reason,omitempty"`
	Description          string                               `protobuf:"bytes,8,opt,name=description,proto3" json:"description,omitempty"`
	Is_3D                bool                                 `protobuf:"varint,9,opt,name=is_3d,json=is3d,proto3" json:"is_3d,omitempty"`
	Note                 string                               `protobuf:"bytes,10,opt,name=note,proto3" json:"note,omitempty"`
	Rrn                  string                               `protobuf:"bytes,11,opt,name=rrn,proto3" json:"rrn,omitempty"`
	Status               string                               `protobuf:"bytes,12,opt,name=status,proto3" json:"status,omitempty"`
	Filing               *CardPayCallbackRecurringDataFilling `protobuf:"bytes,13,opt,name=filing,proto3" json:"filing,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                             `json:"-" bson:"-" structure:"-"`
	XXX_unrecognized     []byte                               `json:"-" bson:"-" structure:"-"`
	XXX_sizecache        int32                                `json:"-" bson:"-" structure:"-"`
}

func (m *CardPayCallbackRecurringData) Reset()         { *m = CardPayCallbackRecurringData{} }
func (m *CardPayCallbackRecurringData) String() string { return proto.CompactTextString(m) }
func (*CardPayCallbackRecurringData) ProtoMessage()    {}
func (*CardPayCallbackRecurringData) Descriptor() ([]byte, []int) {
	return fileDescriptor_cardpay_b3de4ad0513c2c93, []int{9}
}
func (m *CardPayCallbackRecurringData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CardPayCallbackRecurringData.Unmarshal(m, b)
}
func (m *CardPayCallbackRecurringData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CardPayCallbackRecurringData.Marshal(b, m, deterministic)
}
func (dst *CardPayCallbackRecurringData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CardPayCallbackRecurringData.Merge(dst, src)
}
func (m *CardPayCallbackRecurringData) XXX_Size() int {
	return xxx_messageInfo_CardPayCallbackRecurringData.Size(m)
}
func (m *CardPayCallbackRecurringData) XXX_DiscardUnknown() {
	xxx_messageInfo_CardPayCallbackRecurringData.DiscardUnknown(m)
}

var xxx_messageInfo_CardPayCallbackRecurringData proto.InternalMessageInfo

func (m *CardPayCallbackRecurringData) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *CardPayCallbackRecurringData) GetAmount() float64 {
	if m != nil {
		return m.Amount
	}
	return 0
}

func (m *CardPayCallbackRecurringData) GetAuthCode() string {
	if m != nil {
		return m.AuthCode
	}
	return ""
}

func (m *CardPayCallbackRecurringData) GetCreated() string {
	if m != nil {
		return m.Created
	}
	return ""
}

func (m *CardPayCallbackRecurringData) GetCurrency() string {
	if m != nil {
		return m.Currency
	}
	return ""
}

func (m *CardPayCallbackRecurringData) GetDeclineCode() string {
	if m != nil {
		return m.DeclineCode
	}
	return ""
}

func (m *CardPayCallbackRecurringData) GetDeclineReason() string {
	if m != nil {
		return m.DeclineReason
	}
	return ""
}

func (m *CardPayCallbackRecurringData) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *CardPayCallbackRecurringData) GetIs_3D() bool {
	if m != nil {
		return m.Is_3D
	}
	return false
}

func (m *CardPayCallbackRecurringData) GetNote() string {
	if m != nil {
		return m.Note
	}
	return ""
}

func (m *CardPayCallbackRecurringData) GetRrn() string {
	if m != nil {
		return m.Rrn
	}
	return ""
}

func (m *CardPayCallbackRecurringData) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

func (m *CardPayCallbackRecurringData) GetFiling() *CardPayCallbackRecurringDataFilling {
	if m != nil {
		return m.Filing
	}
	return nil
}

type CardPayPaymentCallback struct {
	MerchantOrder         *CardPayMerchantOrder                 `protobuf:"bytes,1,opt,name=merchant_order,json=merchantOrder,proto3" json:"merchant_order,omitempty"`
	PaymentMethod         string                                `protobuf:"bytes,2,opt,name=payment_method,json=paymentMethod,proto3" json:"payment_method,omitempty"`
	CallbackTime          string                                `protobuf:"bytes,3,opt,name=callback_time,json=callbackTime,proto3" json:"callback_time,omitempty"`
	CardAccount           *CallbackCardPayBankCardAccount       `protobuf:"bytes,4,opt,name=card_account,json=cardAccount,proto3" json:"card_account,omitempty"`
	CryptocurrencyAccount *CallbackCardPayCryptoCurrencyAccount `protobuf:"bytes,5,opt,name=cryptocurrency_account,json=cryptocurrencyAccount,proto3" json:"cryptocurrency_account,omitempty"`
	Customer              *CardPayCustomer                      `protobuf:"bytes,6,opt,name=customer,proto3" json:"customer,omitempty"`
	EwalletAccount        *CardPayEWalletAccount                `protobuf:"bytes,7,opt,name=ewallet_account,json=ewalletAccount,proto3" json:"ewallet_account,omitempty"`
	// @inject_tag: json:"payment_data,omitempty"
	PaymentData *CallbackCardPayPaymentData `protobuf:"bytes,8,opt,name=payment_data,json=paymentData,proto3" json:"payment_data,omitempty"`
	// @inject_tag: json:"recurring_data,omitempty"
	RecurringData        *CardPayCallbackRecurringData `protobuf:"bytes,9,opt,name=recurring_data,json=recurringData,proto3" json:"recurring_data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                      `json:"-" bson:"-" structure:"-"`
	XXX_unrecognized     []byte                        `json:"-" bson:"-" structure:"-"`
	XXX_sizecache        int32                         `json:"-" bson:"-" structure:"-"`
}

func (m *CardPayPaymentCallback) Reset()         { *m = CardPayPaymentCallback{} }
func (m *CardPayPaymentCallback) String() string { return proto.CompactTextString(m) }
func (*CardPayPaymentCallback) ProtoMessage()    {}
func (*CardPayPaymentCallback) Descriptor() ([]byte, []int) {
	return fileDescriptor_cardpay_b3de4ad0513c2c93, []int{10}
}
func (m *CardPayPaymentCallback) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CardPayPaymentCallback.Unmarshal(m, b)
}
func (m *CardPayPaymentCallback) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CardPayPaymentCallback.Marshal(b, m, deterministic)
}
func (dst *CardPayPaymentCallback) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CardPayPaymentCallback.Merge(dst, src)
}
func (m *CardPayPaymentCallback) XXX_Size() int {
	return xxx_messageInfo_CardPayPaymentCallback.Size(m)
}
func (m *CardPayPaymentCallback) XXX_DiscardUnknown() {
	xxx_messageInfo_CardPayPaymentCallback.DiscardUnknown(m)
}

var xxx_messageInfo_CardPayPaymentCallback proto.InternalMessageInfo

func (m *CardPayPaymentCallback) GetMerchantOrder() *CardPayMerchantOrder {
	if m != nil {
		return m.MerchantOrder
	}
	return nil
}

func (m *CardPayPaymentCallback) GetPaymentMethod() string {
	if m != nil {
		return m.PaymentMethod
	}
	return ""
}

func (m *CardPayPaymentCallback) GetCallbackTime() string {
	if m != nil {
		return m.CallbackTime
	}
	return ""
}

func (m *CardPayPaymentCallback) GetCardAccount() *CallbackCardPayBankCardAccount {
	if m != nil {
		return m.CardAccount
	}
	return nil
}

func (m *CardPayPaymentCallback) GetCryptocurrencyAccount() *CallbackCardPayCryptoCurrencyAccount {
	if m != nil {
		return m.CryptocurrencyAccount
	}
	return nil
}

func (m *CardPayPaymentCallback) GetCustomer() *CardPayCustomer {
	if m != nil {
		return m.Customer
	}
	return nil
}

func (m *CardPayPaymentCallback) GetEwalletAccount() *CardPayEWalletAccount {
	if m != nil {
		return m.EwalletAccount
	}
	return nil
}

func (m *CardPayPaymentCallback) GetPaymentData() *CallbackCardPayPaymentData {
	if m != nil {
		return m.PaymentData
	}
	return nil
}

func (m *CardPayPaymentCallback) GetRecurringData() *CardPayCallbackRecurringData {
	if m != nil {
		return m.RecurringData
	}
	return nil
}

func init() {
	proto.RegisterType((*CardPayAddress)(nil), "billing.CardPayAddress")
	proto.RegisterType((*CardPayItem)(nil), "billing.CardPayItem")
	proto.RegisterType((*CardPayMerchantOrder)(nil), "billing.CardPayMerchantOrder")
	proto.RegisterType((*CallbackCardPayBankCardAccount)(nil), "billing.CallbackCardPayBankCardAccount")
	proto.RegisterType((*CallbackCardPayCryptoCurrencyAccount)(nil), "billing.CallbackCardPayCryptoCurrencyAccount")
	proto.RegisterType((*CardPayCustomer)(nil), "billing.CardPayCustomer")
	proto.RegisterType((*CardPayEWalletAccount)(nil), "billing.CardPayEWalletAccount")
	proto.RegisterType((*CallbackCardPayPaymentData)(nil), "billing.CallbackCardPayPaymentData")
	proto.RegisterType((*CardPayCallbackRecurringDataFilling)(nil), "billing.CardPayCallbackRecurringDataFilling")
	proto.RegisterType((*CardPayCallbackRecurringData)(nil), "billing.CardPayCallbackRecurringData")
	proto.RegisterType((*CardPayPaymentCallback)(nil), "billing.CardPayPaymentCallback")
}

func init() { proto.RegisterFile("billing/cardpay.proto", fileDescriptor_cardpay_b3de4ad0513c2c93) }

var fileDescriptor_cardpay_b3de4ad0513c2c93 = []byte{
	// 957 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xec, 0x56, 0xdd, 0x6e, 0x23, 0x35,
	0x14, 0x56, 0x92, 0x26, 0x6d, 0xce, 0x34, 0xe9, 0xca, 0xdb, 0x86, 0xd1, 0xc2, 0xae, 0xca, 0x94,
	0x6a, 0x2b, 0x44, 0x5b, 0x94, 0xc2, 0x25, 0x42, 0x6d, 0xca, 0xa2, 0x45, 0xac, 0xa8, 0x46, 0x2b,
	0x21, 0x71, 0x33, 0x72, 0x6d, 0x93, 0x58, 0x9d, 0x19, 0x8f, 0x6c, 0x07, 0x14, 0x9e, 0x83, 0x6b,
	0x78, 0x0d, 0xee, 0x79, 0x07, 0xde, 0x83, 0x37, 0x40, 0xb6, 0x8f, 0x43, 0x92, 0xb2, 0xd5, 0x8a,
	0xeb, 0xbd, 0xf3, 0xf9, 0x6c, 0x9f, 0xf3, 0xf9, 0xfc, 0x7c, 0x33, 0x70, 0x70, 0x2b, 0xcb, 0x52,
	0xd6, 0xd3, 0x73, 0x46, 0x35, 0x6f, 0xe8, 0xe2, 0xac, 0xd1, 0xca, 0x2a, 0xb2, 0x8d, 0x70, 0xf6,
	0x6b, 0x0b, 0x86, 0x13, 0xaa, 0xf9, 0x0d, 0x5d, 0x5c, 0x72, 0xae, 0x85, 0x31, 0x24, 0x85, 0x6d,
	0xa6, 0xe6, 0xb5, 0xd5, 0x8b, 0xb4, 0x75, 0xd8, 0x3a, 0xe9, 0xe7, 0xd1, 0x24, 0x04, 0xb6, 0x98,
	0xb4, 0x8b, 0xb4, 0xed, 0x61, 0xbf, 0x26, 0xfb, 0xd0, 0x6d, 0x66, 0xaa, 0x16, 0x69, 0xc7, 0x83,
	0xc1, 0x70, 0xa8, 0xb1, 0xd4, 0x8a, 0x74, 0x2b, 0xa0, 0xde, 0x20, 0x23, 0xe8, 0x19, 0xab, 0x85,
	0xb0, 0x69, 0xd7, 0xc3, 0x68, 0x91, 0x47, 0xd0, 0xf9, 0x45, 0x36, 0x69, 0xcf, 0x83, 0x6e, 0x99,
	0x29, 0x48, 0x90, 0xd5, 0x4b, 0x2b, 0x2a, 0x17, 0xb8, 0xa6, 0x95, 0x40, 0x3e, 0x7e, 0x4d, 0x0e,
	0x21, 0xe1, 0xc2, 0x30, 0x2d, 0x1b, 0x2b, 0x55, 0x8d, 0x9c, 0x56, 0x21, 0x47, 0xc2, 0x33, 0xf7,
	0xd4, 0xba, 0x79, 0x30, 0x3c, 0x61, 0x2d, 0x59, 0xa0, 0xd6, 0xca, 0x83, 0x91, 0xfd, 0xd1, 0x82,
	0x7d, 0x8c, 0xf8, 0x4a, 0x68, 0x36, 0xa3, 0xb5, 0xfd, 0x4e, 0x73, 0xa1, 0xc9, 0x10, 0xda, 0x92,
	0x63, 0xe0, 0xb6, 0xe4, 0x6f, 0x11, 0xf6, 0x63, 0xe8, 0x4a, 0x2b, 0x2a, 0x93, 0x76, 0x0e, 0x3b,
	0x27, 0xc9, 0x78, 0xff, 0x0c, 0x73, 0x7d, 0xb6, 0xf2, 0xa2, 0x3c, 0x1c, 0x21, 0x57, 0xf0, 0xc8,
	0xcc, 0x64, 0xd3, 0xc8, 0x7a, 0x5a, 0xd0, 0x90, 0x7f, 0xcf, 0x2b, 0x19, 0xbf, 0xb7, 0x79, 0x0d,
	0xcb, 0x93, 0xef, 0xc5, 0x0b, 0x08, 0x64, 0xbf, 0xb7, 0xe0, 0xd9, 0x84, 0x96, 0xe5, 0x2d, 0x65,
	0x77, 0x78, 0xf6, 0x8a, 0xd6, 0x7e, 0x79, 0xc9, 0xc2, 0x9b, 0x47, 0xd0, 0x9b, 0xa9, 0x92, 0x0b,
	0x8d, 0x0f, 0x41, 0x8b, 0x7c, 0x0a, 0xfb, 0xd2, 0x98, 0xb9, 0x8b, 0x8e, 0x35, 0x2e, 0x98, 0xe2,
	0x02, 0x5f, 0x45, 0x70, 0x6f, 0x12, 0xb6, 0x26, 0x8a, 0x0b, 0xf2, 0x14, 0xa0, 0xa2, 0xe6, 0x4e,
	0xf0, 0xa2, 0xa1, 0x35, 0xd6, 0xbc, 0x1f, 0x90, 0x1b, 0xea, 0x53, 0x6e, 0xd5, 0x9d, 0xa8, 0x63,
	0xdd, 0xbd, 0x91, 0xfd, 0xd9, 0x82, 0x8f, 0x36, 0x18, 0x4e, 0xf4, 0xa2, 0xb1, 0x6a, 0x32, 0xd7,
	0x5a, 0xd4, 0x6c, 0x11, 0x79, 0x1e, 0xc3, 0x90, 0xf9, 0x8d, 0x65, 0x32, 0x02, 0xdf, 0x41, 0x40,
	0x63, 0x87, 0x8e, 0xe1, 0x00, 0x8f, 0x59, 0x4d, 0x6b, 0x43, 0x99, 0xcb, 0x7b, 0x21, 0x39, 0xf2,
	0x7e, 0x1c, 0x36, 0x5f, 0xff, 0xbb, 0xf7, 0x92, 0x3b, 0xe2, 0x8d, 0x66, 0x05, 0xad, 0x96, 0x1d,
	0xd1, 0xcf, 0xfb, 0x8d, 0x66, 0x97, 0x1e, 0x20, 0x1f, 0xc2, 0xae, 0xdb, 0x66, 0x48, 0x08, 0xf9,
	0x27, 0x8d, 0x66, 0x91, 0x63, 0x56, 0xc0, 0x5e, 0x24, 0x3f, 0x37, 0x56, 0x55, 0x42, 0xbb, 0xe7,
	0x8a, 0x8a, 0xca, 0x12, 0x69, 0x06, 0xc3, 0xb7, 0x4c, 0x83, 0x5c, 0xda, 0xb2, 0xc1, 0x16, 0xea,
	0x2c, 0x5b, 0x68, 0x04, 0xbd, 0x52, 0x31, 0x5a, 0xc6, 0xe9, 0x40, 0x2b, 0x7b, 0x0e, 0x07, 0x18,
	0xe0, 0xab, 0xef, 0x69, 0x59, 0x0a, 0x1b, 0xd3, 0xb2, 0xd1, 0x83, 0xd9, 0x5f, 0x6d, 0x78, 0xb2,
	0x91, 0xcf, 0x1b, 0xba, 0xa8, 0x44, 0x6d, 0xaf, 0xa9, 0xa5, 0xf7, 0x5a, 0x76, 0x04, 0x3d, 0x7c,
	0x76, 0xdb, 0xb7, 0x3c, 0x5a, 0xe4, 0x7d, 0xe8, 0xd3, 0xb9, 0x9d, 0x85, 0x92, 0x07, 0x7a, 0x3b,
	0x0e, 0xf0, 0x85, 0x76, 0x2a, 0xa0, 0x05, 0xb5, 0x82, 0x23, 0xcb, 0x68, 0x92, 0x27, 0xb0, 0xb3,
	0x4c, 0x53, 0x98, 0xe3, 0xa5, 0xed, 0xd2, 0xc8, 0x05, 0x2b, 0x65, 0x2d, 0x82, 0xd7, 0x5e, 0x1c,
	0x0f, 0x8f, 0x79, 0xc7, 0xc7, 0x30, 0x8c, 0x47, 0xb4, 0xa0, 0x46, 0xd5, 0xe9, 0x76, 0xa8, 0x31,
	0xa2, 0xb9, 0x07, 0x37, 0xe7, 0x6c, 0xe7, 0xfe, 0x9c, 0x3d, 0x86, 0xae, 0x34, 0xc5, 0x05, 0x4f,
	0xfb, 0x87, 0xad, 0x93, 0x9d, 0x7c, 0x4b, 0x9a, 0x0b, 0xee, 0x95, 0x42, 0x59, 0x91, 0x02, 0x2a,
	0x85, 0xb2, 0xc2, 0xc9, 0x8b, 0xd6, 0x75, 0x9a, 0x04, 0x79, 0xd1, 0xba, 0x0e, 0x42, 0x44, 0xed,
	0xdc, 0xa4, 0xbb, 0x51, 0x88, 0x9c, 0x95, 0x7d, 0x0e, 0x47, 0xb1, 0xc4, 0x98, 0xde, 0x5c, 0xb8,
	0xb7, 0xc9, 0x7a, 0xea, 0x32, 0xfb, 0x22, 0x4c, 0xe4, 0xbd, 0x7a, 0xfc, 0xd6, 0x81, 0x0f, 0x1e,
	0xba, 0xf7, 0xae, 0x22, 0xff, 0xb7, 0x22, 0xe4, 0x1a, 0x7a, 0x3f, 0x4a, 0x97, 0xf4, 0x74, 0xe0,
	0x65, 0xf1, 0x93, 0x4d, 0x59, 0x7c, 0xa8, 0x50, 0x39, 0xde, 0xcd, 0xfe, 0xde, 0x82, 0xd1, 0xfa,
	0xa0, 0xc4, 0x6b, 0xe4, 0x1a, 0x86, 0x15, 0x0a, 0x7e, 0xa1, 0x74, 0x94, 0xc8, 0x64, 0xfc, 0x74,
	0x33, 0xd0, 0xda, 0x67, 0x21, 0x1f, 0x54, 0x6b, 0x5f, 0x89, 0x63, 0x18, 0x36, 0xc1, 0x71, 0x51,
	0x09, 0x3b, 0x53, 0x51, 0x8a, 0x06, 0x88, 0xbe, 0xf2, 0x20, 0x39, 0x82, 0x01, 0xc3, 0xc0, 0x85,
	0x95, 0x55, 0xac, 0xf1, 0x6e, 0x04, 0x5f, 0xcb, 0x4a, 0x90, 0x6f, 0x60, 0xd7, 0x7d, 0xac, 0x0b,
	0x1a, 0xa6, 0x1f, 0xbf, 0x07, 0xcf, 0x57, 0xf8, 0x3c, 0xa4, 0xf5, 0x79, 0xc2, 0x56, 0x84, 0x9f,
	0xc3, 0x28, 0x88, 0x61, 0xec, 0x87, 0xa5, 0xd7, 0xae, 0xf7, 0x7a, 0xfa, 0x26, 0xaf, 0xff, 0xa9,
	0xcf, 0xf9, 0xc1, 0xba, 0xb3, 0x18, 0xe5, 0x33, 0xd7, 0x7f, 0x41, 0x12, 0x7d, 0x7f, 0x25, 0xe3,
	0xf4, 0x5e, 0x99, 0x70, 0x3f, 0x5f, 0x9e, 0x24, 0x5f, 0xc3, 0x9e, 0xf8, 0xd9, 0xeb, 0xdc, 0x92,
	0xd4, 0xb6, 0xbf, 0xfc, 0x6c, 0xf3, 0xf2, 0xba, 0x1c, 0xe6, 0x43, 0xbc, 0x16, 0xc3, 0xbf, 0x80,
	0xdd, 0x98, 0x7c, 0x4e, 0x2d, 0xf5, 0x9d, 0x99, 0x8c, 0x8f, 0xde, 0xf4, 0xb4, 0x15, 0xa9, 0xcc,
	0x93, 0x66, 0x45, 0x37, 0xbf, 0x85, 0xa1, 0x8e, 0x5d, 0x14, 0x3c, 0xf5, 0xbd, 0xa7, 0xe3, 0xb7,
	0xea, 0xb9, 0x7c, 0xa0, 0x57, 0xcd, 0xab, 0x2f, 0x7f, 0xf8, 0x62, 0x2a, 0xed, 0x6c, 0x7e, 0x7b,
	0xc6, 0x54, 0x75, 0xde, 0xd0, 0x85, 0x99, 0x37, 0x42, 0x2f, 0x17, 0xa7, 0xe8, 0xf3, 0xd4, 0x08,
	0xfd, 0x93, 0xc3, 0xef, 0xa6, 0xe7, 0xfe, 0xdf, 0xec, 0x1c, 0x37, 0x6e, 0x7b, 0xde, 0xbc, 0xf8,
	0x27, 0x00, 0x00, 0xff, 0xff, 0x44, 0x55, 0x6f, 0x4c, 0xc3, 0x09, 0x00, 0x00,
}

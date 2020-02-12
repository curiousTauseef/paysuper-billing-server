package pkg

import (
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PaymentChannelCostMerchantSet struct {
	Id  string                                  `bson:"_id"`
	Set []*billingpb.PaymentChannelCostMerchant `bson:"set"`
}

type PaymentChannelCostSystemSet struct {
	Id  string                                `bson:"_id"`
	Set []*billingpb.PaymentChannelCostSystem `bson:"set"`
}

type MoneyBackCostMerchantSet struct {
	Id  string                             `bson:"_id"`
	Set []*billingpb.MoneyBackCostMerchant `bson:"set"`
}

type MoneyBackCostSystemSet struct {
	Id  string                           `bson:"_id"`
	Set []*billingpb.MoneyBackCostSystem `bson:"set"`
}

type BinData struct {
	Id                 primitive.ObjectID `bson:"_id"`
	CardBin            int32              `bson:"card_bin"`
	CardBrand          string             `bson:"card_brand"`
	CardType           string             `bson:"card_type"`
	CardCategory       string             `bson:"card_category"`
	BankName           string             `bson:"bank_name"`
	BankCountryName    string             `bson:"bank_country_name"`
	BankCountryIsoCode string             `bson:"bank_country_code_a2"`
	BankSite           string             `bson:"bank_site"`
	BankPhone          string             `bson:"bank_phone"`
}

type BalanceQueryResItem struct {
	Amount float64 `bson:"amount"`
}

type ReserveQueryResItem struct {
	Type   string  `bson:"_id"`
	Amount float64 `bson:"amount"`
}

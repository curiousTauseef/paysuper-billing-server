package pkg

import (
	"github.com/paysuper/paysuper-proto/go/billingpb"
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

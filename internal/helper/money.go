package helper

import (
	"github.com/paysuper/paysuper-proto/go/billingpb"
	tools "github.com/paysuper/paysuper-tools/number"
	"go.uber.org/zap"
)

type Money struct {
	money *tools.Money
}

func NewMoney() *Money {
	return &Money{money: tools.New()}
}

func (m *Money) Round(val float64) (float64, error) {
	rounded, err := m.money.Round(val, 2)

	if err != nil {
		zap.L().Error(
			billingpb.ErrorUnableRound,
			zap.Error(err),
			zap.Float64(billingpb.ErrorFieldValue, val),
		)
	}

	return rounded, err
}

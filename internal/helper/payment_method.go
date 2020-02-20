package helper

import (
	"fmt"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"strings"
)

func GetPaymentMethodKey(currency, mccCode, operatingCompanyId, brand string) string {
	return fmt.Sprintf(
		pkg.PaymentMethodKey,
		strings.ToUpper(currency),
		mccCode,
		strings.ToLower(operatingCompanyId),
		strings.ToUpper(brand),
	)
}

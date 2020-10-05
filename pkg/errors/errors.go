package errors

import (
	"errors"
	"github.com/paysuper/paysuper-proto/go/billingpb"
)

func NewBillingServerErrorMsg(code, msg string, details ...string) *billingpb.ResponseErrorMessage {
	var det string

	if len(details) > 0 && details[0] != "" {
		det = details[0]
	} else {
		det = ""
	}

	return &billingpb.ResponseErrorMessage{Code: code, Message: msg, Details: det}
}

func NewBillingServerResponseError(status int32, message *billingpb.ResponseErrorMessage) *billingpb.ResponseError {
	return &billingpb.ResponseError{
		Status:  status,
		Message: message,
	}
}

var (
	KeyErrorFileProcess    = NewBillingServerErrorMsg("ks000001", "failed to process file")
	KeyErrorNotFound       = NewBillingServerErrorMsg("ks000002", "key not found")
	KeyErrorFailedToInsert = NewBillingServerErrorMsg("ks000003", "failed to insert key")
	KeyErrorCanceled       = NewBillingServerErrorMsg("ks000004", "unable to cancel key")
	KeyErrorFinish         = NewBillingServerErrorMsg("ks000005", "unable to finish key")
	KeyErrorReserve        = NewBillingServerErrorMsg("ks000006", "unable to reserve key")

	ErrBalanceHasMoreOneCurrency = errors.New("merchant balance has more one currency")
)

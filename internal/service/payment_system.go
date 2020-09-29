package service

import (
	"github.com/paysuper/paysuper-billing-server/internal/payment_system"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"sync"
)

const (
	PaymentSystemHandlerMockOk      = "mock_ok"
	PaymentSystemHandlerMockError   = "mock_error"
	PaymentSystemHandlerCardPayMock = "cardpay_mock"
)

var (
	registry = map[string]func() payment_system.PaymentSystemInterface{
		billingpb.PaymentSystemHandlerCardPay: payment_system.NewCardPayHandler,
		PaymentSystemHandlerMockOk:            NewPaymentSystemMockOk,
		PaymentSystemHandlerMockError:         NewPaymentSystemMockError,
		PaymentSystemHandlerCardPayMock:       NewCardPayMock,
	}
)

func NewPaymentSystemGateway() payment_system.PaymentSystemManagerInterface {
	paymentSystem := &Gateway{
		gateways: make(map[string]payment_system.PaymentSystemInterface),
	}
	return paymentSystem
}

type Gateway struct {
	gateways map[string]payment_system.PaymentSystemInterface
	mx       sync.Mutex
}

func (m *Gateway) GetGateway(name string) (payment_system.PaymentSystemInterface, error) {
	initFn, ok := registry[name]

	if !ok {
		return nil, payment_system.PaymentSystemErrorHandlerNotFound
	}

	m.mx.Lock()
	gateway, ok := m.gateways[name]

	if !ok {
		gateway = initFn()
		m.gateways[name] = gateway
	}

	m.mx.Unlock()
	return gateway, nil
}

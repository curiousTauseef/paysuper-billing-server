package service

import (
	"github.com/paysuper/paysuper-billing-server/pkg"
	"go.uber.org/zap"
	"net"
	"net/url"
)

type Repository struct {
	svc *Service
}

type DashboardRepository Repository

type kvIntFloat struct {
	Key   int
	Value float64
}

type kvIntInt struct {
	Key   int
	Value int32
}

func getHostFromUrl(urlString string) string {
	u, err := url.Parse(urlString)
	if err != nil {
		zap.L().Error(
			"url parsing failed",
			zap.Error(err),
		)
		return ""
	}
	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		return u.Host
	}
	return host
}

func getMccByOperationsType(operationsType string) (string, error) {
	mccCode, ok := pkg.MerchantOperationsTypesToMccCodes[operationsType]
	if !ok {
		return "", merchantErrorOperationsTypeNotSupported
	}
	return mccCode, nil
}

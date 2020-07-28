package service

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"log"
	"net"
)

func (s *Service) MigrateCustomers(ctx context.Context) error {
	customers, err := s.customerRepository.FindAll(ctx)

	if err != nil {
		return err
	}

	for _, customer := range customers {
		if customer.Id == "5f202981a4479cab56313f60" {
			log.Println(123)
		}

		if len(customer.Ip) > 0 {
			ip := net.IP(customer.Ip).String()

			if customer.IpString == "" {
				customer.IpString = ip
			}

			if customer.Address == nil {
				if address, err := s.getAddressByIp(ctx, ip); err == nil {
					customer.Address = address
				}
			}

			exists := false

			for key, val := range customer.IpHistory {
				hIp := net.IP(val.Ip).String()

				if val.IpString == "" {
					customer.IpHistory[key].IpString = hIp
				}

				if val.Address == nil {
					if address, err := s.getAddressByIp(ctx, hIp); err == nil {
						customer.IpHistory[key].Address = address
					}
				}

				if hIp == ip {
					exists = true
				}
			}

			if !exists {
				it := &billingpb.CustomerIpHistory{
					Ip:        customer.Ip,
					CreatedAt: customer.CreatedAt,
					IpString:  ip,
					Address:   customer.Address,
				}
				customer.IpHistory = append(customer.IpHistory, it)
			}
		}

		if customer.Address != nil && len(customer.AddressHistory) <= 0 {
			it := &billingpb.CustomerAddressHistory{
				Country:    customer.Address.Country,
				City:       customer.Address.City,
				State:      customer.Address.State,
				PostalCode: customer.Address.PostalCode,
				CreatedAt:  customer.CreatedAt,
			}
			customer.AddressHistory = append(customer.AddressHistory, it)
		}

		if err := s.customerRepository.Update(ctx, customer); err != nil {
			return err
		}
	}

	return nil
}

package service

import (
	"context"
	"fmt"
	"github.com/globalsign/mgo/bson"
	casbinProto "github.com/paysuper/casbin-server/pkg/generated/api/proto/casbinpb"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"go.uber.org/zap"
)

const (
	cacheMerchantId       = "merchant:id:%s"
	cacheMerchantCommonId = "merchant:common:id:%s"

	collectionMerchant                     = "merchant"
	collectionMerchantPaymentMethodHistory = "payment_method_history"
)

func (s *Service) MerchantsMigrate() error {
	merchants, err := s.merchant.GetAll()

	if err != nil {
		zap.L().Error("[task merchants migrate] Unable to get merchants", zap.Error(err))
		return nil
	}

	for _, merchant := range merchants {
		if merchant.User == nil ||
			merchant.User.Id == "" ||
			merchant.User.Email == "" ||
			merchant.User.FirstName == "" ||
			merchant.User.LastName == "" {
			continue
		}

		userRole := &billing.UserRole{
			Id:         bson.NewObjectId().Hex(),
			MerchantId: merchant.Id,
			UserId:     merchant.User.Id,
			Email:      merchant.User.Email,
			FirstName:  merchant.User.FirstName,
			LastName:   merchant.User.LastName,
			Role:       pkg.RoleMerchantOwner,
			Status:     pkg.UserRoleStatusAccepted,
		}

		_, err := s.userRoleRepository.GetMerchantUserByUserId(userRole.MerchantId, userRole.UserId)

		if err != nil {
			err = s.userRoleRepository.AddMerchantUser(userRole)
		}

		if err != nil {
			zap.L().Error("[task merchants migrate] Unable to add merchant user role", zap.Error(err))
			continue
		}

		casbinRole := &casbinProto.UserRoleRequest{
			User: fmt.Sprintf(pkg.CasbinMerchantUserMask, merchant.Id, merchant.User.Id),
			Role: pkg.RoleMerchantOwner,
		}

		roles, err := s.casbinService.GetRolesForUser(context.TODO(), casbinRole)

		if roles == nil || len(roles.Array) < 1 {
			_, err = s.casbinService.AddRoleForUser(context.TODO(), casbinRole)
		}

		if err != nil {
			zap.L().Error("[task merchants migrate] Unable to add user to casbin", zap.Error(err), zap.Any("role", casbinRole))
		}
	}

	zap.L().Info("[task merchants migrate] Finished successfully")

	return nil
}

type MerchantRepositoryInterface interface {
	Update(merchant *billing.Merchant) error
	Insert(merchant *billing.Merchant) error
	Upsert(merchant *billing.Merchant) error
	MultipleInsert(merchants []*billing.Merchant) error
	GetById(id string) (*billing.Merchant, error)
	GetCommonById(id string) (*billing.MerchantCommon, error)
	GetPaymentMethod(merchantId string, method string) (*billing.MerchantPaymentMethod, error)
	GetMerchantsWithAutoPayouts() ([]*billing.Merchant, error)
	GetAll() ([]*billing.Merchant, error)
}

func newMerchantService(svc *Service) MerchantRepositoryInterface {
	return &Merchant{svc: svc}
}

func (h *Merchant) GetMerchantsWithAutoPayouts() (merchants []*billing.Merchant, err error) {
	query := bson.M{
		"manual_payouts_enabled": false,
	}
	err = h.svc.db.Collection(collectionMerchant).Find(query).All(&merchants)
	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionMerchant),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
	}

	return
}

func (h *Merchant) Update(merchant *billing.Merchant) error {
	err := h.svc.db.Collection(collectionMerchant).UpdateId(bson.ObjectIdHex(merchant.Id), merchant)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionMerchant),
			zap.Any(pkg.ErrorDatabaseFieldQuery, merchant),
		)

		return err
	}

	key := fmt.Sprintf(cacheMerchantId, merchant.Id)
	err = h.svc.cacher.Set(key, merchant, 0)

	if err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorDatabaseFieldQuery, merchant),
		)

		return err
	}

	return nil
}

func (h *Merchant) Insert(merchant *billing.Merchant) error {
	err := h.svc.db.Collection(collectionMerchant).Insert(merchant)
	if err != nil {
		return err
	}

	err = h.svc.cacher.Set(fmt.Sprintf(cacheMerchantId, merchant.Id), merchant, 0)

	if err != nil {
		return err
	}

	return nil
}

func (h *Merchant) Upsert(merchant *billing.Merchant) error {
	_, err := h.svc.db.Collection(collectionMerchant).UpsertId(bson.ObjectIdHex(merchant.Id), merchant)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionMerchant),
			zap.Any(pkg.ErrorDatabaseFieldQuery, merchant),
		)

		return err
	}

	key := fmt.Sprintf(cacheMerchantId, merchant.Id)
	err = h.svc.cacher.Set(key, merchant, 0)

	if err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorDatabaseFieldQuery, merchant),
		)

		return err
	}

	return nil
}

func (h *Merchant) MultipleInsert(merchants []*billing.Merchant) error {
	m := make([]interface{}, len(merchants))
	for i, v := range merchants {
		m[i] = v
	}

	if err := h.svc.db.Collection(collectionMerchant).Insert(m...); err != nil {
		return err
	}

	return nil
}

func (h *Merchant) GetById(id string) (*billing.Merchant, error) {
	var c billing.Merchant
	key := fmt.Sprintf(cacheMerchantId, id)

	if err := h.svc.cacher.Get(key, c); err == nil {
		return &c, nil
	}

	query := bson.M{"_id": bson.ObjectIdHex(id)}
	err := h.svc.db.Collection(collectionMerchant).Find(query).One(&c)
	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionMerchant),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, fmt.Errorf(errorNotFound, collectionMerchant)
	}

	if err := h.svc.cacher.Set(key, c, 0); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorCacheFieldData, c),
		)
	}

	return &c, nil
}

func (h *Merchant) GetCommonById(id string) (*billing.MerchantCommon, error) {
	var c billing.MerchantCommon
	key := fmt.Sprintf(cacheMerchantCommonId, id)

	if err := h.svc.cacher.Get(key, c); err == nil {
		return &c, nil
	}

	query := bson.M{"_id": bson.ObjectIdHex(id)}
	err := h.svc.db.Collection(collectionMerchant).Find(query).One(&c)
	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionMerchant),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
		return nil, fmt.Errorf(errorNotFound, collectionMerchant)
	}

	if err := h.svc.cacher.Set(key, c, 0); err != nil {
		zap.L().Error(
			pkg.ErrorCacheQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorCacheFieldCmd, "SET"),
			zap.String(pkg.ErrorCacheFieldKey, key),
			zap.Any(pkg.ErrorCacheFieldData, c),
		)
	}

	return &c, nil
}

func (h Merchant) GetPaymentMethod(merchantId string, method string) (*billing.MerchantPaymentMethod, error) {
	m, err := h.GetById(merchantId)
	if err != nil {
		return nil, err
	}

	merchantPaymentMethods := make(map[string]*billing.MerchantPaymentMethod)
	if len(m.PaymentMethods) > 0 {
		for k, v := range m.PaymentMethods {
			merchantPaymentMethods[k] = v
		}
	}

	pm, err := h.svc.paymentMethod.GetAll()
	if err != nil {
		return nil, err
	}
	if len(merchantPaymentMethods) != len(pm) {
		for k, v := range pm {
			_, ok := merchantPaymentMethods[k]

			if ok {
				continue
			}

			merchantPaymentMethods[k] = &billing.MerchantPaymentMethod{
				PaymentMethod: &billing.MerchantPaymentMethodIdentification{
					Id:   k,
					Name: v.Name,
				},
				Commission: &billing.MerchantPaymentMethodCommissions{
					Fee: DefaultPaymentMethodFee,
					PerTransaction: &billing.MerchantPaymentMethodPerTransactionCommission{
						Fee:      DefaultPaymentMethodPerTransactionFee,
						Currency: DefaultPaymentMethodCurrency,
					},
				},
				Integration: &billing.MerchantPaymentMethodIntegration{},
				IsActive:    true,
			}
		}
	}

	if _, ok := merchantPaymentMethods[method]; !ok {
		return nil, fmt.Errorf(errorNotFound, collectionMerchant)
	}

	return merchantPaymentMethods[method], nil
}

func (h *Merchant) GetAll() ([]*billing.Merchant, error) {
	var c []*billing.Merchant

	err := h.svc.db.Collection(collectionMerchant).Find(nil).All(&c)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionMerchant),
		)
		return nil, fmt.Errorf(errorNotFound, collectionMerchant)
	}

	return c, nil
}

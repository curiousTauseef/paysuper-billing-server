package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/paysuper/paysuper-billing-server/pkg"
	errors2 "github.com/paysuper/paysuper-billing-server/pkg/errors"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	casbinProto "github.com/paysuper/paysuper-proto/go/casbinpb"
	"github.com/paysuper/paysuper-proto/go/postmarkpb"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"time"
)

const (
	defaultCompanyName = "PaySuper"

	claimType   = "type"
	claimEmail  = "email"
	claimRoleId = "role_id"
	claimExpire = "exp"

	roleNameMerchantOwner      = "Owner"
	roleNameMerchantDeveloper  = "Developer"
	roleNameMerchantAccounting = "Accounting"
	roleNameMerchantSupport    = "Support"
	roleNameMerchantViewOnly   = "View only"
	roleNameSystemAdmin        = "Admin"
	roleNameSystemRiskManager  = "Risk manager"
	roleNameSystemFinancial    = "Financial"
	roleNameSystemSupport      = "Support"
	roleNameSystemViewOnly     = "View only"
)

var (
	usersDbInternalError              = errors2.NewBillingServerErrorMsg("uu000001", "unknown database error")
	errorUserAlreadyExist             = errors2.NewBillingServerErrorMsg("uu000002", "user already exist")
	errorUserUnableToAdd              = errors2.NewBillingServerErrorMsg("uu000003", "unable to add user")
	errorUserNotFound                 = errors2.NewBillingServerErrorMsg("uu000004", "user not found")
	errorUserInviteAlreadyAccepted    = errors2.NewBillingServerErrorMsg("uu000005", "user already accepted invite")
	errorUserMerchantNotFound         = errors2.NewBillingServerErrorMsg("uu000006", "merchant not found")
	errorUserUnableToSendInvite       = errors2.NewBillingServerErrorMsg("uu000007", "unable to send invite email")
	errorUserConfirmEmail             = errors2.NewBillingServerErrorMsg("uu000008", "unable to confirm email")
	errorUserUnableToCreateToken      = errors2.NewBillingServerErrorMsg("uu000009", "unable to create invite token")
	errorUserInvalidToken             = errors2.NewBillingServerErrorMsg("uu000010", "invalid token string")
	errorUserInvalidInviteEmail       = errors2.NewBillingServerErrorMsg("uu000011", "email in request and token are not equal")
	errorUserUnableToSave             = errors2.NewBillingServerErrorMsg("uu000013", "can't update user in db")
	errorUserUnableToAddToCasbin      = errors2.NewBillingServerErrorMsg("uu000014", "unable to add user to the casbin server")
	errorUserUnsupportedRoleType      = errors2.NewBillingServerErrorMsg("uu000015", "unsupported role type")
	errorUserUnableToDelete           = errors2.NewBillingServerErrorMsg("uu000016", "unable to delete user")
	errorUserUnableToDeleteFromCasbin = errors2.NewBillingServerErrorMsg("uu000017", "unable to delete user from the casbin server")
	errorUserDontHaveRole             = errors2.NewBillingServerErrorMsg("uu000018", "user dont have any role")
	errorUserUnableResendInvite       = errors2.NewBillingServerErrorMsg("uu000019", "unable to resend invite")
	errorUserGetImplicitPermissions   = errors2.NewBillingServerErrorMsg("uu000020", "unable to get implicit permission for user")
	errorUserProfileNotFound          = errors2.NewBillingServerErrorMsg("uu000021", "unable to get user profile")
	errorUserEmptyNames               = errors2.NewBillingServerErrorMsg("uu000022", "first and last names cannot be empty")
	errorUserEmptyCompanyName         = errors2.NewBillingServerErrorMsg("uu000023", "company name cannot be empty")

	merchantUserRoles = map[string][]*billingpb.RoleListItem{
		pkg.RoleTypeMerchant: {
			{Id: billingpb.RoleMerchantOwner, Name: roleNameMerchantOwner},
			{Id: billingpb.RoleMerchantDeveloper, Name: roleNameMerchantDeveloper},
			{Id: billingpb.RoleMerchantAccounting, Name: roleNameMerchantAccounting},
			{Id: billingpb.RoleMerchantSupport, Name: roleNameMerchantSupport},
			{Id: billingpb.RoleMerchantViewOnly, Name: roleNameMerchantViewOnly},
		},
		pkg.RoleTypeSystem: {
			{Id: billingpb.RoleSystemAdmin, Name: roleNameSystemAdmin},
			{Id: billingpb.RoleSystemRiskManager, Name: roleNameSystemRiskManager},
			{Id: billingpb.RoleSystemFinancial, Name: roleNameSystemFinancial},
			{Id: billingpb.RoleSystemSupport, Name: roleNameSystemSupport},
			{Id: billingpb.RoleSystemViewOnly, Name: roleNameSystemViewOnly},
		},
	}

	merchantUserRolesNames = map[string]string{
		billingpb.RoleMerchantOwner:      roleNameMerchantOwner,
		billingpb.RoleMerchantDeveloper:  roleNameMerchantDeveloper,
		billingpb.RoleMerchantAccounting: roleNameMerchantAccounting,
		billingpb.RoleMerchantSupport:    roleNameMerchantSupport,
		billingpb.RoleMerchantViewOnly:   roleNameMerchantViewOnly,
	}
)

func (s *Service) GetMerchantUsers(ctx context.Context, req *billingpb.GetMerchantUsersRequest, res *billingpb.GetMerchantUsersResponse) error {
	users, err := s.userRoleRepository.GetUsersForMerchant(ctx, req.MerchantId)

	if err != nil {
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = usersDbInternalError
		res.Message.Details = err.Error()

		return nil
	}

	res.Status = billingpb.ResponseStatusOk
	res.Users = users

	return nil
}

func (s *Service) GetAdminUsers(ctx context.Context, _ *billingpb.EmptyRequest, res *billingpb.GetAdminUsersResponse) error {
	users, err := s.userRoleRepository.GetUsersForAdmin(ctx)

	if err != nil {
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = usersDbInternalError
		res.Message.Details = err.Error()

		return nil
	}
	res.Status = billingpb.ResponseStatusOk
	res.Users = users

	return nil
}

func (s *Service) GetMerchantsForUser(ctx context.Context, req *billingpb.GetMerchantsForUserRequest, res *billingpb.GetMerchantsForUserResponse) error {
	users, err := s.userRoleRepository.GetMerchantsForUser(ctx, req.UserId)

	if err != nil {
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = usersDbInternalError
		res.Message.Details = err.Error()

		return nil
	}

	merchants := make([]*billingpb.MerchantForUserInfo, len(users))

	for i, user := range users {
		merchant, err := s.merchantRepository.GetById(ctx, user.MerchantId)
		if err != nil {
			zap.L().Error(
				"Can't get merchant by id",
				zap.Error(err),
				zap.String("merchant_id", user.MerchantId),
			)

			res.Status = billingpb.ResponseStatusSystemError
			res.Message = usersDbInternalError
			res.Message.Details = err.Error()

			return nil
		}

		name := merchant.Id
		if merchant.Company != nil {
			name = merchant.Company.Name
		}

		merchants[i] = &billingpb.MerchantForUserInfo{
			Id:   user.MerchantId,
			Name: name,
			Role: user.Role,
		}
	}

	res.Status = billingpb.ResponseStatusOk
	res.Merchants = merchants
	return nil
}

func (s *Service) InviteUserMerchant(
	ctx context.Context,
	req *billingpb.InviteUserMerchantRequest,
	res *billingpb.InviteUserMerchantResponse,
) error {
	if req.Role == billingpb.RoleMerchantOwner {
		zap.L().Error(errorUserUnsupportedRoleType.Message, zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserUnsupportedRoleType

		return nil
	}

	merchant, err := s.merchantRepository.GetById(ctx, req.MerchantId)

	if err != nil {
		zap.L().Error(errorUserMerchantNotFound.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserMerchantNotFound

		return nil
	}

	if merchant.Company == nil || merchant.Company.Name == "" {
		zap.L().Error(errorUserEmptyCompanyName.Message, zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserEmptyCompanyName

		return nil
	}

	owner, err := s.userRoleRepository.GetMerchantOwner(ctx, merchant.Id)

	if err != nil {
		zap.L().Error(errorUserNotFound.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserNotFound

		return nil
	}

	user, err := s.userRoleRepository.GetMerchantUserByEmail(ctx, merchant.Id, req.Email)
	zap.L().Error("[InviteUserMerchant] GetMerchantUserByEmail", zap.Error(err), zap.Any("req", req), zap.Any("user", user))
	if (err != nil && err != mongo.ErrNoDocuments) || user != nil {
		zap.L().Error(errorUserAlreadyExist.Message, zap.Error(err), zap.Any("req", req), zap.Any("user", user))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserAlreadyExist

		return nil
	}

	role := &billingpb.UserRole{
		Id:         primitive.NewObjectID().Hex(),
		MerchantId: merchant.Id,
		Role:       req.Role,
		Status:     pkg.UserRoleStatusInvited,
		Email:      req.Email,
	}

	if err = s.userRoleRepository.AddMerchantUser(ctx, role); err != nil {
		zap.L().Error(errorUserUnableToAdd.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserUnableToAdd

		return nil
	}

	expire := time.Now().Add(time.Hour * time.Duration(s.cfg.UserInviteTokenTimeout)).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		claimType:   pkg.RoleTypeMerchant,
		claimEmail:  req.Email,
		claimRoleId: role.Id,
		claimExpire: expire,
	})
	tokenString, err := token.SignedString([]byte(s.cfg.UserInviteTokenSecret))

	if err != nil {
		zap.L().Error(errorUserUnableToCreateToken.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserUnableToCreateToken

		return nil
	}

	err = s.sendInviteEmail(req.Email, owner.Email, owner.FirstName, owner.LastName, merchant.Company.Name, tokenString, role.Role)

	if err != nil {
		zap.L().Error(
			errorUserUnableToSendInvite.Message,
			zap.Error(err),
			zap.String("receiverEmail", req.Email),
			zap.String("senderEmail", owner.Email),
			zap.String("senderFirstName", owner.FirstName),
			zap.String("senderLastName", owner.LastName),
			zap.String("senderCompany", merchant.Company.Name),
			zap.String("token", tokenString),
		)
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserUnableToSendInvite

		return nil
	}

	res.Status = billingpb.ResponseStatusOk
	res.Role = role

	return nil
}

func (s *Service) InviteUserAdmin(
	ctx context.Context,
	req *billingpb.InviteUserAdminRequest,
	res *billingpb.InviteUserAdminResponse,
) error {
	owner, err := s.userRoleRepository.GetSystemAdmin(ctx)

	if err != nil {
		zap.L().Error(errorUserNotFound.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserNotFound

		return nil
	}

	user, err := s.userRoleRepository.GetAdminUserByEmail(ctx, req.Email)

	if (err != nil && err != mongo.ErrNoDocuments) || user != nil {
		zap.L().Error(errorUserAlreadyExist.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserAlreadyExist

		return nil
	}

	role := &billingpb.UserRole{
		Id:     primitive.NewObjectID().Hex(),
		Role:   req.Role,
		Status: pkg.UserRoleStatusInvited,
		Email:  req.Email,
	}

	if err = s.userRoleRepository.AddAdminUser(ctx, role); err != nil {
		zap.L().Error(errorUserUnableToAdd.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserUnableToAdd

		return nil
	}

	expire := time.Now().Add(time.Hour * time.Duration(s.cfg.UserInviteTokenTimeout)).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		claimType:   pkg.RoleTypeSystem,
		claimEmail:  req.Email,
		claimRoleId: role.Id,
		claimExpire: expire,
	})
	tokenString, err := token.SignedString([]byte(s.cfg.UserInviteTokenSecret))

	if err != nil {
		zap.L().Error(errorUserUnableToCreateToken.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserUnableToCreateToken

		return nil
	}

	err = s.sendInviteEmail(req.Email, owner.Email, owner.FirstName, owner.LastName, defaultCompanyName, tokenString, "")

	if err != nil {
		zap.L().Error(
			errorUserUnableToSendInvite.Message,
			zap.Error(err),
			zap.String("receiverEmail", req.Email),
			zap.String("senderEmail", owner.Email),
			zap.String("senderFirstName", owner.FirstName),
			zap.String("senderLastName", owner.LastName),
			zap.String("token", tokenString),
		)
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserUnableToSendInvite

		return nil
	}

	res.Status = billingpb.ResponseStatusOk
	res.Role = role

	return nil
}

func (s *Service) ResendInviteMerchant(
	ctx context.Context,
	req *billingpb.ResendInviteMerchantRequest,
	res *billingpb.EmptyResponseWithStatus,
) error {
	merchant, err := s.merchantRepository.GetById(ctx, req.MerchantId)

	if err != nil {
		zap.L().Error(errorUserMerchantNotFound.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserMerchantNotFound

		return nil
	}

	owner, err := s.userRoleRepository.GetMerchantOwner(ctx, merchant.Id)

	if err != nil {
		zap.L().Error(errorUserNotFound.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserNotFound

		return nil
	}

	role, err := s.userRoleRepository.GetMerchantUserByEmail(ctx, merchant.Id, req.Email)

	if err != nil {
		zap.L().Error(errorUserNotFound.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserNotFound

		return nil
	}

	if role.Status != pkg.UserRoleStatusInvited {
		zap.L().Error(errorUserUnableResendInvite.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserUnableResendInvite

		return nil
	}

	token, err := s.createInviteToken(role)

	if err != nil {
		zap.L().Error(errorUserUnableToCreateToken.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserUnableToCreateToken

		return nil
	}

	err = s.sendInviteEmail(role.Email, owner.Email, owner.FirstName, owner.LastName, merchant.Company.Name, token, role.Role)

	if err != nil {
		zap.L().Error(
			errorUserUnableToSendInvite.Message,
			zap.Error(err),
			zap.String("receiverEmail", role.Email),
			zap.String("senderEmail", owner.Email),
			zap.String("senderFirstName", owner.FirstName),
			zap.String("senderLastName", owner.LastName),
			zap.String("senderCompany", merchant.Company.Name),
			zap.String("tokenString", token),
		)
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserUnableToSendInvite

		return nil
	}

	res.Status = billingpb.ResponseStatusOk

	return nil
}

func (s *Service) ResendInviteAdmin(
	ctx context.Context,
	req *billingpb.ResendInviteAdminRequest,
	res *billingpb.EmptyResponseWithStatus,
) error {
	owner, err := s.userRoleRepository.GetSystemAdmin(ctx)

	if err != nil {
		zap.L().Error(errorUserNotFound.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserNotFound

		return nil
	}

	role, err := s.userRoleRepository.GetAdminUserByEmail(ctx, req.Email)

	if err != nil {
		zap.L().Error(errorUserNotFound.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserNotFound

		return nil
	}

	token, err := s.createInviteToken(role)

	if err != nil {
		zap.L().Error(errorUserUnableToCreateToken.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserUnableToCreateToken

		return nil
	}

	err = s.sendInviteEmail(role.Email, owner.Email, owner.FirstName, owner.LastName, defaultCompanyName, token, "")

	if err != nil {
		zap.L().Error(
			errorUserUnableToSendInvite.Message,
			zap.Error(err),
			zap.String("receiverEmail", role.Email),
			zap.String("senderEmail", owner.Email),
			zap.String("senderFirstName", owner.FirstName),
			zap.String("senderLastName", owner.LastName),
			zap.String("tokenString", token),
		)
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserUnableToSendInvite

		return nil
	}

	res.Status = billingpb.ResponseStatusOk

	return nil
}

func (s *Service) AcceptInvite(
	ctx context.Context,
	req *billingpb.AcceptInviteRequest,
	res *billingpb.AcceptInviteResponse,
) error {
	claims, err := s.parseInviteToken(req.Token)

	if err != nil {
		zap.L().Error("Error on parse invite token", zap.Error(err), zap.String("token", req.Token), zap.String("email", req.Email))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserInvalidToken

		return nil
	}

	if claims[claimEmail] != req.Email {
		zap.L().Error(errorUserInvalidInviteEmail.Message, zap.String("token email", claims[claimEmail].(string)), zap.String("email", req.Email))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserInvalidInviteEmail

		return nil
	}

	profile, err := s.userProfileRepository.GetByUserId(ctx, req.UserId)

	if err != nil {
		zap.L().Error(errorUserProfileNotFound.Message, zap.Error(err), zap.String("user_id", req.UserId))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserProfileNotFound

		return nil
	}

	if profile.Personal == nil || profile.Personal.FirstName == "" || profile.Personal.LastName == "" {
		zap.L().Error(errorUserEmptyNames.Message, zap.Error(err), zap.String("user_id", req.UserId))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserEmptyNames

		return nil
	}

	var user *billingpb.UserRole

	switch claims[claimType] {
	case pkg.RoleTypeMerchant:
		user, err = s.userRoleRepository.GetMerchantUserById(ctx, claims[claimRoleId].(string))
		break
	case pkg.RoleTypeSystem:
		user, err = s.userRoleRepository.GetAdminUserById(ctx, claims[claimRoleId].(string))
		break
	default:
		err = errors.New(errorUserUnsupportedRoleType.Message)
	}

	if err != nil {
		zap.L().Error(errorUserNotFound.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserNotFound
		return nil
	}

	if user.Status != pkg.UserRoleStatusInvited {
		zap.L().Error(errorUserInviteAlreadyAccepted.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserInviteAlreadyAccepted
		return nil
	}

	user.Status = pkg.UserRoleStatusAccepted
	user.UserId = req.UserId
	user.FirstName = profile.Personal.FirstName
	user.LastName = profile.Personal.LastName

	switch claims[claimType] {
	case pkg.RoleTypeMerchant:
		err = s.userRoleRepository.UpdateMerchantUser(ctx, user)
		break
	case pkg.RoleTypeSystem:
		err = s.userRoleRepository.UpdateAdminUser(ctx, user)
		break
	default:
		err = errors.New(errorUserUnsupportedRoleType.Message)
	}

	if err != nil {
		zap.L().Error(errorUserUnableToAdd.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserUnableToAdd
		return nil
	}

	casbinUserId := user.UserId

	if claims[claimType] == pkg.RoleTypeMerchant {
		casbinUserId = fmt.Sprintf(pkg.CasbinMerchantUserMask, user.MerchantId, user.UserId)
	}

	_, err = s.casbinService.AddRoleForUser(ctx, &casbinProto.UserRoleRequest{
		User: casbinUserId,
		Role: user.Role,
	})

	if err != nil {
		zap.L().Error(errorUserUnableToAddToCasbin.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserUnableToAddToCasbin
		return nil
	}

	if err = s.emailConfirmedSuccessfully(ctx, profile); err != nil {
		zap.L().Error(errorUserConfirmEmail.Message, zap.Error(err), zap.Any("profile", profile))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserConfirmEmail
		return nil
	}

	res.Status = billingpb.ResponseStatusOk
	res.Role = user

	return nil
}

func (s *Service) CheckInviteToken(
	_ context.Context,
	req *billingpb.CheckInviteTokenRequest,
	res *billingpb.CheckInviteTokenResponse,
) error {
	claims, err := s.parseInviteToken(req.Token)

	if err != nil {
		zap.L().Error("Error on parse invite token", zap.Error(err), zap.String("token", req.Token), zap.String("email", req.Email))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserInvalidToken

		return nil
	}

	if claims[claimEmail] != req.Email {
		zap.L().Error(errorUserInvalidInviteEmail.Message, zap.String("token email", claims[claimEmail].(string)), zap.String("email", req.Email))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserInvalidInviteEmail

		return nil
	}

	res.Status = billingpb.ResponseStatusOk
	res.RoleId = claims[claimRoleId].(string)
	res.RoleType = claims[claimType].(string)

	return nil
}

func (s *Service) sendInviteEmail(
	receiverEmail,
	senderEmail,
	senderFirstName,
	senderLastName,
	senderCompany,
	token,
	roleKey string,
) error {
	roleName := ""
	ok := false

	if roleKey != "" {
		roleName, ok = merchantUserRolesNames[roleKey]

		if !ok {
			return errorUserUnsupportedRoleType
		}
	}

	payload := &postmarkpb.Payload{
		TemplateAlias: s.cfg.EmailTemplates.UserInvite,
		TemplateModel: map[string]string{
			"sender_first_name": senderFirstName,
			"sender_last_name":  senderLastName,
			"sender_email":      senderEmail,
			"sender_company":    senderCompany,
			"invite_link":       s.cfg.GetUserInviteUrl(token),
			"role_name":         roleName,
		},
		To: receiverEmail,
	}
	err := s.postmarkBroker.Publish(postmarkpb.PostmarkSenderTopicName, payload, amqp.Table{})

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ChangeRoleForMerchantUser(ctx context.Context, req *billingpb.ChangeRoleForMerchantUserRequest, res *billingpb.EmptyResponseWithStatus) error {
	if req.Role == billingpb.RoleMerchantOwner {
		zap.L().Error(errorUserUnsupportedRoleType.Message, zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserUnsupportedRoleType

		return nil
	}

	user, err := s.userRoleRepository.GetMerchantUserById(ctx, req.RoleId)

	if err != nil || user.MerchantId != req.MerchantId {
		zap.L().Error(errorUserNotFound.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserNotFound

		return nil
	}

	user.Role = req.Role
	err = s.userRoleRepository.UpdateMerchantUser(ctx, user)

	if err != nil {
		zap.L().Error(errorUserUnableToSave.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorUserUnableToSave

		return nil
	}

	if user.UserId != "" {
		casbinUserId := fmt.Sprintf(pkg.CasbinMerchantUserMask, user.MerchantId, user.UserId)

		_, err = s.casbinService.DeleteUser(ctx, &casbinProto.UserRoleRequest{User: casbinUserId})

		if err != nil {
			zap.L().Error(errorUserUnableToDeleteFromCasbin.Message, zap.Error(err), zap.Any("req", req))
			res.Status = billingpb.ResponseStatusSystemError
			res.Message = errorUserUnableToDeleteFromCasbin
			return nil
		}

		_, err = s.casbinService.AddRoleForUser(ctx, &casbinProto.UserRoleRequest{
			User: casbinUserId,
			Role: req.Role,
		})

		if err != nil {
			zap.L().Error(errorUserUnableToDeleteFromCasbin.Message, zap.Error(err), zap.Any("req", req))
			res.Status = billingpb.ResponseStatusSystemError
			res.Message = errorUserUnableToDeleteFromCasbin
			return nil
		}
	}

	res.Status = billingpb.ResponseStatusOk

	return nil
}

func (s *Service) ChangeRoleForAdminUser(ctx context.Context, req *billingpb.ChangeRoleForAdminUserRequest, res *billingpb.EmptyResponseWithStatus) error {
	user, err := s.userRoleRepository.GetAdminUserById(ctx, req.RoleId)

	if err != nil {
		zap.L().Error(errorUserNotFound.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserNotFound

		return nil
	}

	user.Role = req.Role
	err = s.userRoleRepository.UpdateAdminUser(ctx, user)

	if err != nil {
		zap.L().Error(errorUserUnableToSave.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusSystemError
		res.Message = errorUserUnableToSave
		res.Message.Details = err.Error()

		return nil
	}

	if user.UserId != "" {
		_, err = s.casbinService.DeleteUser(ctx, &casbinProto.UserRoleRequest{User: user.UserId})

		if err != nil {
			zap.L().Error(errorUserUnableToDeleteFromCasbin.Message, zap.Error(err), zap.Any("req", req))
			res.Status = billingpb.ResponseStatusSystemError
			res.Message = errorUserUnableToDeleteFromCasbin
			return nil
		}

		_, err = s.casbinService.AddRoleForUser(ctx, &casbinProto.UserRoleRequest{
			User: user.UserId,
			Role: req.Role,
		})

		if err != nil {
			zap.L().Error(errorUserUnableToDeleteFromCasbin.Message, zap.Error(err), zap.Any("req", req))
			res.Status = billingpb.ResponseStatusSystemError
			res.Message = errorUserUnableToDeleteFromCasbin
			return nil
		}
	}

	res.Status = billingpb.ResponseStatusOk

	return nil
}

func (s *Service) GetRoleList(_ context.Context, req *billingpb.GetRoleListRequest, res *billingpb.GetRoleListResponse) error {
	list, ok := merchantUserRoles[req.Type]

	if !ok {
		zap.L().Error("unsupported user type", zap.Any("req", req))
		return nil
	}

	res.Items = list

	return nil
}

func (s *Service) DeleteMerchantUser(
	ctx context.Context,
	req *billingpb.MerchantRoleRequest,
	res *billingpb.EmptyResponseWithStatus,
) error {
	user, err := s.userRoleRepository.GetMerchantUserById(ctx, req.RoleId)

	if err != nil || user.MerchantId != req.MerchantId {
		zap.L().Error(errorUserNotFound.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserNotFound

		return nil
	}

	if err = s.userRoleRepository.DeleteMerchantUser(ctx, user); err != nil {
		zap.L().Error(errorUserUnableToDelete.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserUnableToDelete

		return nil
	}

	if user.UserId != "" {
		_, err = s.casbinService.DeleteUser(ctx, &casbinProto.UserRoleRequest{
			User: fmt.Sprintf(pkg.CasbinMerchantUserMask, user.MerchantId, user.UserId),
		})

		if err != nil {
			zap.L().Error(errorUserUnableToDeleteFromCasbin.Message, zap.Error(err), zap.Any("req", req))
			res.Status = billingpb.ResponseStatusBadData
			res.Message = errorUserUnableToDeleteFromCasbin
			return nil
		}
	}

	profile, err := s.userProfileRepository.GetByUserId(ctx, user.UserId)

	if profile != nil {
		if err = s.emailConfirmedTruncate(ctx, profile); err != nil {
			zap.L().Error(errorUserConfirmEmail.Message, zap.Error(err), zap.Any("profile", profile))
			res.Status = billingpb.ResponseStatusBadData
			res.Message = errorUserConfirmEmail
			return nil
		}
	}

	res.Status = billingpb.ResponseStatusOk

	return nil
}

func (s *Service) DeleteAdminUser(
	ctx context.Context,
	req *billingpb.AdminRoleRequest,
	res *billingpb.EmptyResponseWithStatus,
) error {
	user, err := s.userRoleRepository.GetAdminUserById(ctx, req.RoleId)

	if err != nil {
		zap.L().Error(errorUserNotFound.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserNotFound

		return nil
	}

	if err = s.userRoleRepository.DeleteAdminUser(ctx, user); err != nil {
		zap.L().Error(errorUserUnableToDelete.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserUnableToDelete

		return nil
	}

	if user.UserId != "" {
		_, err = s.casbinService.DeleteUser(ctx, &casbinProto.UserRoleRequest{User: user.UserId})

		if err != nil {
			zap.L().Error(errorUserUnableToDeleteFromCasbin.Message, zap.Error(err), zap.Any("req", req))
			res.Status = billingpb.ResponseStatusBadData
			res.Message = errorUserUnableToDeleteFromCasbin
			return nil
		}
	}

	profile, err := s.userProfileRepository.GetByUserId(ctx, user.UserId)

	if profile != nil {
		if err = s.emailConfirmedTruncate(ctx, profile); err != nil {
			zap.L().Error(errorUserConfirmEmail.Message, zap.Error(err), zap.Any("profile", profile))
			res.Status = billingpb.ResponseStatusBadData
			res.Message = errorUserConfirmEmail
			return nil
		}
	}

	res.Status = billingpb.ResponseStatusOk

	return nil
}

func (s *Service) GetMerchantUserRole(
	ctx context.Context,
	req *billingpb.MerchantRoleRequest,
	res *billingpb.UserRoleResponse,
) error {
	user, err := s.userRoleRepository.GetMerchantUserById(ctx, req.RoleId)

	if err != nil || user.MerchantId != req.MerchantId {
		zap.L().Error(errorUserNotFound.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserNotFound

		return nil
	}

	res.Status = billingpb.ResponseStatusOk
	res.UserRole = user

	return nil
}

func (s *Service) parseInviteToken(t string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(errorUserInvalidToken.Message)
		}

		return []byte(s.cfg.UserInviteTokenSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("token isn't valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return nil, errors.New("cannot read claims")
	}

	return claims, nil
}

func (s *Service) createInviteToken(role *billingpb.UserRole) (string, error) {
	roleType := pkg.RoleTypeMerchant
	if role.MerchantId == "" {
		roleType = pkg.RoleTypeSystem
	}

	expire := time.Now().Add(time.Hour * time.Duration(s.cfg.UserInviteTokenTimeout)).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		claimType:   roleType,
		claimEmail:  role.Email,
		claimRoleId: role.Id,
		claimExpire: expire,
	})

	return token.SignedString([]byte(s.cfg.UserInviteTokenSecret))
}

func (s *Service) getUserPermissions(ctx context.Context, userId string, merchantId string) ([]*billingpb.Permission, error) {
	id := userId

	if len(merchantId) > 0 {
		id = fmt.Sprintf(pkg.CasbinMerchantUserMask, merchantId, userId)
	}

	rsp, err := s.casbinService.GetImplicitPermissionsForUser(ctx, &casbinProto.PermissionRequest{User: id})

	if err != nil {
		zap.L().Error(errorUserGetImplicitPermissions.Message, zap.Error(err), zap.String("userId", id))
		return nil, errorUserGetImplicitPermissions
	}

	if len(rsp.D2) == 0 {
		zap.L().Error(errorUserDontHaveRole.Message, zap.String("userId", id))
		return nil, errorUserDontHaveRole
	}

	permissions := make([]*billingpb.Permission, len(rsp.D2))
	for i, p := range rsp.D2 {
		permissions[i] = &billingpb.Permission{
			Name:   p.D1[0],
			Access: p.D1[1],
		}
	}

	return permissions, nil
}

func (s *Service) GetAdminUserRole(
	ctx context.Context,
	req *billingpb.AdminRoleRequest,
	res *billingpb.UserRoleResponse,
) error {
	user, err := s.userRoleRepository.GetAdminUserById(ctx, req.RoleId)

	if err != nil {
		zap.L().Error(errorUserNotFound.Message, zap.Error(err), zap.Any("req", req))
		res.Status = billingpb.ResponseStatusBadData
		res.Message = errorUserNotFound

		return nil
	}

	res.Status = billingpb.ResponseStatusOk
	res.UserRole = user

	return nil
}

func (s *Service) GetAdminByUserId(
	ctx context.Context,
	req *billingpb.CommonUserProfileRequest,
	rsp *billingpb.UserRoleResponse,
) error {
	user, err := s.userRoleRepository.GetAdminUserByUserId(ctx, req.UserId)

	if err != nil {
		rsp.Status = billingpb.ResponseStatusNotFound
		rsp.Message = errorUserNotFound

		return nil
	}

	rsp.Status = billingpb.ResponseStatusOk
	rsp.UserRole = user

	return nil
}

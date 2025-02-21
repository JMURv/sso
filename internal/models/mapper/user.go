package mapper

import (
	"github.com/JMURv/sso/api/grpc/v1/gen"
	md "github.com/JMURv/sso/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ListModelToProto(u []*md.User) []*gen.SSO_User {
	res := make([]*gen.SSO_User, len(u))
	for i, v := range u {
		res[i] = ModelToProto(v)
	}
	return res
}

func ModelToProto(u *md.User) *gen.SSO_User {
	perms := make([]*gen.SSO_Permission, 0, len(u.Permissions))
	for _, v := range u.Permissions {
		perms = append(
			perms, &gen.SSO_Permission{
				Name:  v.Name,
				Value: v.Value,
			},
		)
	}

	return &gen.SSO_User{
		Id:          u.ID.String(),
		Name:        u.Name,
		Password:    u.Password,
		Email:       u.Email,
		Avatar:      u.Avatar,
		Address:     u.Address,
		Phone:       u.Phone,
		Permissions: perms,
		CreatedAt:   timestamppb.New(u.CreatedAt),
		UpdatedAt:   timestamppb.New(u.UpdatedAt),
	}
}

func ProtoToModel(u *gen.SSO_User) *md.User {
	perms := make([]md.Permission, 0, len(u.Permissions))
	for _, v := range u.Permissions {
		perms = append(
			perms, md.Permission{
				Name:  v.Name,
				Value: v.Value,
			},
		)
	}

	modelUser := &md.User{
		Name:        u.Name,
		Password:    u.Password,
		Email:       u.Email,
		Avatar:      u.Avatar,
		Address:     u.Address,
		Phone:       u.Phone,
		Permissions: perms,
		CreatedAt:   u.CreatedAt.AsTime(),
		UpdatedAt:   u.UpdatedAt.AsTime(),
	}

	uid, err := uuid.Parse(u.Id)
	if err != nil {
		zap.L().Debug("failed to parse user id")
	} else {
		modelUser.ID = uid
	}
	return modelUser
}

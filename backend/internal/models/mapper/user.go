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
	roles := make([]*gen.SSO_Role, 0, len(u.Roles))
	for _, v := range u.Roles {
		roles = append(roles, RoleToProto(&v))
	}
	return &gen.SSO_User{
		Id:        u.ID.String(),
		Name:      u.Name,
		Password:  u.Password,
		Email:     u.Email,
		Avatar:    u.Avatar,
		Roles:     roles,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}

func ProtoToModel(u *gen.SSO_User) *md.User {
	perms := make([]md.Role, 0, len(u.Roles))
	for _, v := range u.Roles {
		perms = append(perms, *RoleFromProto(v))
	}

	modelUser := &md.User{
		Name:      u.Name,
		Password:  u.Password,
		Email:     u.Email,
		Avatar:    u.Avatar,
		Roles:     perms,
		CreatedAt: u.CreatedAt.AsTime(),
		UpdatedAt: u.UpdatedAt.AsTime(),
	}

	uid, err := uuid.Parse(u.Id)
	if err != nil {
		zap.L().Warn("failed to parse user id")
	} else {
		modelUser.ID = uid
	}
	return modelUser
}

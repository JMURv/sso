package utils

import (
	pb "github.com/JMURv/sso/api/pb"
	md "github.com/JMURv/sso/pkg/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ListPermissionsToProto(req []*md.Permission) []*pb.SSO_Permission {
	res := make([]*pb.SSO_Permission, len(req))
	for i, v := range req {
		res[i] = &pb.SSO_Permission{
			Id:    v.ID,
			Name:  v.Name,
			Value: v.Value,
		}
	}
	return res
}

func PermissionToProto(req *md.Permission) *pb.SSO_Permission {
	return &pb.SSO_Permission{
		Id:    req.ID,
		Name:  req.Name,
		Value: req.Value,
	}
}

func PermissionFromProto(req *pb.SSO_Permission) *md.Permission {
	return &md.Permission{
		ID:    req.Id,
		Name:  req.Name,
		Value: req.Value,
	}
}

func ListModelToProto(u []*md.User) []*pb.SSO_User {
	res := make([]*pb.SSO_User, len(u))
	for i, v := range u {
		res[i] = ModelToProto(v)
	}
	return res
}

func ModelToProto(u *md.User) *pb.SSO_User {
	perms := make([]*pb.SSO_Permission, 0, len(u.Permissions))
	for _, v := range u.Permissions {
		perms = append(
			perms, &pb.SSO_Permission{
				Name:  v.Name,
				Value: v.Value,
			},
		)
	}

	return &pb.SSO_User{
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

func ProtoToModel(u *pb.SSO_User) *md.User {
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

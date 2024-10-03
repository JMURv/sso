package utils

import (
	pb "github.com/JMURv/sso/api/pb"
	md "github.com/JMURv/sso/pkg/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ListModelToProto(u []*md.User) []*pb.User {
	res := make([]*pb.User, len(u))
	for i, v := range u {
		res[i] = ModelToProto(v)
	}
	return res
}

func ModelToProto(u *md.User) *pb.User {
	return &pb.User{
		Id:        u.ID.String(),
		Name:      u.Name,
		Password:  u.Password,
		Email:     u.Email,
		Avatar:    u.Avatar,
		Address:   u.Address,
		Phone:     u.Phone,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}

func ProtoToModel(u *pb.User) *md.User {
	modelUser := &md.User{
		Name:      u.Name,
		Password:  u.Password,
		Email:     u.Email,
		Avatar:    u.Avatar,
		Address:   u.Address,
		Phone:     u.Phone,
		CreatedAt: u.CreatedAt.AsTime(),
		UpdatedAt: u.UpdatedAt.AsTime(),
	}

	uid, err := uuid.Parse(u.Id)
	if err != nil {
		zap.L().Debug("failed to parse user id")
	} else {
		modelUser.ID = uid
	}
	return modelUser
}

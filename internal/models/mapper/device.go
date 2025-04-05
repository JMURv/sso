package mapper

import (
	pb "github.com/JMURv/sso/api/grpc/v1/gen"
	md "github.com/JMURv/sso/internal/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ListDevicesToProto(req []md.Device) []*pb.SSO_Device {
	res := make([]*pb.SSO_Device, len(req))
	for i := 0; i < len(req); i++ {
		res[i] = DeviceToProto(&req[i])
	}
	return res
}

func DeviceToProto(d *md.Device) *pb.SSO_Device {
	return &pb.SSO_Device{
		Id:         d.ID,
		UserId:     d.UserID.String(),
		Name:       d.Name,
		DeviceType: d.DeviceType,
		Os:         d.OS,
		Browser:    d.Browser,
		Ua:         d.UA,
		Ip:         d.IP,
		LastActive: timestamppb.New(d.LastActive),
		CreatedAt:  timestamppb.New(d.CreatedAt),
	}
}

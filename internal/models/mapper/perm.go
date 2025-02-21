package mapper

import (
	"github.com/JMURv/sso/api/grpc/v1/gen"
	md "github.com/JMURv/sso/internal/models"
)

func ListPermissionsToProto(req []*md.Permission) []*gen.SSO_Permission {
	res := make([]*gen.SSO_Permission, len(req))
	for i, v := range req {
		res[i] = &gen.SSO_Permission{
			Id:    v.ID,
			Name:  v.Name,
			Value: v.Value,
		}
	}
	return res
}

func PermissionToProto(req *md.Permission) *gen.SSO_Permission {
	return &gen.SSO_Permission{
		Id:    req.ID,
		Name:  req.Name,
		Value: req.Value,
	}
}

func PermissionFromProto(req *gen.SSO_Permission) *md.Permission {
	return &md.Permission{
		ID:    req.Id,
		Name:  req.Name,
		Value: req.Value,
	}
}

package mapper

import (
	"github.com/JMURv/sso/api/grpc/v1/gen"
	md "github.com/JMURv/sso/internal/models"
)

func ListRolesToProto(req []md.Role) []*gen.SSO_Role {
	res := make([]*gen.SSO_Role, len(req))
	for i, v := range req {
		res[i] = RoleToProto(&v)
	}
	return res
}

func ListRolesToProtoFromPointer(req []*md.Role) []*gen.SSO_Role {
	res := make([]*gen.SSO_Role, len(req))
	for i, v := range req {
		res[i] = RoleToProto(v)
	}
	return res
}

func RoleToProto(req *md.Role) *gen.SSO_Role {
	return &gen.SSO_Role{
		Id:          req.ID,
		Name:        req.Name,
		Description: req.Description,
	}
}

func RoleFromProto(req *gen.SSO_Role) *md.Role {
	return &md.Role{
		ID:          req.Id,
		Name:        req.Name,
		Description: req.Description,
	}
}

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.1
// source: api/pb/sso.proto

package users

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	SSO_SendLoginCode_FullMethodName            = "/user.SSO/SendLoginCode"
	SSO_CheckLoginCode_FullMethodName           = "/user.SSO/CheckLoginCode"
	SSO_CheckEmail_FullMethodName               = "/user.SSO/CheckEmail"
	SSO_Logout_FullMethodName                   = "/user.SSO/Logout"
	SSO_SendForgotPasswordEmail_FullMethodName  = "/user.SSO/SendForgotPasswordEmail"
	SSO_CheckForgotPasswordEmail_FullMethodName = "/user.SSO/CheckForgotPasswordEmail"
	SSO_SendSupportEmail_FullMethodName         = "/user.SSO/SendSupportEmail"
	SSO_Me_FullMethodName                       = "/user.SSO/Me"
	SSO_UpdateMe_FullMethodName                 = "/user.SSO/UpdateMe"
)

// SSOClient is the client API for SSO service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SSOClient interface {
	SendLoginCode(ctx context.Context, in *SendLoginCodeReq, opts ...grpc.CallOption) (*Empty, error)
	CheckLoginCode(ctx context.Context, in *CheckLoginCodeReq, opts ...grpc.CallOption) (*CheckLoginCodeRes, error)
	CheckEmail(ctx context.Context, in *EmailMsg, opts ...grpc.CallOption) (*CheckEmailRes, error)
	Logout(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error)
	SendForgotPasswordEmail(ctx context.Context, in *EmailMsg, opts ...grpc.CallOption) (*Empty, error)
	CheckForgotPasswordEmail(ctx context.Context, in *CheckForgotPasswordEmailReq, opts ...grpc.CallOption) (*Empty, error)
	SendSupportEmail(ctx context.Context, in *SendSupportEmailReq, opts ...grpc.CallOption) (*Empty, error)
	Me(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*User, error)
	UpdateMe(ctx context.Context, in *User, opts ...grpc.CallOption) (*User, error)
}

type sSOClient struct {
	cc grpc.ClientConnInterface
}

func NewSSOClient(cc grpc.ClientConnInterface) SSOClient {
	return &sSOClient{cc}
}

func (c *sSOClient) SendLoginCode(ctx context.Context, in *SendLoginCodeReq, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, SSO_SendLoginCode_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sSOClient) CheckLoginCode(ctx context.Context, in *CheckLoginCodeReq, opts ...grpc.CallOption) (*CheckLoginCodeRes, error) {
	out := new(CheckLoginCodeRes)
	err := c.cc.Invoke(ctx, SSO_CheckLoginCode_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sSOClient) CheckEmail(ctx context.Context, in *EmailMsg, opts ...grpc.CallOption) (*CheckEmailRes, error) {
	out := new(CheckEmailRes)
	err := c.cc.Invoke(ctx, SSO_CheckEmail_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sSOClient) Logout(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, SSO_Logout_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sSOClient) SendForgotPasswordEmail(ctx context.Context, in *EmailMsg, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, SSO_SendForgotPasswordEmail_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sSOClient) CheckForgotPasswordEmail(ctx context.Context, in *CheckForgotPasswordEmailReq, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, SSO_CheckForgotPasswordEmail_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sSOClient) SendSupportEmail(ctx context.Context, in *SendSupportEmailReq, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, SSO_SendSupportEmail_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sSOClient) Me(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*User, error) {
	out := new(User)
	err := c.cc.Invoke(ctx, SSO_Me_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sSOClient) UpdateMe(ctx context.Context, in *User, opts ...grpc.CallOption) (*User, error) {
	out := new(User)
	err := c.cc.Invoke(ctx, SSO_UpdateMe_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SSOServer is the server API for SSO service.
// All implementations must embed UnimplementedSSOServer
// for forward compatibility
type SSOServer interface {
	SendLoginCode(context.Context, *SendLoginCodeReq) (*Empty, error)
	CheckLoginCode(context.Context, *CheckLoginCodeReq) (*CheckLoginCodeRes, error)
	CheckEmail(context.Context, *EmailMsg) (*CheckEmailRes, error)
	Logout(context.Context, *Empty) (*Empty, error)
	SendForgotPasswordEmail(context.Context, *EmailMsg) (*Empty, error)
	CheckForgotPasswordEmail(context.Context, *CheckForgotPasswordEmailReq) (*Empty, error)
	SendSupportEmail(context.Context, *SendSupportEmailReq) (*Empty, error)
	Me(context.Context, *Empty) (*User, error)
	UpdateMe(context.Context, *User) (*User, error)
	mustEmbedUnimplementedSSOServer()
}

// UnimplementedSSOServer must be embedded to have forward compatible implementations.
type UnimplementedSSOServer struct {
}

func (UnimplementedSSOServer) SendLoginCode(context.Context, *SendLoginCodeReq) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendLoginCode not implemented")
}
func (UnimplementedSSOServer) CheckLoginCode(context.Context, *CheckLoginCodeReq) (*CheckLoginCodeRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckLoginCode not implemented")
}
func (UnimplementedSSOServer) CheckEmail(context.Context, *EmailMsg) (*CheckEmailRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckEmail not implemented")
}
func (UnimplementedSSOServer) Logout(context.Context, *Empty) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Logout not implemented")
}
func (UnimplementedSSOServer) SendForgotPasswordEmail(context.Context, *EmailMsg) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendForgotPasswordEmail not implemented")
}
func (UnimplementedSSOServer) CheckForgotPasswordEmail(context.Context, *CheckForgotPasswordEmailReq) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckForgotPasswordEmail not implemented")
}
func (UnimplementedSSOServer) SendSupportEmail(context.Context, *SendSupportEmailReq) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendSupportEmail not implemented")
}
func (UnimplementedSSOServer) Me(context.Context, *Empty) (*User, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Me not implemented")
}
func (UnimplementedSSOServer) UpdateMe(context.Context, *User) (*User, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateMe not implemented")
}
func (UnimplementedSSOServer) mustEmbedUnimplementedSSOServer() {}

// UnsafeSSOServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SSOServer will
// result in compilation errors.
type UnsafeSSOServer interface {
	mustEmbedUnimplementedSSOServer()
}

func RegisterSSOServer(s grpc.ServiceRegistrar, srv SSOServer) {
	s.RegisterService(&SSO_ServiceDesc, srv)
}

func _SSO_SendLoginCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendLoginCodeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SSOServer).SendLoginCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SSO_SendLoginCode_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SSOServer).SendLoginCode(ctx, req.(*SendLoginCodeReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _SSO_CheckLoginCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckLoginCodeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SSOServer).CheckLoginCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SSO_CheckLoginCode_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SSOServer).CheckLoginCode(ctx, req.(*CheckLoginCodeReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _SSO_CheckEmail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmailMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SSOServer).CheckEmail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SSO_CheckEmail_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SSOServer).CheckEmail(ctx, req.(*EmailMsg))
	}
	return interceptor(ctx, in, info, handler)
}

func _SSO_Logout_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SSOServer).Logout(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SSO_Logout_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SSOServer).Logout(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _SSO_SendForgotPasswordEmail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmailMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SSOServer).SendForgotPasswordEmail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SSO_SendForgotPasswordEmail_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SSOServer).SendForgotPasswordEmail(ctx, req.(*EmailMsg))
	}
	return interceptor(ctx, in, info, handler)
}

func _SSO_CheckForgotPasswordEmail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckForgotPasswordEmailReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SSOServer).CheckForgotPasswordEmail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SSO_CheckForgotPasswordEmail_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SSOServer).CheckForgotPasswordEmail(ctx, req.(*CheckForgotPasswordEmailReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _SSO_SendSupportEmail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendSupportEmailReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SSOServer).SendSupportEmail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SSO_SendSupportEmail_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SSOServer).SendSupportEmail(ctx, req.(*SendSupportEmailReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _SSO_Me_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SSOServer).Me(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SSO_Me_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SSOServer).Me(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _SSO_UpdateMe_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(User)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SSOServer).UpdateMe(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SSO_UpdateMe_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SSOServer).UpdateMe(ctx, req.(*User))
	}
	return interceptor(ctx, in, info, handler)
}

// SSO_ServiceDesc is the grpc.ServiceDesc for SSO service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SSO_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "user.SSO",
	HandlerType: (*SSOServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendLoginCode",
			Handler:    _SSO_SendLoginCode_Handler,
		},
		{
			MethodName: "CheckLoginCode",
			Handler:    _SSO_CheckLoginCode_Handler,
		},
		{
			MethodName: "CheckEmail",
			Handler:    _SSO_CheckEmail_Handler,
		},
		{
			MethodName: "Logout",
			Handler:    _SSO_Logout_Handler,
		},
		{
			MethodName: "SendForgotPasswordEmail",
			Handler:    _SSO_SendForgotPasswordEmail_Handler,
		},
		{
			MethodName: "CheckForgotPasswordEmail",
			Handler:    _SSO_CheckForgotPasswordEmail_Handler,
		},
		{
			MethodName: "SendSupportEmail",
			Handler:    _SSO_SendSupportEmail_Handler,
		},
		{
			MethodName: "Me",
			Handler:    _SSO_Me_Handler,
		},
		{
			MethodName: "UpdateMe",
			Handler:    _SSO_UpdateMe_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/pb/sso.proto",
}

const (
	Users_UserSearch_FullMethodName = "/user.Users/UserSearch"
	Users_ListUsers_FullMethodName  = "/user.Users/ListUsers"
	Users_Register_FullMethodName   = "/user.Users/Register"
	Users_GetUser_FullMethodName    = "/user.Users/GetUser"
	Users_UpdateUser_FullMethodName = "/user.Users/UpdateUser"
	Users_DeleteUser_FullMethodName = "/user.Users/DeleteUser"
)

// UsersClient is the client API for Users service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UsersClient interface {
	UserSearch(ctx context.Context, in *UserSearchReq, opts ...grpc.CallOption) (*PaginatedUsersRes, error)
	ListUsers(ctx context.Context, in *ListUsersReq, opts ...grpc.CallOption) (*PaginatedUsersRes, error)
	Register(ctx context.Context, in *RegisterReq, opts ...grpc.CallOption) (*RegisterRes, error)
	GetUser(ctx context.Context, in *UuidMsg, opts ...grpc.CallOption) (*User, error)
	UpdateUser(ctx context.Context, in *UserWithUid, opts ...grpc.CallOption) (*User, error)
	DeleteUser(ctx context.Context, in *UuidMsg, opts ...grpc.CallOption) (*Empty, error)
}

type usersClient struct {
	cc grpc.ClientConnInterface
}

func NewUsersClient(cc grpc.ClientConnInterface) UsersClient {
	return &usersClient{cc}
}

func (c *usersClient) UserSearch(ctx context.Context, in *UserSearchReq, opts ...grpc.CallOption) (*PaginatedUsersRes, error) {
	out := new(PaginatedUsersRes)
	err := c.cc.Invoke(ctx, Users_UserSearch_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usersClient) ListUsers(ctx context.Context, in *ListUsersReq, opts ...grpc.CallOption) (*PaginatedUsersRes, error) {
	out := new(PaginatedUsersRes)
	err := c.cc.Invoke(ctx, Users_ListUsers_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usersClient) Register(ctx context.Context, in *RegisterReq, opts ...grpc.CallOption) (*RegisterRes, error) {
	out := new(RegisterRes)
	err := c.cc.Invoke(ctx, Users_Register_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usersClient) GetUser(ctx context.Context, in *UuidMsg, opts ...grpc.CallOption) (*User, error) {
	out := new(User)
	err := c.cc.Invoke(ctx, Users_GetUser_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usersClient) UpdateUser(ctx context.Context, in *UserWithUid, opts ...grpc.CallOption) (*User, error) {
	out := new(User)
	err := c.cc.Invoke(ctx, Users_UpdateUser_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usersClient) DeleteUser(ctx context.Context, in *UuidMsg, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, Users_DeleteUser_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UsersServer is the server API for Users service.
// All implementations must embed UnimplementedUsersServer
// for forward compatibility
type UsersServer interface {
	UserSearch(context.Context, *UserSearchReq) (*PaginatedUsersRes, error)
	ListUsers(context.Context, *ListUsersReq) (*PaginatedUsersRes, error)
	Register(context.Context, *RegisterReq) (*RegisterRes, error)
	GetUser(context.Context, *UuidMsg) (*User, error)
	UpdateUser(context.Context, *UserWithUid) (*User, error)
	DeleteUser(context.Context, *UuidMsg) (*Empty, error)
	mustEmbedUnimplementedUsersServer()
}

// UnimplementedUsersServer must be embedded to have forward compatible implementations.
type UnimplementedUsersServer struct {
}

func (UnimplementedUsersServer) UserSearch(context.Context, *UserSearchReq) (*PaginatedUsersRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UserSearch not implemented")
}
func (UnimplementedUsersServer) ListUsers(context.Context, *ListUsersReq) (*PaginatedUsersRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListUsers not implemented")
}
func (UnimplementedUsersServer) Register(context.Context, *RegisterReq) (*RegisterRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedUsersServer) GetUser(context.Context, *UuidMsg) (*User, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUser not implemented")
}
func (UnimplementedUsersServer) UpdateUser(context.Context, *UserWithUid) (*User, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUser not implemented")
}
func (UnimplementedUsersServer) DeleteUser(context.Context, *UuidMsg) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUser not implemented")
}
func (UnimplementedUsersServer) mustEmbedUnimplementedUsersServer() {}

// UnsafeUsersServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UsersServer will
// result in compilation errors.
type UnsafeUsersServer interface {
	mustEmbedUnimplementedUsersServer()
}

func RegisterUsersServer(s grpc.ServiceRegistrar, srv UsersServer) {
	s.RegisterService(&Users_ServiceDesc, srv)
}

func _Users_UserSearch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserSearchReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersServer).UserSearch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Users_UserSearch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersServer).UserSearch(ctx, req.(*UserSearchReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Users_ListUsers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListUsersReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersServer).ListUsers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Users_ListUsers_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersServer).ListUsers(ctx, req.(*ListUsersReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Users_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Users_Register_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersServer).Register(ctx, req.(*RegisterReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Users_GetUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UuidMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersServer).GetUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Users_GetUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersServer).GetUser(ctx, req.(*UuidMsg))
	}
	return interceptor(ctx, in, info, handler)
}

func _Users_UpdateUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserWithUid)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersServer).UpdateUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Users_UpdateUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersServer).UpdateUser(ctx, req.(*UserWithUid))
	}
	return interceptor(ctx, in, info, handler)
}

func _Users_DeleteUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UuidMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersServer).DeleteUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Users_DeleteUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersServer).DeleteUser(ctx, req.(*UuidMsg))
	}
	return interceptor(ctx, in, info, handler)
}

// Users_ServiceDesc is the grpc.ServiceDesc for Users service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Users_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "user.Users",
	HandlerType: (*UsersServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "UserSearch",
			Handler:    _Users_UserSearch_Handler,
		},
		{
			MethodName: "ListUsers",
			Handler:    _Users_ListUsers_Handler,
		},
		{
			MethodName: "Register",
			Handler:    _Users_Register_Handler,
		},
		{
			MethodName: "GetUser",
			Handler:    _Users_GetUser_Handler,
		},
		{
			MethodName: "UpdateUser",
			Handler:    _Users_UpdateUser_Handler,
		},
		{
			MethodName: "DeleteUser",
			Handler:    _Users_DeleteUser_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/pb/sso.proto",
}
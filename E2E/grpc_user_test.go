package tests

import (
	"context"
	"fmt"
	"github.com/JMURv/sso/E2E/helpers"
	pb "github.com/JMURv/sso/api/pb"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/cache/redis"
	ctrl "github.com/JMURv/sso/internal/controller"
	hdl "github.com/JMURv/sso/internal/handler/grpc"
	"github.com/JMURv/sso/internal/repository/db"
	"github.com/JMURv/sso/internal/smtp"
	"github.com/JMURv/sso/internal/validation"
	md "github.com/JMURv/sso/pkg/model"
	utils "github.com/JMURv/sso/pkg/utils/grpc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"testing"
)

func init() {
	cache := redis.New(helpers.Conf.Redis)

	helpers.Conf.DB.Database += "_test"
	repo := db.New(helpers.Conf.DB)
	email := smtp.New(helpers.Conf.Email, helpers.Conf.Server)

	au := auth.New(helpers.Conf.Auth.Secret)
	svc := ctrl.New(au, repo, cache, email)
	h := hdl.New(au, svc)

	go h.Start(helpers.Conf.Server.Port)
}

func CreateUser() (*md.User, string) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%v", helpers.Conf.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewUsersClient(conn)

	resp, err := client.Register(context.Background(), &pb.RegisterReq{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "secret1234",
	})
	if err != nil {
		stat, _ := status.FromError(err)
		panic(stat.Message())
	}

	u := utils.ProtoToModel(resp.User)
	return u, resp.Access
}

func TestUserSearchGrpc(t *testing.T) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%v", helpers.Conf.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewUsersClient(conn)

	tests := []struct {
		q            string
		expectedCode codes.Code
	}{
		{
			q:            "jmurv",
			expectedCode: codes.OK,
		},
		{
			q:            "jmu",
			expectedCode: codes.OK,
		},
		{
			q:            "test",
			expectedCode: codes.OK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.q, func(t *testing.T) {
			resp, err := client.UserSearch(context.Background(), &pb.UserSearchReq{
				Query: tc.q,
				Page:  1,
				Size:  1,
			})
			if err != nil {
				stat, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, stat.Code())
			} else {
				assert.Equal(t, codes.OK, tc.expectedCode)
			}

			if tc.q == "jmurv" {
				assert.Equal(t, 1, len(resp.Data))
			}
		})
	}
}

func TestListUsersGrpc(t *testing.T) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%v", helpers.Conf.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewUsersClient(conn)
	resp, err := client.ListUsers(context.Background(), &pb.ListUsersReq{
		Page: 1,
		Size: 10,
	})
	if err != nil {
		stat, _ := status.FromError(err)
		t.Log(stat.String())
	}

	assert.Equal(t, 2, len(resp.Data))
}

func TestGetUserGrpc(t *testing.T) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%v", helpers.Conf.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewUsersClient(conn)
	tests := []struct {
		name         string
		id           string
		expectedCode string
	}{
		{
			name:         "Invalid UUID",
			id:           "123",
			expectedCode: codes.InvalidArgument.String(),
		},
		{
			name:         "Not Found",
			id:           uuid.New().String(),
			expectedCode: codes.NotFound.String(),
		},
		{
			name:         "Success",
			id:           "16d6ef09-4c97-4bf6-92dc-e4291d75170c",
			expectedCode: codes.OK.String(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := client.GetUser(context.Background(), &pb.UuidMsg{
				Uuid: tc.id,
			})
			if err != nil {
				stat, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, stat.Code().String())
			} else {
				assert.Equal(t, tc.expectedCode, codes.OK.String())
				assert.Equal(t, "Владимир", resp.Name)
			}
		})
	}
}

func TestUserRegistrationGrpc(t *testing.T) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%v", helpers.Conf.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewUsersClient(conn)
	defer helpers.CleanDB(t)

	tests := []struct {
		name         string
		userData     pb.RegisterReq
		expectedCode string
		expectedMsg  string
		filePath     string
	}{
		{
			name: "Successful Registration",
			userData: pb.RegisterReq{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "secret1234",
			},
			expectedCode: codes.OK.String(),
			expectedMsg:  "",
		},
		{
			name: "Missing Email",
			userData: pb.RegisterReq{
				Name:     "John Doe",
				Password: "secret123",
			},
			expectedCode: codes.InvalidArgument.String(),
			expectedMsg:  validation.ErrMissingEmail.Error(),
		},
		{
			name: "Invalid Email Format",
			userData: pb.RegisterReq{
				Name:     "John Doe",
				Email:    "invalid-email",
				Password: "secret123",
			},
			expectedCode: codes.InvalidArgument.String(),
			expectedMsg:  validation.ErrInvalidEmail.Error(),
		},
		{
			name: "Short Password",
			userData: pb.RegisterReq{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "short",
			},
			expectedCode: codes.InvalidArgument.String(),
			expectedMsg:  validation.ErrPassTooShort.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := client.Register(context.Background(), &pb.RegisterReq{
				Name:     tc.userData.Name,
				Email:    tc.userData.Email,
				Password: tc.userData.Password,
			})
			if err != nil {
				stat, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, stat.Code().String())
				assert.Equal(t, tc.expectedMsg, stat.Message())
			} else {
				assert.Equal(t, tc.expectedCode, codes.OK.String())

				assert.Equal(t, tc.userData.Name, resp.User.Name)
				assert.Equal(t, tc.userData.Email, resp.User.Email)
				assert.NotEqual(t, tc.userData.Password, resp.User.Password)
				assert.NotEqual(t, "", resp.User.Id)

				assert.NotNil(t, resp.Access, "Expected access cookie to be set")
				assert.NotEqual(t, "", resp.Access, "Expected access cookie to have a value")

				assert.NotNil(t, resp.Refresh, "Expected refresh cookie to be set")
				assert.NotEqual(t, "", resp.Refresh, "Expected refresh cookie to have a value")
			}
		})
	}
}

func TestUpdateUsersGrpc(t *testing.T) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%v", helpers.Conf.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewUsersClient(conn)
	defer helpers.CleanDB(t)

	user, access := CreateUser()
	tests := []struct {
		name         string
		access       string
		expectedCode string
		expectedMsg  string
		userData     *pb.UserWithUid
	}{
		{
			name:         "Unauthenticated",
			access:       "",
			expectedCode: codes.Unauthenticated.String(),
			expectedMsg:  ctrl.ErrUnauthorized.Error(),
			userData: &pb.UserWithUid{
				Uid: "123",
				User: &pb.User{
					Email: "john@example.com",
				},
			},
		},
		{
			name:         "Invalid UUID",
			access:       access,
			expectedCode: codes.InvalidArgument.String(),
			expectedMsg:  ctrl.ErrParseUUID.Error(),
			userData: &pb.UserWithUid{
				Uid: "123",
				User: &pb.User{
					Email: "john@example.com",
				},
			},
		},
		{
			name:         "Missing Name",
			access:       access,
			expectedCode: codes.InvalidArgument.String(),
			expectedMsg:  validation.ErrMissingName.Error(),
			userData: &pb.UserWithUid{
				Uid: user.ID.String(),
				User: &pb.User{
					Email: "john@example.com",
				},
			},
		},
		{
			name:         "Missing Email",
			access:       access,
			expectedCode: codes.InvalidArgument.String(),
			expectedMsg:  validation.ErrMissingEmail.Error(),
			userData: &pb.UserWithUid{
				Uid: user.ID.String(),
				User: &pb.User{
					Name: "John Doe",
				},
			},
		},
		{
			name:         "Not Found",
			access:       access,
			expectedCode: codes.NotFound.String(),
			expectedMsg:  ctrl.ErrNotFound.Error(),
			userData: &pb.UserWithUid{
				Uid: uuid.New().String(),
				User: &pb.User{
					Name:  "John Doe",
					Email: "john@example.com",
				},
			},
		},
		{
			name:         "Valid",
			access:       access,
			expectedCode: codes.OK.String(),
			userData: &pb.UserWithUid{
				Uid: user.ID.String(),
				User: &pb.User{
					Name:  "John Doe UPDATED",
					Email: "johnUPDATED@example.com",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := metadata.NewOutgoingContext(
				context.Background(),
				metadata.Pairs(
					"authorization", fmt.Sprintf("Bearer %s", tc.access),
				),
			)

			resp, err := client.UpdateUser(ctx, tc.userData)
			if err != nil {
				stat, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, stat.Code().String())
				assert.Equal(t, tc.expectedMsg, stat.Message())
			} else {
				assert.Equal(t, tc.expectedCode, codes.OK.String())

				assert.Equal(t, tc.userData.User.Name, resp.Name)
				assert.Equal(t, tc.userData.User.Email, resp.Email)
				assert.NotEqual(t, tc.userData.User.Password, resp.Password)
				assert.NotEqual(t, user.Name, resp.Name)
				assert.NotEqual(t, user.Email, resp.Email)
				assert.NotEqual(t, "", resp.Id)
			}
		})
	}
}

func TestDeleteUserGrpc(t *testing.T) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%v", helpers.Conf.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.FailNow()
	}
	defer conn.Close()

	client := pb.NewUsersClient(conn)
	defer helpers.CleanDB(t)

	user, access := CreateUser()
	tests := []struct {
		name         string
		id           string
		access       string
		expectedCode string
		expectedMsg  string
	}{
		{
			name:         "Invalid UUID",
			id:           "123",
			access:       access,
			expectedCode: codes.InvalidArgument.String(),
			expectedMsg:  ctrl.ErrParseUUID.Error(),
		},
		{
			name:         "Invalid Token",
			id:           user.ID.String(),
			access:       "",
			expectedCode: codes.Unauthenticated.String(),
			expectedMsg:  ctrl.ErrUnauthorized.Error(),
		},
		{
			name:         "Valid",
			id:           user.ID.String(),
			access:       access,
			expectedCode: codes.OK.String(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mdt := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tc.access))
			ctx := metadata.NewOutgoingContext(context.Background(), mdt)

			_, err := client.DeleteUser(ctx, &pb.UuidMsg{
				Uuid: tc.id,
			})
			if err != nil {
				stat, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, stat.Code().String())
				assert.Equal(t, tc.expectedMsg, stat.Message())
			} else {
				assert.Equal(t, tc.expectedCode, codes.OK.String())
			}
		})
	}
}

// AUTH HERE

func TestSendLoginCode(t *testing.T) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%v", helpers.Conf.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	defer helpers.CleanDB(t)

	client := pb.NewSSOClient(conn)

	user, _ := CreateUser()
	tests := []struct {
		name         string
		payload      *pb.SendLoginCodeReq
		expectedCode string
		expectedMsg  string
	}{
		{
			name: "Empty req",
			payload: &pb.SendLoginCodeReq{
				Email:    "",
				Password: "",
			},
			expectedCode: codes.InvalidArgument.String(),
			expectedMsg:  ctrl.ErrDecodeRequest.Error(),
		},
		{
			name: "Invalid Email",
			payload: &pb.SendLoginCodeReq{
				Email:    "invalid-email",
				Password: "test-password",
			},
			expectedCode: codes.InvalidArgument.String(),
			expectedMsg:  ctrl.ErrInvalidCredentials.Error(),
		},
		{
			name: "No password",
			payload: &pb.SendLoginCodeReq{
				Email: user.Email,
			},
			expectedCode: codes.InvalidArgument.String(),
			expectedMsg:  ctrl.ErrDecodeRequest.Error(),
		},
		{
			name: "Valid",
			payload: &pb.SendLoginCodeReq{
				Email:    user.Email,
				Password: "secret1234",
			},
			expectedCode: codes.OK.String(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.SendLoginCode(context.Background(), tc.payload)
			if err != nil {
				stat, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, stat.Code().String())
				assert.Equal(t, tc.expectedMsg, stat.Message())
			} else {
				assert.Equal(t, tc.expectedCode, codes.OK.String())
			}
		})
	}
}

func TestCheckEmailGrpc(t *testing.T) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%v", helpers.Conf.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	defer helpers.CleanDB(t)

	client := pb.NewSSOClient(conn)

	user, _ := CreateUser()
	tests := []struct {
		name         string
		expectedCode string
		expectedMsg  string
		expectedRes  bool
		emailData    *pb.EmailMsg
	}{
		{
			name:         "InvalidArgument",
			expectedCode: codes.InvalidArgument.String(),
			expectedMsg:  ctrl.ErrDecodeRequest.Error(),
			emailData: &pb.EmailMsg{
				Email: "",
			},
		},
		{
			name:         "Valid - true",
			expectedCode: codes.OK.String(),
			expectedRes:  true,
			emailData: &pb.EmailMsg{
				Email: user.Email,
			},
		},
		{
			name:         "Valid - false",
			expectedCode: codes.OK.String(),
			expectedRes:  false,
			emailData: &pb.EmailMsg{
				Email: user.Email[:len(user.Email)-1],
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := client.CheckEmail(context.Background(), tc.emailData)
			if err != nil {
				stat, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, stat.Code().String())
				assert.Equal(t, tc.expectedMsg, stat.Message())
			} else {
				assert.Equal(t, tc.expectedCode, codes.OK.String())
				assert.Equal(t, tc.expectedRes, resp.IsExist)
			}
		})
	}
}

func TestLogoutGrpc(t *testing.T) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%v", helpers.Conf.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.FailNow()
	}
	defer conn.Close()

	client := pb.NewSSOClient(conn)
	defer helpers.CleanDB(t)

	_, access := CreateUser()
	tests := []struct {
		name         string
		access       string
		expectedCode string
		expectedMsg  string
	}{
		{
			name:         "Unauthenticated",
			access:       "",
			expectedCode: codes.Unauthenticated.String(),
			expectedMsg:  ctrl.ErrUnauthorized.Error(),
		},
		{
			name:         "Valid",
			access:       access,
			expectedCode: codes.OK.String(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mdt := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tc.access))
			ctx := metadata.NewOutgoingContext(context.Background(), mdt)

			_, err := client.Logout(ctx, &pb.Empty{})
			if err != nil {
				stat, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, stat.Code().String())
				assert.Equal(t, tc.expectedMsg, stat.Message())
			} else {
				assert.Equal(t, tc.expectedCode, codes.OK.String())
			}
		})
	}
}

func TestSendForgotPasswordEmailGrpc(t *testing.T) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%v", helpers.Conf.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.FailNow()
	}
	defer conn.Close()

	client := pb.NewSSOClient(conn)
	defer helpers.CleanDB(t)

	u, _ := CreateUser()
	tests := []struct {
		name         string
		expectedCode string
		expectedMsg  string
		payload      *pb.EmailMsg
	}{
		{
			name:         "InvalidArgument",
			expectedCode: codes.InvalidArgument.String(),
			expectedMsg:  ctrl.ErrDecodeRequest.Error(),
			payload: &pb.EmailMsg{
				Email: "",
			},
		},
		{
			name:         "InvalidCredentials",
			expectedCode: codes.InvalidArgument.String(),
			expectedMsg:  ctrl.ErrInvalidCredentials.Error(),
			payload: &pb.EmailMsg{
				Email: "invalid-email@example.com",
			},
		},
		{
			name:         "Valid",
			expectedCode: codes.OK.String(),
			payload: &pb.EmailMsg{
				Email: u.Email,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.SendForgotPasswordEmail(context.Background(), tc.payload)
			if err != nil {
				stat, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, stat.Code().String())
				assert.Equal(t, tc.expectedMsg, stat.Message())
			} else {
				assert.Equal(t, tc.expectedCode, codes.OK.String())
			}
		})
	}
}

func TestSendSupportEmailGrpc(t *testing.T) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%v", helpers.Conf.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.FailNow()
	}
	defer conn.Close()

	client := pb.NewSSOClient(conn)
	defer helpers.CleanDB(t)

	_, access := CreateUser()
	tests := []struct {
		name         string
		access       string
		expectedCode string
		expectedMsg  string
		payload      *pb.SendSupportEmailReq
	}{
		{
			name:         "Invalid Token",
			access:       "",
			expectedCode: codes.Unauthenticated.String(),
			expectedMsg:  ctrl.ErrUnauthorized.Error(),
			payload: &pb.SendSupportEmailReq{
				Text:  "",
				Theme: "",
			},
		},
		{
			name:         "ErrDecodeRequest Text",
			access:       access,
			expectedCode: codes.InvalidArgument.String(),
			expectedMsg:  ctrl.ErrDecodeRequest.Error(),
			payload: &pb.SendSupportEmailReq{
				Text:  "",
				Theme: "test-theme",
			},
		},
		{
			name:         "ErrDecodeRequest Theme",
			access:       access,
			expectedCode: codes.InvalidArgument.String(),
			expectedMsg:  ctrl.ErrDecodeRequest.Error(),
			payload: &pb.SendSupportEmailReq{
				Text:  "test-text",
				Theme: "",
			},
		},
		{
			name:         "Valid",
			access:       access,
			expectedCode: codes.OK.String(),
			payload: &pb.SendSupportEmailReq{
				Text:  "test-text",
				Theme: "test-theme",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mdt := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tc.access))
			ctx := metadata.NewOutgoingContext(context.Background(), mdt)

			_, err := client.SendSupportEmail(ctx, tc.payload)
			if err != nil {
				stat, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, stat.Code().String())
				assert.Equal(t, tc.expectedMsg, stat.Message())
			} else {
				assert.Equal(t, tc.expectedCode, codes.OK.String())
			}
		})
	}
}

func TestMeGrpc(t *testing.T) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%v", helpers.Conf.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.FailNow()
	}
	defer conn.Close()

	client := pb.NewSSOClient(conn)
	defer helpers.CleanDB(t)

	_, access := CreateUser()
	tests := []struct {
		name         string
		access       string
		expectedCode string
		expectedMsg  string
		payload      *pb.Empty
	}{
		{
			name:         "Invalid Token",
			access:       "",
			expectedCode: codes.Unauthenticated.String(),
			expectedMsg:  ctrl.ErrUnauthorized.Error(),
			payload:      &pb.Empty{},
		},
		{
			name:         "Valid",
			access:       access,
			expectedCode: codes.OK.String(),
			payload:      &pb.Empty{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mdt := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tc.access))
			ctx := metadata.NewOutgoingContext(context.Background(), mdt)

			_, err := client.Me(ctx, tc.payload)
			if err != nil {
				stat, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, stat.Code().String())
				assert.Equal(t, tc.expectedMsg, stat.Message())
			} else {
				assert.Equal(t, tc.expectedCode, codes.OK.String())
			}
		})
	}
}

func TestUpdateMeGrpc(t *testing.T) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%v", helpers.Conf.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.FailNow()
	}
	defer conn.Close()

	client := pb.NewSSOClient(conn)
	defer helpers.CleanDB(t)

	u, access := CreateUser()
	tests := []struct {
		name         string
		access       string
		expectedCode string
		expectedMsg  string
		payload      *pb.User
	}{
		{
			name:         "Invalid Token",
			access:       "",
			expectedCode: codes.Unauthenticated.String(),
			expectedMsg:  ctrl.ErrUnauthorized.Error(),
			payload:      &pb.User{},
		},
		{
			name:         "ErrMissingName",
			access:       access,
			expectedCode: codes.InvalidArgument.String(),
			expectedMsg:  validation.ErrMissingName.Error(),
			payload: &pb.User{
				Name:  "",
				Email: "test-email@test.com",
			},
		},
		{
			name:         "ErrMissingEmail",
			access:       access,
			expectedCode: codes.InvalidArgument.String(),
			expectedMsg:  validation.ErrMissingEmail.Error(),
			payload: &pb.User{
				Name:  "new-name",
				Email: "",
			},
		},
		{
			name:         "ErrInvalidEmail",
			access:       access,
			expectedCode: codes.InvalidArgument.String(),
			expectedMsg:  validation.ErrInvalidEmail.Error(),
			payload: &pb.User{
				Name:  "new-name",
				Email: "invalid-email",
			},
		},
		{
			name:         "Valid",
			access:       access,
			expectedCode: codes.OK.String(),
			payload: &pb.User{
				Name:  "new-name",
				Email: "test-email@test.com",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mdt := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tc.access))
			ctx := metadata.NewOutgoingContext(context.Background(), mdt)

			resp, err := client.UpdateMe(ctx, tc.payload)
			if err != nil {
				stat, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, stat.Code().String())
				assert.Equal(t, tc.expectedMsg, stat.Message())
			} else {
				assert.Equal(t, tc.expectedCode, codes.OK.String())
				assert.NotEqual(t, u.Name, resp.Name)
				assert.NotEqual(t, u.Email, resp.Email)
				assert.Equal(t, tc.payload.Name, resp.Name)
				assert.Equal(t, tc.payload.Email, resp.Email)
			}
		})
	}
}

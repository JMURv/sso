package ctrl

import (
	"context"
	"errors"
	"fmt"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/cache"
	"github.com/JMURv/sso/internal/controller/mocks"
	repo "github.com/JMURv/sso/internal/repository"
	"github.com/JMURv/sso/pkg/consts"
	md "github.com/JMURv/sso/pkg/model"
	utils "github.com/JMURv/sso/pkg/utils/http"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)

func TestIsUserExist(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	authRepo := mocks.NewMockAuth(ctrlMock)
	mockRepo := mocks.NewMockappRepo(ctrlMock)
	mockCache := mocks.NewMockCacheRepo(ctrlMock)
	mockSMTP := mocks.NewMockSMTPRepo(ctrlMock)

	ctrl := New(authRepo, mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	email := "test@example.com"

	// Test case 1: User exists
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(&md.User{}, nil).Times(1)

	isExist, err := ctrl.IsUserExist(ctx, email)
	assert.Nil(t, err)
	assert.True(t, isExist)

	// Test case 2: User does not exist (ErrNotFound)
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(nil, repo.ErrNotFound).Times(1)

	isExist, err = ctrl.IsUserExist(ctx, email)
	assert.Nil(t, err)
	assert.False(t, isExist)

	// Test case 3: Repo error (other than ErrNotFound)
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(nil, errors.New("some repo error")).Times(1)

	isExist, err = ctrl.IsUserExist(ctx, email)
	assert.NotNil(t, err)
	assert.True(t, isExist)
}

func TestUserSearch(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	authRepo := mocks.NewMockAuth(ctrlMock)
	mockRepo := mocks.NewMockappRepo(ctrlMock)
	mockCache := mocks.NewMockCacheRepo(ctrlMock)
	mockSMTP := mocks.NewMockSMTPRepo(ctrlMock)

	ctrl := New(authRepo, mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	query := "test"
	page := 1
	size := 10
	expectedData := &utils.PaginatedData{}

	// Simulate cache hit
	mockCache.EXPECT().GetToStruct(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

	mockRepo.EXPECT().UserSearch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	data, err := ctrl.UserSearch(ctx, query, page, size)
	assert.Nil(t, err)
	assert.Equal(t, expectedData, data)

	// Simulate cache miss
	mockCache.EXPECT().GetToStruct(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("cache miss")).Times(1)
	mockRepo.EXPECT().UserSearch(gomock.Any(), query, page, size).Return(expectedData, nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

	data, err = ctrl.UserSearch(ctx, query, page, size)
	assert.Nil(t, err)
	assert.Equal(t, expectedData, data)

	// Simulate cache miss and repo returns error
	mockCache.EXPECT().GetToStruct(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("cache miss")).Times(1)
	mockRepo.EXPECT().UserSearch(gomock.Any(), query, page, size).Return(nil, repo.ErrNotFound).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	data, err = ctrl.UserSearch(ctx, query, page, size)
	assert.Nil(t, data)
	assert.Equal(t, repo.ErrNotFound, err)

	// Simulate cache miss and cache set failure
	mockCache.EXPECT().GetToStruct(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("cache miss")).Times(1)
	mockRepo.EXPECT().UserSearch(gomock.Any(), query, page, size).Return(expectedData, nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("cache set failure")).Times(1)

	data, err = ctrl.UserSearch(ctx, query, page, size)
	assert.Nil(t, err)
	assert.Equal(t, expectedData, data)
}

func TestListUsers(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	authRepo := mocks.NewMockAuth(ctrlMock)
	mockRepo := mocks.NewMockappRepo(ctrlMock)
	mockCache := mocks.NewMockCacheRepo(ctrlMock)
	mockSMTP := mocks.NewMockSMTPRepo(ctrlMock)

	ctrl := New(authRepo, mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	page := 1
	size := 10
	cacheKey := fmt.Sprintf(usersListKey, page, size)

	expectedData := &utils.PaginatedData{}

	// Test case 1: Cache hit
	mockCache.EXPECT().GetToStruct(gomock.Any(), cacheKey, gomock.Any()).DoAndReturn(
		func(_ context.Context, _ string, dest **utils.PaginatedData) error {
			*dest = expectedData
			return nil
		},
	).Times(1)

	data, err := ctrl.ListUsers(ctx, page, size)
	assert.Nil(t, err)
	assert.Equal(t, expectedData, data)

	// Test case 2: Cache miss, repo success, cache set success
	mockCache.EXPECT().GetToStruct(gomock.Any(), cacheKey, gomock.Any()).Return(errors.New("cache miss")).Times(1)
	mockRepo.EXPECT().ListUsers(gomock.Any(), page, size).Return(expectedData, nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), consts.DefaultCacheTime, cacheKey, gomock.Any()).Return(nil).Times(1)

	data, err = ctrl.ListUsers(ctx, page, size)
	assert.Nil(t, err)
	assert.Equal(t, expectedData, data)

	// Test case 3: Cache miss, repo returns ErrNotFound
	mockCache.EXPECT().GetToStruct(gomock.Any(), cacheKey, gomock.Any()).Return(errors.New("cache miss")).Times(1)
	mockRepo.EXPECT().ListUsers(gomock.Any(), page, size).Return(nil, repo.ErrNotFound).Times(1)

	data, err = ctrl.ListUsers(ctx, page, size)
	assert.Nil(t, data)
	assert.Equal(t, ErrNotFound, err)

	// Test case 4: Cache miss, repo error (other than ErrNotFound)
	mockCache.EXPECT().GetToStruct(gomock.Any(), cacheKey, gomock.Any()).Return(errors.New("cache miss")).Times(1)
	mockRepo.EXPECT().ListUsers(gomock.Any(), page, size).Return(nil, errors.New("some repo error")).Times(1)

	data, err = ctrl.ListUsers(ctx, page, size)
	assert.Nil(t, data)
	assert.NotNil(t, err)

	// Test case 5: Cache miss, repo success, cache set failure
	mockCache.EXPECT().GetToStruct(gomock.Any(), cacheKey, gomock.Any()).Return(errors.New("cache miss")).Times(1)
	mockRepo.EXPECT().ListUsers(gomock.Any(), page, size).Return(expectedData, nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), consts.DefaultCacheTime, cacheKey, gomock.Any()).Return(errors.New("cache set failure")).Times(1)

	data, err = ctrl.ListUsers(ctx, page, size)
	assert.Nil(t, err)
	assert.Equal(t, expectedData, data)
}

func TestGetUserByID(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	authRepo := mocks.NewMockAuth(ctrlMock)
	mockRepo := mocks.NewMockappRepo(ctrlMock)
	mockCache := mocks.NewMockCacheRepo(ctrlMock)
	mockSMTP := mocks.NewMockSMTPRepo(ctrlMock)

	ctrl := New(authRepo, mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	userID := uuid.New()
	cacheKey := fmt.Sprintf(userCacheKey, userID)
	expectedUser := &md.User{}

	// Test case 1: Cache hit
	mockCache.EXPECT().GetToStruct(gomock.Any(), cacheKey, gomock.Any()).DoAndReturn(
		func(_ context.Context, _ string, dest *md.User) error {
			*dest = *expectedUser
			return nil
		},
	).Times(1)

	user, err := ctrl.GetUserByID(ctx, userID)
	assert.Nil(t, err)
	assert.Equal(t, expectedUser, user)

	// Test case 2: Cache miss, repo success, cache set success
	mockCache.EXPECT().GetToStruct(gomock.Any(), cacheKey, gomock.Any()).Return(errors.New("cache miss")).Times(1)
	mockRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(expectedUser, nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), consts.DefaultCacheTime, cacheKey, gomock.Any()).Return(nil).Times(1)

	user, err = ctrl.GetUserByID(ctx, userID)
	assert.Nil(t, err)
	assert.Equal(t, expectedUser, user)

	// Test case 3: Cache miss, repo returns ErrNotFound
	mockCache.EXPECT().GetToStruct(gomock.Any(), cacheKey, gomock.Any()).Return(errors.New("cache miss")).Times(1)
	mockRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(nil, repo.ErrNotFound).Times(1)

	user, err = ctrl.GetUserByID(ctx, userID)
	assert.Nil(t, user)
	assert.Equal(t, ErrNotFound, err)

	// Test case 4: Cache miss, repo error (other than ErrNotFound)
	mockCache.EXPECT().GetToStruct(gomock.Any(), cacheKey, gomock.Any()).Return(errors.New("cache miss")).Times(1)
	mockRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(nil, errors.New("some repo error")).Times(1)

	user, err = ctrl.GetUserByID(ctx, userID)
	assert.Nil(t, user)
	assert.NotNil(t, err)

	// Test case 5: Cache miss, repo success, cache set failure
	mockCache.EXPECT().GetToStruct(gomock.Any(), cacheKey, gomock.Any()).Return(errors.New("cache miss")).Times(1)
	mockRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(expectedUser, nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), consts.DefaultCacheTime, cacheKey, gomock.Any()).Return(errors.New("cache set failure")).Times(1)

	user, err = ctrl.GetUserByID(ctx, userID)
	assert.Nil(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestGetUserByEmail(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	authRepo := mocks.NewMockAuth(ctrlMock)
	mockRepo := mocks.NewMockappRepo(ctrlMock)
	mockCache := mocks.NewMockCacheRepo(ctrlMock)
	mockSMTP := mocks.NewMockSMTPRepo(ctrlMock)

	ctrl := New(authRepo, mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	email := "test@example.com"
	cacheKey := fmt.Sprintf(userCacheKey, email)
	expectedUser := &md.User{}

	// Test case 1: Cache hit
	mockCache.EXPECT().GetToStruct(gomock.Any(), cacheKey, gomock.Any()).DoAndReturn(
		func(_ context.Context, _ string, dest *md.User) error {
			*dest = *expectedUser
			return nil
		},
	).Times(1)

	user, err := ctrl.GetUserByEmail(ctx, email)
	assert.Nil(t, err)
	assert.Equal(t, expectedUser, user)

	// Test case 2: Cache miss, repo success, cache set success
	mockCache.EXPECT().GetToStruct(gomock.Any(), cacheKey, gomock.Any()).Return(errors.New("cache miss")).Times(1)
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(expectedUser, nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), consts.DefaultCacheTime, cacheKey, gomock.Any()).Return(nil).Times(1)

	user, err = ctrl.GetUserByEmail(ctx, email)
	assert.Nil(t, err)
	assert.Equal(t, expectedUser, user)

	// Test case 3: Cache miss, repo returns ErrNotFound
	mockCache.EXPECT().GetToStruct(gomock.Any(), cacheKey, gomock.Any()).Return(errors.New("cache miss")).Times(1)
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(nil, repo.ErrNotFound).Times(1)

	user, err = ctrl.GetUserByEmail(ctx, email)
	assert.Nil(t, user)
	assert.Equal(t, ErrNotFound, err)

	// Test case 4: Cache miss, repo error (other than ErrNotFound)
	mockCache.EXPECT().GetToStruct(gomock.Any(), cacheKey, gomock.Any()).Return(errors.New("cache miss")).Times(1)
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(nil, errors.New("some repo error")).Times(1)

	user, err = ctrl.GetUserByEmail(ctx, email)
	assert.Nil(t, user)
	assert.NotNil(t, err)

	// Test case 5: Cache miss, repo success, cache set failure
	mockCache.EXPECT().GetToStruct(gomock.Any(), cacheKey, gomock.Any()).Return(errors.New("cache miss")).Times(1)
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(expectedUser, nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), consts.DefaultCacheTime, cacheKey, gomock.Any()).Return(errors.New("cache set failure")).Times(1)

	user, err = ctrl.GetUserByEmail(ctx, email)
	assert.Nil(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestCreateUser(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	authRepo := mocks.NewMockAuth(ctrlMock)
	mockRepo := mocks.NewMockappRepo(ctrlMock)
	mockCache := mocks.NewMockCacheRepo(ctrlMock)
	mockSMTP := mocks.NewMockSMTPRepo(ctrlMock)

	ctrl := New(authRepo, mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	user := &md.User{ID: uuid.New(), Email: "test@example.com"}
	fileName := "welcome.pdf"
	fileBytes := []byte("some file content")
	expectedAccessToken := "access-token"
	expectedRefreshToken := "refresh-token"

	mockCache.EXPECT().InvalidateKeysByPattern(gomock.Any(), "users-*").AnyTimes()

	// Test case 1: User already exists
	mockRepo.EXPECT().CreateUser(gomock.Any(), user).Return(repo.ErrAlreadyExists).Times(1)

	createdUser, accessToken, refreshToken, err := ctrl.CreateUser(ctx, user, "", nil)
	assert.Nil(t, createdUser)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	assert.Equal(t, ErrAlreadyExists, err)

	// Test case 2: Repo returns a different error
	mockRepo.EXPECT().CreateUser(gomock.Any(), user).Return(errors.New("repo error")).Times(1)

	createdUser, accessToken, refreshToken, err = ctrl.CreateUser(ctx, user, "", nil)
	assert.Nil(t, createdUser)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	assert.Equal(t, errors.New("repo error"), err)

	// Test case 3: Successful user creation, no file, cache set success
	mockRepo.EXPECT().CreateUser(gomock.Any(), user).Return(nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), consts.DefaultCacheTime, gomock.Any(), gomock.Any()).Return(nil).Times(1)

	mockSMTP.EXPECT().SendOptFile(gomock.Any(), user.Email, fileName, fileBytes).Times(0)

	authRepo.EXPECT().NewToken(user, auth.AccessTokenDuration).Return(expectedAccessToken, nil).Times(1)
	authRepo.EXPECT().NewToken(user, auth.RefreshTokenDuration).Return(expectedRefreshToken, nil).Times(1)

	createdUser, accessToken, refreshToken, err = ctrl.CreateUser(ctx, user, "", nil)
	assert.Equal(t, user, createdUser)
	assert.Equal(t, expectedAccessToken, accessToken)
	assert.Equal(t, expectedRefreshToken, refreshToken)
	assert.Nil(t, err)

	// Test case 4: User created, file sent, cache set failure
	mockRepo.EXPECT().CreateUser(gomock.Any(), user).Return(nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), consts.DefaultCacheTime, gomock.Any(), gomock.Any()).Return(errors.New("cache error")).Times(1)

	mockSMTP.EXPECT().SendOptFile(gomock.Any(), user.Email, fileName, fileBytes).Return(nil).Times(1)

	authRepo.EXPECT().NewToken(user, auth.AccessTokenDuration).Return(expectedAccessToken, nil).Times(1)
	authRepo.EXPECT().NewToken(user, auth.RefreshTokenDuration).Return(expectedRefreshToken, nil).Times(1)

	createdUser, accessToken, refreshToken, err = ctrl.CreateUser(ctx, user, fileName, fileBytes)
	assert.Equal(t, user, createdUser)
	assert.Equal(t, expectedAccessToken, accessToken)
	assert.Equal(t, expectedRefreshToken, refreshToken)
	assert.Nil(t, err)

	// Test case 5: Failed to generate access token
	mockRepo.EXPECT().CreateUser(gomock.Any(), user).Return(nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), consts.DefaultCacheTime, gomock.Any(), gomock.Any()).Return(nil).Times(1)

	mockSMTP.EXPECT().SendOptFile(gomock.Any(), user.Email, fileName, fileBytes).Times(0)

	authRepo.EXPECT().NewToken(user, auth.AccessTokenDuration).Return("", errors.New("token generation error")).Times(1)

	createdUser, accessToken, refreshToken, err = ctrl.CreateUser(ctx, user, "", nil)
	assert.Equal(t, user, createdUser)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	assert.Equal(t, ErrWhileGeneratingToken, err)
}

func TestUpdateUser(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	authRepo := mocks.NewMockAuth(ctrlMock)
	mockRepo := mocks.NewMockappRepo(ctrlMock)
	mockCache := mocks.NewMockCacheRepo(ctrlMock)
	mockSMTP := mocks.NewMockSMTPRepo(ctrlMock)

	ctrl := New(authRepo, mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	userID := uuid.New()
	newUserData := &md.User{ID: userID, Email: "updated@example.com"}
	updatedUser := &md.User{ID: userID, Email: "updated@example.com"}

	mockCache.EXPECT().InvalidateKeysByPattern(gomock.Any(), "users-*").AnyTimes()

	// Test case 1: User not found
	mockRepo.EXPECT().UpdateUser(gomock.Any(), userID, newUserData).Return(nil, repo.ErrNotFound).Times(1)

	user, err := ctrl.UpdateUser(ctx, userID, newUserData)
	assert.Nil(t, user)
	assert.Equal(t, ErrNotFound, err)

	// Test case 2: Repo returns an error
	mockRepo.EXPECT().UpdateUser(gomock.Any(), userID, newUserData).Return(nil, errors.New("repo error")).Times(1)

	user, err = ctrl.UpdateUser(ctx, userID, newUserData)
	assert.Nil(t, user)
	assert.Equal(t, errors.New("repo error"), err)

	// Test case 3: Successful update, cache set success
	mockRepo.EXPECT().UpdateUser(gomock.Any(), userID, newUserData).Return(updatedUser, nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), consts.DefaultCacheTime, fmt.Sprintf(userCacheKey, updatedUser.ID), gomock.Any()).Return(nil).Times(1)

	user, err = ctrl.UpdateUser(ctx, userID, newUserData)
	assert.Equal(t, updatedUser, user)
	assert.Nil(t, err)

	// Test case 4: Successful update, cache set failure
	mockRepo.EXPECT().UpdateUser(gomock.Any(), userID, newUserData).Return(updatedUser, nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), consts.DefaultCacheTime, fmt.Sprintf(userCacheKey, updatedUser.ID), gomock.Any()).Return(errors.New("cache error")).Times(1)

	user, err = ctrl.UpdateUser(ctx, userID, newUserData)
	assert.Equal(t, updatedUser, user)
	assert.Nil(t, err)
}

func TestDeleteUser(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	authRepo := mocks.NewMockAuth(ctrlMock)
	mockRepo := mocks.NewMockappRepo(ctrlMock)
	mockCache := mocks.NewMockCacheRepo(ctrlMock)
	mockSMTP := mocks.NewMockSMTPRepo(ctrlMock)

	ctrl := New(authRepo, mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	userID := uuid.New()

	// Test case 1: Simulate successful deletion
	mockRepo.EXPECT().DeleteUser(gomock.Any(), userID).Return(nil).Times(1)
	mockCache.EXPECT().Delete(gomock.Any(), fmt.Sprintf(userCacheKey, userID)).Return(nil).Times(1)
	mockCache.EXPECT().InvalidateKeysByPattern(gomock.Any(), gomock.Any()).AnyTimes()

	err := ctrl.DeleteUser(ctx, userID)
	assert.Nil(t, err)

	// Test case 2: Simulate deletion error
	mockRepo.EXPECT().DeleteUser(gomock.Any(), userID).Return(errors.New("delete error")).Times(1)

	err = ctrl.DeleteUser(ctx, userID)
	assert.NotNil(t, err)
	assert.Equal(t, "delete error", err.Error())

	// Test case 3: Simulate cache deletion error
	mockRepo.EXPECT().DeleteUser(gomock.Any(), userID).Return(nil).Times(1)
	mockCache.EXPECT().Delete(gomock.Any(), fmt.Sprintf(userCacheKey, userID)).Return(errors.New("cache delete error")).Times(1)

	err = ctrl.DeleteUser(ctx, userID)
	assert.Nil(t, err)
}

func TestSendSupportEmail(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	authRepo := mocks.NewMockAuth(ctrlMock)
	mockRepo := mocks.NewMockappRepo(ctrlMock)
	mockCache := mocks.NewMockCacheRepo(ctrlMock)
	mockSMTP := mocks.NewMockSMTPRepo(ctrlMock)

	ctrl := New(authRepo, mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	userID := uuid.New()
	theme := "Test Theme"
	text := "Test message"
	user := &md.User{ID: userID}

	// Simulate successful user retrieval and email sending
	mockRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(user, nil).Times(1)
	mockSMTP.EXPECT().SendSupportEmail(gomock.Any(), user, theme, text).Return(nil).Times(1)

	err := ctrl.SendSupportEmail(ctx, userID, theme, text)
	assert.Nil(t, err)

	// Simulate user not found
	mockRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(nil, repo.ErrNotFound).Times(1)

	err = ctrl.SendSupportEmail(ctx, userID, theme, text)
	assert.Equal(t, ErrNotFound, err)

	// Simulate error retrieving user
	mockRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(nil, errors.New("db error")).Times(1)

	err = ctrl.SendSupportEmail(ctx, userID, theme, text)
	assert.NotNil(t, err)
	assert.Equal(t, "db error", err.Error())

	// Simulate error sending email
	mockRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(user, nil).Times(1)
	mockSMTP.EXPECT().SendSupportEmail(gomock.Any(), user, theme, text).Return(errors.New("smtp error")).Times(1)

	err = ctrl.SendSupportEmail(ctx, userID, theme, text)
	assert.NotNil(t, err)
	assert.Equal(t, "smtp error", err.Error())
}

func TestCheckForgotPasswordEmail(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	authRepo := mocks.NewMockAuth(ctrlMock)
	mockRepo := mocks.NewMockappRepo(ctrlMock)
	mockCache := mocks.NewMockCacheRepo(ctrlMock)
	mockSMTP := mocks.NewMockSMTPRepo(ctrlMock)

	ctrl := New(authRepo, mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	userID := uuid.New()
	password := "newPassword"
	code := 1234

	// Simulate successful code retrieval and user update
	mockCache.EXPECT().GetCode(gomock.Any(), fmt.Sprintf(recoveryCacheKey, userID)).Return(code, nil).Times(1)
	user := &md.User{ID: userID}
	mockRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(user, nil).Times(1)

	// Simulate successful password generation
	mockRepo.EXPECT().UpdateUser(gomock.Any(), userID, user).Return(nil, nil).Times(1)
	mockCache.EXPECT().Delete(gomock.Any(), fmt.Sprintf(userCacheKey, userID)).Return(nil).Times(1)

	err := ctrl.CheckForgotPasswordEmail(ctx, password, userID, code)
	assert.Nil(t, err)

	// Simulate invalid code
	mockCache.EXPECT().GetCode(gomock.Any(), fmt.Sprintf(recoveryCacheKey, userID)).Return(9999, nil).Times(1)
	err = ctrl.CheckForgotPasswordEmail(ctx, password, userID, code)
	assert.Equal(t, ErrCodeIsNotValid, err)

	// Simulate user not found
	mockCache.EXPECT().GetCode(gomock.Any(), fmt.Sprintf(recoveryCacheKey, userID)).Return(code, nil).Times(1)
	mockRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(nil, repo.ErrNotFound).Times(1)
	err = ctrl.CheckForgotPasswordEmail(ctx, password, userID, code)
	assert.Equal(t, ErrNotFound, err)

	// Simulate error retrieving user
	mockCache.EXPECT().GetCode(gomock.Any(), fmt.Sprintf(recoveryCacheKey, userID)).Return(code, nil).Times(1)
	mockRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(nil, errors.New("db error")).Times(1)
	err = ctrl.CheckForgotPasswordEmail(ctx, password, userID, code)
	assert.NotNil(t, err)
	assert.Equal(t, "db error", err.Error())

	// Simulate password generation error
	mockCache.EXPECT().GetCode(gomock.Any(), fmt.Sprintf(recoveryCacheKey, userID)).Return(code, nil).Times(1)
	mockRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(user, nil).Times(1)
	mockRepo.EXPECT().UpdateUser(gomock.Any(), userID, user).Return(nil, errors.New("update error")).Times(1)
	err = ctrl.CheckForgotPasswordEmail(ctx, password, userID, code)
	assert.NotNil(t, err)
	assert.Equal(t, "update error", err.Error())

	// Simulate error deleting from cache
	mockCache.EXPECT().GetCode(gomock.Any(), fmt.Sprintf(recoveryCacheKey, userID)).Return(code, nil).Times(1)
	mockRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(user, nil).Times(1)
	mockRepo.EXPECT().UpdateUser(gomock.Any(), userID, user).Return(nil, nil).Times(1)
	mockCache.EXPECT().Delete(gomock.Any(), fmt.Sprintf(userCacheKey, userID)).Return(errors.New("cache delete error")).Times(1)

	err = ctrl.CheckForgotPasswordEmail(ctx, password, userID, code)
	assert.Nil(t, err)

}

func TestSendForgotPasswordEmail(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	authRepo := mocks.NewMockAuth(ctrlMock)
	mockRepo := mocks.NewMockappRepo(ctrlMock)
	mockCache := mocks.NewMockCacheRepo(ctrlMock)
	mockSMTP := mocks.NewMockSMTPRepo(ctrlMock)

	ctrl := New(authRepo, mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	email := "test@example.com"
	userID := uuid.New()
	user := &md.User{ID: userID, Email: email}

	// Simulate successful user retrieval and email sending
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(user, nil).Times(1)

	mockCache.EXPECT().Set(gomock.Any(), time.Minute*15, fmt.Sprintf(recoveryCacheKey, userID.String()), gomock.Any()).Return(nil).Times(1)
	mockSMTP.EXPECT().SendForgotPasswordEmail(gomock.Any(), gomock.Any(), userID.String(), email).Return(nil).Times(1)

	err := ctrl.SendForgotPasswordEmail(ctx, email)
	assert.Nil(t, err)

	// Simulate user not found
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(nil, repo.ErrNotFound).Times(1)

	err = ctrl.SendForgotPasswordEmail(ctx, email)
	assert.Equal(t, ErrInvalidCredentials, err)

	// Simulate error retrieving user
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(nil, errors.New("db error")).Times(1)

	err = ctrl.SendForgotPasswordEmail(ctx, email)
	assert.NotNil(t, err)
	assert.Equal(t, "db error", err.Error())

	// Simulate error setting code in cache
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(user, nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), time.Minute*15, fmt.Sprintf(recoveryCacheKey, userID.String()), gomock.Any()).Return(errors.New("cache error")).Times(1)

	err = ctrl.SendForgotPasswordEmail(ctx, email)
	assert.NotNil(t, err)
	assert.Equal(t, "cache error", err.Error())

	// Simulate error sending email
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(user, nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), time.Minute*15, fmt.Sprintf(recoveryCacheKey, userID.String()), gomock.Any()).Return(nil).Times(1)
	mockSMTP.EXPECT().SendForgotPasswordEmail(gomock.Any(), gomock.Any(), userID.String(), email).Return(errors.New("smtp error")).Times(1)

	err = ctrl.SendForgotPasswordEmail(ctx, email)
	assert.NotNil(t, err)
	assert.Equal(t, "smtp error", err.Error())
}

func TestSendLoginCode(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	authRepo := mocks.NewMockAuth(ctrlMock)
	mockRepo := mocks.NewMockappRepo(ctrlMock)
	mockCache := mocks.NewMockCacheRepo(ctrlMock)
	mockSMTP := mocks.NewMockSMTPRepo(ctrlMock)

	ctrl := New(authRepo, mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	email := "test@example.com"
	password := "correctPassword"
	userID := uuid.New()
	genPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.Nil(t, err)

	user := &md.User{ID: userID, Email: email, Password: string(genPass)}

	// Simulate successful user retrieval and email sending
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(user, nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), time.Minute*15, fmt.Sprintf(codeCacheKey, email), gomock.Any()).Return(nil).Times(1)
	mockSMTP.EXPECT().SendLoginEmail(gomock.Any(), gomock.Any(), email).Return(nil).Times(1)

	err = ctrl.SendLoginCode(ctx, email, password)
	assert.Nil(t, err)

	// Simulate invalid credentials (user not found)
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(nil, repo.ErrNotFound).Times(1)
	err = ctrl.SendLoginCode(ctx, email, password)
	assert.Equal(t, ErrInvalidCredentials, err)

	// Simulate invalid credentials (password mismatch)
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(user, nil).Times(1)
	err = ctrl.SendLoginCode(ctx, email, "wrongPassword")
	assert.Equal(t, ErrInvalidCredentials, err)

	// Simulate error retrieving user
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(nil, errors.New("db error")).Times(1)
	err = ctrl.SendLoginCode(ctx, email, password)
	assert.NotNil(t, err)
	assert.Equal(t, "db error", err.Error())

	// Simulate error setting code in cache
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(user, nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), time.Minute*15, fmt.Sprintf(codeCacheKey, email), gomock.Any()).Return(errors.New("cache error")).Times(1)
	err = ctrl.SendLoginCode(ctx, email, password)
	assert.NotNil(t, err)
	assert.Equal(t, "cache error", err.Error())

	// Simulate error sending email
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(user, nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), time.Minute*15, fmt.Sprintf(codeCacheKey, email), gomock.Any()).Return(nil).Times(1)
	mockSMTP.EXPECT().SendLoginEmail(gomock.Any(), gomock.Any(), email).Return(errors.New("smtp error")).Times(1)
	err = ctrl.SendLoginCode(ctx, email, password)
	assert.NotNil(t, err)
	assert.Equal(t, "smtp error", err.Error())
}

func TestCheckLoginCode(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	authRepo := mocks.NewMockAuth(ctrlMock)
	mockRepo := mocks.NewMockappRepo(ctrlMock)
	mockCache := mocks.NewMockCacheRepo(ctrlMock)
	mockSMTP := mocks.NewMockSMTPRepo(ctrlMock)

	ctrl := New(authRepo, mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	email := "test@example.com"
	code := 1234
	userID := uuid.New()
	user := &md.User{ID: userID, Email: email}

	// Simulate successful code retrieval and user lookup
	mockCache.EXPECT().GetCode(gomock.Any(), fmt.Sprintf(codeCacheKey, email)).Return(code, nil).Times(1)
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(user, nil).Times(1)

	// Simulate successful token generation
	authRepo.EXPECT().NewToken(user, auth.AccessTokenDuration).Return("accessToken", nil).Times(1)
	authRepo.EXPECT().NewToken(user, auth.RefreshTokenDuration).Return("refreshToken", nil).Times(1)

	accessToken, refreshToken, err := ctrl.CheckLoginCode(ctx, email, code)
	assert.Nil(t, err)
	assert.Equal(t, "accessToken", accessToken)
	assert.Equal(t, "refreshToken", refreshToken)

	// Simulate error retrieving code from cache
	mockCache.EXPECT().GetCode(gomock.Any(), fmt.Sprintf(codeCacheKey, email)).Return(0, cache.ErrNotFoundInCache).Times(1)
	accessToken, refreshToken, err = ctrl.CheckLoginCode(ctx, email, code)
	assert.Equal(t, ErrNotFound, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)

	// Simulate error retrieving user
	mockCache.EXPECT().GetCode(gomock.Any(), fmt.Sprintf(codeCacheKey, email)).Return(code, nil).Times(1)
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(nil, errors.New("db error")).Times(1)
	accessToken, refreshToken, err = ctrl.CheckLoginCode(ctx, email, code)
	assert.NotNil(t, err)
	assert.Equal(t, "db error", err.Error())

	// Simulate stored code mismatch
	mockCache.EXPECT().GetCode(gomock.Any(), fmt.Sprintf(codeCacheKey, email)).Return(9999, nil).Times(1)
	accessToken, refreshToken, err = ctrl.CheckLoginCode(ctx, email, code)
	assert.Equal(t, ErrCodeIsNotValid, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)

	// Simulate error generating access token
	mockCache.EXPECT().GetCode(gomock.Any(), fmt.Sprintf(codeCacheKey, email)).Return(code, nil).Times(1)
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(user, nil).Times(1)
	authRepo.EXPECT().NewToken(user, auth.AccessTokenDuration).Return("", errors.New("token error")).Times(1)
	accessToken, refreshToken, err = ctrl.CheckLoginCode(ctx, email, code)
	assert.NotNil(t, err)
	assert.Equal(t, ErrWhileGeneratingToken.Error(), err.Error())

	// Simulate error generating refresh token
	mockCache.EXPECT().GetCode(gomock.Any(), fmt.Sprintf(codeCacheKey, email)).Return(code, nil).Times(1)
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(user, nil).Times(1)
	authRepo.EXPECT().NewToken(user, auth.AccessTokenDuration).Return("accessToken", nil).Times(1)
	authRepo.EXPECT().NewToken(user, auth.RefreshTokenDuration).Return("", errors.New("token error")).Times(1)
	accessToken, refreshToken, err = ctrl.CheckLoginCode(ctx, email, code)
	assert.NotNil(t, err)
	assert.Equal(t, ErrWhileGeneratingToken.Error(), err.Error())
}
package ctrl

import (
	"context"
	"errors"
	"fmt"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/config"
	md "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo"
	"github.com/JMURv/sso/tests/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestIsUserExist(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mockRepo := mocks.NewMockAppRepo(mock)
	mockCache := mocks.NewMockCacheService(mock)
	mockSMTP := mocks.NewMockEmailService(mock)

	ctrl := New(mockRepo, mockCache, mockSMTP)

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
	mock := gomock.NewController(t)
	defer mock.Finish()

	mockRepo := mocks.NewMockAppRepo(mock)
	mockCache := mocks.NewMockCacheService(mock)
	mockSMTP := mocks.NewMockEmailService(mock)

	ctrl := New(mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	query := "test"
	page := 1
	size := 10
	expectedData := &md.PaginatedUser{}

	// Simulate cache hit
	mockCache.EXPECT().GetToStruct(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

	mockRepo.EXPECT().SearchUser(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	data, err := ctrl.SearchUser(ctx, query, page, size)
	assert.Nil(t, err)
	assert.Equal(t, expectedData, data)

	// Simulate cache miss
	mockCache.EXPECT().GetToStruct(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("cache miss")).Times(1)
	mockRepo.EXPECT().SearchUser(gomock.Any(), query, page, size).Return(expectedData, nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

	data, err = ctrl.SearchUser(ctx, query, page, size)
	assert.Nil(t, err)
	assert.Equal(t, expectedData, data)

	// Simulate cache miss and repo returns error
	mockCache.EXPECT().GetToStruct(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("cache miss")).Times(1)
	mockRepo.EXPECT().SearchUser(gomock.Any(), query, page, size).Return(nil, repo.ErrNotFound).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	data, err = ctrl.SearchUser(ctx, query, page, size)
	assert.Nil(t, data)
	assert.Equal(t, repo.ErrNotFound, err)

	// Simulate cache miss and cache set failure
	mockCache.EXPECT().GetToStruct(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("cache miss")).Times(1)
	mockRepo.EXPECT().SearchUser(gomock.Any(), query, page, size).Return(expectedData, nil).Times(1)
	mockCache.EXPECT().Set(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(errors.New("cache set failure")).Times(1)

	data, err = ctrl.SearchUser(ctx, query, page, size)
	assert.Nil(t, err)
	assert.Equal(t, expectedData, data)
}

func TestListUsers(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mockRepo := mocks.NewMockAppRepo(mock)
	mockCache := mocks.NewMockCacheService(mock)
	mockSMTP := mocks.NewMockEmailService(mock)

	ctrl := New(mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	page := 1
	size := 10
	cacheKey := fmt.Sprintf(usersListKey, page, size)

	expectedData := &md.PaginatedUser{}

	// Test case 1: Cache hit
	mockCache.EXPECT().GetToStruct(gomock.Any(), cacheKey, gomock.Any()).DoAndReturn(
		func(_ context.Context, _ string, dest **md.PaginatedUser) error {
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
	mockCache.EXPECT().Set(gomock.Any(), config.DefaultCacheTime, cacheKey, gomock.Any()).Return(nil).Times(1)

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
	mockCache.EXPECT().Set(
		gomock.Any(),
		config.DefaultCacheTime,
		cacheKey,
		gomock.Any(),
	).Return(errors.New("cache set failure")).Times(1)

	data, err = ctrl.ListUsers(ctx, page, size)
	assert.Nil(t, err)
	assert.Equal(t, expectedData, data)
}

func TestGetUserByID(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mockRepo := mocks.NewMockAppRepo(mock)
	mockCache := mocks.NewMockCacheService(mock)
	mockSMTP := mocks.NewMockEmailService(mock)

	ctrl := New(mockRepo, mockCache, mockSMTP)

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
	mockCache.EXPECT().Set(gomock.Any(), config.DefaultCacheTime, cacheKey, gomock.Any()).Return(nil).Times(1)

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
	mockCache.EXPECT().Set(
		gomock.Any(),
		config.DefaultCacheTime,
		cacheKey,
		gomock.Any(),
	).Return(errors.New("cache set failure")).Times(1)

	user, err = ctrl.GetUserByID(ctx, userID)
	assert.Nil(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestGetUserByEmail(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mockRepo := mocks.NewMockAppRepo(mock)
	mockCache := mocks.NewMockCacheService(mock)
	mockSMTP := mocks.NewMockEmailService(mock)

	ctrl := New(mockRepo, mockCache, mockSMTP)

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
	mockCache.EXPECT().Set(gomock.Any(), config.DefaultCacheTime, cacheKey, gomock.Any()).Return(nil).Times(1)

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
	mockCache.EXPECT().Set(
		gomock.Any(),
		config.DefaultCacheTime,
		cacheKey,
		gomock.Any(),
	).Return(errors.New("cache set failure")).Times(1)

	user, err = ctrl.GetUserByEmail(ctx, email)
	assert.Nil(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestCreateUser(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	authRepo := mocks.NewMockAuthService(mock)
	mockRepo := mocks.NewMockAppRepo(mock)
	mockCache := mocks.NewMockCacheService(mock)
	mockSMTP := mocks.NewMockEmailService(mock)

	ctrl := New(mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	idx := uuid.New()
	user := &md.User{ID: idx, Email: "test@example.com"}
	fileName := "welcome.pdf"
	fileBytes := []byte("some file content")
	expectedAccessToken := "access-token"
	expectedRefreshToken := "refresh-token"

	mockCache.EXPECT().InvalidateKeysByPattern(gomock.Any(), "users-*").AnyTimes()

	// Test case 1: User already exists
	mockRepo.EXPECT().CreateUser(gomock.Any(), user).Return(idx, repo.ErrAlreadyExists).Times(1)

	retID, accessToken, refreshToken, err := ctrl.CreateUser(ctx, user, "", nil)
	assert.Equal(t, uuid.Nil, retID)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	assert.Equal(t, ErrAlreadyExists, err)

	// Test case 2: Repo returns a different error
	mockRepo.EXPECT().CreateUser(gomock.Any(), user).Return(uuid.Nil, errors.New("repo error")).Times(1)

	retID, accessToken, refreshToken, err = ctrl.CreateUser(ctx, user, "", nil)
	assert.Equal(t, uuid.Nil, retID)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	assert.Equal(t, errors.New("repo error"), err)

	// Test case 3: Successful user creation, no file, cache set success
	mockRepo.EXPECT().CreateUser(gomock.Any(), user).Return(idx, nil).Times(1)
	mockCache.EXPECT().Set(gomock.Any(), config.DefaultCacheTime, gomock.Any(), gomock.Any()).Return(nil).Times(1)

	mockSMTP.EXPECT().SendOptFile(gomock.Any(), user.Email, fileName, fileBytes).Times(0)
	mockSMTP.EXPECT().SendUserCredentials(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	authRepo.EXPECT().NewToken(user, auth.AccessTokenDuration).Return(expectedAccessToken, nil).Times(1)
	authRepo.EXPECT().NewToken(user, auth.RefreshTokenDuration).Return(expectedRefreshToken, nil).Times(1)

	retID, accessToken, refreshToken, err = ctrl.CreateUser(ctx, user, "", nil)
	assert.Equal(t, idx, retID)
	assert.Equal(t, expectedAccessToken, accessToken)
	assert.Equal(t, expectedRefreshToken, refreshToken)
	assert.Nil(t, err)

	// Test case 4: User created, file sent, cache set failure
	mockRepo.EXPECT().CreateUser(gomock.Any(), user).Return(idx, nil).Times(1)
	mockCache.EXPECT().Set(
		gomock.Any(),
		config.DefaultCacheTime,
		gomock.Any(),
		gomock.Any(),
	).Return(errors.New("cache error")).Times(1)

	mockSMTP.EXPECT().SendOptFile(gomock.Any(), user.Email, fileName, fileBytes).Return(nil).Times(1)
	mockSMTP.EXPECT().SendUserCredentials(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	authRepo.EXPECT().NewToken(user, auth.AccessTokenDuration).Return(expectedAccessToken, nil).Times(1)
	authRepo.EXPECT().NewToken(user, auth.RefreshTokenDuration).Return(expectedRefreshToken, nil).Times(1)

	retID, accessToken, refreshToken, err = ctrl.CreateUser(ctx, user, fileName, fileBytes)
	assert.Equal(t, idx, retID)
	assert.Equal(t, expectedAccessToken, accessToken)
	assert.Equal(t, expectedRefreshToken, refreshToken)
	assert.Nil(t, err)

	t.Run(
		"SendOptFile Error", func(t *testing.T) {
			mockRepo.EXPECT().CreateUser(gomock.Any(), user).Return(idx, nil).Times(1)
			mockCache.EXPECT().Set(
				gomock.Any(),
				config.DefaultCacheTime,
				gomock.Any(),
				gomock.Any(),
			).Return(errors.New("cache error")).Times(1)

			mockSMTP.EXPECT().SendOptFile(
				gomock.Any(),
				user.Email,
				fileName,
				fileBytes,
			).Return(errors.New("file error")).Times(1)
			mockSMTP.EXPECT().SendUserCredentials(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

			authRepo.EXPECT().NewToken(user, auth.AccessTokenDuration).Return(expectedAccessToken, nil).Times(1)
			authRepo.EXPECT().NewToken(user, auth.RefreshTokenDuration).Return(expectedRefreshToken, nil).Times(1)

			retID, accessToken, refreshToken, err = ctrl.CreateUser(ctx, user, fileName, fileBytes)
			assert.Equal(t, idx, retID)
			assert.Equal(t, expectedAccessToken, accessToken)
			assert.Equal(t, expectedRefreshToken, refreshToken)
			assert.Nil(t, err)
		},
	)

	t.Run(
		"Failed to generate access token", func(t *testing.T) {
			mockRepo.EXPECT().CreateUser(gomock.Any(), user).Return(idx, nil).Times(1)
			mockCache.EXPECT().Set(
				gomock.Any(),
				config.DefaultCacheTime,
				gomock.Any(),
				gomock.Any(),
			).Return(nil).Times(1)

			mockSMTP.EXPECT().SendOptFile(gomock.Any(), user.Email, fileName, fileBytes).Times(0)

			authRepo.EXPECT().NewToken(user, auth.AccessTokenDuration).Return(
				"",
				errors.New("token generation error"),
			).Times(1)

			retID, accessToken, refreshToken, err = ctrl.CreateUser(ctx, user, "", nil)
			assert.Equal(t, idx, retID)
			assert.Empty(t, accessToken)
			assert.Empty(t, refreshToken)
			assert.Equal(t, ErrWhileGeneratingToken, err)
		},
	)

	t.Run(
		"Failed to generate refresh token", func(t *testing.T) {
			mockRepo.EXPECT().CreateUser(gomock.Any(), user).Return(idx, nil).Times(1)
			mockCache.EXPECT().Set(
				gomock.Any(),
				config.DefaultCacheTime,
				gomock.Any(),
				gomock.Any(),
			).Return(nil).Times(1)

			mockSMTP.EXPECT().SendOptFile(gomock.Any(), user.Email, fileName, fileBytes).Times(0)

			authRepo.EXPECT().NewToken(user, auth.AccessTokenDuration).Return("accessToken", nil).Times(1)
			authRepo.EXPECT().NewToken(user, auth.RefreshTokenDuration).Return(
				"",
				errors.New("token generation error"),
			).Times(1)

			retID, accessToken, refreshToken, err = ctrl.CreateUser(ctx, user, "", nil)
			assert.Equal(t, idx, retID)
			assert.Empty(t, accessToken)
			assert.Empty(t, refreshToken)
			assert.Equal(t, ErrWhileGeneratingToken, err)
		},
	)
}

func TestUpdateUser(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mockRepo := mocks.NewMockAppRepo(mock)
	mockCache := mocks.NewMockCacheService(mock)
	mockSMTP := mocks.NewMockEmailService(mock)

	ctrl := New(mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	userID := uuid.New()
	newUserData := &md.User{ID: userID, Email: "updated@example.com"}

	mockCache.EXPECT().InvalidateKeysByPattern(gomock.Any(), "users-*").AnyTimes()

	// Test case 1: User not found
	mockRepo.EXPECT().UpdateUser(gomock.Any(), userID, newUserData).Return(repo.ErrNotFound).Times(1)

	err := ctrl.UpdateUser(ctx, userID, newUserData)
	assert.Equal(t, ErrNotFound, err)

	// Test case 2: Repo returns an error
	mockRepo.EXPECT().UpdateUser(gomock.Any(), userID, newUserData).Return(errors.New("repo error")).Times(1)

	err = ctrl.UpdateUser(ctx, userID, newUserData)
	assert.Equal(t, errors.New("repo error"), err)

	// Test case 3: Successful update, cache set success
	mockRepo.EXPECT().UpdateUser(gomock.Any(), userID, newUserData).Return(nil).Times(1)
	mockCache.EXPECT().Delete(gomock.Any(), fmt.Sprintf(userCacheKey, userID)).Return(nil).Times(1)

	err = ctrl.UpdateUser(ctx, userID, newUserData)
	assert.Nil(t, err)

	// Test case 4: Successful update, cache set failure
	mockRepo.EXPECT().UpdateUser(gomock.Any(), userID, newUserData).Return(nil).Times(1)
	mockCache.EXPECT().Delete(
		gomock.Any(),
		fmt.Sprintf(userCacheKey, userID),
	).Return(errors.New("cache error")).Times(1)

	err = ctrl.UpdateUser(ctx, userID, newUserData)
	assert.Nil(t, err)
}

func TestDeleteUser(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mockRepo := mocks.NewMockAppRepo(mock)
	mockCache := mocks.NewMockCacheService(mock)
	mockSMTP := mocks.NewMockEmailService(mock)

	ctrl := New(mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	userID := uuid.New()

	t.Run(
		"Success", func(t *testing.T) {
			mockRepo.EXPECT().DeleteUser(gomock.Any(), userID).Return(nil).Times(1)
			mockCache.EXPECT().Delete(gomock.Any(), fmt.Sprintf(userCacheKey, userID)).Return(nil).Times(1)
			mockCache.EXPECT().InvalidateKeysByPattern(gomock.Any(), gomock.Any()).AnyTimes()

			err := ctrl.DeleteUser(ctx, userID)
			assert.Nil(t, err)
		},
	)

	t.Run(
		"repo.ErrNotFound", func(t *testing.T) {
			mockRepo.EXPECT().DeleteUser(gomock.Any(), userID).Return(repo.ErrNotFound).Times(1)

			err := ctrl.DeleteUser(ctx, userID)
			assert.NotNil(t, err)
			assert.Equal(t, repo.ErrNotFound.Error(), err.Error())
		},
	)

	t.Run(
		"Internal Repo Error", func(t *testing.T) {
			mockRepo.EXPECT().DeleteUser(gomock.Any(), userID).Return(errors.New("delete error")).Times(1)

			err := ctrl.DeleteUser(ctx, userID)
			assert.NotNil(t, err)
			assert.Equal(t, "delete error", err.Error())
		},
	)

	t.Run(
		"Internal Cache Error", func(t *testing.T) {
			mockRepo.EXPECT().DeleteUser(gomock.Any(), userID).Return(nil).Times(1)
			mockCache.EXPECT().Delete(
				gomock.Any(),
				fmt.Sprintf(userCacheKey, userID),
			).Return(errors.New("cache delete error")).Times(1)

			err := ctrl.DeleteUser(ctx, userID)
			assert.Nil(t, err)
		},
	)
}

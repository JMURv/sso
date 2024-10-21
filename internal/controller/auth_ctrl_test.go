package ctrl

import (
	"fmt"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/cache"
	repo "github.com/JMURv/sso/internal/repository"
	"github.com/JMURv/sso/mocks"
	md "github.com/JMURv/sso/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)

func TestSendSupportEmail(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	authRepo := mocks.NewMockAuth(ctrlMock)
	mockRepo := mocks.NewMockuserRepo(ctrlMock)
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
	mockRepo := mocks.NewMockuserRepo(ctrlMock)
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
	mockRepo.EXPECT().UpdateUser(gomock.Any(), userID, user).Return(nil).Times(1)
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
	mockRepo.EXPECT().UpdateUser(gomock.Any(), userID, user).Return(errors.New("update error")).Times(1)
	err = ctrl.CheckForgotPasswordEmail(ctx, password, userID, code)
	assert.NotNil(t, err)
	assert.Equal(t, "update error", err.Error())

	// Simulate error deleting from cache
	mockCache.EXPECT().GetCode(gomock.Any(), fmt.Sprintf(recoveryCacheKey, userID)).Return(code, nil).Times(1)
	mockRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(user, nil).Times(1)
	mockRepo.EXPECT().UpdateUser(gomock.Any(), userID, user).Return(nil).Times(1)
	mockCache.EXPECT().Delete(
		gomock.Any(),
		fmt.Sprintf(userCacheKey, userID),
	).Return(errors.New("cache delete error")).Times(1)

	err = ctrl.CheckForgotPasswordEmail(ctx, password, userID, code)
	assert.Nil(t, err)

}

func TestSendForgotPasswordEmail(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	authRepo := mocks.NewMockAuth(ctrlMock)
	mockRepo := mocks.NewMockuserRepo(ctrlMock)
	mockCache := mocks.NewMockCacheRepo(ctrlMock)
	mockSMTP := mocks.NewMockSMTPRepo(ctrlMock)

	ctrl := New(authRepo, mockRepo, mockCache, mockSMTP)

	ctx := context.Background()
	email := "test@example.com"
	userID := uuid.New()
	user := &md.User{ID: userID, Email: email}

	// Simulate successful user retrieval and email sending
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(user, nil).Times(1)

	mockCache.EXPECT().Set(
		gomock.Any(),
		time.Minute*15,
		fmt.Sprintf(recoveryCacheKey, userID.String()),
		gomock.Any(),
	).Return(nil).Times(1)
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
	mockCache.EXPECT().Set(
		gomock.Any(),
		time.Minute*15,
		fmt.Sprintf(recoveryCacheKey, userID.String()),
		gomock.Any(),
	).Return(errors.New("cache error")).Times(1)

	err = ctrl.SendForgotPasswordEmail(ctx, email)
	assert.NotNil(t, err)
	assert.Equal(t, "cache error", err.Error())

	// Simulate error sending email
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(user, nil).Times(1)
	mockCache.EXPECT().Set(
		gomock.Any(),
		time.Minute*15,
		fmt.Sprintf(recoveryCacheKey, userID.String()),
		gomock.Any(),
	).Return(nil).Times(1)
	mockSMTP.EXPECT().SendForgotPasswordEmail(
		gomock.Any(),
		gomock.Any(),
		userID.String(),
		email,
	).Return(errors.New("smtp error")).Times(1)

	err = ctrl.SendForgotPasswordEmail(ctx, email)
	assert.NotNil(t, err)
	assert.Equal(t, "smtp error", err.Error())
}

func TestSendLoginCode(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	authRepo := mocks.NewMockAuth(ctrlMock)
	mockRepo := mocks.NewMockuserRepo(ctrlMock)
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
	mockCache.EXPECT().Set(
		gomock.Any(),
		time.Minute*15,
		fmt.Sprintf(codeCacheKey, email),
		gomock.Any(),
	).Return(nil).Times(1)
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
	mockCache.EXPECT().Set(
		gomock.Any(),
		time.Minute*15,
		fmt.Sprintf(codeCacheKey, email),
		gomock.Any(),
	).Return(errors.New("cache error")).Times(1)
	err = ctrl.SendLoginCode(ctx, email, password)
	assert.NotNil(t, err)
	assert.Equal(t, "cache error", err.Error())

	// Simulate error sending email
	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(user, nil).Times(1)
	mockCache.EXPECT().Set(
		gomock.Any(),
		time.Minute*15,
		fmt.Sprintf(codeCacheKey, email),
		gomock.Any(),
	).Return(nil).Times(1)
	mockSMTP.EXPECT().SendLoginEmail(gomock.Any(), gomock.Any(), email).Return(errors.New("smtp error")).Times(1)
	err = ctrl.SendLoginCode(ctx, email, password)
	assert.NotNil(t, err)
	assert.Equal(t, "smtp error", err.Error())
}

func TestCheckLoginCode(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	authRepo := mocks.NewMockAuth(ctrlMock)
	mockRepo := mocks.NewMockuserRepo(ctrlMock)
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
	mockCache.EXPECT().GetCode(gomock.Any(), fmt.Sprintf(codeCacheKey, email)).Return(
		0,
		cache.ErrNotFoundInCache,
	).Times(1)
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

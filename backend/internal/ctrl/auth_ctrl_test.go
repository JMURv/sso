package ctrl

//import (
//	"context"
//	"errors"
//	"github.com/JMURv/sso/tests/mocks"
//	"github.com/google/uuid"
//	"github.com/stretchr/testify/assert"
//	"go.uber.org/mock/gomock"
//	"testing"
//)
//
//func TestValidateToken(t *testing.T) {
//	mock := gomock.NewController(t)
//	defer mock.Finish()
//
//	authRepo := mocks.NewMockAuthService(mock)
//	mockRepo := mocks.NewMockAppRepo(mock)
//	mockCache := mocks.NewMockCacheService(mock)
//	mockSMTP := mocks.NewMockEmailService(mock)
//
//	ctrl := New(authRepo, mockRepo, mockCache, mockSMTP)
//
//}

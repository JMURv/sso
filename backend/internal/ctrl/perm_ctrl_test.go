package ctrl

//import (
//	"context"
//	"errors"
//	"github.com/JMURv/sso/internal/repo"
//	"github.com/JMURv/sso/mocks"
//	"github.com/stretchr/testify/assert"
//	"go.uber.org/mock/gomock"
//	"testing"
//)
//
//var cacheErr = errors.New("cache error")
//var repoErr = errors.New("repo error")
//
//func TestController_ListPermissions(t *testing.T) {
//	mock := gomock.NewController(t)
//	defer mock.Finish()
//
//	ctx := context.Background()
//	page, size := 1, 40
//	mockAuth := mocks.NewMockAuthService(mock)
//	mockRepo := mocks.NewMockAppRepo(mock)
//	mockCache := mocks.NewMockCacheService(mock)
//	mockEmail := mocks.NewMockEmailService(mock)
//	ctrl := New(mockAuth, mockRepo, mockCache, mockEmail)
//
//	t.Run(
//		"cache.GetToStruct - Success", func(t *testing.T) {
//			mockCache.EXPECT().
//				GetToStruct(gomock.Any(), gomock.Any(), gomock.Any()).
//				Return(nil).
//				Times(1)
//
//			res, err := ctrl.ListPermissions(ctx, page, size)
//			assert.Nil(t, err)
//			assert.IsType(t, &model.PaginatedPermission{}, res)
//		},
//	)
//
//	t.Run(
//		"repo.ListPermissions - Internal Err", func(t *testing.T) {
//			mockCache.EXPECT().
//				GetToStruct(gomock.Any(), gomock.Any(), gomock.Any()).
//				Return(cacheErr).
//				Times(1)
//
//			mockRepo.EXPECT().
//				ListPermissions(gomock.Any(), page, size).
//				Return(nil, repoErr).
//				Times(1)
//
//			res, err := ctrl.ListPermissions(ctx, page, size)
//			assert.IsType(t, err, repoErr)
//			assert.Nil(t, res)
//		},
//	)
//
//	t.Run(
//		"cache.Set - Internal Err", func(t *testing.T) {
//			mockCache.EXPECT().
//				GetToStruct(gomock.Any(), gomock.Any(), gomock.Any()).
//				Return(cacheErr).
//				Times(1)
//
//			mockRepo.EXPECT().
//				ListPermissions(gomock.Any(), page, size).
//				Return(&model.PaginatedPermission{}, nil).
//				Times(1)
//
//			mockCache.EXPECT().
//				Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
//				Return(cacheErr).
//				Times(1)
//
//			res, err := ctrl.ListPermissions(ctx, page, size)
//			assert.Nil(t, err)
//			assert.NotNil(t, res)
//		},
//	)
//
//}
//
//func TestController_GetPermission(t *testing.T) {
//	mock := gomock.NewController(t)
//	defer mock.Finish()
//
//	ctx := context.Background()
//	id := uint64(1)
//	mockAuth := mocks.NewMockAuthService(mock)
//	mockRepo := mocks.NewMockAppRepo(mock)
//	mockCache := mocks.NewMockCacheService(mock)
//	mockEmail := mocks.NewMockEmailService(mock)
//	ctrl := New(mockAuth, mockRepo, mockCache, mockEmail)
//
//	t.Run(
//		"cache.GetToStruct - Success", func(t *testing.T) {
//			mockCache.EXPECT().
//				GetToStruct(gomock.Any(), gomock.Any(), gomock.Any()).
//				Return(nil).
//				Times(1)
//
//			res, err := ctrl.GetPermission(ctx, id)
//			assert.Nil(t, err)
//			assert.IsType(t, &model.Permission{}, res)
//		},
//	)
//
//	t.Run(
//		"repo.ListPermissions - ErrNotFound", func(t *testing.T) {
//			mockCache.EXPECT().
//				GetToStruct(gomock.Any(), gomock.Any(), gomock.Any()).
//				Return(cacheErr).
//				Times(1)
//
//			mockRepo.EXPECT().
//				GetPermission(gomock.Any(), id).
//				Return(nil, repo.ErrNotFound).
//				Times(1)
//
//			res, err := ctrl.GetPermission(ctx, id)
//			assert.IsType(t, err, repo.ErrNotFound)
//			assert.Nil(t, res)
//		},
//	)
//
//	t.Run(
//		"repo.ListPermissions - Internal Err", func(t *testing.T) {
//			mockCache.EXPECT().
//				GetToStruct(gomock.Any(), gomock.Any(), gomock.Any()).
//				Return(cacheErr).
//				Times(1)
//
//			mockRepo.EXPECT().
//				GetPermission(gomock.Any(), id).
//				Return(nil, repoErr).
//				Times(1)
//
//			res, err := ctrl.GetPermission(ctx, id)
//			assert.IsType(t, err, repoErr)
//			assert.Nil(t, res)
//		},
//	)
//
//	t.Run(
//		"cache.Set - Internal Err", func(t *testing.T) {
//			mockCache.EXPECT().
//				GetToStruct(gomock.Any(), gomock.Any(), gomock.Any()).
//				Return(cacheErr).
//				Times(1)
//
//			mockRepo.EXPECT().
//				GetPermission(gomock.Any(), id).
//				Return(&model.Permission{}, nil).
//				Times(1)
//
//			mockCache.EXPECT().
//				Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
//				Return(cacheErr).
//				Times(1)
//
//			res, err := ctrl.GetPermission(ctx, id)
//			assert.Nil(t, err)
//			assert.NotNil(t, res)
//		},
//	)
//
//}
//
//func TestController_CreatePerm(t *testing.T) {
//	mock := gomock.NewController(t)
//	defer mock.Finish()
//
//	ctx := context.Background()
//	mockAuth := mocks.NewMockAuthService(mock)
//	mockRepo := mocks.NewMockAppRepo(mock)
//	mockCache := mocks.NewMockCacheService(mock)
//	mockEmail := mocks.NewMockEmailService(mock)
//	ctrl := New(mockAuth, mockRepo, mockCache, mockEmail)
//
//	id := uint64(1)
//	zero := uint64(0)
//
//	perm := &model.Permission{Name: "test-perm"}
//
//	t.Run(
//		"repo.CreatePerm - ErrAlreadyExists", func(t *testing.T) {
//			mockRepo.EXPECT().
//				CreatePerm(gomock.Any(), perm).
//				Return(zero, repo.ErrAlreadyExists).
//				Times(1)
//
//			res, err := ctrl.CreatePerm(ctx, perm)
//			assert.Equal(t, zero, res)
//			assert.IsType(t, ErrAlreadyExists, err)
//		},
//	)
//
//	t.Run(
//		"repo.CreatePerm - Internal Err", func(t *testing.T) {
//			mockRepo.EXPECT().
//				CreatePerm(gomock.Any(), perm).
//				Return(zero, repoErr).
//				Times(1)
//
//			res, err := ctrl.CreatePerm(ctx, perm)
//			assert.Equal(t, zero, res)
//			assert.IsType(t, err, repoErr)
//		},
//	)
//
//	t.Run(
//		"Success", func(t *testing.T) {
//			mockCache.EXPECT().
//				InvalidateKeysByPattern(gomock.Any(), gomock.Any()).
//				AnyTimes()
//
//			mockRepo.EXPECT().
//				CreatePerm(gomock.Any(), perm).
//				Return(id, nil).
//				Times(1)
//
//			res, err := ctrl.CreatePerm(ctx, perm)
//			assert.Nil(t, err)
//			assert.Equal(t, id, res)
//		},
//	)
//
//}
//
//func TestController_UpdatePermission(t *testing.T) {
//	mock := gomock.NewController(t)
//	defer mock.Finish()
//
//	ctx := context.Background()
//	mockAuth := mocks.NewMockAuthService(mock)
//	mockRepo := mocks.NewMockAppRepo(mock)
//	mockCache := mocks.NewMockCacheService(mock)
//	mockEmail := mocks.NewMockEmailService(mock)
//	ctrl := New(mockAuth, mockRepo, mockCache, mockEmail)
//
//	id := uint64(1)
//	perm := &model.Permission{Name: "test-perm"}
//
//	t.Run(
//		"repo.UpdatePerm - ErrNotFound", func(t *testing.T) {
//			mockRepo.EXPECT().
//				UpdatePerm(gomock.Any(), id, perm).
//				Return(repo.ErrNotFound).
//				Times(1)
//
//			err := ctrl.UpdatePerm(ctx, id, perm)
//			assert.IsType(t, ErrNotFound, err)
//		},
//	)
//
//	t.Run(
//		"repo.UpdatePerm - Internal Err", func(t *testing.T) {
//			mockRepo.EXPECT().
//				UpdatePerm(gomock.Any(), id, perm).
//				Return(repoErr).
//				Times(1)
//
//			err := ctrl.UpdatePerm(ctx, id, perm)
//			assert.IsType(t, err, repoErr)
//		},
//	)
//
//	t.Run(
//		"cache.Delete - Internal Err", func(t *testing.T) {
//			mockRepo.EXPECT().
//				UpdatePerm(gomock.Any(), id, perm).
//				Return(repoErr).
//				Times(1)
//
//			mockCache.EXPECT().
//				Delete(gomock.Any(), gomock.Any()).
//				Return(cacheErr).
//				Times(1)
//
//			err := ctrl.UpdatePerm(ctx, id, perm)
//			assert.IsType(t, err, repoErr)
//		},
//	)
//
//	t.Run(
//		"Success", func(t *testing.T) {
//			mockCache.EXPECT().
//				InvalidateKeysByPattern(gomock.Any(), gomock.Any()).
//				AnyTimes()
//
//			mockRepo.EXPECT().
//				UpdatePerm(gomock.Any(), id, perm).
//				Return(nil).
//				Times(1)
//
//			err := ctrl.UpdatePerm(ctx, id, perm)
//			assert.Nil(t, err)
//		},
//	)
//}
//
//func TestController_DeletePermission(t *testing.T) {
//	mock := gomock.NewController(t)
//	defer mock.Finish()
//
//	ctx := context.Background()
//	mockAuth := mocks.NewMockAuthService(mock)
//	mockRepo := mocks.NewMockAppRepo(mock)
//	mockCache := mocks.NewMockCacheService(mock)
//	mockEmail := mocks.NewMockEmailService(mock)
//	ctrl := New(mockAuth, mockRepo, mockCache, mockEmail)
//
//	id := uint64(1)
//
//	t.Run(
//		"repo.DeletePerm - ErrNotFound", func(t *testing.T) {
//			mockRepo.EXPECT().
//				DeletePerm(gomock.Any(), id).
//				Return(repo.ErrNotFound).
//				Times(1)
//
//			err := ctrl.DeletePerm(ctx, id)
//			assert.IsType(t, ErrNotFound, err)
//		},
//	)
//
//	t.Run(
//		"repo.DeletePerm - Internal Err", func(t *testing.T) {
//			mockRepo.EXPECT().
//				DeletePerm(gomock.Any(), id).
//				Return(repoErr).
//				Times(1)
//
//			err := ctrl.DeletePerm(ctx, id)
//			assert.IsType(t, err, repoErr)
//		},
//	)
//
//	t.Run(
//		"cache.Delete - Internal Err", func(t *testing.T) {
//			mockRepo.EXPECT().
//				DeletePerm(gomock.Any(), id).
//				Return(repoErr).
//				Times(1)
//
//			mockCache.EXPECT().
//				Delete(gomock.Any(), gomock.Any()).
//				Return(cacheErr).
//				Times(1)
//
//			err := ctrl.DeletePerm(ctx, id)
//			assert.IsType(t, err, repoErr)
//		},
//	)
//
//	t.Run(
//		"Success", func(t *testing.T) {
//			mockCache.EXPECT().
//				InvalidateKeysByPattern(gomock.Any(), gomock.Any()).
//				AnyTimes()
//
//			mockRepo.EXPECT().
//				DeletePerm(gomock.Any(), id).
//				Return(nil).
//				Times(1)
//
//			err := ctrl.DeletePerm(ctx, id)
//			assert.Nil(t, err)
//		},
//	)
//}

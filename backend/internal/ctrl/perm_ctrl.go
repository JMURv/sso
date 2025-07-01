package ctrl

import (
	"context"
	"errors"
	"fmt"

	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo"
	"github.com/goccy/go-json"
	"github.com/opentracing/opentracing-go"
)

const (
	permKey     = "perm:%v"
	permListKey = "perms-list:%v:%v:%v"
	permPattern = "perms-*"
)

type permCtrl interface {
	ListPermissions(ctx context.Context, page, size int, filters map[string]any) (*dto.PaginatedPermissionResponse, error)
	GetPermission(ctx context.Context, id uint64) (*md.Permission, error)
	CreatePerm(ctx context.Context, req *dto.CreatePermissionRequest) (uint64, error)
	UpdatePerm(ctx context.Context, id uint64, req *dto.UpdatePermissionRequest) error
	DeletePerm(ctx context.Context, id uint64) error
}

type permRepo interface {
	ListPermissions(ctx context.Context, page, size int, filters map[string]any) (*dto.PaginatedPermissionResponse, error)
	GetPermission(ctx context.Context, id uint64) (*md.Permission, error)
	CreatePerm(ctx context.Context, req *dto.CreatePermissionRequest) (uint64, error)
	UpdatePerm(ctx context.Context, id uint64, req *dto.UpdatePermissionRequest) error
	DeletePerm(ctx context.Context, id uint64) error
}

func (c *Controller) ListPermissions(ctx context.Context, page, size int, filters map[string]any) (*dto.PaginatedPermissionResponse, error) {
	const op = "perms.ListPermissions.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	cached := &dto.PaginatedPermissionResponse{}
	key := fmt.Sprintf(permListKey, page, size, filters)
	if err := c.cache.GetToStruct(ctx, key, &cached); err == nil {
		return cached, nil
	}

	res, err := c.repo.ListPermissions(ctx, page, size, filters)
	if err != nil {
		return nil, err
	}

	var bytes []byte
	if bytes, err = json.Marshal(res); err == nil {
		c.cache.Set(ctx, config.DefaultCacheTime, key, bytes)
	}
	return res, nil
}

func (c *Controller) GetPermission(ctx context.Context, id uint64) (*md.Permission, error) {
	const op = "perms.GetPermission.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	cached := &md.Permission{}
	cacheKey := fmt.Sprintf(permKey, id)
	if err := c.cache.GetToStruct(ctx, cacheKey, cached); err == nil {
		return cached, nil
	}

	res, err := c.repo.GetPermission(ctx, id)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	var bytes []byte
	if bytes, err = json.Marshal(res); err == nil {
		c.cache.Set(ctx, config.DefaultCacheTime, cacheKey, bytes)
	}
	return res, nil
}

func (c *Controller) CreatePerm(ctx context.Context, req *dto.CreatePermissionRequest) (uint64, error) {
	const op = "perms.CreatePerm.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := c.repo.CreatePerm(ctx, req)
	if err != nil && errors.Is(err, repo.ErrAlreadyExists) {
		return 0, ErrAlreadyExists
	} else if err != nil {
		return 0, err
	}

	go c.cache.InvalidateKeysByPattern(ctx, permPattern)
	return res, nil
}

func (c *Controller) UpdatePerm(ctx context.Context, id uint64, req *dto.UpdatePermissionRequest) error {
	const op = "perms.UpdatePerm.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	err := c.repo.UpdatePerm(ctx, id, req)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	c.cache.Delete(ctx, fmt.Sprintf(permKey, id))
	go c.cache.InvalidateKeysByPattern(ctx, permPattern)
	return nil
}

func (c *Controller) DeletePerm(ctx context.Context, id uint64) error {
	const op = "perms.DeletePerm.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	err := c.repo.DeletePerm(ctx, id)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	c.cache.Delete(ctx, fmt.Sprintf(permKey, id))
	go c.cache.InvalidateKeysByPattern(ctx, permPattern)
	return nil
}

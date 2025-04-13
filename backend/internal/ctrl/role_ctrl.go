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

const roleKey = "role:%v"
const roleListKey = "roles-list:%v:%v:%v"
const rolePattern = "roles-*"

type roleCtrl interface {
	ListRoles(ctx context.Context, page, size int, filters map[string]any) (*dto.PaginatedRoleResponse, error)
	GetRole(ctx context.Context, uid uint64) (*md.Role, error)
	CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (uint64, error)
	UpdateRole(ctx context.Context, uid uint64, req *dto.UpdateRoleRequest) error
	DeleteRole(ctx context.Context, uid uint64) error
}

type roleRepo interface {
	ListRoles(ctx context.Context, page, size int, filters map[string]any) (*dto.PaginatedRoleResponse, error)
	GetRole(ctx context.Context, id uint64) (*md.Role, error)
	CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (uint64, error)
	UpdateRole(ctx context.Context, id uint64, req *dto.UpdateRoleRequest) error
	DeleteRole(ctx context.Context, id uint64) error
}

func (c *Controller) ListRoles(ctx context.Context, page, size int, filters map[string]any) (*dto.PaginatedRoleResponse, error) {
	const op = "roles.ListRoles.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	cached := &dto.PaginatedRoleResponse{}
	key := fmt.Sprintf(roleListKey, page, size, filters)
	if err := c.cache.GetToStruct(ctx, key, &cached); err == nil {
		return cached, nil
	}

	res, err := c.repo.ListRoles(ctx, page, size, filters)
	if err != nil {
		return nil, err
	}

	var bytes []byte
	if bytes, err = json.Marshal(res); err == nil {
		c.cache.Set(ctx, config.DefaultCacheTime, key, bytes)
	}

	return res, nil
}

func (c *Controller) GetRole(ctx context.Context, id uint64) (*md.Role, error) {
	const op = "roles.GetRole.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	cached := &md.Role{}
	cacheKey := fmt.Sprintf(roleKey, id)
	if err := c.cache.GetToStruct(ctx, cacheKey, cached); err == nil {
		return cached, nil
	}

	res, err := c.repo.GetRole(ctx, id)
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

func (c *Controller) CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (uint64, error) {
	const op = "roles.CreateRole.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := c.repo.CreateRole(ctx, req)
	if err != nil && errors.Is(err, repo.ErrAlreadyExists) {
		return 0, ErrAlreadyExists
	} else if err != nil {
		return 0, err
	}

	go c.cache.InvalidateKeysByPattern(ctx, rolePattern)
	return res, nil
}

func (c *Controller) UpdateRole(ctx context.Context, id uint64, req *dto.UpdateRoleRequest) error {
	const op = "roles.UpdateRole.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	err := c.repo.UpdateRole(ctx, id, req)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	c.cache.Delete(ctx, fmt.Sprintf(roleKey, id))
	go c.cache.InvalidateKeysByPattern(ctx, rolePattern)
	return nil
}

func (c *Controller) DeleteRole(ctx context.Context, id uint64) error {
	const op = "roles.DeleteRole.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	err := c.repo.DeleteRole(ctx, id)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		return ErrNotFound
	} else if err != nil {

		return err
	}

	c.cache.Delete(ctx, fmt.Sprintf(roleKey, id))
	go c.cache.InvalidateKeysByPattern(ctx, rolePattern)
	return nil
}

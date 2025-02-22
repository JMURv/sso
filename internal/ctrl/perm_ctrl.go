package ctrl

import (
	"context"
	"errors"
	"fmt"
	"github.com/JMURv/sso/internal/config"
	md "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo"
	"github.com/goccy/go-json"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

const permKey = "perm:%v"
const permListKey = "perms-list:%v:%v"
const permPattern = "perms-*"

type permRepo interface {
	ListPermissions(ctx context.Context, page, size int) (*md.PaginatedPermission, error)
	GetPermission(ctx context.Context, id uint64) (*md.Permission, error)
	CreatePerm(ctx context.Context, req *md.Permission) (uint64, error)
	UpdatePerm(ctx context.Context, id uint64, req *md.Permission) error
	DeletePerm(ctx context.Context, id uint64) error
}

func (c *Controller) ListPermissions(ctx context.Context, page, size int) (*md.PaginatedPermission, error) {
	const op = "perms.ListPermissions.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	cached := &md.PaginatedPermission{}
	key := fmt.Sprintf(permListKey, page, size)
	if err := c.cache.GetToStruct(ctx, key, &cached); err == nil {
		return cached, nil
	}

	res, err := c.repo.ListPermissions(ctx, page, size)
	if err != nil {
		zap.L().Debug(
			"failed to list permissions",
			zap.Error(err), zap.String("op", op),
			zap.Int("page", page), zap.Int("size", size),
		)
		return nil, err
	}

	if bytes, err := json.Marshal(res); err == nil {
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
		zap.L().Debug(
			"failed to find permission",
			zap.Error(err), zap.String("op", op),
			zap.Uint64("id", id),
		)
		return nil, ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to get permission",
			zap.Error(err), zap.String("op", op),
			zap.Uint64("id", id),
		)
		return nil, err
	}

	if bytes, err := json.Marshal(res); err == nil {
		c.cache.Set(ctx, config.DefaultCacheTime, cacheKey, bytes)
	}
	return res, nil
}

func (c *Controller) CreatePerm(ctx context.Context, req *md.Permission) (uint64, error) {
	const op = "perms.CreatePerm.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := c.repo.CreatePerm(ctx, req)
	if err != nil && errors.Is(err, repo.ErrAlreadyExists) {
		return 0, ErrAlreadyExists
	} else if err != nil {
		zap.L().Debug(
			"failed to create permission",
			zap.Error(err), zap.String("op", op),
		)
		return 0, err
	}

	go c.cache.InvalidateKeysByPattern(ctx, permPattern)
	return res, nil
}

func (c *Controller) UpdatePerm(ctx context.Context, id uint64, req *md.Permission) error {
	const op = "perms.UpdatePerm.ctrl"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	err := c.repo.UpdatePerm(ctx, id, req)
	if err != nil && errors.Is(err, repo.ErrNotFound) {
		zap.L().Debug(
			"failed to find permission",
			zap.Error(err), zap.String("op", op),
			zap.Uint64("id", id),
		)
		return ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to update permission",
			zap.Error(err), zap.String("op", op),
			zap.Uint64("id", id),
		)
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
		zap.L().Debug(
			"failed to delete permission",
			zap.Error(err), zap.String("op", op),
			zap.Uint64("id", id),
		)
		return ErrNotFound
	} else if err != nil {
		zap.L().Debug(
			"failed to delete permission",
			zap.Error(err), zap.String("op", op),
			zap.Uint64("id", id),
		)
		return err
	}

	c.cache.Delete(ctx, fmt.Sprintf(permKey, id))
	go c.cache.InvalidateKeysByPattern(ctx, permPattern)
	return nil
}

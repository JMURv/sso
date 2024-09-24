package memory

import (
	"context"
	repo "github.com/JMURv/sso/internal/repository"
	"github.com/JMURv/sso/pkg/model"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/crypto/bcrypt"
)

func (r *Repository) ListUsers(ctx context.Context) ([]*model.User, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "users.ListUsers.repo")
	defer span.Finish()

	r.RLock()
	defer r.RUnlock()

	res := make([]*model.User, 0, len(r.usersData))
	for _, v := range r.usersData {
		res = append(res, v)
	}
	return res, nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "users.GetByID.repo")
	defer span.Finish()

	r.RLock()
	defer r.RUnlock()
	for _, v := range r.usersData {
		if v.ID == userID {
			return v, nil
		}
	}
	return nil, repo.ErrNotFound
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "users.GetByEmail.repo")
	defer span.Finish()

	r.RLock()
	defer r.RUnlock()
	for _, v := range r.usersData {
		if v.Email == email {
			return v, nil
		}
	}
	return nil, repo.ErrNotFound
}

func (r *Repository) CreateUser(ctx context.Context, u *model.User) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "users.CreateUser.repo")
	defer span.Finish()

	u.ID = uuid.New()
	if u.Name == "" {
		return repo.ErrUsernameIsRequired
	}

	if u.Email == "" {
		return repo.ErrEmailIsRequired
	}

	if u.Password == "" {
		return repo.ErrPasswordIsRequired
	}
	password, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return repo.ErrGeneratingPassword
	}
	u.Password = string(password)

	r.Lock()
	r.usersData[u.ID] = u
	defer r.Unlock()
	return nil
}

func (r *Repository) UpdateUser(ctx context.Context, userID uuid.UUID, u *model.User) (*model.User, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "users.UpdateUser.repo")
	defer span.Finish()

	if u.Name == "" {
		return nil, repo.ErrUsernameIsRequired
	}

	if u.Email == "" {
		return nil, repo.ErrEmailIsRequired
	}

	r.Lock()
	defer r.Unlock()
	if _, ok := r.usersData[userID]; ok {
		r.usersData[userID] = u
		return u, nil
	}
	return nil, repo.ErrNotFound
}

func (r *Repository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "users.DeleteUser.repo")
	defer span.Finish()

	r.Lock()
	defer r.Unlock()

	for k, v := range r.usersData {
		if v.ID == userID {
			delete(r.usersData, k)
			return nil
		}
	}
	return repo.ErrNotFound
}

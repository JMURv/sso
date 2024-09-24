package memory

import (
	"github.com/JMURv/sso/pkg/model"
	"github.com/google/uuid"
	"sync"
)

type Repository struct {
	sync.RWMutex
	usersData map[uuid.UUID]*model.User
}

func New() *Repository {
	return &Repository{
		usersData: make(map[uuid.UUID]*model.User),
	}
}

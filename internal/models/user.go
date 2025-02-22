package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Password    string       `json:"password"`
	Email       string       `json:"email"`
	Avatar      string       `json:"avatar"`
	Address     string       `json:"address"`
	Phone       string       `json:"phone"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

package model

import (
	"database/sql"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"os"
	"time"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Password  string    `json:"password"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Permission struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
	Value  bool      `json:"value"`
}

func MustPrecreateUsers(db *sql.DB) {
	var count int64
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		panic(err)
	}

	if count == 0 {
		type usrWithPerms struct {
			ID       uuid.UUID    `json:"id"`
			Name     string       `json:"name"`
			Password string       `json:"password"`
			Email    string       `json:"email"`
			Avatar   string       `json:"avatar"`
			Address  string       `json:"address"`
			Phone    string       `json:"phone"`
			Perms    []Permission `json:"permissions"`
		}
		bytes, err := os.ReadFile("db/precreate/users.json")
		if err != nil {
			panic(err)
		}

		p := make([]*usrWithPerms, 0, 2)
		if err = json.Unmarshal(bytes, &p); err != nil {
			panic(err)
		}

		for _, v := range p {
			password, err := bcrypt.GenerateFromPassword([]byte(v.Password), bcrypt.DefaultCost)
			if err != nil {
				panic(err)
			}
			v.Password = string(password)

			var userID uuid.UUID
			err = db.QueryRow(`INSERT INTO users (id, name, password, email, avatar, address, phone) 
				VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
				v.ID,
				v.Name,
				v.Password,
				v.Email,
				v.Avatar,
				v.Address,
				v.Phone,
			).Scan(&userID)

			if err != nil {
				panic(err)
			}

			for _, perm := range v.Perms {
				_, err := db.Exec(`INSERT INTO permissions (user_id, name, value) VALUES ($1, $2, $3)`,
					v.ID,
					perm.Name,
					perm.Value,
				)
				if err != nil {
					panic(err)
				}
			}
		}

		zap.L().Debug("Users and permissions have been created")
	} else {
		zap.L().Debug("Users and permissions already exist")
	}
}

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

type PaginatedUser struct {
	Data        []*User `json:"data"`
	Count       int64   `json:"count"`
	TotalPages  int     `json:"total_pages"`
	CurrentPage int     `json:"current_page"`
	HasNextPage bool    `json:"has_next_page"`
}

func MustPrecreateUsersAndPerms(db *sql.DB) {
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
			tx, err := db.Begin()
			if err != nil {
				panic(err)
			}

			password, err := bcrypt.GenerateFromPassword([]byte(v.Password), bcrypt.DefaultCost)
			if err != nil {
				panic(err)
			}
			v.Password = string(password)

			var userID uuid.UUID
			err = tx.QueryRow(
				`INSERT INTO users (id, name, password, email, avatar, address, phone) 
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
				var permID uint64

				err := tx.QueryRow(`SELECT FROM permission WHERE name = $1`, perm.Name).Scan(&permID)
				if err != nil && err == sql.ErrNoRows {
					if err := tx.QueryRow(
						`INSERT INTO permission (name) VALUES ($1) RETURNING id`,
						perm.Name,
					).Scan(&permID); err != nil {
						tx.Rollback()
						panic(err)
					}
				} else if err != nil {
					tx.Rollback()
					panic(err)
				}

				_, err = tx.Exec(
					`INSERT INTO user_permission (user_id, permission_id, value) VALUES ($1, $2, $3)`,
					userID, permID, true,
				)
				if err != nil {
					tx.Rollback()
					panic(err)
				}
			}
		}

		zap.L().Debug("Users and permissions have been created")
	} else {
		zap.L().Debug("Users and permissions already exist")
	}
}

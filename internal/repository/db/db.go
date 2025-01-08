package db

import (
	"database/sql"
	"fmt"
	conf "github.com/JMURv/sso/pkg/config"
	"github.com/JMURv/sso/pkg/model"
	dbutils "github.com/JMURv/sso/pkg/utils/db"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"os"
)

type Repository struct {
	conn *sql.DB
}

func New(conf *conf.DBConfig) *Repository {
	conn, err := sql.Open(
		"postgres", fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			conf.User,
			conf.Password,
			conf.Host,
			conf.Port,
			conf.Database,
		),
	)
	if err != nil {
		zap.L().Fatal("Failed to connect to the database", zap.Error(err))
	}

	if err = conn.Ping(); err != nil {
		zap.L().Fatal("Failed to ping the database", zap.Error(err))
	}

	if err = dbutils.ApplyMigrations(conn, conf); err != nil {
		zap.L().Fatal("Failed to apply migrations", zap.Error(err))
	}

	MustPrecreate(conn)
	return &Repository{conn: conn}
}

func (r *Repository) Close() error {
	return r.conn.Close()
}

func MustPrecreate(db *sql.DB) {
	var count int64
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		panic(err)
	}

	if count == 0 {
		type usrWithPerms struct {
			Name     string             `json:"name"`
			Password string             `json:"password"`
			Email    string             `json:"email"`
			Perms    []model.Permission `json:"permissions"`
		}
		bytes, err := os.ReadFile("precreate.json")
		if err != nil {
			panic(err)
		}

		p := make([]usrWithPerms, 0, 2)
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
				`INSERT INTO users (name, password, email) 
				VALUES ($1, $2, $3) RETURNING id`,
				v.Name,
				v.Password,
				v.Email,
			).Scan(&userID)

			if err != nil {
				panic(err)
			}

			for _, perm := range v.Perms {
				var permID uint64

				err := tx.QueryRow(`SELECT FROM permission WHERE name = $1`, perm.Name).Scan(&permID)
				if err != nil && err == sql.ErrNoRows {
					if err := tx.QueryRow(permCreate, perm.Name).Scan(&permID); err != nil {
						tx.Rollback()
						panic(err)
					}
				} else if err != nil {
					tx.Rollback()
					panic(err)
				}

				if _, err = tx.Exec(userCreatePermQ, userID, permID, true); err != nil {
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

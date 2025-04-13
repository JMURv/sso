package db

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"github.com/JMURv/sso/internal/config"
	"github.com/golang-migrate/migrate/v4"
	pgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
)

func applyMigrations(db *sql.DB, conf config.Config) error {
	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		return err
	}

	path := os.Getenv("MIGRATIONS_PATH")
	if path == "" {
		path = filepath.ToSlash(
			filepath.Join("internal", "repo", "db", "migration"),
		)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+path, conf.DB.Database, driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			zap.L().Info("No migrations to apply")
			return nil
		} else {
			zap.L().Error("Failed to apply migrations", zap.Error(err))
			return err
		}
	}

	zap.L().Info("Applied migrations")
	return nil
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateRandomPassword(n int) (string, error) {
	password := make([]byte, n)
	l := big.NewInt(int64(len(letters)))
	for i := range password {
		num, err := rand.Int(rand.Reader, l)
		if err != nil {
			return "", err
		}
		password[i] = letters[num.Int64()]
	}
	return string(password), nil
}

func mustPrecreate(conf config.Config, db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			panic(err)
		}
	}()

	roleID := 1
	err = tx.QueryRow(`SELECT id FROM roles WHERE name='admin'`).Scan(&roleID)
	if err != nil {
		zap.L().Error("failed to get admin role", zap.Error(err))
		return
	}

	for i := 0; i < len(conf.Auth.Admins); i++ {
		email := conf.Auth.Admins[i]
		name := strings.Split(email, "@")[0]

		var userID uuid.UUID
		if err = tx.QueryRow(`SELECT id FROM users WHERE email = $1`, email).Scan(&userID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				randomPassword, err := GenerateRandomPassword(12)
				if err != nil {
					zap.L().Error("failed to generate random password", zap.Error(err))
					return
				}

				password, err := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)
				if err != nil {
					zap.L().Error("failed to generate password", zap.Error(err))
					return
				}

				if err = tx.QueryRow(
					`INSERT INTO users (name, password, email, avatar, is_active, is_email_verified) 
			 		 VALUES ($1, $2, $3, $4, $5, $6)
			 		 RETURNING id`, name, password, email, "", true, true,
				).Scan(&userID); err != nil {
					zap.L().Error("failed to insert user", zap.Error(err))
					return
				}

				if _, err = tx.Exec(
					`INSERT INTO user_roles (user_id, role_id) 
			 		 VALUES ($1, $2)
			 		 ON CONFLICT (user_id, role_id) DO NOTHING`, userID, roleID,
				); err != nil {
					zap.L().Error("failed to insert user role", zap.Error(err))
					return
				}

				zap.L().Info(
					"user has been created",
					zap.String("email", email),
					zap.String("pass", randomPassword),
				)
			}
		}
	}
	if err = tx.Commit(); err != nil {
		zap.L().Error("failed to commit transaction", zap.Error(err))
		return
	}
}

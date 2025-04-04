package db

import (
	"database/sql"
	"errors"
	"fmt"
	conf "github.com/JMURv/sso/internal/config"
	md "github.com/JMURv/sso/internal/models"
	"github.com/goccy/go-json"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"os"
	"path/filepath"
)

func applyMigrations(db *sql.DB, conf conf.Config) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	rootDir, err := findRootDir()
	if err != nil {
		return err
	}

	path := filepath.ToSlash(
		filepath.Join(rootDir, "internal", "repo", "db", "migration"),
	)

	m, err := migrate.NewWithDatabaseInstance("file://"+path, conf.DB.Database, driver)
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil && errors.Is(err, migrate.ErrNoChange) {
		zap.L().Info("no migrations to apply")
		return nil
	} else if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	zap.L().Info("applied migrations")
	return nil
}

func findRootDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		if dir == "/" {
			break
		}
		dir = filepath.Dir(dir)
	}
	return "", fmt.Errorf("go.mod not found")
}

func mustPrecreate(db *sql.DB) {
	var count int64
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		panic(err)
	}

	if count == 0 {
		type usrWithPerms struct {
			Name     string          `json:"name"`
			Password string          `json:"password"`
			Email    string          `json:"email"`
			Avatar   string          `json:"avatar"`
			Address  string          `json:"address"`
			Phone    string          `json:"phone"`
			Perms    []md.Permission `json:"permissions"`
		}
		bytes, err := os.ReadFile("precreate.json")
		if err != nil && errors.Is(err, os.ErrNotExist) {
			zap.L().Info("precreate.json not found")
			return
		} else if err != nil {
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
				`INSERT INTO users (name, password, email, avatar, address, phone) 
				VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
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
				zap.L().Debug("debug", zap.Any("perm", perm))
			}

			if err = tx.Commit(); err != nil {
				panic(err)
			}
		}

		zap.L().Info("users and permissions have been created")
	} else {
		zap.L().Info("users and permissions already exist")
	}
}

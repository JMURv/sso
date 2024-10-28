package utils

import (
	"database/sql"
	"errors"
	conf "github.com/JMURv/sso/pkg/config"
	md "github.com/JMURv/sso/pkg/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"path/filepath"
	"strconv"
	"strings"
)

func ApplyMigrations(db *sql.DB, conf *conf.DBConfig) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	path, _ := filepath.Abs("db/migration")
	path = filepath.ToSlash(path)

	m, err := migrate.NewWithDatabaseInstance("file://"+path, conf.Database, driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && errors.Is(err, migrate.ErrNoChange) {
		zap.L().Info("No migrations to apply")
	} else if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func ScanPermissions(perms []string) ([]md.Permission, error) {
	permissions := make([]md.Permission, 0, len(perms))
	for _, perm := range perms {
		parts := strings.Split(perm, "|")
		if len(parts) != 3 {
			continue
		}

		id, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return nil, err
		}

		value, err := strconv.ParseBool(parts[2])
		if err != nil {
			return nil, err
		}

		permissions = append(
			permissions, md.Permission{
				ID:    id,
				Name:  parts[1],
				Value: value,
			},
		)
	}
	return permissions, nil
}

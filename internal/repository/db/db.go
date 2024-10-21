package db

import (
	"database/sql"
	"fmt"
	conf "github.com/JMURv/sso/pkg/config"
	"github.com/JMURv/sso/pkg/model"
	dbutils "github.com/JMURv/sso/pkg/utils/db"
	"go.uber.org/zap"
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

	if err := conn.Ping(); err != nil {
		zap.L().Fatal("Failed to ping the database", zap.Error(err))
	}

	if err := dbutils.ApplyMigrations(conn, conf); err != nil {
		zap.L().Fatal("Failed to apply migrations", zap.Error(err))
	}

	model.MustPrecreateUsersAndPerms(conn)
	return &Repository{conn: conn}
}

func (r *Repository) Close() error {
	return r.conn.Close()
}

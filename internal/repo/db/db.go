package db

import (
	"database/sql"
	"fmt"
	conf "github.com/JMURv/sso/internal/config"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type Repository struct {
	conn *sql.DB
}

func New(conf conf.Config) *Repository {
	conn, err := sql.Open(
		"postgres", fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			conf.DB.User,
			conf.DB.Password,
			conf.DB.Host,
			conf.DB.Port,
			conf.DB.Database,
		),
	)
	if err != nil {
		zap.L().Fatal("failed to connect to the database", zap.Error(err))
	}

	if err = conn.Ping(); err != nil {
		zap.L().Fatal("failed to ping the database", zap.Error(err))
	}

	if err = applyMigrations(conn, conf); err != nil {
		zap.L().Fatal("failed to apply migrations", zap.Error(err))
	}

	mustPrecreate(conn)
	return &Repository{conn: conn}
}

func (r *Repository) Close() error {
	return r.conn.Close()
}

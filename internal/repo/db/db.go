package db

import (
	"fmt"
	conf "github.com/JMURv/sso/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Repository struct {
	conn *sqlx.DB
}

func New(conf conf.Config) *Repository {
	conn, err := sqlx.Open(
		"pgx", fmt.Sprintf(
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

	if err = applyMigrations(conn.DB, conf); err != nil {
		zap.L().Fatal("failed to apply migrations", zap.Error(err))
	}

	mustPrecreate(conn.DB)
	return &Repository{conn: conn}
}

func (r *Repository) Close() error {
	return r.conn.Close()
}

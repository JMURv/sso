package http

import (
	"database/sql"
	"fmt"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/cache/redis"
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/ctrl"
	hdl "github.com/JMURv/sso/internal/hdl/http"
	"github.com/JMURv/sso/internal/repo/db"
	"github.com/JMURv/sso/internal/repo/s3"
	"github.com/JMURv/sso/internal/smtp"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
)

const getTables = `
SELECT tablename 
FROM pg_tables 
WHERE schemaname = 'public';
`

func setupTestServer() (*httptest.Server, func()) {
	zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))

	conf := config.MustLoad("config/.env")
	au := auth.New(conf)

	repo := db.New(conf)
	cache := redis.New(conf)
	svc := ctrl.New(repo, au, cache, s3.New(conf), smtp.New(conf))
	handler := hdl.New(svc, au)

	mux := http.NewServeMux()
	handler.RegisterAuthRoutes()
	handler.RegisterOAuth2Routes()
	handler.RegisterOIDCRoutes()
	handler.RegisterWebAuthnRoutes()
	handler.RegisterUserRoutes()
	handler.RegisterPermRoutes()
	handler.RegisterRoleRoutes()
	handler.RegisterDeviceRoutes()

	cleanupFunc := func() {
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
			zap.L().Fatal("Failed to connect to the database", zap.Error(err))
		}

		if err = conn.Ping(); err != nil {
			zap.L().Fatal("Failed to ping the database", zap.Error(err))
		}

		rows, err := conn.Query(getTables)
		if err != nil {
			zap.L().Fatal("Failed to fetch table names", zap.Error(err))
		}
		defer func(rows *sql.Rows) {
			if err := rows.Close(); err != nil {
				zap.L().Error(
					"error while closing rows",
					zap.Error(err),
				)
			}
		}(rows)

		var tables []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				zap.L().Fatal("Failed to scan table name", zap.Error(err))
			}
			tables = append(tables, name)
		}

		if len(tables) == 0 {
			return
		}

		_, err = conn.Exec(fmt.Sprintf("TRUNCATE TABLE %v RESTART IDENTITY CASCADE;", strings.Join(tables, ", ")))
		if err != nil {
			zap.L().Fatal("Failed to truncate tables", zap.Error(err))
		}
	}

	return httptest.NewServer(mux), cleanupFunc
}

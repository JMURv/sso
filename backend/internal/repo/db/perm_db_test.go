package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	rrepo "github.com/JMURv/sso/internal/repo"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"regexp"
	"testing"
)

var internalErr = errors.New("internal error")

func TestRepository_ListPermissions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := Repository{conn: &sqlx.DB{DB: db}}
	page, size := 1, 10

	mockData := []md.Permission{
		{ID: 1, Name: "admin"},
		{ID: 2, Name: "moderator"},
	}

	t.Run(
		"Success", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(permSelect)).WillReturnRows(
				sqlmock.NewRows([]string{"count"}).
					AddRow(int64(len(mockData))),
			)

			mock.ExpectQuery(regexp.QuoteMeta(permList)).
				WithArgs(size, (page-1)*size).
				WillReturnRows(
					sqlmock.NewRows([]string{"id", "name"}).
						AddRow(mockData[0].ID, mockData[0].Name).
						AddRow(mockData[1].ID, mockData[1].Name),
				)

			res, err := repo.ListPermissions(context.Background(), page, size, map[string]any{})
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, int64(len(mockData)), res.Count)
			assert.Len(t, res.Data, len(mockData))
		},
	)

	t.Run(
		"SelectErr", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(permSelect)).WillReturnError(internalErr)

			res, err := repo.ListPermissions(context.Background(), page, size, map[string]any{})
			require.Nil(t, res)
			require.Equal(t, internalErr, err)
		},
	)

	t.Run(
		"QueryErr", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(permSelect)).WillReturnRows(
				sqlmock.NewRows([]string{"count"}).
					AddRow(int64(len(mockData))),
			)

			mock.ExpectQuery(regexp.QuoteMeta(permList)).
				WithArgs(size, (page-1)*size).
				WillReturnError(internalErr)

			res, err := repo.ListPermissions(context.Background(), page, size, map[string]any{})
			require.Nil(t, res)
			require.Equal(t, internalErr, err)
		},
	)

	t.Run(
		"RowsIterationError", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(permSelect)).WillReturnRows(
				sqlmock.NewRows([]string{"count"}).
					AddRow(int64(len(mockData))),
			)

			mock.ExpectQuery(regexp.QuoteMeta(permList)).
				WithArgs(size, (page-1)*size).
				WillReturnRows(
					sqlmock.NewRows([]string{"id", "name"}).
						AddRow(mockData[0].ID, mockData[0].Name).
						RowError(0, internalErr),
				)

			res, err := repo.ListPermissions(context.Background(), page, size, map[string]any{})
			require.Nil(t, res)
			require.Equal(t, internalErr, err)
		},
	)
}

func TestRepository_GetPermission(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := Repository{conn: &sqlx.DB{DB: db}}
	mockData := md.Permission{ID: 1, Name: "admin"}

	t.Run(
		"Success", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(permGet)).
				WithArgs(mockData.ID).
				WillReturnRows(
					sqlmock.NewRows([]string{"id", "name"}).
						AddRow(mockData.ID, mockData.Name),
				)

			res, err := repo.GetPermission(context.Background(), mockData.ID)
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, mockData.Name, res.Name)
		},
	)

	t.Run(
		"ErrNotFound", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(permGet)).
				WithArgs(mockData.ID).
				WillReturnError(sql.ErrNoRows)

			res, err := repo.GetPermission(context.Background(), mockData.ID)
			require.ErrorIs(t, rrepo.ErrNotFound, err)
			require.Nil(t, res)
		},
	)

	t.Run(
		"ErrInternal", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(permGet)).
				WithArgs(mockData.ID).
				WillReturnError(internalErr)

			res, err := repo.GetPermission(context.Background(), mockData.ID)
			require.ErrorIs(t, internalErr, err)
			require.Nil(t, res)
		},
	)
}

func TestRepository_CreatePerm(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := Repository{conn: &sqlx.DB{DB: db}}
	id := uint64(1)
	mockData := dto.CreatePermissionRequest{Name: "admin"}

	t.Run(
		"Success", func(t *testing.T) {
			mock.ExpectBegin()

			mock.ExpectQuery(regexp.QuoteMeta(permCreate)).
				WithArgs(mockData.Name).
				WillReturnRows(
					sqlmock.NewRows([]string{"id"}).
						AddRow(id),
				)

			mock.ExpectCommit()

			res, err := repo.CreatePerm(context.Background(), &mockData)
			require.NoError(t, err)
			require.NotNil(t, res)
			require.NotNil(t, uint64(0))
			assert.Equal(t, id, res)
		},
	)

	t.Run(
		"BeginError", func(t *testing.T) {
			mock.ExpectBegin().WillReturnError(internalErr)

			res, err := repo.CreatePerm(context.Background(), &mockData)
			require.Equal(t, uint64(0), res)
			require.Equal(t, internalErr, err)
		},
	)

	t.Run(
		"ErrAlreadyExists", func(t *testing.T) {
			mock.ExpectBegin()

			mock.ExpectQuery(regexp.QuoteMeta(permCreate)).
				WithArgs(mockData.Name).
				WillReturnError(errors.New("unique constraint"))

			mock.ExpectRollback()

			res, err := repo.CreatePerm(context.Background(), &mockData)
			require.ErrorIs(t, rrepo.ErrAlreadyExists, err)
			require.Equal(t, uint64(0), res)
		},
	)

	t.Run(
		"ErrInternal", func(t *testing.T) {
			mock.ExpectBegin()

			mock.ExpectQuery(regexp.QuoteMeta(permCreate)).
				WithArgs(mockData.Name).
				WillReturnError(internalErr)

			mock.ExpectRollback()

			res, err := repo.CreatePerm(context.Background(), &mockData)
			require.ErrorIs(t, internalErr, err)
			require.Equal(t, uint64(0), res)
		},
	)

	t.Run(
		"CommitError", func(t *testing.T) {
			mock.ExpectBegin()

			mock.ExpectQuery(regexp.QuoteMeta(permCreate)).
				WithArgs(mockData.Name).
				WillReturnRows(
					sqlmock.NewRows([]string{"id"}).
						AddRow(id),
				)

			mock.ExpectCommit().WillReturnError(internalErr)

			res, err := repo.CreatePerm(context.Background(), &mockData)
			require.Equal(t, uint64(0), res)
			require.Equal(t, internalErr, err)
		},
	)
}

func TestRepository_UpdatePerm(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := Repository{conn: &sqlx.DB{DB: db}}
	id := uint64(1)
	mockData := dto.UpdatePermissionRequest{Name: "new-admin"}

	t.Run(
		"Success", func(t *testing.T) {
			mock.ExpectBegin()

			mock.ExpectExec(regexp.QuoteMeta(permUpdate)).
				WithArgs(mockData.Name, id).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mock.ExpectCommit()

			err := repo.UpdatePerm(context.Background(), id, &mockData)
			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		},
	)

	t.Run(
		"BeginError", func(t *testing.T) {
			mock.ExpectBegin().WillReturnError(internalErr)

			err := repo.UpdatePerm(context.Background(), id, &mockData)
			require.Equal(t, internalErr, err)
			require.NoError(t, mock.ExpectationsWereMet())
		},
	)

	t.Run(
		"ErrNotFound", func(t *testing.T) {
			mock.ExpectBegin()

			mock.ExpectExec(regexp.QuoteMeta(permUpdate)).
				WithArgs(mockData.Name, id).
				WillReturnResult(sqlmock.NewResult(1, 0))

			err := repo.UpdatePerm(context.Background(), id, &mockData)
			require.ErrorIs(t, rrepo.ErrNotFound, err)
			require.NoError(t, mock.ExpectationsWereMet())
		},
	)

	t.Run(
		"ErrInternal", func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectExec(regexp.QuoteMeta(permUpdate)).
				WithArgs(mockData.Name, id).
				WillReturnError(internalErr)

			mock.ExpectRollback()

			err := repo.UpdatePerm(context.Background(), id, &mockData)
			require.ErrorIs(t, internalErr, err)
			require.NoError(t, mock.ExpectationsWereMet())
		},
	)

	t.Run(
		"CommitError", func(t *testing.T) {
			mock.ExpectBegin()

			mock.ExpectExec(regexp.QuoteMeta(permUpdate)).
				WithArgs(mockData.Name, id).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mock.ExpectCommit().WillReturnError(internalErr)

			err := repo.UpdatePerm(context.Background(), id, &mockData)
			require.Equal(t, internalErr, err)
		},
	)
}

func TestRepository_DeletePerm(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := Repository{conn: &sqlx.DB{DB: db}}
	id := uint64(1)

	t.Run(
		"Success", func(t *testing.T) {
			mock.ExpectExec(regexp.QuoteMeta(permDelete)).
				WithArgs(id).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := repo.DeletePerm(context.Background(), id)
			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		},
	)

	t.Run(
		"ErrNotFound", func(t *testing.T) {
			mock.ExpectExec(regexp.QuoteMeta(permDelete)).
				WithArgs(id).
				WillReturnResult(sqlmock.NewResult(1, 0))

			err := repo.DeletePerm(context.Background(), id)
			require.ErrorIs(t, rrepo.ErrNotFound, err)
			require.NoError(t, mock.ExpectationsWereMet())
		},
	)

	t.Run(
		"ErrInternal", func(t *testing.T) {
			mock.ExpectExec(regexp.QuoteMeta(permDelete)).
				WithArgs(id).
				WillReturnError(internalErr)

			err := repo.DeletePerm(context.Background(), id)
			require.ErrorIs(t, internalErr, err)
			require.NoError(t, mock.ExpectationsWereMet())
		},
	)
}

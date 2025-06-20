package db

import (
	"context"
	"errors"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	rrepo "github.com/JMURv/sso/internal/repo"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"regexp"
	"testing"
	"time"
)

func TestListUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := Repository{conn: &sqlx.DB{DB: db}}
	page := 1
	size := 10
	expectedCount := int64(20)
	expectedTotalPages := int((expectedCount + int64(size) - 1) / int64(size))

	mockUsers := []md.User{
		{
			ID:        uuid.New(),
			Name:      "John Doe",
			Password:  "password",
			Email:     "johndoe@example.com",
			Avatar:    "https://example.com/avatar.jpg",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "Jane Smith",
			Password:  "password",
			Email:     "janesmith@example.com",
			Avatar:    "https://example.com/avatar.jpg",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	t.Run(
		"Success", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(userSelectQ)).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

			rows := sqlmock.NewRows(
				[]string{
					"id",
					"name",
					"password",
					"email",
					"avatar",
					"address",
					"phone",
					"created_at",
					"updated_at",
				},
			).
				AddRow(
					mockUsers[0].ID.String(),
					mockUsers[0].Name,
					mockUsers[0].Password,
					mockUsers[0].Email,
					mockUsers[0].Avatar,
					mockUsers[0].CreatedAt,
					mockUsers[0].UpdatedAt,
				).
				AddRow(
					mockUsers[1].ID.String(),
					mockUsers[1].Name,
					mockUsers[1].Password,
					mockUsers[1].Email,
					mockUsers[1].Avatar,
					mockUsers[1].CreatedAt,
					mockUsers[1].UpdatedAt,
				)

			mock.ExpectQuery(regexp.QuoteMeta(userListQ)).WithArgs(size, (page-1)*size).WillReturnRows(rows)

			resp, err := repo.ListUsers(context.Background(), page, size, map[string]any{})

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, expectedCount, resp.Count)
			assert.Len(t, resp.Data, len(mockUsers))
			assert.Equal(t, expectedTotalPages, resp.TotalPages)
			assert.True(t, resp.HasNextPage)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		},
	)

	t.Run(
		"Count error", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(userSelectQ)).
				WillReturnError(errors.New("count error"))

			resp, err := repo.ListUsers(context.Background(), page, size, map[string]any{})

			assert.Error(t, err)
			assert.Nil(t, resp)
			assert.Equal(t, "count error", err.Error())

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		},
	)

	t.Run(
		"ErrInternal", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(userSelectQ)).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

			mock.ExpectQuery(regexp.QuoteMeta(userListQ)).
				WillReturnError(errors.New("find error"))

			resp, err := repo.ListUsers(context.Background(), page, size, map[string]any{})
			assert.Error(t, err)
			assert.Nil(t, resp)
			assert.Equal(t, "find error", err.Error())

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		},
	)

	t.Run(
		"Empty", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(userSelectQ)).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

			mock.ExpectQuery(regexp.QuoteMeta(userListQ)).
				WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}))

			resp, err := repo.ListUsers(context.Background(), page, size, map[string]any{})
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, int64(0), resp.Count)
			assert.Len(t, resp.Data, 0)
			assert.Equal(t, 0, resp.TotalPages)
			assert.False(t, resp.HasNextPage)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		},
	)
}

func TestGetUserByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := Repository{conn: &sqlx.DB{DB: db}}

	testUser := md.User{
		ID:        uuid.New(),
		Name:      "John Doe",
		Password:  "password",
		Email:     "johndoe@example.com",
		Avatar:    "https://example.com/avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run(
		"Success", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(userGetByIDQ)).
				WithArgs(testUser.ID.String()).
				WillReturnRows(
					sqlmock.NewRows(
						[]string{
							"id",
							"name",
							"password",
							"email",
							"avatar",
							"address",
							"phone",
							"created_at",
							"updated_at",
						},
					).
						AddRow(
							testUser.ID.String(),
							testUser.Name,
							testUser.Password,
							testUser.Email,
							testUser.Avatar,
							testUser.CreatedAt,
							testUser.UpdatedAt,
							"{2|admin|true}",
						),
				)

			result, err := repo.GetUserByID(context.Background(), testUser.ID)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, testUser.ID, result.ID)
			assert.Equal(t, testUser.Name, result.Name)
			assert.Equal(t, testUser.Email, result.Email)
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		},
	)

	t.Run(
		"ErrNotFound", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(userGetByIDQ)).
				WithArgs(testUser.ID.String()).
				WillReturnError(rrepo.ErrNotFound)

			result, err := repo.GetUserByID(context.Background(), testUser.ID)
			assert.Nil(t, result)
			assert.Equal(t, rrepo.ErrNotFound, err)
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		},
	)

	var notExpectedError = errors.New("not expected error")
	t.Run(
		"ErrInternal", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(userGetByIDQ)).
				WithArgs(testUser.ID.String()).
				WillReturnError(notExpectedError)

			result, err := repo.GetUserByID(context.Background(), testUser.ID)
			assert.Nil(t, result)
			assert.Equal(t, notExpectedError, err)
			assert.NotEqual(t, rrepo.ErrNotFound, err)
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		},
	)
}

func TestGetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := Repository{conn: &sqlx.DB{DB: db}}
	testUser := md.User{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "testuser@example.com",
	}

	t.Run(
		"Success", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(userGetByEmailQ)).
				WithArgs(testUser.Email).
				WillReturnRows(
					sqlmock.NewRows(
						[]string{
							"id",
							"name",
							"password",
							"email",
							"avatar",
							"address",
							"phone",
							"created_at",
							"updated_at",
						},
					).
						AddRow(
							testUser.ID.String(),
							testUser.Name,
							testUser.Password,
							testUser.Email,
							testUser.Avatar,
							testUser.CreatedAt,
							testUser.UpdatedAt,
						),
				)

			result, err := repo.GetUserByEmail(context.Background(), testUser.Email)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, testUser.ID, result.ID)
			assert.Equal(t, testUser.Name, result.Name)
			assert.Equal(t, testUser.Email, result.Email)
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		},
	)

	t.Run(
		"ErrNotFound", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(userGetByEmailQ)).
				WithArgs(testUser.Email).
				WillReturnError(rrepo.ErrNotFound)

			result, err := repo.GetUserByEmail(context.Background(), testUser.Email)
			assert.Nil(t, result)
			assert.Equal(t, rrepo.ErrNotFound, err)
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		},
	)

	var notExpectedError = errors.New("not expected error")
	t.Run(
		"ErrInternal", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(userGetByEmailQ)).
				WithArgs(testUser.Email).
				WillReturnError(notExpectedError)

			result, err := repo.GetUserByEmail(context.Background(), testUser.Email)
			assert.Nil(t, result)
			assert.Equal(t, notExpectedError, err)
			assert.NotEqual(t, rrepo.ErrNotFound, err)
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		},
	)
}

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := Repository{conn: &sqlx.DB{DB: db}}
	id := uuid.New()
	testUser := &dto.CreateUserRequest{
		Name:     "Test User",
		Email:    "testuser@example.com",
		Password: "securepassword",
	}

	t.Run(
		"Success", func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectQuery(
				regexp.QuoteMeta(userCreateQ),
			).
				WillReturnRows(
					sqlmock.NewRows([]string{"id"}).
						AddRow(id.String()),
				)
			mock.ExpectCommit()

			_, err := repo.CreateUser(context.Background(), testUser)
			assert.NoError(t, err)
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		},
	)

	t.Run(
		"ErrAlreadyExists", func(t *testing.T) {
			testUser.Email = "existinguser@example.com"

			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta(userCreateQ)).
				WillReturnError(rrepo.ErrAlreadyExists)
			mock.ExpectRollback()

			_, err := repo.CreateUser(context.Background(), testUser)
			assert.Equal(t, rrepo.ErrAlreadyExists, err)
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		},
	)
}

func TestUpdateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := Repository{conn: &sqlx.DB{DB: db}}

	testUserID := uuid.New()
	testUser := &dto.UpdateUserRequest{
		Name:     "Original User",
		Email:    "originaluser@example.com",
		Password: "securepassword",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(userCreateQ)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID.String()))
	mock.ExpectCommit()

	err = repo.UpdateUser(context.Background(), testUserID, testUser)
	require.NoError(t, err)

	newData := &dto.UpdateUserRequest{
		Name:     "Updated User",
		Password: "newsecurepassword",
		Email:    "updateduser@example.com",
		Avatar:   "new_avatar.png",
	}

	t.Run(
		"Success", func(t *testing.T) {
			mock.ExpectBegin()

			mock.ExpectExec(
				regexp.QuoteMeta(userUpdateQ),
			).
				WithArgs(
					newData.Name,
					newData.Password,
					newData.Email,
					newData.Avatar,
					testUserID.String(),
				).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mock.ExpectExec(
				regexp.QuoteMeta(permDelete),
			).
				WithArgs(testUserID.String()).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mock.ExpectCommit()

			err := repo.UpdateUser(context.Background(), testUserID, newData)
			assert.NoError(t, err)
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		},
	)

	t.Run(
		"ErrNotFound", func(t *testing.T) {
			mock.ExpectBegin()

			mock.ExpectExec(
				regexp.QuoteMeta(
					userUpdateQ,
				),
			).
				WithArgs(
					newData.Name,
					newData.Password,
					newData.Email,
					newData.Avatar,
					testUserID.String(),
				).
				WillReturnResult(sqlmock.NewResult(1, 0))

			err := repo.UpdateUser(context.Background(), testUserID, newData)
			assert.ErrorIs(t, err, rrepo.ErrNotFound)
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		},
	)
}

func TestDeleteUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := Repository{conn: &sqlx.DB{DB: db}}
	testUserID := uuid.New()
	t.Run(
		"Success", func(t *testing.T) {
			mock.ExpectExec(regexp.QuoteMeta(userDeleteQ)).
				WithArgs(testUserID.String()).
				WillReturnResult(sqlmock.NewResult(0, 1))

			err := repo.DeleteUser(context.Background(), testUserID)
			assert.NoError(t, err)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		},
	)

	t.Run(
		"ErrNotFound", func(t *testing.T) {
			mock.ExpectExec(regexp.QuoteMeta(userDeleteQ)).
				WithArgs(testUserID.String()).
				WillReturnResult(sqlmock.NewResult(1, 0))

			err := repo.DeleteUser(context.Background(), testUserID)
			assert.ErrorIs(t, err, rrepo.ErrNotFound)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		},
	)

	t.Run(
		"ErrInternal", func(t *testing.T) {
			mock.ExpectExec(regexp.QuoteMeta(userDeleteQ)).
				WithArgs(testUserID.String()).
				WillReturnError(errors.New("db error"))

			err := repo.DeleteUser(context.Background(), testUserID)
			assert.Error(t, err)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		},
	)
}

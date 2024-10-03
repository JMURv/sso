package db

import (
	"context"
	"database/sql"
	"errors"
	rrepo "github.com/JMURv/sso/internal/repository"
	md "github.com/JMURv/sso/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"regexp"
	"testing"
	"time"
)

func TestUserSearch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := Repository{conn: db}

	const countQ = `SELECT COUNT(*) FROM users WHERE name ILIKE $1 OR email ILIKE $2`
	const q = `SELECT id, name, password, email, avatar, address, phone, created_at, updated_at FROM users WHERE name ILIKE $1 OR email ILIKE $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`

	query := "test"
	page := 1
	size := 10
	expectedCount := int64(2)
	expectedTotalPages := int((expectedCount + int64(size) - 1) / int64(size))

	mockUsers := []md.User{
		{
			ID:        uuid.New(),
			Name:      "John Doe",
			Password:  "password",
			Email:     "johndoe@example.com",
			Avatar:    "https://example.com/avatar.jpg",
			Address:   "123 Main St",
			Phone:     "123-456-7890",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "Jane Smith",
			Password:  "password",
			Email:     "janesmith@example.com",
			Avatar:    "https://example.com/avatar.jpg",
			Address:   "123 Main St",
			Phone:     "123-456-7890",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	t.Run("Success case", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(countQ)).
			WithArgs("%"+query+"%", "%"+query+"%").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		rows := sqlmock.NewRows([]string{"id", "name", "password", "email", "avatar", "address", "phone", "created_at", "updated_at"}).
			AddRow(
				mockUsers[0].ID.String(),
				mockUsers[0].Name,
				mockUsers[0].Password,
				mockUsers[0].Email,
				mockUsers[0].Avatar,
				mockUsers[0].Address,
				mockUsers[0].Phone,
				mockUsers[0].CreatedAt,
				mockUsers[0].UpdatedAt,
			).
			AddRow(
				mockUsers[1].ID.String(),
				mockUsers[1].Name,
				mockUsers[1].Password,
				mockUsers[1].Email,
				mockUsers[1].Avatar,
				mockUsers[1].Address,
				mockUsers[1].Phone,
				mockUsers[1].CreatedAt,
				mockUsers[1].UpdatedAt,
			)

		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WithArgs("%"+query+"%", "%"+query+"%", size, (page-1)*size).
			WillReturnRows(rows)

		resp, err := repo.UserSearch(context.Background(), query, page, size)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedCount, resp.Count)
		assert.Len(t, resp.Data, len(mockUsers))
		assert.Equal(t, expectedTotalPages, resp.TotalPages)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Count error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(countQ)).
			WithArgs("%"+query+"%", "%"+query+"%").
			WillReturnError(errors.New("count error"))

		resp, err := repo.UserSearch(context.Background(), query, page, size)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "count error", err.Error())

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Find error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(countQ)).
			WithArgs("%"+query+"%", "%"+query+"%").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WithArgs("%"+query+"%", "%"+query+"%", size, (page-1)*size).
			WillReturnError(errors.New("find error"))

		resp, err := repo.UserSearch(context.Background(), query, page, size)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "find error", err.Error())

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Empty result", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(countQ)).
			WithArgs("%"+query+"%", "%"+query+"%").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WithArgs("%"+query+"%", "%"+query+"%", size, (page-1)*size).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}))

		resp, err := repo.UserSearch(context.Background(), query, page, size)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int64(0), resp.Count)
		assert.Len(t, resp.Data, 0)
		assert.Equal(t, 0, resp.TotalPages)
		assert.False(t, resp.HasNextPage)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestListUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := Repository{conn: db}
	const countQ = `SELECT COUNT(*) FROM users`
	const q = `SELECT id, name, password, email, avatar, address, phone, created_at, updated_at 
		 FROM users 
		 ORDER BY created_at DESC 
		 LIMIT $1 OFFSET $2`

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
			Address:   "123 Main St",
			Phone:     "123-456-7890",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "Jane Smith",
			Password:  "password",
			Email:     "janesmith@example.com",
			Avatar:    "https://example.com/avatar.jpg",
			Address:   "123 Main St",
			Phone:     "123-456-7890",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	t.Run("Success case", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(countQ)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		rows := sqlmock.NewRows([]string{"id", "name", "password", "email", "avatar", "address", "phone", "created_at", "updated_at"}).
			AddRow(
				mockUsers[0].ID.String(),
				mockUsers[0].Name,
				mockUsers[0].Password,
				mockUsers[0].Email,
				mockUsers[0].Avatar,
				mockUsers[0].Address,
				mockUsers[0].Phone,
				mockUsers[0].CreatedAt,
				mockUsers[0].UpdatedAt,
			).
			AddRow(
				mockUsers[1].ID.String(),
				mockUsers[1].Name,
				mockUsers[1].Password,
				mockUsers[1].Email,
				mockUsers[1].Avatar,
				mockUsers[1].Address,
				mockUsers[1].Phone,
				mockUsers[1].CreatedAt,
				mockUsers[1].UpdatedAt,
			)

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(size, (page-1)*size).WillReturnRows(rows)

		resp, err := repo.ListUsers(context.Background(), page, size)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedCount, resp.Count)
		assert.Len(t, resp.Data, len(mockUsers))
		assert.Equal(t, expectedTotalPages, resp.TotalPages)
		assert.True(t, resp.HasNextPage)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Count error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(countQ)).
			WillReturnError(errors.New("count error"))

		resp, err := repo.ListUsers(context.Background(), page, size)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "count error", err.Error())

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Find error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(countQ)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WillReturnError(errors.New("find error"))

		resp, err := repo.ListUsers(context.Background(), page, size)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "find error", err.Error())

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Empty users", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(countQ)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}))

		resp, err := repo.ListUsers(context.Background(), page, size)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int64(0), resp.Count)
		assert.Len(t, resp.Data, 0)
		assert.Equal(t, 0, resp.TotalPages)
		assert.False(t, resp.HasNextPage)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestGetUserByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := Repository{conn: db}

	const q = `SELECT id, name, password, email, avatar, address, phone, created_at, updated_at
		FROM users
		WHERE id = $1`
	testUser := md.User{
		ID:        uuid.New(),
		Name:      "John Doe",
		Password:  "password",
		Email:     "johndoe@example.com",
		Avatar:    "https://example.com/avatar.jpg",
		Address:   "123 Main St",
		Phone:     "123-456-7890",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("Success case", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WithArgs(testUser.ID.String()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "password", "email", "avatar", "address", "phone", "created_at", "updated_at"}).
				AddRow(
					testUser.ID.String(),
					testUser.Name,
					testUser.Password,
					testUser.Email,
					testUser.Avatar,
					testUser.Address,
					testUser.Phone,
					testUser.CreatedAt,
					testUser.UpdatedAt,
				))

		result, err := repo.GetUserByID(context.Background(), testUser.ID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testUser.ID, result.ID)
		assert.Equal(t, testUser.Name, result.Name)
		assert.Equal(t, testUser.Email, result.Email)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("User not found case", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WithArgs(testUser.ID.String()).
			WillReturnError(rrepo.ErrNotFound)

		result, err := repo.GetUserByID(context.Background(), testUser.ID)
		assert.Nil(t, result)
		assert.Equal(t, rrepo.ErrNotFound, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	var notExpectedError = errors.New("not expected error")
	t.Run("Unexpected error case", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WithArgs(testUser.ID.String()).
			WillReturnError(notExpectedError)

		result, err := repo.GetUserByID(context.Background(), testUser.ID)
		assert.Nil(t, result)
		assert.Equal(t, notExpectedError, err)
		assert.NotEqual(t, rrepo.ErrNotFound, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestGetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := Repository{conn: db}

	const q = `SELECT id, name, password, email, avatar, address, phone, created_at, updated_at
		FROM users
		WHERE email = $1`
	testUser := md.User{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "testuser@example.com",
	}

	t.Run("Success case", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WithArgs(testUser.Email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "password", "email", "avatar", "address", "phone", "created_at", "updated_at"}).
				AddRow(
					testUser.ID.String(),
					testUser.Name,
					testUser.Password,
					testUser.Email,
					testUser.Avatar,
					testUser.Address,
					testUser.Phone,
					testUser.CreatedAt,
					testUser.UpdatedAt,
				))

		result, err := repo.GetUserByEmail(context.Background(), testUser.Email)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testUser.ID, result.ID)
		assert.Equal(t, testUser.Name, result.Name)
		assert.Equal(t, testUser.Email, result.Email)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("User not found case", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WithArgs(testUser.Email).
			WillReturnError(rrepo.ErrNotFound)

		result, err := repo.GetUserByEmail(context.Background(), testUser.Email)
		assert.Nil(t, result)
		assert.Equal(t, rrepo.ErrNotFound, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	var notExpectedError = errors.New("not expected error")
	t.Run("Unexpected error case", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WithArgs(testUser.Email).
			WillReturnError(notExpectedError)

		result, err := repo.GetUserByEmail(context.Background(), testUser.Email)
		assert.Nil(t, result)
		assert.Equal(t, notExpectedError, err)
		assert.NotEqual(t, rrepo.ErrNotFound, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := Repository{conn: db}

	const selectQ = `SELECT id FROM users WHERE id=$1 OR email=$2`
	const q = `INSERT INTO users (name, password, email, avatar, address, phone) 
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`
	testUser := &md.User{
		ID:       uuid.New(),
		Name:     "Test User",
		Email:    "testuser@example.com",
		Password: "securepassword",
	}

	t.Run("Success case", func(t *testing.T) {

		mock.ExpectQuery(regexp.QuoteMeta(selectQ)).
			WithArgs(testUser.ID.String(), testUser.Email).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectQuery(
			regexp.QuoteMeta(q)).
			WillReturnRows(
				sqlmock.NewRows([]string{"id"}).
					AddRow(testUser.ID.String()))

		_, err := repo.CreateUser(context.Background(), testUser)
		assert.NoError(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Password is required", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(selectQ)).
			WithArgs(testUser.ID.String(), testUser.Email).
			WillReturnError(sql.ErrNoRows)

		testUser.Password = ""

		_, err := repo.CreateUser(context.Background(), testUser)
		assert.Equal(t, rrepo.ErrPasswordIsRequired, err)
	})

	t.Run("Username is required", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(selectQ)).
			WithArgs(testUser.ID.String(), testUser.Email).
			WillReturnError(sql.ErrNoRows)

		testUser.Password = "securepassword"
		testUser.Name = ""

		_, err := repo.CreateUser(context.Background(), testUser)
		assert.Equal(t, rrepo.ErrUsernameIsRequired, err)
	})

	t.Run("Email is required", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(selectQ)).
			WithArgs(testUser.ID.String(), "").
			WillReturnError(sql.ErrNoRows)

		testUser.Name = "Test User"
		testUser.Email = ""

		_, err := repo.CreateUser(context.Background(), testUser)
		assert.Equal(t, rrepo.ErrEmailIsRequired, err)
	})

	t.Run("Error generating password", func(t *testing.T) {
		testUser.Email = "testuser@example.com"
		mock.ExpectQuery(regexp.QuoteMeta(selectQ)).
			WithArgs(testUser.ID.String(), testUser.Email).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WillReturnError(rrepo.ErrGeneratingPassword)

		_, err := repo.CreateUser(context.Background(), testUser)
		assert.Equal(t, rrepo.ErrGeneratingPassword, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("User already exists", func(t *testing.T) {
		testUser.Email = "existinguser@example.com"
		mock.ExpectQuery(regexp.QuoteMeta(selectQ)).
			WithArgs(testUser.ID.String(), testUser.Email).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WillReturnError(rrepo.ErrAlreadyExists)

		_, err := repo.CreateUser(context.Background(), testUser)
		assert.Equal(t, rrepo.ErrAlreadyExists, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestUpdateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := Repository{conn: db}

	testUserID := uuid.New()
	testUser := &md.User{
		ID:       testUserID,
		Name:     "Original User",
		Email:    "originaluser@example.com",
		Password: "securepassword",
	}

	const insertQ = `INSERT INTO users (name, password, email, avatar, address, phone) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	const selectQ = `SELECT id FROM users WHERE id=$1 OR email=$2`

	mock.ExpectQuery(regexp.QuoteMeta(selectQ)).
		WithArgs(testUser.ID.String(), testUser.Email).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(regexp.QuoteMeta(insertQ)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID.String()))

	_, err = repo.CreateUser(context.Background(), testUser)
	require.NoError(t, err)

	newData := &md.User{
		Name:     "Updated User",
		Password: "newsecurepassword",
		Email:    "updateduser@example.com",
		Avatar:   "new_avatar.png",
		Address:  "123 Updated St.",
		Phone:    "123-456-7890",
	}

	t.Run("Success case", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE users 
		 SET name = $1, password = $2, email = $3, avatar = $4, address = $5, phone = $6 
		 WHERE id = $7`)).
			WithArgs(
				newData.Name,
				newData.Password,
				newData.Email,
				newData.Avatar,
				newData.Address,
				newData.Phone,
				testUserID.String(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.UpdateUser(context.Background(), testUserID, newData)
		assert.NoError(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE users 
         SET name = $1, password = $2, email = $3, avatar = $4, address = $5, phone = $6 
         WHERE id = $7`)).
			WithArgs(
				newData.Name,
				newData.Password,
				newData.Email,
				newData.Avatar,
				newData.Address,
				newData.Phone,
				testUserID.String(),
			).
			WillReturnResult(sqlmock.NewResult(1, 0))

		err := repo.UpdateUser(context.Background(), testUserID, newData)
		assert.ErrorIs(t, err, rrepo.ErrNotFound)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestDeleteUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := Repository{conn: db}

	testUserID := uuid.New()
	const deleteQ = `DELETE FROM users WHERE id = $1`

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(deleteQ)).
			WithArgs(testUserID.String()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteUser(context.Background(), testUserID)
		assert.NoError(t, err)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("User not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(deleteQ)).
			WithArgs(testUserID.String()).
			WillReturnResult(sqlmock.NewResult(1, 0))

		err := repo.DeleteUser(context.Background(), testUserID)
		assert.ErrorIs(t, err, rrepo.ErrNotFound)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Database Error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(deleteQ)).
			WithArgs(testUserID.String()).
			WillReturnError(errors.New("db error"))

		err := repo.DeleteUser(context.Background(), testUserID)
		assert.Error(t, err)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

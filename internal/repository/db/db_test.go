package db

import (
	"context"
	"errors"
	rrepo "github.com/JMURv/sso/internal/repository"
	md "github.com/JMURv/sso/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"regexp"
	"testing"
)

func TestUserSearch(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	dialector := postgres.New(postgres.Config{
		Conn: mockDB,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	repo := Repository{conn: db}

	const countQ = `^SELECT count(.+) FROM "users"`
	const q = `^SELECT \* FROM "users" WHERE`
	query := "test"
	page := 1
	size := 10
	expectedCount := int64(2)
	expectedTotalPages := int((expectedCount + int64(size) - 1) / int64(size))

	mockUsers := []md.User{
		{ID: uuid.New(), Name: "John Doe", Email: "johndoe@example.com"},
		{ID: uuid.New(), Name: "Jane Smith", Email: "janesmith@example.com"},
	}

	t.Run("Success case", func(t *testing.T) {
		mock.ExpectQuery(countQ).
			WithArgs("%"+query+"%", "%"+query+"%").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		rows := sqlmock.NewRows([]string{"id", "name", "email"}).
			AddRow(mockUsers[0].ID.String(), mockUsers[0].Name, mockUsers[0].Email).
			AddRow(mockUsers[1].ID.String(), mockUsers[1].Name, mockUsers[1].Email)

		mock.ExpectQuery(q).
			WithArgs("%"+query+"%", "%"+query+"%", size).
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
		mock.ExpectQuery(countQ).
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
		mock.ExpectQuery(countQ).
			WithArgs("%"+query+"%", "%"+query+"%").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		mock.ExpectQuery(q).
			WithArgs("%"+query+"%", "%"+query+"%", size).
			WillReturnError(errors.New("find error"))

		resp, err := repo.UserSearch(context.Background(), query, page, size)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "find error", err.Error())

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Empty result", func(t *testing.T) {
		mock.ExpectQuery(countQ).
			WithArgs("%"+query+"%", "%"+query+"%").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		mock.ExpectQuery(q).
			WithArgs("%"+query+"%", "%"+query+"%", size).
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
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	dialector := postgres.New(postgres.Config{
		Conn: mockDB,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	repo := Repository{conn: db}
	const countQ = `^SELECT count(.+) FROM "users"`
	const q = `^SELECT \* FROM "users" ORDER BY created_at desc LIMIT`

	page := 1
	size := 10
	expectedCount := int64(20)
	expectedTotalPages := int((expectedCount + int64(size) - 1) / int64(size))

	mockUsers := []md.User{
		{ID: uuid.New(), Name: "John Doe", Email: "johndoe@example.com"},
		{ID: uuid.New(), Name: "Jane Smith", Email: "janesmith@example.com"},
	}

	t.Run("Success case", func(t *testing.T) {
		mock.ExpectQuery(countQ).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		rows := sqlmock.NewRows([]string{"id", "name", "email"}).
			AddRow(mockUsers[0].ID.String(), mockUsers[0].Name, mockUsers[0].Email).
			AddRow(mockUsers[1].ID.String(), mockUsers[1].Name, mockUsers[1].Email)

		mock.ExpectQuery(q).
			WillReturnRows(rows)

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
		mock.ExpectQuery(countQ).
			WillReturnError(errors.New("count error"))

		resp, err := repo.ListUsers(context.Background(), page, size)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "count error", err.Error())

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Find error", func(t *testing.T) {
		mock.ExpectQuery(countQ).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		mock.ExpectQuery(q).
			WillReturnError(errors.New("find error"))

		resp, err := repo.ListUsers(context.Background(), page, size)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "find error", err.Error())

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Empty users", func(t *testing.T) {
		mock.ExpectQuery(countQ).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		mock.ExpectQuery(q).
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
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	dialector := postgres.New(postgres.Config{
		Conn: mockDB,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	repo := Repository{conn: db}

	const q = `^SELECT \* FROM "users" WHERE id=\$1 ORDER BY "users"."id" LIMIT \$2`
	testUser := md.User{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "testuser@example.com",
	}

	t.Run("Success case", func(t *testing.T) {
		mock.ExpectQuery(q).
			WithArgs(testUser.ID.String(), 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(testUser.ID.String(), testUser.Name, testUser.Email))

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
		mock.ExpectQuery(q).
			WithArgs(testUser.ID.String(), 1).
			WillReturnError(gorm.ErrRecordNotFound)

		result, err := repo.GetUserByID(context.Background(), testUser.ID)
		assert.Nil(t, result)
		assert.Equal(t, rrepo.ErrNotFound, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	var notExpectedError = errors.New("not expected error")
	t.Run("Unexpected error case", func(t *testing.T) {
		mock.ExpectQuery(q).
			WithArgs(testUser.ID.String(), 1).
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
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	dialector := postgres.New(postgres.Config{
		Conn: mockDB,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	repo := Repository{conn: db}

	const q = `^SELECT \* FROM "users" WHERE email=\$1 ORDER BY "users"."id" LIMIT \$2`
	testUser := md.User{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "testuser@example.com",
	}

	t.Run("Success case", func(t *testing.T) {
		mock.ExpectQuery(q).
			WithArgs(testUser.Email, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(testUser.ID.String(), testUser.Name, testUser.Email))

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
		mock.ExpectQuery(q).
			WithArgs(testUser.Email, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		result, err := repo.GetUserByEmail(context.Background(), testUser.Email)
		assert.Nil(t, result)
		assert.Equal(t, rrepo.ErrNotFound, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	var notExpectedError = errors.New("not expected error")
	t.Run("Unexpected error case", func(t *testing.T) {
		mock.ExpectQuery(q).
			WithArgs(testUser.Email, 1).
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
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	dialector := postgres.New(postgres.Config{
		Conn: mockDB,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	repo := Repository{conn: db}

	const q = `INSERT INTO "users"`
	testUser := &md.User{
		ID:       uuid.New(),
		Name:     "Test User",
		Email:    "testuser@example.com",
		Password: "securepassword",
	}

	t.Run("Success case", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery(
			regexp.QuoteMeta(q)).
			WillReturnRows(
				sqlmock.NewRows([]string{"id"}).
					AddRow(testUser.ID.String()))

		mock.ExpectCommit()

		err := repo.CreateUser(context.Background(), testUser)
		assert.NoError(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Password is required", func(t *testing.T) {
		testUser.Password = ""

		err := repo.CreateUser(context.Background(), testUser)
		assert.Equal(t, rrepo.ErrPasswordIsRequired, err)
	})

	t.Run("Username is required", func(t *testing.T) {
		testUser.Password = "securepassword"
		testUser.Name = ""

		err := repo.CreateUser(context.Background(), testUser)
		assert.Equal(t, rrepo.ErrUsernameIsRequired, err)
	})

	t.Run("Email is required", func(t *testing.T) {
		testUser.Name = "Test User"
		testUser.Email = ""

		err := repo.CreateUser(context.Background(), testUser)
		assert.Equal(t, rrepo.ErrEmailIsRequired, err)
	})

	t.Run("Error generating password", func(t *testing.T) {
		testUser.Email = "testuser@example.com"
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WillReturnError(rrepo.ErrGeneratingPassword)

		mock.ExpectRollback()

		err := repo.CreateUser(context.Background(), testUser)
		assert.Equal(t, rrepo.ErrGeneratingPassword, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("User already exists", func(t *testing.T) {
		testUser.Email = "existinguser@example.com"

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
			WillReturnError(gorm.ErrDuplicatedKey)
		mock.ExpectRollback()

		err := repo.CreateUser(context.Background(), testUser)
		assert.Equal(t, rrepo.ErrAlreadyExists, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestUpdateUser(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	dialector := postgres.New(postgres.Config{
		Conn: mockDB,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	repo := Repository{conn: db}

	testUserID := uuid.New()
	testUser := &md.User{
		ID:       testUserID,
		Name:     "Original User",
		Email:    "originaluser@example.com",
		Password: "securepassword",
	}

	// Insert a user to update later
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID.String()))
	mock.ExpectCommit()

	err = repo.CreateUser(context.Background(), testUser)
	require.NoError(t, err)

	newData := &md.User{
		Name:    "Updated User",
		Email:   "updateduser@example.com",
		Avatar:  "new_avatar.png",
		Address: "123 Updated St.",
		Phone:   "123-456-7890",
		IsAdmin: true,
		IsOpt:   true,
	}

	t.Run("Success case", func(t *testing.T) {
		mock.ExpectQuery(`^SELECT \* FROM "users" WHERE id=\$1`).
			WithArgs(testUserID.String(), 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "avatar", "address", "phone", "is_opt", "is_admin"}).
				AddRow(testUser.ID.String(), testUser.Name, testUser.Email, "", "", "", false, false))

		//mock.ExpectQuery(regexp.QuoteMeta(`UPDATE "users" SET`)).
		//	WithArgs(newData.Name, newData.Email, newData.Avatar, newData.Address, newData.Phone, newData.IsOpt, newData.IsAdmin, testUserID.String()).
		//	WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "avatar", "address", "phone", "is_opt", "is_admin"}).
		//		AddRow(testUser.ID.String(), testUser.Name, testUser.Email, "", "", "", false, false))

		updatedUser, err := repo.UpdateUser(context.Background(), testUserID, newData)
		assert.NoError(t, err)
		assert.Equal(t, newData.Name, updatedUser.Name)
		assert.Equal(t, newData.Email, updatedUser.Email)
		assert.Equal(t, newData.Avatar, updatedUser.Avatar)
		assert.Equal(t, newData.Address, updatedUser.Address)
		assert.Equal(t, newData.Phone, updatedUser.Phone)
		assert.Equal(t, newData.IsAdmin, updatedUser.IsAdmin)
		assert.Equal(t, newData.IsOpt, updatedUser.IsOpt)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("User not found case", func(t *testing.T) {
		mock.ExpectQuery(`^SELECT \* FROM "users" WHERE id=\$1`).
			WithArgs(testUserID.String(), 1).
			WillReturnError(gorm.ErrRecordNotFound)

		updatedUser, err := repo.UpdateUser(context.Background(), testUserID, newData)
		assert.Nil(t, updatedUser)
		assert.Equal(t, rrepo.ErrNotFound, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestDeleteUser(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	dialector := postgres.New(postgres.Config{
		Conn: mockDB,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	repo := Repository{conn: db}

	testUserID := uuid.New()

	t.Run("Success case", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "users"`)).
			WithArgs(testUserID.String()).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.DeleteUser(context.Background(), testUserID)
		assert.NoError(t, err)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("User not found case", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`^DELETE FROM "users"`).
			WithArgs(testUserID.String()).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.DeleteUser(context.Background(), testUserID)
		assert.NoError(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

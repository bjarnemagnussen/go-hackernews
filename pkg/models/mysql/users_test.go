package mysql_test

import (
	"testing"

	"github.com/bjarnemagnussen/go-hackernews/pkg/models"
	"github.com/bjarnemagnussen/go-hackernews/pkg/models/mysql"
	"github.com/bjarnemagnussen/go-hackernews/pkg/models/mysql/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) *models.User {
	arg := mysql.CreateUserParams{
		Username: util.RandomUsername(),
		Email:    util.RandomEmail(),
		Password: "password123",
	}

	id, err := userModel.Insert(arg)
	require.NoError(t, err)
	require.NotZero(t, id)

	user, err := userModel.GetByID(id)
	require.NoError(t, err)

	require.NotZero(t, user.ID)
	require.NotZero(t, user.Created)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, true, user.Active)

	return user
}

func TestUser_Insert(t *testing.T) {
	id, err := userModel.Insert(mysql.CreateUserParams{
		Username: "admin",
		Email:    "email@address.com",
		Password: "password123",
	})
	require.NoError(t, err)
	require.NotZero(t, id)
}

func TestUser_InsertDuplicates(t *testing.T) {
	_, err := userModel.Insert(mysql.CreateUserParams{
		Username: "dupe",
		Email:    "email@address.com",
		Password: "password123",
	})
	require.Error(t, err, models.ErrDuplicateEmail)

	_, err = userModel.Insert(mysql.CreateUserParams{
		Username: "Admin",
		Email:    "alternative@address.com",
		Password: "password123",
	})
	require.Error(t, err, models.ErrDuplicateUsername)
}

func TestGet(t *testing.T) {
	createRandomUser(t)
}

func TestUser_Update(t *testing.T) {
	user := createRandomUser(t)
	arg := mysql.UpdateUserParams{
		ID:    user.ID,
		About: "about me page 404 not found!",
	}

	err := userModel.Update(arg)
	require.NoError(t, err)

	got, err := userModel.GetByID(user.ID)
	require.NoError(t, err)
	require.NotEmpty(t, got)

	require.Equal(t, user.ID, got.ID)
	require.Equal(t, user.Username, got.Username)
	require.Equal(t, arg.About, got.About)
}
func TestUser_Authenticate(t *testing.T) {
	user := createRandomUser(t)

	id, err := userModel.Authenticate(user.Username, "password123")
	require.NoError(t, err)
	require.Equal(t, user.ID, id)

	_, err = userModel.Authenticate(user.Username, "wrongPassword")
	require.Error(t, err, models.ErrInvalidCredentials)
}

func TestUser_GetFromName(t *testing.T) {
	user := createRandomUser(t)

	id, err := userModel.GetIDByUsername(user.Username)
	require.NoError(t, err)
	require.Equal(t, user.ID, id)
}
func TestUser_GetFull(t *testing.T) {
	user := createRandomUser(t)

	got, err := userModel.GetFull(user.ID)
	require.NoError(t, err)
	require.NotEmpty(t, got)
	require.Equal(t, user.ID, got.ID)
	require.Equal(t, user.Username, got.Username)
	require.Equal(t, user.Email, got.Email)
	require.Equal(t, user.Active, got.Active)
	require.Equal(t, user.Created, got.Created)
	// TODO: add karma to test if it calculates correctly!
	require.Equal(t, 0, got.Karma)
}

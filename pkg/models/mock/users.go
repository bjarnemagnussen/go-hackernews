package mock

import (
	"time"

	"github.com/bjarnemagnussen/go-hackernews/pkg/models"
	"github.com/bjarnemagnussen/go-hackernews/pkg/models/mysql"
)

var mockUser = &models.User{
	ID:       1,
	Username: "Alice",
	About:    "",
	Email:    "alice@example.com",
	Created:  time.Now().UTC(),
	Active:   true,
}

type UserModel struct{}

func (m *UserModel) Insert(arg mysql.CreateUserParams) (int, error) {
	switch arg.Email {
	case "dupe@example.com":
		return 0, models.ErrDuplicateEmail
	default:
		return 2, nil
	}
}
func (m *UserModel) Update(arg mysql.UpdateUserParams) error {
	return nil
}

func (m *UserModel) Authenticate(username, password string) (int, error) {
	switch username {
	case "alice":
		return 1, nil
	default:
		return 0, models.ErrInvalidCredentials
	}
}

func (m *UserModel) GetByID(id int) (*models.User, error) {
	switch id {
	case 1:
		return mockUser, nil
	default:
		return nil, models.ErrNoRecord
	}
}

func (m *UserModel) GetIDByUsername(name string) (int, error) {
	switch name {
	case "alice":
		return 1, nil
	default:
		return 0, models.ErrNoRecord
	}
}

func (m *UserModel) GetFull(id int) (*models.User, error) {
	switch id {
	case 1:
		mockUser := mockUser
		mockUser.Karma = 100
		return mockUser, nil
	default:
		return nil, models.ErrNoRecord
	}
}

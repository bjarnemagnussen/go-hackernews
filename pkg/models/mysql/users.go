package mysql

import (
	"database/sql"
	"strings"

	"github.com/bjarnemagnussen/go-hackernews/pkg/models"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	DB *sql.DB
}

type CreateUserParams struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserParams struct {
	ID    int    `json:"id"`
	About string `json:"about_me"`
}

func (m *UserModel) Insert(arg CreateUserParams) (int, error) {
	stmt := `INSERT INTO users (username, username_unique, email, password_hash, created)
	VALUES (?, ?, ?, ?, UTC_TIMESTAMP())`

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(arg.Password), 12)
	if err != nil {
		return 0, err
	}
	usernameUnique := strings.ToLower(strings.Replace(arg.Username, ".", "", -1))

	result, err := m.DB.Exec(stmt, arg.Username, usernameUnique, strings.ToLower(arg.Email), string(hashedPassword))
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 {
				switch {
				case strings.Contains(mysqlErr.Message, "uc_users_email"):
					return 0, models.ErrDuplicateEmail
				case strings.Contains(mysqlErr.Message, "uc_users_name"):
					return 0, models.ErrDuplicateUsername
				}
			}
		}
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *UserModel) Update(arg UpdateUserParams) error {
	stmt := `UPDATE users SET about_me = ? WHERE id = ?`
	_, err := m.DB.Exec(stmt, arg.About, arg.ID)
	if err != nil {
		return err
	}
	return nil
}

func (m *UserModel) Authenticate(username, password string) (int, error) {
	var id int
	var hashedPassword []byte
	var migrate bool
	stmt := `SELECT id, password_hash, migrate FROM users WHERE username = ? AND active = TRUE`
	row := m.DB.QueryRow(stmt, username)
	err := row.Scan(&id, &hashedPassword, &migrate)
	if err == sql.ErrNoRows {
		return 0, models.ErrInvalidCredentials
	} else if err != nil {
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, models.ErrInvalidCredentials
	} else if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *UserModel) GetByID(id int) (*models.User, error) {
	u := &models.User{}

	stmt := `SELECT id, username, email, about_me, created, active FROM users WHERE id = ?`
	err := m.DB.QueryRow(stmt, id).Scan(&u.ID, &u.Username, &u.Email, &u.About, &u.Created, &u.Active)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}

	return u, nil
}

func (m *UserModel) GetIDByUsername(name string) (int, error) {
	var id int

	stmt := `SELECT id FROM users WHERE username = ?`
	err := m.DB.QueryRow(stmt, name).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, models.ErrNoRecord
	} else if err != nil {
		return 0, err
	}
	// TODO: Better workflow in handler possible?
	return id, nil
}

func (m *UserModel) GetFull(id int) (*models.User, error) {
	u := &models.User{}

	stmt := `WITH 
		karma_posts (user_id, karma) AS (
			SELECT 
				p.user_id,
				COUNT(v.id) AS karma
			FROM posts AS p
			LEFT JOIN votes AS v ON v.post_id = p.id
			WHERE p.user_id = ?
			GROUP BY p.user_id
		), karma_comments (user_id, karma) AS (
			SELECT 
				c.user_id,
				COUNT(v.id) AS karma
			FROM comments AS c
			LEFT JOIN comment_votes AS v ON v.comment_id = c.id
			WHERE c.user_id = ?
			GROUP BY c.user_id
		)
		SELECT u.id, u.username, u.email, u.about_me, u.created, u.active, IFNULL(pv.karma, 0)+IFNULL(cv.karma, 0) AS karma
		FROM users AS u
		LEFT JOIN karma_posts AS pv ON pv.user_id = u.id
		LEFT JOIN karma_comments AS cv ON cv.user_id = u.id
		WHERE u.id = ?`

	err := m.DB.QueryRow(stmt, id, id, id).Scan(&u.ID, &u.Username, &u.Email, &u.About, &u.Created, &u.Active, &u.Karma)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}

	return u, nil
}

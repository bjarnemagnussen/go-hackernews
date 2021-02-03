package mock

import (
	"database/sql"
	"time"

	"github.com/bjarnemagnussen/go-hackernews/pkg/models"
	"github.com/bjarnemagnussen/go-hackernews/pkg/models/mysql"
)

var mockPostURL = &models.Post{
	ID:        1,
	Username:  "user123",
	Title:     "Post Test Title",
	URLScheme: sql.NullString{"https", true},
	URLBase:   sql.NullString{"google.com", true},
	URI:       sql.NullString{"google.com/ref/some_link", true},
	Content:   sql.NullString{"", false},
	Created:   time.Now().UTC().Add(-2 * time.Hour),
	Deleted:   0,

	Comments:  5,
	Votes:     10,
	Retrieved: time.Now().UTC(),
}

type PostModel struct{}

const (
	StandardPost = iota
	AskPost
	ShowPost
)

func (m *PostModel) Insert(arg mysql.CreatePostParams) (int, error) {
	return 2, nil
}

func (m *PostModel) GetByID(loggedInUserID, id int) (*models.Post, error) {
	switch id {
	case 1:
		return mockPostURL, nil
	default:
		return nil, models.ErrNoRecord
	}
}

func (m *PostModel) GetIDFromURL(s string) (int, error) {
	switch s {
	case "https://google.com/ref/some_link":
		return 1, nil
	default:
		return 0, models.ErrNoRecord
	}
}

func (m *PostModel) GetFromDomain(loggedInUserID, lastID int, domain string) ([]*models.Post, error) {
	switch domain {
	case "google.com":
		return []*models.Post{mockPostURL}, nil
	default:
		return nil, models.ErrNoRecord
	}
}

func (m *PostModel) GetForType(loggedInUserID, lastID int, popoular bool, postType mysql.PostType) ([]*models.Post, error) {
	return nil, nil
}

func (m *PostModel) Latest(loggedInUserID, lastPostID, filterByUserID int) ([]*models.Post, error) {
	return []*models.Post{mockPostURL}, nil
}

func (m *PostModel) Popular(loggedInUserID, offset int) ([]*models.Post, error) {
	return []*models.Post{mockPostURL}, nil
}

func (m *PostModel) Upvote(postID, userID int, created time.Time) (int, error) {
	return 1, nil
}

package mock

import (
	"database/sql"
	"time"

	"github.com/bjarnemagnussen/go-hackernews/pkg/models"
	"github.com/bjarnemagnussen/go-hackernews/pkg/models/mysql"
)

var mockCommentParent = &models.Comment{
	ID:        1,
	Username:  "alice",
	Content:   "This is a test comment.",
	Created:   time.Now().UTC(),
	ParentID:  sql.NullInt64{0, false},
	UserID:    1,
	PostID:    1,
	PostTitle: "Post Test Title",
	Leaf:      0,
	Deleted:   0,
}

var mockCommentChild = &models.Comment{
	ID:        2,
	Username:  "alice",
	Content:   "This is a test reply.",
	Created:   time.Now().UTC(),
	ParentID:  sql.NullInt64{1, true},
	UserID:    1,
	PostID:    1,
	PostTitle: "Post Test Title",
	Leaf:      1,
	Deleted:   0,
}

type CommentModel struct{}

func (m *CommentModel) Insert(arg mysql.CreateCommentParams) (int, error) {
	return 3, nil
}

func (m *CommentModel) GetByID(loggedInUserID, id int) (*models.Comment, error) {
	switch id {
	case 1:
		return mockCommentParent, nil
	case 2:
		return mockCommentChild, nil
	default:
		return nil, models.ErrNoRecord
	}
}

func (m *CommentModel) ForPost(loggedInUserID, postID int) ([]*models.Comment, error) {
	switch postID {
	case 1:
		return []*models.Comment{mockCommentParent, mockCommentChild}, nil
	default:
		return nil, nil
	}
}

func (m *CommentModel) ForUser(loggedInUserID, lastCommentID, id int) ([]*models.Comment, error) {
	switch id {
	case 1:
		return []*models.Comment{mockCommentParent, mockCommentChild}, nil
	default:
		return nil, nil
	}
}

func (m *CommentModel) ForComment(userID, postID, commentID int) ([]*models.Comment, error) {
	switch commentID {
	case 1:
		return []*models.Comment{mockCommentChild}, nil
	default:
		return nil, nil
	}
}

func (m *CommentModel) Latest(userId, lastId int) ([]*models.Comment, error) {
	return []*models.Comment{mockCommentChild, mockCommentParent}, nil
}

func (m *CommentModel) Upvote(commentID, userID int, created time.Time) (int, error) {
	return 0, nil
}

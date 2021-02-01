package models

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNoRecord           = errors.New("models: no matching record found")
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrDuplicatePost      = errors.New("models: duplicate entry")
	ErrDuplicateUpvote    = errors.New("models: duplicate entry")
	ErrDuplicateEmail     = errors.New("models: duplicate email")
	ErrDuplicateUsername  = errors.New("models: duplicate username")
)

type Post struct {
	ID        int
	Username  string
	Title     string
	URLScheme sql.NullString
	URLBase   sql.NullString
	URI       sql.NullString
	Content   sql.NullString
	Created   time.Time
	Deleted   int

	// Comments contains the number of comments this post has.
	Comments int

	// Votes contains total votes.
	Votes int

	// Votable defines if a user can vote on the post or not
	Votable int

	// Retrieved is used to calculate the time difference from Created.
	Retrieved time.Time
}

type Comment struct {
	ID        int
	Username  string
	Content   string
	Created   time.Time
	ParentID  sql.NullInt64
	UserID    int
	PostID    int
	PostTitle string
	Leaf      int
	Deleted   int

	// The following denotes values that are calculated on-the-fly using CTE.

	// Votes contains total votes.
	Votes int

	// Level denotes how many ancestors this comment has.
	Level int

	// Votable defines if a user can vote on the comment or not
	Votable int

	// Retrieved is used to calculate the time difference from Created.
	Retrieved time.Time
}

type User struct {
	ID             int
	Username       string
	About          string
	Email          string
	HashedPassword []byte
	Created        time.Time
	Active         bool

	// Is a count of all upvotes on the users posts and comments.
	Karma int
}

package mysql

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/bjarnemagnussen/go-hackernews/pkg/models"
	"github.com/bjarnemagnussen/go-hackernews/pkg/models/mysql/util"
	"github.com/go-sql-driver/mysql"
)

type PostModel struct {
	DB *sql.DB
}

type CreatePostParams struct {
	UserID  int       `json:"user_id"`
	Title   string    `json:"title"`
	Url     string    `json:"url"`
	Content string    `json:"content"`
	Created time.Time `json:"created"`
}

type PostType int

const (
	StandardPost = iota
	AskPost
	ShowPost
)

func (m *PostModel) Insert(arg CreatePostParams) (int, error) {

	stmt := `INSERT INTO posts (user_id, title, url_scheme, url_base, uri, content, created, post_type)
	VALUES(?, ?, ?, ?, ?, ?, ?, ?)`

	var uri, urlScheme, urlBase sql.NullString
	content := sql.NullString{arg.Content, true}
	var postType PostType

	if arg.Url != "" {
		// Post is a link.
		var err error
		uri.String, urlScheme.String, urlBase.String, err = util.ExtractURL(arg.Url)
		if err != nil {
			return 0, err
		}
		uri.Valid, urlScheme.Valid, urlBase.Valid = true, true, true

		// Ignore content if URL was set.
		content = sql.NullString{"", false}

		// Cleanup title.
		if strings.HasSuffix(uri.String, ".pdf") && !strings.HasSuffix(arg.Title, "[pdf]") {
			arg.Title += " [pdf]"
		}
		arg.Title = strings.TrimPrefix(strings.TrimPrefix(arg.Title, "Show Us:"), "Ask Us:")

	} else {
		// Post is content.
		switch {
		case strings.HasPrefix(arg.Title, "Show Us:"):
			postType = ShowPost
		case strings.HasPrefix(arg.Title, "Ask Us:"):
			postType = AskPost
		case strings.HasSuffix(arg.Title, "?"):
			postType = AskPost
			arg.Title = "Ask Us: " + arg.Title
		}
	}

	result, err := m.DB.Exec(stmt, arg.UserID, arg.Title, urlScheme, urlBase, uri, content, arg.Created, postType)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "uc_posts_uri") {
				return 0, models.ErrDuplicatePost
			}
		} else {
			return 0, err
		}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *PostModel) GetByID(loggedInUserID, id int) (*models.Post, error) {

	stmt := `SELECT p.id, u.username, p.title, p.url_scheme, p.url_base, p.uri, p.content, p.created,
	COUNT(DISTINCT v.id) AS votes,
	IF(p.user_id = ?, 2, NOT EXISTS (SELECT * FROM votes WHERE user_id = ? AND post_id = p.id)) AS votable,
	COUNT(DISTINCT c.id) AS comments
	FROM posts AS p
	LEFT JOIN votes AS v ON p.id = v.post_id
	LEFT JOIN comments AS c ON p.id = c.post_id
	JOIN users AS u ON p.user_id = u.id
	WHERE p.id = ? AND p.deleted = 0
	GROUP BY p.id`

	p := &models.Post{}
	p.Retrieved = time.Now()

	err := m.DB.QueryRow(stmt, loggedInUserID, loggedInUserID, id).Scan(
		&p.ID, &p.Username, &p.Title, &p.URLScheme, &p.URLBase, &p.URI, &p.Content, &p.Created, &p.Votes,
		&p.Votable, &p.Comments,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}

	return p, nil
}

func (m *PostModel) GetIDFromURL(s string) (int, error) {

	stmt := `SELECT id FROM posts WHERE uri = ?`

	p := &models.Post{}

	u, err := url.Parse(s)
	if err != nil {
		return 0, err
	}

	// TODO: Regarding ports: how to handle 80/443, see inserting above
	err = m.DB.QueryRow(stmt, u.Host+strings.TrimRight(u.RequestURI(), "/")).Scan(&p.ID)
	if err == sql.ErrNoRows {
		return 0, models.ErrNoRecord
	} else if err != nil {
		return 0, err
	}

	return p.ID, nil
}

func (m *PostModel) GetFromDomain(loggedInUserID, lastID int, domain string) ([]*models.Post, error) {
	stmt := `SELECT p.id, u.username, p.title, p.url_scheme, p.url_base, p.uri, p.created,
		COUNT(DISTINCT v.id) AS votes,
		IF(p.user_id = ?, 2, NOT EXISTS (SELECT * FROM votes WHERE user_id = ? AND post_id = p.id)) AS votable,
		COUNT(DISTINCT c.id) AS comments
	FROM posts AS p
	LEFT JOIN votes AS v ON p.id = v.post_id
	LEFT JOIN comments AS c ON p.id = c.post_id
	JOIN users AS u ON p.user_id = u.id
	WHERE p.url_base = ? AND p.id >= ? AND p.deleted = 0
	GROUP BY p.id
	ORDER BY p.created DESC, p.id DESC
	LIMIT 31`

	now := time.Now()

	rows, err := m.DB.Query(stmt, loggedInUserID, loggedInUserID, domain, lastID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []*models.Post{}

	for rows.Next() {
		p := &models.Post{}

		err := rows.Scan(
			&p.ID, &p.Username, &p.Title, &p.URLScheme, &p.URLBase, &p.URI, &p.Created, &p.Votes,
			&p.Votable, &p.Comments,
		)
		if err != nil {
			return nil, err
		}

		p.Retrieved = now

		posts = append(posts, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (m *PostModel) GetForType(loggedInUserID, offset int, postType PostType) ([]*models.Post, error) {

	stmt := `SELECT p.id, u.username, p.title, p.url_scheme, p.url_base, p.uri, p.created,
		COUNT(DISTINCT v.id) AS votes,
		IF(p.user_id = ?, 2, NOT EXISTS (SELECT * FROM votes WHERE user_id = ? AND post_id = p.id)) AS votable,
		COUNT(DISTINCT c.id) AS comments
	FROM posts AS p
	LEFT JOIN votes AS v ON p.id = v.post_id
	LEFT JOIN comments AS c ON p.id = c.post_id
	JOIN users AS u ON p.user_id = u.id
	WHERE p.post_type = ? AND p.deleted = 0
	GROUP BY p.id
	ORDER BY (COUNT(DISTINCT v.id)) / POW((TIMESTAMPDIFF(HOUR, p.created, NOW()) + 2), 1.2) DESC, p.created DESC, p.ID DESC
	LIMIT ?, 31`

	now := time.Now()

	var err error
	var rows *sql.Rows
	rows, err = m.DB.Query(stmt, loggedInUserID, loggedInUserID, postType, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []*models.Post{}

	for rows.Next() {
		p := &models.Post{}

		err := rows.Scan(
			&p.ID, &p.Username, &p.Title, &p.URLScheme, &p.URLBase, &p.URI, &p.Created, &p.Votes,
			&p.Votable, &p.Comments,
		)
		if err != nil {
			return nil, err
		}

		p.Retrieved = now

		posts = append(posts, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (m *PostModel) Latest(loggedInUserID, lastPostID, filterByUserID int) ([]*models.Post, error) {

	stmt := `SELECT p.id, u.username, p.title, p.url_scheme, p.url_base, p.uri, p.created,
		COUNT(DISTINCT v.id) AS votes,
		IF(p.user_id = ?, 2, NOT EXISTS (SELECT * FROM votes WHERE user_id = ? AND post_id = p.id)) AS votable,
		COUNT(DISTINCT c.id) AS comments
	FROM posts AS p
	LEFT JOIN votes AS v ON p.id = v.post_id
	LEFT JOIN comments AS c ON p.id = c.post_id
	JOIN users AS u ON p.user_id = u.id
	%s AND p.deleted = 0
	GROUP BY p.id
	ORDER BY p.created DESC, p.id DESC
	LIMIT 31`

	now := time.Now()

	var rows *sql.Rows
	var err error
	if filterByUserID > 0 {
		// Using the post ID as a proxy for creation timestamp.
		stmt = fmt.Sprintf(stmt, "WHERE p.id >= ? AND p.user_id = ?")
		rows, err = m.DB.Query(stmt, loggedInUserID, loggedInUserID, lastPostID, filterByUserID)
	} else {
		// Using the post ID as a proxy for creation timestamp.
		stmt = fmt.Sprintf(stmt, "WHERE p.id >= ?")
		rows, err = m.DB.Query(stmt, loggedInUserID, loggedInUserID, lastPostID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []*models.Post{}

	for rows.Next() {
		p := &models.Post{}

		err := rows.Scan(
			&p.ID, &p.Username, &p.Title, &p.URLScheme, &p.URLBase, &p.URI, &p.Created, &p.Votes,
			&p.Votable, &p.Comments,
		)
		if err != nil {
			return nil, err
		}

		p.Retrieved = now

		posts = append(posts, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (m *PostModel) Popular(loggedInUserID, offset int) ([]*models.Post, error) {

	stmt := `SELECT p.id, u.username, p.title, p.url_scheme, p.url_base, p.uri, p.created,
		COUNT(DISTINCT v.id) AS votes,
		IF(p.user_id = ?, 2, NOT EXISTS (SELECT * FROM votes WHERE user_id = ? AND post_id = p.id)) AS votable,
		COUNT(DISTINCT c.id) AS comments
	FROM posts AS p
	LEFT JOIN votes AS v ON p.id = v.post_id
	LEFT JOIN comments AS c ON p.id = c.post_id
	JOIN users AS u ON p.user_id = u.id
	WHERE p.deleted = 0
	GROUP BY p.id
	ORDER BY (COUNT(DISTINCT v.id)) / POW((TIMESTAMPDIFF(HOUR, p.created, NOW()) + 2), 1.2) DESC, p.created DESC, p.ID DESC
	LIMIT ?,31`

	now := time.Now()

	var err error
	var rows *sql.Rows
	rows, err = m.DB.Query(stmt, loggedInUserID, loggedInUserID, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []*models.Post{}

	for rows.Next() {
		p := &models.Post{}

		err := rows.Scan(
			&p.ID, &p.Username, &p.Title, &p.URLScheme, &p.URLBase, &p.URI, &p.Created, &p.Votes,
			&p.Votable, &p.Comments,
		)
		if err != nil {
			return nil, err
		}

		p.Retrieved = now

		posts = append(posts, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// TODO: For the other queries igore upvotes from deactivated user accounts!
func (m *PostModel) Upvote(postID, userID int, created time.Time) (int, error) {
	stmt := `INSERT INTO votes (post_id, user_id, created)
	VALUES(?, ?, ?)`

	result, err := m.DB.Exec(stmt, postID, userID, created)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "uc_votes") {
				return 0, models.ErrDuplicateUpvote
			}
		} else {
			return 0, err
		}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

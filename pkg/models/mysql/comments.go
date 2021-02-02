package mysql

import (
	"database/sql"
	"strings"
	"time"

	"github.com/bjarnemagnussen/go-hackernews/pkg/models"
	"github.com/go-sql-driver/mysql"
)

type CommentModel struct {
	DB             *sql.DB
	UnderPostStmt  *sql.Stmt
	ForUserStmt    *sql.Stmt
	ForCommentStmt *sql.Stmt
}

type CreateCommentParams struct {
	// content string, parentId, userId, postId int, created time.Time
	Content  string    `json:"content"`
	ParentID int       `json:"parent_id"`
	UserID   int       `json:"user_id"`
	PostID   int       `json:"post_id"`
	Created  time.Time `json:"created"`
}

func NewCommentModel(db *sql.DB) (*CommentModel, error) {

	// From:
	// https://stackoverflow.com/questions/62384927/mysql-define-data-type-for-concat-with-alias
	// and
	// https://stackoverflow.com/questions/24529166/order-comments-by-thread-path-and-by-number-of-total-votes
	// and
	// https://www.peterspython.com/en/blog/threaded-comments-using-common-table-expressions-cte-for-a-mysql-flask-blog-or-cms
	// We initiate a table `first_comments` that contain the top-level
	// (no parents) comments. A `path` is constructed using their number of
	// votes concatenated with their id.
	// This table is iterated over by joining with comments with their parents
	// found in `first_comments` and extending the parent's path with its own
	// number of votes and id.

	// TODO: Optimize/Make it less complicated
	// TODO: Add paging (last and next)
	stmt := `WITH RECURSIVE
		first_comments (id, content, parent_id, post_id, user_id, created, deleted, level, votes, votable, path) AS (
			SELECT
				c.id,
				c.content,
				c.parent_id,
				c.post_id,
				c.user_id,
				c.created,
				c.deleted,
				0 as level,
				COUNT(DISTINCT v.id) AS votes,
				IF(c.user_id = ?, 2, NOT EXISTS (SELECT * FROM comment_votes WHERE user_id = ? AND comment_id = c.id)) AS votable,
				JSON_ARRAY(LPAD(COUNT(DISTINCT v.id), 6, 0), LPAD(c.id, 6, 0)) as path
			FROM comments AS c
			LEFT JOIN comment_votes AS v ON c.id = v.comment_id
			WHERE c.post_id = ? AND c.parent_id IS NULL
			GROUP BY c.id
			UNION ALL
			SELECT
				r.id,
				r.content,
				r.parent_id,
				r.post_id,
				r.user_id,
				r.created,
				r.deleted,
				fle.level+1 as level,
				r.votes,
				r.votable,
				JSON_ARRAY_APPEND(
					fle.path, 
					'$', LPAD(r.votes, 6, 0), 
					'$', LPAD(r.id, 6, 0)
				) as path
			FROM first_comments AS fle
			JOIN (
				SELECT
					c.id,
					c.content,
					c.parent_id,
					c.post_id,
					c.user_id,
					c.created,
					c.deleted,
					COUNT(DISTINCT v.id) AS votes,
					IF(c.user_id = ?, 2, NOT EXISTS (SELECT * FROM comment_votes WHERE user_id = ? AND comment_id = c.id)) AS votable
				FROM comments AS c
				LEFT JOIN comment_votes AS v ON c.id = v.comment_id
				WHERE c.post_id = ?
				GROUP BY c.id
			) AS r ON fle.id = r.parent_id
		)

		SELECT c.id, u.username, IF(c.deleted = 0, c.content, "[deleted]") AS content, c.parent_id, c.post_id, c.user_id, c.created, c.level, (d.id IS NOT NULL) AS leaf, c.votes, c.votable, c.deleted
		FROM first_comments AS c
		JOIN  users AS u ON c.user_id = u.id
		LEFT JOIN comments AS d ON c.id = d.parent_id
		WHERE c.deleted = 0 OR (c.deleted = 1 AND d.id IS NOT NULL)
		GROUP BY c.id, u.username, c.content, c.parent_id, c.post_id, c.user_id, c.created, c.level, (d.id IS NOT NULL), c.votes, c.votable, c.deleted, c.path
		ORDER BY GROUP_CONCAT(c.path) DESC`

	underPostStmt, err := db.Prepare(stmt)
	if err != nil {
		return nil, err
	}

	// TODO: Allow pagination (next and last using path?)
	stmt2 := `WITH RECURSIVE
		first_comments (id, content, parent_id, post_id, user_id, created, deleted, level, votes, votable, path) AS (
			(SELECT
				c.id,
				c.content,
				c.parent_id,
				c.post_id,
				c.user_id,
				c.created,
				c.deleted,
				0 as level,
				COUNT(DISTINCT v.id) AS votes,
				IF(c.user_id = ?, 2, NOT EXISTS (SELECT * FROM comment_votes WHERE user_id = ? AND comment_id = c.id)) AS votable,
				JSON_ARRAY(LPAD(COUNT(DISTINCT v.id), 6, 0), LPAD(c.id, 6, 0)) as path
			FROM comments AS c
			LEFT JOIN comment_votes AS v ON c.id = v.comment_id
			WHERE c.user_id = ? AND c.id <= ? AND c.deleted = 0
			GROUP BY c.id
			ORDER BY path ASC)
			UNION DISTINCT
			SELECT
				r.id,
				r.content,
				r.parent_id,
				r.post_id,
				r.user_id,
				r.created,
				r.deleted,
				fle.level+1 as level,
				r.votes,
				r.votable,
				JSON_ARRAY_APPEND(
					fle.path, 
					'$', LPAD(r.votes, 6, 0), 
					'$', LPAD(r.id, 6, 0)
				) as path
			FROM first_comments AS fle
			JOIN (
				SELECT
					c.id,
					c.content,
					c.parent_id,
					c.post_id,
					c.user_id,
					c.created,
					c.deleted,
					COUNT(DISTINCT v.id) AS votes,
					IF(c.user_id = ?, 2, NOT EXISTS (SELECT * FROM comment_votes WHERE user_id = ? AND comment_id = c.id)) AS votable
				FROM comments AS c
				LEFT JOIN comment_votes AS v ON c.id = v.comment_id
				WHERE c.deleted = 0
				GROUP BY c.id
			) AS r ON fle.id = r.parent_id
		)

		SELECT c.id, u.username, c.content, c.parent_id, c.post_id, c.user_id, c.created, c.level, (d.id IS NULL) AS leaf, c.votes, c.votable, p.title, c.deleted
		FROM first_comments AS c
		JOIN  users AS u ON c.user_id = u.id
		LEFT JOIN comments AS d ON c.parent_id = d.id
		LEFT JOIN posts AS p ON p.id = c.post_id
		WHERE LENGTH(c.path) = (
			SELECT MAX(LENGTH(f.path))
			FROM first_comments AS f
			WHERE f.id = c.id
		) AND p.deleted = 0
		GROUP BY c.id, u.username, c.content, c.parent_id, c.post_id, c.user_id, c.created, c.level, leaf, c.votes, c.votable, c.deleted, d.id, c.path
		ORDER BY GROUP_CONCAT(c.path) DESC`
	// TODO: Is the leaf var necessary?

	forUserStmt, err := db.Prepare(stmt2)
	if err != nil {
		return nil, err
	}

	stmt3 := `WITH RECURSIVE
		first_comments (id, content, parent_id, post_id, user_id, created, deleted, level, votes, votable, path) AS (
			SELECT
				c.id,
				c.content,
				c.parent_id,
				c.post_id,
				c.user_id,
				c.created,
				c.deleted,
				0 as level,
				COUNT(DISTINCT v.id) AS votes,
				IF(c.user_id = ?, 2, NOT EXISTS (SELECT * FROM comment_votes WHERE user_id = ? AND comment_id = c.id)) AS votable,
				JSON_ARRAY(LPAD(COUNT(DISTINCT v.id), 6, 0), LPAD(c.id, 6, 0)) as path
			FROM comments AS c
			LEFT JOIN comment_votes AS v ON c.id = v.comment_id
			WHERE c.parent_id = ?
			GROUP BY c.id
			UNION ALL
			SELECT
				r.id,
				r.content,
				r.parent_id,
				r.post_id,
				r.user_id,
				r.created,
				r.deleted,
				fle.level+1 as level,
				r.votes,
				r.votable,
				JSON_ARRAY_APPEND(
					fle.path, 
					'$', LPAD(r.votes, 6, 0), 
					'$', LPAD(r.id, 6, 0)
				) as path
			FROM first_comments AS fle
			JOIN (
				SELECT
					c.id,
					c.content,
					c.parent_id,
					c.post_id,
					c.user_id,
					c.created,
					c.deleted,
					COUNT(DISTINCT v.id) AS votes,
					IF(c.user_id = ?, 2, NOT EXISTS (SELECT * FROM comment_votes WHERE user_id = ? AND comment_id = c.id)) AS votable
				FROM comments AS c
				LEFT JOIN comment_votes AS v ON c.id = v.comment_id
				WHERE c.post_id = ?
				GROUP BY c.id
			) AS r ON fle.id = r.parent_id
		)
		
		SELECT c.id, u.username, IF(c.deleted = 0, c.content, "[deleted]") AS content, c.parent_id, c.post_id, c.user_id, c.created, c.level, (d.id IS NOT NULL) AS leaf, c.votes, c.votable, c.deleted
		FROM first_comments AS c
		JOIN  users AS u ON c.user_id = u.id
		LEFT JOIN comments AS d ON c.id = d.parent_id
		WHERE c.deleted = 0 OR (c.deleted = 1 AND d.id IS NOT NULL)
		GROUP BY c.id, u.username, c.content, c.parent_id, c.post_id, c.user_id, c.created, c.level, (d.id IS NOT NULL), c.votes, c.votable, c.deleted, c.path
		ORDER BY GROUP_CONCAT(c.path) DESC`
	// TODO: Is using WHERE-c.post_id clause in the recursive call really an optimization? (I think so!)

	forCommentStmt, err := db.Prepare(stmt3)
	if err != nil {
		return nil, err
	}

	return &CommentModel{DB: db, UnderPostStmt: underPostStmt, ForUserStmt: forUserStmt, ForCommentStmt: forCommentStmt}, nil
}

func (m *CommentModel) Insert(arg CreateCommentParams) (int, error) {
	stmt := `INSERT INTO comments (content, parent_id, user_id, post_id, created)
	VALUES(?, ?, ?, ?, ?)`

	pID := sql.NullInt64{int64(arg.ParentID), true}
	if arg.ParentID == 0 {
		pID = sql.NullInt64{0, false}
	}
	result, err := m.DB.Exec(stmt, arg.Content, pID, arg.UserID, arg.PostID, arg.Created)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *CommentModel) GetByID(loggedInUserID, id int) (*models.Comment, error) {
	stmt := `SELECT c.id, IF(c.deleted = 0, c.content, "[deleted]") AS content, c.parent_id, c.user_id, c.post_id, c.created, c.deleted,
		COUNT(DISTINCT v.id) AS votes,
		IF(c.user_id = ?, 2, NOT EXISTS (SELECT * FROM comment_votes WHERE user_id = ? AND comment_id = c.id)) AS votable,
		p.title
	FROM comments AS c
	LEFT JOIN posts AS p ON c.post_id = p.id
	LEFT JOIN comment_votes AS v ON c.id = v.comment_id
	WHERE c.id = ?
	GROUP BY c.id`

	c := &models.Comment{}

	err := m.DB.QueryRow(stmt, loggedInUserID, loggedInUserID, id).Scan(
		&c.ID, &c.Content, &c.ParentID, &c.UserID, &c.PostID, &c.Created, &c.Deleted, &c.Votes, &c.Votable, &c.PostTitle,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}

	c.Retrieved = time.Now()

	return c, nil
}

func (m *CommentModel) ForPost(loggedInUserID, postID int) ([]*models.Comment, error) {

	rows, err := m.UnderPostStmt.Query(loggedInUserID, loggedInUserID, postID, loggedInUserID, loggedInUserID, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []*models.Comment{}

	now := time.Now()
	for rows.Next() {
		c := &models.Comment{}

		err := rows.Scan(
			&c.ID, &c.Username, &c.Content, &c.ParentID, &c.PostID, &c.UserID, &c.Created, &c.Level, &c.Leaf, &c.Votes, &c.Votable, &c.Deleted,
		)
		if err != nil {
			return nil, err
		}

		c.Retrieved = now

		comments = append(comments, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (m *CommentModel) ForUser(loggedInUserID, lastCommentID, id int) ([]*models.Comment, error) {
	if lastCommentID == 0 {
		lastCommentID = 999999999
	}

	rows, err := m.ForUserStmt.Query(loggedInUserID, loggedInUserID, id, lastCommentID, loggedInUserID, loggedInUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []*models.Comment{}

	now := time.Now()
	for rows.Next() {
		c := &models.Comment{}

		err := rows.Scan(
			&c.ID, &c.Username, &c.Content, &c.ParentID, &c.PostID, &c.UserID, &c.Created, &c.Level, &c.Leaf, &c.Votes, &c.Votable, &c.PostTitle, &c.Deleted,
		)
		if err != nil {
			return nil, err
		}

		c.Retrieved = now

		comments = append(comments, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (m *CommentModel) ForComment(userID, postID, commentID int) ([]*models.Comment, error) {

	rows, err := m.ForCommentStmt.Query(userID, userID, commentID, userID, userID, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []*models.Comment{}

	now := time.Now()
	for rows.Next() {
		c := &models.Comment{}

		err := rows.Scan(
			&c.ID, &c.Username, &c.Content, &c.ParentID, &c.PostID, &c.UserID, &c.Created, &c.Level, &c.Leaf, &c.Votes, &c.Votable, &c.Deleted,
		)
		if err != nil {
			return nil, err
		}

		c.Retrieved = now

		comments = append(comments, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (m *CommentModel) Latest(userID, lastID int) ([]*models.Comment, error) {
	stmt := `SELECT c.id, u.username, c.content, c.parent_id, c.user_id, c.post_id, p.title, c.created, c.deleted, 
		IF(c.user_id = ?, 2, NOT EXISTS (SELECT * FROM comment_votes WHERE user_id = ? AND comment_id = c.id)) AS votable,
		COUNT(DISTINCT v.id) AS votes
		FROM comments AS c
		JOIN  users AS u ON c.user_id = u.id
		LEFT JOIN comment_votes AS v ON c.id = v.comment_id
		LEFT JOIN posts AS p ON p.id = c.post_id 
		WHERE c.id > ? AND c.deleted = 0 AND p.deleted = 0
		GROUP BY c.id
		ORDER BY c.id DESC
		LIMIT 30`

	now := time.Now()
	rows, err := m.DB.Query(stmt, userID, userID, lastID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []*models.Comment{}

	for rows.Next() {
		c := &models.Comment{}

		err := rows.Scan(
			&c.ID, &c.Username, &c.Content, &c.ParentID, &c.UserID, &c.PostID, &c.PostTitle, &c.Created,
			&c.Deleted, &c.Votable, &c.Votes,
		)
		if err != nil {
			return nil, err
		}

		c.Retrieved = now

		comments = append(comments, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

// TODO: For the other queries igore upvotes from deactivated user accounts!
func (m *CommentModel) Upvote(commentID, userID int, created time.Time) (int, error) {
	stmt := `INSERT INTO comment_votes (comment_id, user_id, created)
	VALUES(?, ?, ?)`

	result, err := m.DB.Exec(stmt, commentID, userID, created)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "uc_commentvotes") {
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

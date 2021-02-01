package mysql_test

import (
	"database/sql"
	"log"
	"testing"
	"time"

	"github.com/bjarnemagnussen/go-hackernews/pkg/models"
	"github.com/bjarnemagnussen/go-hackernews/pkg/models/mysql"
	"github.com/bjarnemagnussen/go-hackernews/pkg/models/mysql/util"
	"github.com/stretchr/testify/require"
)

type createTestCommentArg struct {
	content  string
	userID   int
	postID   int
	parentID int
	created  time.Time
}

func createTestComment(t *testing.T, arg createTestCommentArg) *models.Comment {
	id, err := commentModel.Insert(mysql.CreateCommentParams{
		Content:  arg.content,
		UserID:   arg.userID,
		PostID:   arg.postID,
		ParentID: arg.parentID,
		Created:  arg.created,
	})
	if err != nil {
		log.Fatalf("Could not insert comment in helper function: %s", err)
	}

	var pID sql.NullInt64
	if arg.parentID != 0 {
		pID = sql.NullInt64{int64(arg.parentID), true}
	}

	return &models.Comment{
		ID:       id,
		Content:  arg.content,
		ParentID: pID,
		UserID:   arg.userID,
		PostID:   arg.postID,
		Deleted:  0,
	}
}

func TestComment_Insert(t *testing.T) {
	user := createRandomUser(t)
	post := createTestPost(t, createTestPostArg{
		title:  util.RandomString(10),
		url:    util.RandomDomain(),
		userID: user.ID,
	})

	id, err := commentModel.Insert(mysql.CreateCommentParams{
		Content: util.RandomString(200),
		UserID:  user.ID,
		PostID:  post.ID,
		Created: time.Now().UTC(),
	})
	require.NoError(t, err)
	require.NotZero(t, id)
}

func TestComment_GetRoot(t *testing.T) {
	user := createRandomUser(t)
	post := createTestPost(t, createTestPostArg{
		title:  util.RandomString(10),
		url:    util.RandomDomain(),
		userID: user.ID,
	})
	comment := createTestComment(t, createTestCommentArg{
		content: util.RandomString(10),
		userID:  user.ID,
		postID:  post.ID,
		created: time.Now().UTC(),
	})

	got, err := commentModel.GetByID(0, comment.ID)
	require.NoError(t, err)
	require.NotEmpty(t, got)
	require.Equal(t, comment.ID, got.ID)
	require.Equal(t, comment.ParentID, got.ParentID)
	require.Equal(t, comment.UserID, got.UserID)
	require.Equal(t, comment.PostID, got.PostID)
	require.Equal(t, comment.Content, got.Content)
	require.Equal(t, comment.Deleted, got.Deleted)
}

func TestComment_GetChild(t *testing.T) {
	user := createRandomUser(t)
	post := createTestPost(t, createTestPostArg{
		title:  util.RandomString(10),
		url:    util.RandomDomain(),
		userID: user.ID,
	})
	comment := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user.ID,
		postID:   post.ID,
		parentID: 1,
		created:  time.Now().UTC(),
	})

	got, err := commentModel.GetByID(0, comment.ID)
	require.NoError(t, err)
	require.NotEmpty(t, got)
	require.Equal(t, comment.ID, got.ID)
	require.Equal(t, comment.ParentID, got.ParentID)
	require.Equal(t, comment.UserID, got.UserID)
	require.Equal(t, comment.PostID, got.PostID)
	require.Equal(t, comment.Content, got.Content)
	require.Equal(t, comment.Deleted, got.Deleted)
}

func TestComment_GetForPost(t *testing.T) {
	user := createRandomUser(t)
	post := createTestPost(t, createTestPostArg{
		title:  util.RandomString(10),
		url:    util.RandomDomain(),
		userID: user.ID,
	})
	comment1 := createTestComment(t, createTestCommentArg{
		content: util.RandomString(10),
		userID:  user.ID,
		postID:  post.ID,
		created: time.Now().UTC(),
	})
	comment2 := createTestComment(t, createTestCommentArg{
		content: util.RandomString(10),
		userID:  user.ID,
		postID:  post.ID,
		created: time.Now().UTC(),
	})
	commentModel.Upvote(comment2.ID, user.ID, time.Now().UTC())

	got, err := commentModel.ForPost(0, post.ID)
	require.NoError(t, err)
	require.Equal(t, 2, len(got))
	// We expect the comment with the most upvoates to come first.
	require.Equal(t, comment2.ID, got[0].ID)
	require.Equal(t, comment1.ID, got[1].ID)
	require.Equal(t, comment1.Content, got[1].Content)
	require.Equal(t, comment1.UserID, got[1].UserID)
}

func TestComment_GetForPostThreading(t *testing.T) {
	user1 := createRandomUser(t)
	user2 := createRandomUser(t)
	post := createTestPost(t, createTestPostArg{
		title:  util.RandomString(10),
		url:    util.RandomDomain(),
		userID: user1.ID,
	})
	commentParent := createTestComment(t, createTestCommentArg{
		content: util.RandomString(10),
		userID:  user1.ID,
		postID:  post.ID,
		created: time.Now().UTC(),
	})
	commentChild1 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentParent.ID,
		created:  time.Now().UTC(),
	})
	commentChild1Child1 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentChild1.ID,
		created:  time.Now().UTC(),
	})
	commentChild1Child1Child1 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentChild1Child1.ID,
		created:  time.Now().UTC(),
	})
	commentChild1Child1Child2 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentChild1Child1.ID,
		created:  time.Now().UTC(),
	})
	commentChild1Child2 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentChild1.ID,
		created:  time.Now().UTC(),
	})
	commentChild1Child2Child1 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentChild1Child2.ID,
		created:  time.Now().UTC(),
	})
	commentChild1Child2Child2 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentChild1Child2.ID,
		created:  time.Now().UTC(),
	})
	commentChild2 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentParent.ID,
		created:  time.Now().UTC(),
	})
	commentModel.Upvote(commentChild2.ID, user2.ID, time.Now().UTC())

	commentModel.Upvote(commentChild1Child2.ID, user2.ID, time.Now().UTC())

	commentModel.Upvote(commentChild1Child1Child2.ID, user2.ID, time.Now().UTC())

	got, err := commentModel.ForPost(0, post.ID)
	require.NoError(t, err)
	require.Equal(t, 9, len(got))
	// We expect the comment with the most upvoates to come first.
	require.Equal(t, commentParent.ID, got[0].ID)
	require.Equal(t, commentChild2.ID, got[1].ID)
	require.Equal(t, commentChild1.ID, got[2].ID)
	require.Equal(t, commentChild1Child2.ID, got[3].ID)
	require.Equal(t, commentChild1Child2Child2.ID, got[4].ID) // Both children have score 0, but higher id number has priority.
	require.Equal(t, commentChild1Child2Child1.ID, got[5].ID)
	require.Equal(t, commentChild1Child1.ID, got[6].ID)
	require.Equal(t, commentChild1Child1Child2.ID, got[7].ID)
	require.Equal(t, commentChild1Child1Child1.ID, got[8].ID)
}

func TestComment_GetForUser(t *testing.T) {
	user := createRandomUser(t)
	post := createTestPost(t, createTestPostArg{
		title:  util.RandomString(10),
		url:    util.RandomDomain(),
		userID: user.ID,
	})
	comment1 := createTestComment(t, createTestCommentArg{
		content: util.RandomString(10),
		userID:  user.ID,
		postID:  post.ID,
		created: time.Now().UTC(),
	})
	comment2 := createTestComment(t, createTestCommentArg{
		content: util.RandomString(10),
		userID:  user.ID,
		postID:  post.ID,
		created: time.Now().UTC(),
	})
	commentModel.Upvote(comment2.ID, user.ID, time.Now().UTC())

	got, err := commentModel.ForUser(0, 0, user.ID)
	require.NoError(t, err)
	require.Equal(t, 2, len(got))
	// We expect the comment with the most upvoates to come first.
	require.Equal(t, comment2.ID, got[0].ID)
	require.Equal(t, comment1.ID, got[1].ID)
	require.Equal(t, comment1.Content, got[1].Content)
	require.Equal(t, comment1.UserID, got[1].UserID)
}

func TestComment_GetForUserThreading(t *testing.T) {
	user1 := createRandomUser(t)
	user2 := createRandomUser(t)
	post := createTestPost(t, createTestPostArg{
		title:  util.RandomString(10),
		url:    util.RandomDomain(),
		userID: user1.ID,
	})
	commentParent := createTestComment(t, createTestCommentArg{
		content: util.RandomString(10),
		userID:  user1.ID,
		postID:  post.ID,
		created: time.Now().UTC(),
	})
	commentChild1 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentParent.ID,
		created:  time.Now().UTC(),
	})
	commentChild1Child1 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentChild1.ID,
		created:  time.Now().UTC(),
	})
	commentChild1Child1Child1 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentChild1Child1.ID,
		created:  time.Now().UTC(),
	})
	commentChild1Child1Child2 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentChild1Child1.ID,
		created:  time.Now().UTC(),
	})
	commentChild1Child2 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentChild1.ID,
		created:  time.Now().UTC(),
	})
	commentChild1Child2Child1 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentChild1Child2.ID,
		created:  time.Now().UTC(),
	})
	commentChild1Child2Child2 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentChild1Child2.ID,
		created:  time.Now().UTC(),
	})
	commentChild2 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentParent.ID,
		created:  time.Now().UTC(),
	})
	commentModel.Upvote(commentChild2.ID, user2.ID, time.Now().UTC())

	commentModel.Upvote(commentChild1Child2.ID, user2.ID, time.Now().UTC())

	commentModel.Upvote(commentChild1Child1Child2.ID, user2.ID, time.Now().UTC())

	got, err := commentModel.ForUser(0, 0, user1.ID)
	require.NoError(t, err)
	require.Equal(t, 9, len(got))
	// We expect the comment with the most upvoates to come first.
	require.Equal(t, commentParent.ID, got[0].ID)
	require.Equal(t, commentChild2.ID, got[1].ID)
	require.Equal(t, commentChild1.ID, got[2].ID)
	require.Equal(t, commentChild1Child2.ID, got[3].ID)
	require.Equal(t, commentChild1Child2Child2.ID, got[4].ID) // Both children have score 0, but higher id number has priority.
	require.Equal(t, commentChild1Child2Child1.ID, got[5].ID)
	require.Equal(t, commentChild1Child1.ID, got[6].ID)
	require.Equal(t, commentChild1Child1Child2.ID, got[7].ID)
	require.Equal(t, commentChild1Child1Child1.ID, got[8].ID)
}

func TestComment_GetForComment(t *testing.T) {
	user := createRandomUser(t)
	post := createTestPost(t, createTestPostArg{
		title:  util.RandomString(10),
		url:    util.RandomDomain(),
		userID: user.ID,
	})
	commentParent := createTestComment(t, createTestCommentArg{
		content: util.RandomString(10),
		userID:  user.ID,
		postID:  post.ID,
		created: time.Now().UTC(),
	})
	comment1 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user.ID,
		postID:   post.ID,
		parentID: commentParent.ID,
		created:  time.Now().UTC(),
	})
	comment2 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user.ID,
		postID:   post.ID,
		parentID: commentParent.ID,
		created:  time.Now().UTC(),
	})
	commentModel.Upvote(comment2.ID, user.ID, time.Now().UTC())

	got, err := commentModel.ForComment(0, post.ID, commentParent.ID)
	require.NoError(t, err)
	require.Equal(t, 2, len(got))
	// We expect the comment with the most upvoates to come first.
	require.Equal(t, comment2.ID, got[0].ID)
	require.Equal(t, comment1.ID, got[1].ID)
	require.Equal(t, comment1.Content, got[1].Content)
	require.Equal(t, comment1.UserID, got[1].UserID)
}

func TestComment_GetForCommentThreading(t *testing.T) {
	user1 := createRandomUser(t)
	user2 := createRandomUser(t)
	post := createTestPost(t, createTestPostArg{
		title:  util.RandomString(10),
		url:    util.RandomDomain(),
		userID: user1.ID,
	})
	commentParent := createTestComment(t, createTestCommentArg{
		content: util.RandomString(10),
		userID:  user1.ID,
		postID:  post.ID,
		created: time.Now().UTC(),
	})
	commentChild1 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentParent.ID,
		created:  time.Now().UTC(),
	})
	commentChild1Child1 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentChild1.ID,
		created:  time.Now().UTC(),
	})
	commentChild1Child1Child1 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentChild1Child1.ID,
		created:  time.Now().UTC(),
	})
	commentChild1Child1Child2 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentChild1Child1.ID,
		created:  time.Now().UTC(),
	})
	commentChild1Child2 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentChild1.ID,
		created:  time.Now().UTC(),
	})
	commentChild1Child2Child1 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentChild1Child2.ID,
		created:  time.Now().UTC(),
	})
	commentChild1Child2Child2 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentChild1Child2.ID,
		created:  time.Now().UTC(),
	})
	commentChild2 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user1.ID,
		postID:   post.ID,
		parentID: commentParent.ID,
		created:  time.Now().UTC(),
	})
	commentModel.Upvote(commentChild2.ID, user2.ID, time.Now().UTC())

	commentModel.Upvote(commentChild1Child2.ID, user2.ID, time.Now().UTC())

	commentModel.Upvote(commentChild1Child1Child2.ID, user2.ID, time.Now().UTC())

	got, err := commentModel.ForComment(0, post.ID, commentParent.ID)
	require.NoError(t, err)
	require.Equal(t, 8, len(got))
	// We expect the comment with the most upvoates to come first.
	require.Equal(t, commentChild2.ID, got[0].ID)
	require.Equal(t, commentChild1.ID, got[1].ID)
	require.Equal(t, commentChild1Child2.ID, got[2].ID)
	require.Equal(t, commentChild1Child2Child2.ID, got[3].ID) // Both children have score 0, but higher id number has priority.
	require.Equal(t, commentChild1Child2Child1.ID, got[4].ID)
	require.Equal(t, commentChild1Child1.ID, got[5].ID)
	require.Equal(t, commentChild1Child1Child2.ID, got[6].ID)
	require.Equal(t, commentChild1Child1Child1.ID, got[7].ID)
}

func TestComment_Latest(t *testing.T) {
	user1 := createRandomUser(t)
	user2 := createRandomUser(t)
	post := createTestPost(t, createTestPostArg{
		title:   util.RandomString(10),
		content: util.RandomString(200),
		userID:  user1.ID,
	})
	comment1 := createTestComment(t, createTestCommentArg{
		content: util.RandomString(10),
		userID:  user1.ID,
		postID:  post.ID,
		created: time.Now().UTC(),
	})
	comment2 := createTestComment(t, createTestCommentArg{
		content:  util.RandomString(10),
		userID:   user2.ID,
		postID:   post.ID,
		parentID: comment1.ID,
		created:  time.Now().UTC(),
	})
	comment3 := createTestComment(t, createTestCommentArg{
		content: util.RandomString(10),
		userID:  user2.ID,
		postID:  post.ID,
		created: time.Now().UTC(),
	})
	commentModel.Upvote(comment2.ID, user1.ID, time.Now())
	commentModel.Upvote(comment2.ID, user2.ID, time.Now())

	latests, err := commentModel.Latest(0, 0)
	require.NoError(t, err)
	require.Greater(t, 31, len(latests))
	require.Equal(t, comment3.ID, latests[0].ID)
	require.Equal(t, comment2.ID, latests[1].ID)
	require.Equal(t, comment1.ID, latests[2].ID)
}

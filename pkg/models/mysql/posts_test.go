package mysql_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/bjarnemagnussen/go-hackernews/pkg/models"
	"github.com/bjarnemagnussen/go-hackernews/pkg/models/mysql"
	"github.com/bjarnemagnussen/go-hackernews/pkg/models/mysql/util"
	"github.com/stretchr/testify/require"
)

type createTestPostArg struct {
	title   string
	url     string
	content string
	userID  int
	created time.Time
}

func createTestPost(t *testing.T, arg createTestPostArg) *models.Post {
	if arg.created == (time.Time{}) {
		// Set current time without nanoseconds, as those are trimmed from MySQL's
		// datetime
		arg.created = time.Unix(time.Now().Unix(), 0).UTC()
	}
	argPost := mysql.CreatePostParams{
		UserID:  arg.userID,
		Title:   arg.title,
		Url:     arg.url,
		Content: arg.content,
		Created: arg.created,
	}

	id, err := postModel.Insert(argPost)
	require.NoError(t, err)
	require.NotZero(t, id)

	var content, urlBase, urlScheme, uri sql.NullString
	if arg.url != "" {
		content = sql.NullString{"", false}
		uri.String, urlScheme.String, urlBase.String, err = util.ExtractURL(arg.url)
		uri.Valid, urlScheme.Valid, urlBase.Valid = true, true, true
		require.NoError(t, err)
	} else {
		content = sql.NullString{arg.content, true}
	}

	return &models.Post{
		ID:        id,
		Title:     arg.title,
		URLScheme: urlScheme,
		URLBase:   urlBase,
		URI:       uri,
		Content:   content,
		Created:   arg.created,
		Deleted:   0,
	}
}

func TestPost_Insert(t *testing.T) {
	user := createRandomUser(t)
	arg := mysql.CreatePostParams{
		UserID:  user.ID,
		Title:   "Title",
		Url:     "https://google.com/sub",
		Created: time.Now(),
	}

	id, err := postModel.Insert(arg)
	require.NoError(t, err)
	require.NotZero(t, id)
}

func TestPost_InsertDuplicate(t *testing.T) {
	u := util.RandomURI()

	user1 := createRandomUser(t)
	argPost := mysql.CreatePostParams{
		UserID:  user1.ID,
		Title:   util.RandomString(6),
		Url:     u,
		Created: time.Now().UTC(),
	}

	_, err := postModel.Insert(argPost)
	require.NoError(t, err)

	user2 := createRandomUser(t)
	argPost = mysql.CreatePostParams{
		UserID:  user2.ID,
		Title:   util.RandomString(6),
		Url:     u + "/",
		Created: time.Now().UTC(),
	}

	_, err = postModel.Insert(argPost)
	require.Error(t, err, models.ErrDuplicatePost)
}
func TestPost_GetLinkPost(t *testing.T) {
	user := createRandomUser(t)
	post := createTestPost(t, createTestPostArg{
		title:  util.RandomString(10),
		url:    util.RandomDomain(),
		userID: user.ID,
	})

	got, err := postModel.GetByID(0, post.ID)
	require.NoError(t, err)
	require.Equal(t, post.ID, got.ID)
	require.Equal(t, post.Title, got.Title)
	require.Equal(t, post.URI, got.URI)
	require.Equal(t, post.URLBase, got.URLBase)
	require.Equal(t, post.URLScheme, got.URLScheme)
	require.Equal(t, post.Created, got.Created)
	require.Equal(t, post.Deleted, got.Deleted)
	require.Equal(t, user.Username, got.Username)
	require.Equal(t, "", got.Content.String)
}

func TestPost_GetStandardPost(t *testing.T) {
	user := createRandomUser(t)
	post := createTestPost(t, createTestPostArg{
		title:   util.RandomString(10),
		content: util.RandomString(200),
		userID:  user.ID,
	})

	got, err := postModel.GetByID(0, post.ID)
	require.NoError(t, err)
	require.Equal(t, post.ID, got.ID)
	require.Equal(t, post.Title, got.Title)
	require.Equal(t, "", got.URI.String)
	require.Equal(t, "", got.URLBase.String)
	require.Equal(t, "", got.URLScheme.String)
	require.Equal(t, post.Created, got.Created)
	require.Equal(t, post.Deleted, got.Deleted)
	require.Equal(t, user.Username, got.Username)
	require.Equal(t, post.Content, got.Content)
}

func TestPost_GetPDF(t *testing.T) {
	user := createRandomUser(t)
	post := createTestPost(t, createTestPostArg{
		title:  util.RandomString(10),
		url:    util.RandomURI() + ".pdf",
		userID: user.ID,
	})

	got, err := postModel.GetByID(0, post.ID)
	require.NoError(t, err)
	require.Equal(t, post.ID, got.ID)
	require.Equal(t, post.Title+" [pdf]", got.Title)
}

func TestPost_GetIDFromURL(t *testing.T) {
	u := util.RandomURI()
	user := createRandomUser(t)
	post := createTestPost(t, createTestPostArg{
		title:  util.RandomString(10),
		url:    u,
		userID: user.ID,
	})

	id1, err := postModel.GetIDFromURL(u + "/")
	require.NoError(t, err)
	require.Equal(t, post.ID, id1)
}

func TestPost_GetForTypeShow(t *testing.T) {
	user := createRandomUser(t)
	createTestPost(t, createTestPostArg{
		title:   "Show BH: " + util.RandomString(10),
		content: util.RandomString(200),
		userID:  user.ID,
	})
	createTestPost(t, createTestPostArg{
		title:   "Show BH: " + util.RandomString(10),
		content: util.RandomString(200),
		userID:  user.ID,
	})

	showPosts, err := postModel.GetForType(user.ID, 0, mysql.ShowPost)
	require.NoError(t, err)
	require.Equal(t, 2, len(showPosts))
}

func TestPost_GetForTypeAsk(t *testing.T) {
	user := createRandomUser(t)
	createTestPost(t, createTestPostArg{
		title:   "Ask BH: " + util.RandomString(10),
		content: util.RandomString(200),
		userID:  user.ID,
	})
	createTestPost(t, createTestPostArg{
		title:   "Ask BH: " + util.RandomString(10),
		content: util.RandomString(200),
		userID:  user.ID,
	})
	createTestPost(t, createTestPostArg{
		title:   "Ask BH: " + util.RandomString(10),
		content: util.RandomString(200),
		userID:  user.ID,
	})

	askPosts, err := postModel.GetForType(user.ID, 0, mysql.AskPost)
	require.NoError(t, err)
	require.Equal(t, 3, len(askPosts))
}

func TestPost_Latest(t *testing.T) {
	user1 := createRandomUser(t)
	user2 := createRandomUser(t)
	user3 := createRandomUser(t)
	post1 := createTestPost(t, createTestPostArg{
		title:   util.RandomString(10),
		content: util.RandomString(200),
		userID:  user1.ID,
	})
	post2 := createTestPost(t, createTestPostArg{
		title:   util.RandomString(10),
		content: util.RandomString(200),
		userID:  user1.ID,
	})
	post3 := createTestPost(t, createTestPostArg{
		title:   util.RandomString(10),
		content: util.RandomString(200),
		userID:  user1.ID,
	})
	postModel.Upvote(post1.ID, user1.ID, time.Now())
	postModel.Upvote(post1.ID, user2.ID, time.Now())
	postModel.Upvote(post1.ID, user3.ID, time.Now())

	postModel.Upvote(post2.ID, user1.ID, time.Now())

	postModel.Upvote(post3.ID, user1.ID, time.Now())
	postModel.Upvote(post3.ID, user2.ID, time.Now())

	latests, err := postModel.Latest(0, 0, 0)
	require.NoError(t, err)
	require.Greater(t, 31, len(latests))
	require.Equal(t, post3.ID, latests[0].ID)
	require.Equal(t, post2.ID, latests[1].ID)
	require.Equal(t, post1.ID, latests[2].ID)
}

func TestPost_Popular(t *testing.T) {
	created := time.Now()
	user1 := createRandomUser(t)
	user2 := createRandomUser(t)
	user3 := createRandomUser(t)
	user4 := createRandomUser(t)
	user5 := createRandomUser(t)

	post1 := createTestPost(t, createTestPostArg{
		title:   util.RandomString(10),
		content: util.RandomString(200),
		userID:  user1.ID,
		created: created,
	})
	post2 := createTestPost(t, createTestPostArg{
		title:   util.RandomString(10),
		content: util.RandomString(200),
		userID:  user1.ID,
		created: created,
	})
	post3 := createTestPost(t, createTestPostArg{
		title:   util.RandomString(10),
		content: util.RandomString(200),
		userID:  user1.ID,
		created: created,
	})

	postModel.Upvote(post1.ID, user1.ID, time.Now())
	postModel.Upvote(post1.ID, user2.ID, time.Now())
	postModel.Upvote(post1.ID, user3.ID, time.Now())
	postModel.Upvote(post1.ID, user4.ID, time.Now())

	postModel.Upvote(post2.ID, user1.ID, time.Now())
	postModel.Upvote(post2.ID, user2.ID, time.Now())
	postModel.Upvote(post2.ID, user3.ID, time.Now())
	postModel.Upvote(post2.ID, user4.ID, time.Now())
	postModel.Upvote(post2.ID, user5.ID, time.Now())

	postModel.Upvote(post3.ID, user1.ID, time.Now())
	postModel.Upvote(post3.ID, user2.ID, time.Now())
	postModel.Upvote(post3.ID, user3.ID, time.Now())

	popular, err := postModel.Popular(0, 0)
	require.NoError(t, err)
	require.Greater(t, 31, len(popular))
	require.Equal(t, post2.ID, popular[0].ID)
	require.Equal(t, post1.ID, popular[1].ID)
	require.Equal(t, post3.ID, popular[2].ID)
}

// db/sqlc/post_test.go

package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/haotianxu2021/newPortfolio/util"
	"github.com/stretchr/testify/require"
)

func createRandomPost(t *testing.T, user User) Post {
	arg := CreatePostParams{
		UserID: sql.NullInt32{
			Int32: user.ID,
			Valid: true,
		},
		Title:   "test post " + util.RandomString(5),
		Content: "test content " + util.RandomString(10),
		Type:    "blog",
		Status: sql.NullString{
			String: "published",
			Valid:  true,
		},
	}

	post, err := testQueries.CreatePost(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, post)

	require.Equal(t, arg.UserID, post.UserID)
	require.Equal(t, arg.Title, post.Title)
	require.Equal(t, arg.Content, post.Content)
	require.Equal(t, arg.Type, post.Type)
	require.Equal(t, arg.Status, post.Status)

	require.NotZero(t, post.ID)
	require.NotZero(t, post.CreatedAt)

	return post
}

func TestListPosts(t *testing.T) {
	user := createRandomUser(t)

	// Create 10 random posts
	for i := 0; i < 10; i++ {
		createRandomPost(t, user)
	}

	arg := ListPostsParams{
		Column1: "published",
		Limit:   5,
		Offset:  0,
	}

	posts, err := testQueries.ListPosts(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, posts, 5)

	for _, post := range posts {
		require.NotEmpty(t, post)
		require.Equal(t, "published", post.Status.String)
		require.NotEmpty(t, post.Username)
	}
}

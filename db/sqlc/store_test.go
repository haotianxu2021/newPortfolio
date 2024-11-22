// db/sqlc/store_test.go

package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/haotianxu2021/newPortfolio/util"
	"github.com/stretchr/testify/require"
)

func createRandomTag(t *testing.T) Tag {
	name := "tag_" + util.RandomString(5)
	tag, err := testQueries.CreateTag(context.Background(), name)
	require.NoError(t, err)
	require.NotEmpty(t, tag)
	require.Equal(t, name, tag.Name)
	return tag
}

func TestCreatePostTx(t *testing.T) {
	store := NewStore(testDB)

	// Create a test user first
	user := createRandomUser(t)

	// Create some test tags
	tag1 := createRandomTag(t)
	tag2 := createRandomTag(t)

	// Test creating a post with images and tags
	arg := CreatePostTxParams{
		UserID:  user.ID,
		Title:   "test title" + util.RandomString(5),
		Content: "test content" + util.RandomString(10),
		Images: []CreatePostImage{
			{
				FilePath: "/test/path1.jpg",
				AltText:  "test alt 1",
			},
			{
				FilePath: "/test/path2.jpg",
				AltText:  "test alt 2",
			},
		},
		Tags: []int32{tag1.ID, tag2.ID},
	}

	post, err := store.CreatePostTx(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, post)

	// Verify post details
	require.Equal(t, arg.Title, post.Title)
	require.Equal(t, arg.Content, post.Content)
	require.Equal(t, "blog", post.Type)
	require.Equal(t, sql.NullString{String: "published", Valid: true}, post.Status)

	// Verify images were created
	listImagesArg := ListUserImagesParams{
		UserID: sql.NullInt32{Int32: user.ID, Valid: true},
		Limit:  10,
		Offset: 0,
	}
	images, err := testQueries.ListUserImages(context.Background(), listImagesArg)
	require.NoError(t, err)
	require.Len(t, images, 2)

	// Verify tags were linked
	postTags, err := testQueries.ListPostTags(context.Background(), post.ID)
	require.NoError(t, err)
	require.Len(t, postTags, 2)
}

func TestUpdatePostTx(t *testing.T) {
	store := NewStore(testDB)

	// Create initial post with user and tags
	user := createRandomUser(t)
	tag1 := createRandomTag(t)
	tag2 := createRandomTag(t)
	tag3 := createRandomTag(t)

	initialPost, err := store.CreatePostTx(context.Background(), CreatePostTxParams{
		UserID:  user.ID,
		Title:   "initial title",
		Content: "initial content",
		Tags:    []int32{tag1.ID, tag2.ID},
	})
	require.NoError(t, err)

	// Update post with new title, content, and different tags
	arg := UpdatePostTxParams{
		ID:      initialPost.ID,
		Title:   "updated title",
		Content: "updated content",
		Tags:    []int32{tag2.ID, tag3.ID}, // Remove tag1, keep tag2, add tag3
	}

	result, err := store.UpdatePostTx(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, result.Post)
	require.NotEmpty(t, result.Tags)

	// Verify updated post details
	require.Equal(t, arg.Title, result.Post.Title)
	require.Equal(t, arg.Content, result.Post.Content)

	// Verify tags were updated
	require.Len(t, result.Tags, 2)
	tagIDs := make(map[int32]bool)
	for _, tag := range result.Tags {
		tagIDs[tag.TagID] = true
	}
	require.True(t, tagIDs[tag2.ID])
	require.True(t, tagIDs[tag3.ID])
	require.False(t, tagIDs[tag1.ID])
}

func TestAddPostTagTx(t *testing.T) {
	store := NewStore(testDB)

	// Create test user and post
	user := createRandomUser(t)
	tag := createRandomTag(t)

	post, err := store.CreatePostTx(context.Background(), CreatePostTxParams{
		UserID:  user.ID,
		Title:   "test title",
		Content: "test content",
	})
	require.NoError(t, err)

	// Test adding a single tag
	arg := PostTagTxParams{
		PostID: post.ID,
		TagID:  tag.ID,
	}

	postTag, err := store.AddPostTagTx(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, postTag)
	require.Equal(t, arg.PostID, postTag.PostID)
	require.Equal(t, arg.TagID, postTag.TagID)
}

func TestBatchAddPostTagsTx(t *testing.T) {
	store := NewStore(testDB)

	// Create test user, post, and multiple tags
	user := createRandomUser(t)
	tag1 := createRandomTag(t)
	tag2 := createRandomTag(t)
	tag3 := createRandomTag(t)

	post, err := store.CreatePostTx(context.Background(), CreatePostTxParams{
		UserID:  user.ID,
		Title:   "test title",
		Content: "test content",
	})
	require.NoError(t, err)

	// Test adding multiple tags
	arg := BatchAddPostTagsParams{
		PostID: post.ID,
		TagIDs: []int32{tag1.ID, tag2.ID, tag3.ID},
	}

	postTags, err := store.BatchAddPostTagsTx(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, postTags, 3)

	// Verify all tags were added
	tagIDs := make(map[int32]bool)
	for _, pt := range postTags {
		tagIDs[pt.TagID] = true
	}
	require.True(t, tagIDs[tag1.ID])
	require.True(t, tagIDs[tag2.ID])
	require.True(t, tagIDs[tag3.ID])
}

func TestUpdatePostTagsTx(t *testing.T) {
	store := NewStore(testDB)

	// Create test data
	user := createRandomUser(t)
	tag1 := createRandomTag(t)
	tag2 := createRandomTag(t)
	tag3 := createRandomTag(t)

	// Create post with initial tags
	post, err := store.CreatePostTx(context.Background(), CreatePostTxParams{
		UserID:  user.ID,
		Title:   "test title",
		Content: "test content",
		Tags:    []int32{tag1.ID, tag2.ID},
	})
	require.NoError(t, err)

	// Update tags (remove tag1, keep tag2, add tag3)
	arg := UpdatePostTagsParams{
		PostID: post.ID,
		TagIDs: []int32{tag2.ID, tag3.ID},
	}

	err = store.UpdatePostTagsTx(context.Background(), arg)
	require.NoError(t, err)

	// Verify updated tags
	postTags, err := testQueries.ListPostTags(context.Background(), post.ID)
	require.NoError(t, err)
	require.Len(t, postTags, 2)

	tagIDs := make(map[int32]bool)
	for _, pt := range postTags {
		tagIDs[pt.ID] = true
	}
	require.False(t, tagIDs[tag1.ID])
	require.True(t, tagIDs[tag2.ID])
	require.True(t, tagIDs[tag3.ID])
}

// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"context"
	"database/sql"
)

type Querier interface {
	AddPostImage(ctx context.Context, arg AddPostImageParams) error
	AddPostTag(ctx context.Context, arg AddPostTagParams) error
	CreateComment(ctx context.Context, arg CreateCommentParams) (Comment, error)
	CreateImage(ctx context.Context, arg CreateImageParams) (Image, error)
	CreatePost(ctx context.Context, arg CreatePostParams) (Post, error)
	CreatePostTag(ctx context.Context, arg CreatePostTagParams) (PostTag, error)
	CreateTag(ctx context.Context, name string) (Tag, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	DecrementPostLikes(ctx context.Context, id int32) (Post, error)
	DeleteImage(ctx context.Context, id int32) error
	DeletePost(ctx context.Context, id int32) error
	DeletePostTag(ctx context.Context, arg DeletePostTagParams) error
	DeletePostTags(ctx context.Context, postID int32) error
	DeleteTag(ctx context.Context, id int32) error
	DeleteTagFromPosts(ctx context.Context, tagID int32) error
	GetImage(ctx context.Context, id int32) (Image, error)
	GetPost(ctx context.Context, id int32) (GetPostRow, error)
	GetPostTag(ctx context.Context, arg GetPostTagParams) (PostTag, error)
	GetPostsByTagID(ctx context.Context, tagID int32) ([]GetPostsByTagIDRow, error)
	GetTag(ctx context.Context, id int32) (Tag, error)
	GetUser(ctx context.Context, id int32) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByUsername(ctx context.Context, username string) (User, error)
	IncrementPostLikes(ctx context.Context, id int32) (Post, error)
	ListPostComments(ctx context.Context, postID sql.NullInt32) ([]ListPostCommentsRow, error)
	ListPostTags(ctx context.Context, postID int32) ([]Tag, error)
	ListPosts(ctx context.Context, arg ListPostsParams) ([]ListPostsRow, error)
	ListPostsByUser(ctx context.Context, arg ListPostsByUserParams) ([]ListPostsByUserRow, error)
	ListPostsOrderByLikes(ctx context.Context, arg ListPostsOrderByLikesParams) ([]ListPostsOrderByLikesRow, error)
	ListTags(ctx context.Context) ([]Tag, error)
	ListUserImages(ctx context.Context, arg ListUserImagesParams) ([]Image, error)
	ListUsers(ctx context.Context, arg ListUsersParams) ([]ListUsersRow, error)
	ListUsersOrderByPostLikes(ctx context.Context, arg ListUsersOrderByPostLikesParams) ([]ListUsersOrderByPostLikesRow, error)
	UpdatePost(ctx context.Context, arg UpdatePostParams) (Post, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
	UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) (UpdateUserPasswordRow, error)
}

var _ Querier = (*Queries)(nil)

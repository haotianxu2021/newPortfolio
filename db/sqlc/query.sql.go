// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package db

import (
	"context"
	"database/sql"
)

const addPostImage = `-- name: AddPostImage :exec
INSERT INTO post_images (
  post_id,
  image_id,
  display_order
) VALUES (
  $1, $2, $3
)
`

type AddPostImageParams struct {
	PostID       int32         `json:"post_id"`
	ImageID      int32         `json:"image_id"`
	DisplayOrder sql.NullInt32 `json:"display_order"`
}

func (q *Queries) AddPostImage(ctx context.Context, arg AddPostImageParams) error {
	_, err := q.db.ExecContext(ctx, addPostImage, arg.PostID, arg.ImageID, arg.DisplayOrder)
	return err
}

const addPostTag = `-- name: AddPostTag :exec
INSERT INTO post_tags (post_id, tag_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
`

type AddPostTagParams struct {
	PostID int32 `json:"post_id"`
	TagID  int32 `json:"tag_id"`
}

func (q *Queries) AddPostTag(ctx context.Context, arg AddPostTagParams) error {
	_, err := q.db.ExecContext(ctx, addPostTag, arg.PostID, arg.TagID)
	return err
}

const createComment = `-- name: CreateComment :one
INSERT INTO comments (
  post_id,
  user_id,
  content
) VALUES (
  $1, $2, $3
) RETURNING id, post_id, user_id, content, created_at
`

type CreateCommentParams struct {
	PostID  sql.NullInt32 `json:"post_id"`
	UserID  sql.NullInt32 `json:"user_id"`
	Content string        `json:"content"`
}

func (q *Queries) CreateComment(ctx context.Context, arg CreateCommentParams) (Comment, error) {
	row := q.db.QueryRowContext(ctx, createComment, arg.PostID, arg.UserID, arg.Content)
	var i Comment
	err := row.Scan(
		&i.ID,
		&i.PostID,
		&i.UserID,
		&i.Content,
		&i.CreatedAt,
	)
	return i, err
}

const createImage = `-- name: CreateImage :one
INSERT INTO images (
  user_id,
  file_path,
  alt_text,
  uploaded_at
) VALUES (
  $1, $2, $3, CURRENT_TIMESTAMP
) RETURNING id, user_id, file_path, alt_text, uploaded_at
`

type CreateImageParams struct {
	UserID   sql.NullInt32  `json:"user_id"`
	FilePath string         `json:"file_path"`
	AltText  sql.NullString `json:"alt_text"`
}

func (q *Queries) CreateImage(ctx context.Context, arg CreateImageParams) (Image, error) {
	row := q.db.QueryRowContext(ctx, createImage, arg.UserID, arg.FilePath, arg.AltText)
	var i Image
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.FilePath,
		&i.AltText,
		&i.UploadedAt,
	)
	return i, err
}

const createPost = `-- name: CreatePost :one
INSERT INTO posts (
  user_id,
  title,
  content,
  type,
  status
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING id, user_id, title, content, type, status, created_at, updated_at, likes
`

type CreatePostParams struct {
	UserID  sql.NullInt32  `json:"user_id"`
	Title   string         `json:"title"`
	Content string         `json:"content"`
	Type    string         `json:"type"`
	Status  sql.NullString `json:"status"`
}

func (q *Queries) CreatePost(ctx context.Context, arg CreatePostParams) (Post, error) {
	row := q.db.QueryRowContext(ctx, createPost,
		arg.UserID,
		arg.Title,
		arg.Content,
		arg.Type,
		arg.Status,
	)
	var i Post
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Title,
		&i.Content,
		&i.Type,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Likes,
	)
	return i, err
}

const createPostTag = `-- name: CreatePostTag :one
INSERT INTO post_tags (
    post_id,
    tag_id
) VALUES (
    $1, $2
) RETURNING post_id, tag_id
`

type CreatePostTagParams struct {
	PostID int32 `json:"post_id"`
	TagID  int32 `json:"tag_id"`
}

func (q *Queries) CreatePostTag(ctx context.Context, arg CreatePostTagParams) (PostTag, error) {
	row := q.db.QueryRowContext(ctx, createPostTag, arg.PostID, arg.TagID)
	var i PostTag
	err := row.Scan(&i.PostID, &i.TagID)
	return i, err
}

const createTag = `-- name: CreateTag :one
INSERT INTO tags (name)
VALUES ($1)
ON CONFLICT (name) DO UPDATE
SET name = EXCLUDED.name
RETURNING id, name
`

func (q *Queries) CreateTag(ctx context.Context, name string) (Tag, error) {
	row := q.db.QueryRowContext(ctx, createTag, name)
	var i Tag
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (
  username,
  email,
  password_hash,
  first_name,
  last_name,
  bio
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING id, username, email, password_hash, first_name, last_name, bio, created_at, updated_at
`

type CreateUserParams struct {
	Username     string         `json:"username"`
	Email        string         `json:"email"`
	PasswordHash string         `json:"password_hash"`
	FirstName    sql.NullString `json:"first_name"`
	LastName     sql.NullString `json:"last_name"`
	Bio          sql.NullString `json:"bio"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.Username,
		arg.Email,
		arg.PasswordHash,
		arg.FirstName,
		arg.LastName,
		arg.Bio,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.PasswordHash,
		&i.FirstName,
		&i.LastName,
		&i.Bio,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const decrementPostLikes = `-- name: DecrementPostLikes :one
UPDATE posts
SET likes = GREATEST(likes - 1, 0)
WHERE id = $1
RETURNING id, user_id, title, content, type, status, created_at, updated_at, likes
`

func (q *Queries) DecrementPostLikes(ctx context.Context, id int32) (Post, error) {
	row := q.db.QueryRowContext(ctx, decrementPostLikes, id)
	var i Post
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Title,
		&i.Content,
		&i.Type,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Likes,
	)
	return i, err
}

const deleteImage = `-- name: DeleteImage :exec
DELETE FROM images WHERE id = $1
`

func (q *Queries) DeleteImage(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, deleteImage, id)
	return err
}

const deletePost = `-- name: DeletePost :exec
DELETE FROM posts WHERE id = $1
`

func (q *Queries) DeletePost(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, deletePost, id)
	return err
}

const deletePostTag = `-- name: DeletePostTag :exec
DELETE FROM post_tags 
WHERE post_id = $1 AND tag_id = $2
`

type DeletePostTagParams struct {
	PostID int32 `json:"post_id"`
	TagID  int32 `json:"tag_id"`
}

func (q *Queries) DeletePostTag(ctx context.Context, arg DeletePostTagParams) error {
	_, err := q.db.ExecContext(ctx, deletePostTag, arg.PostID, arg.TagID)
	return err
}

const deletePostTags = `-- name: DeletePostTags :exec
DELETE FROM post_tags WHERE post_id = $1
`

func (q *Queries) DeletePostTags(ctx context.Context, postID int32) error {
	_, err := q.db.ExecContext(ctx, deletePostTags, postID)
	return err
}

const deleteTag = `-- name: DeleteTag :exec
DELETE FROM tags WHERE id = $1
`

func (q *Queries) DeleteTag(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, deleteTag, id)
	return err
}

const deleteTagFromPosts = `-- name: DeleteTagFromPosts :exec
DELETE FROM post_tags 
WHERE tag_id = $1
`

func (q *Queries) DeleteTagFromPosts(ctx context.Context, tagID int32) error {
	_, err := q.db.ExecContext(ctx, deleteTagFromPosts, tagID)
	return err
}

const getImage = `-- name: GetImage :one
SELECT id, user_id, file_path, alt_text, uploaded_at FROM images
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetImage(ctx context.Context, id int32) (Image, error) {
	row := q.db.QueryRowContext(ctx, getImage, id)
	var i Image
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.FilePath,
		&i.AltText,
		&i.UploadedAt,
	)
	return i, err
}

const getPost = `-- name: GetPost :one
SELECT 
  p.id, p.user_id, p.title, p.content, p.type, p.status, p.created_at, p.updated_at, p.likes,
  u.username,
  u.first_name,
  u.last_name,
  array_agg(DISTINCT t.name) as tags,
  array_agg(DISTINCT i.file_path) as images,
  p.likes
FROM posts p
LEFT JOIN users u ON p.user_id = u.id
LEFT JOIN post_tags pt ON p.id = pt.post_id
LEFT JOIN tags t ON pt.tag_id = t.id
LEFT JOIN post_images pi ON p.id = pi.post_id
LEFT JOIN images i ON pi.image_id = i.id
WHERE p.id = $1
GROUP BY p.id, u.id
`

type GetPostRow struct {
	ID        int32          `json:"id"`
	UserID    sql.NullInt32  `json:"user_id"`
	Title     string         `json:"title"`
	Content   string         `json:"content"`
	Type      string         `json:"type"`
	Status    sql.NullString `json:"status"`
	CreatedAt sql.NullTime   `json:"created_at"`
	UpdatedAt sql.NullTime   `json:"updated_at"`
	Likes     int32          `json:"likes"`
	Username  sql.NullString `json:"username"`
	FirstName sql.NullString `json:"first_name"`
	LastName  sql.NullString `json:"last_name"`
	Tags      interface{}    `json:"tags"`
	Images    interface{}    `json:"images"`
	Likes_2   int32          `json:"likes_2"`
}

func (q *Queries) GetPost(ctx context.Context, id int32) (GetPostRow, error) {
	row := q.db.QueryRowContext(ctx, getPost, id)
	var i GetPostRow
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Title,
		&i.Content,
		&i.Type,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Likes,
		&i.Username,
		&i.FirstName,
		&i.LastName,
		&i.Tags,
		&i.Images,
		&i.Likes_2,
	)
	return i, err
}

const getPostTag = `-- name: GetPostTag :one
SELECT post_id, tag_id FROM post_tags
WHERE post_id = $1 AND tag_id = $2 LIMIT 1
`

type GetPostTagParams struct {
	PostID int32 `json:"post_id"`
	TagID  int32 `json:"tag_id"`
}

func (q *Queries) GetPostTag(ctx context.Context, arg GetPostTagParams) (PostTag, error) {
	row := q.db.QueryRowContext(ctx, getPostTag, arg.PostID, arg.TagID)
	var i PostTag
	err := row.Scan(&i.PostID, &i.TagID)
	return i, err
}

const getPostsByTagID = `-- name: GetPostsByTagID :many
SELECT 
  p.id, p.user_id, p.title, p.content, p.type, p.status, p.created_at, p.updated_at, p.likes,
  u.username,
  u.first_name,
  u.last_name
FROM posts p
JOIN post_tags pt ON p.id = pt.post_id
JOIN users u ON p.user_id = u.id
WHERE pt.tag_id = $1
`

type GetPostsByTagIDRow struct {
	ID        int32          `json:"id"`
	UserID    sql.NullInt32  `json:"user_id"`
	Title     string         `json:"title"`
	Content   string         `json:"content"`
	Type      string         `json:"type"`
	Status    sql.NullString `json:"status"`
	CreatedAt sql.NullTime   `json:"created_at"`
	UpdatedAt sql.NullTime   `json:"updated_at"`
	Likes     int32          `json:"likes"`
	Username  string         `json:"username"`
	FirstName sql.NullString `json:"first_name"`
	LastName  sql.NullString `json:"last_name"`
}

func (q *Queries) GetPostsByTagID(ctx context.Context, tagID int32) ([]GetPostsByTagIDRow, error) {
	rows, err := q.db.QueryContext(ctx, getPostsByTagID, tagID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetPostsByTagIDRow{}
	for rows.Next() {
		var i GetPostsByTagIDRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Title,
			&i.Content,
			&i.Type,
			&i.Status,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Likes,
			&i.Username,
			&i.FirstName,
			&i.LastName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTag = `-- name: GetTag :one
SELECT id, name FROM tags WHERE id = $1 LIMIT 1
`

func (q *Queries) GetTag(ctx context.Context, id int32) (Tag, error) {
	row := q.db.QueryRowContext(ctx, getTag, id)
	var i Tag
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const getUser = `-- name: GetUser :one
SELECT id, username, email, password_hash, first_name, last_name, bio, created_at, updated_at FROM users
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetUser(ctx context.Context, id int32) (User, error) {
	row := q.db.QueryRowContext(ctx, getUser, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.PasswordHash,
		&i.FirstName,
		&i.LastName,
		&i.Bio,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, username, email, password_hash, first_name, last_name, bio, created_at, updated_at FROM users
WHERE email = $1 LIMIT 1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.PasswordHash,
		&i.FirstName,
		&i.LastName,
		&i.Bio,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByUsername = `-- name: GetUserByUsername :one
SELECT id, username, email, password_hash, first_name, last_name, bio, created_at, updated_at FROM users
WHERE username = $1 LIMIT 1
`

func (q *Queries) GetUserByUsername(ctx context.Context, username string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByUsername, username)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.PasswordHash,
		&i.FirstName,
		&i.LastName,
		&i.Bio,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const incrementPostLikes = `-- name: IncrementPostLikes :one
UPDATE posts
SET likes = likes + 1
WHERE id = $1
RETURNING id, user_id, title, content, type, status, created_at, updated_at, likes
`

func (q *Queries) IncrementPostLikes(ctx context.Context, id int32) (Post, error) {
	row := q.db.QueryRowContext(ctx, incrementPostLikes, id)
	var i Post
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Title,
		&i.Content,
		&i.Type,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Likes,
	)
	return i, err
}

const listPostComments = `-- name: ListPostComments :many
SELECT 
  c.id, c.post_id, c.user_id, c.content, c.created_at,
  u.username,
  u.first_name,
  u.last_name
FROM comments c
JOIN users u ON c.user_id = u.id
WHERE c.post_id = $1
ORDER BY c.created_at DESC
`

type ListPostCommentsRow struct {
	ID        int32          `json:"id"`
	PostID    sql.NullInt32  `json:"post_id"`
	UserID    sql.NullInt32  `json:"user_id"`
	Content   string         `json:"content"`
	CreatedAt sql.NullTime   `json:"created_at"`
	Username  string         `json:"username"`
	FirstName sql.NullString `json:"first_name"`
	LastName  sql.NullString `json:"last_name"`
}

func (q *Queries) ListPostComments(ctx context.Context, postID sql.NullInt32) ([]ListPostCommentsRow, error) {
	rows, err := q.db.QueryContext(ctx, listPostComments, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListPostCommentsRow{}
	for rows.Next() {
		var i ListPostCommentsRow
		if err := rows.Scan(
			&i.ID,
			&i.PostID,
			&i.UserID,
			&i.Content,
			&i.CreatedAt,
			&i.Username,
			&i.FirstName,
			&i.LastName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listPostTags = `-- name: ListPostTags :many
SELECT t.id, t.name 
FROM tags t
JOIN post_tags pt ON t.id = pt.tag_id
WHERE pt.post_id = $1
`

func (q *Queries) ListPostTags(ctx context.Context, postID int32) ([]Tag, error) {
	rows, err := q.db.QueryContext(ctx, listPostTags, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Tag{}
	for rows.Next() {
		var i Tag
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listPosts = `-- name: ListPosts :many
SELECT 
  p.id, p.user_id, p.title, p.content, p.type, p.status, p.created_at, p.updated_at, p.likes,
  u.username,
  COUNT(DISTINCT c.id) as comment_count,
  COALESCE(array_agg(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL), ARRAY[]::text[]) as tags,
  p.likes
FROM posts p
LEFT JOIN users u ON p.user_id = u.id
LEFT JOIN comments c ON p.id = c.post_id
LEFT JOIN post_tags pt ON p.id = pt.post_id
LEFT JOIN tags t ON pt.tag_id = t.id
WHERE ($1::text IS NULL OR p.status = $1)
GROUP BY p.id, u.id
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3
`

type ListPostsParams struct {
	Column1 string `json:"column_1"`
	Limit   int32  `json:"limit"`
	Offset  int32  `json:"offset"`
}

type ListPostsRow struct {
	ID           int32          `json:"id"`
	UserID       sql.NullInt32  `json:"user_id"`
	Title        string         `json:"title"`
	Content      string         `json:"content"`
	Type         string         `json:"type"`
	Status       sql.NullString `json:"status"`
	CreatedAt    sql.NullTime   `json:"created_at"`
	UpdatedAt    sql.NullTime   `json:"updated_at"`
	Likes        int32          `json:"likes"`
	Username     sql.NullString `json:"username"`
	CommentCount int64          `json:"comment_count"`
	Tags         interface{}    `json:"tags"`
	Likes_2      int32          `json:"likes_2"`
}

func (q *Queries) ListPosts(ctx context.Context, arg ListPostsParams) ([]ListPostsRow, error) {
	rows, err := q.db.QueryContext(ctx, listPosts, arg.Column1, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListPostsRow{}
	for rows.Next() {
		var i ListPostsRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Title,
			&i.Content,
			&i.Type,
			&i.Status,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Likes,
			&i.Username,
			&i.CommentCount,
			&i.Tags,
			&i.Likes_2,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listPostsByUser = `-- name: ListPostsByUser :many
SELECT 
    p.id, p.user_id, p.title, p.content, p.type, p.status, p.created_at, p.updated_at, p.likes,
    u.username,
    COUNT(DISTINCT c.id) as comment_count,
    COALESCE(array_agg(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL), ARRAY[]::text[]) as tags,
    p.likes
FROM posts p
LEFT JOIN users u ON p.user_id = u.id
LEFT JOIN comments c ON p.id = c.post_id
LEFT JOIN post_tags pt ON p.id = pt.post_id
LEFT JOIN tags t ON pt.tag_id = t.id
WHERE p.user_id = $1
    AND ($2::text IS NULL OR p.status = $2)
GROUP BY p.id, u.id
ORDER BY p.created_at DESC
LIMIT $3 OFFSET $4
`

type ListPostsByUserParams struct {
	UserID  sql.NullInt32 `json:"user_id"`
	Column2 string        `json:"column_2"`
	Limit   int32         `json:"limit"`
	Offset  int32         `json:"offset"`
}

type ListPostsByUserRow struct {
	ID           int32          `json:"id"`
	UserID       sql.NullInt32  `json:"user_id"`
	Title        string         `json:"title"`
	Content      string         `json:"content"`
	Type         string         `json:"type"`
	Status       sql.NullString `json:"status"`
	CreatedAt    sql.NullTime   `json:"created_at"`
	UpdatedAt    sql.NullTime   `json:"updated_at"`
	Likes        int32          `json:"likes"`
	Username     sql.NullString `json:"username"`
	CommentCount int64          `json:"comment_count"`
	Tags         interface{}    `json:"tags"`
	Likes_2      int32          `json:"likes_2"`
}

func (q *Queries) ListPostsByUser(ctx context.Context, arg ListPostsByUserParams) ([]ListPostsByUserRow, error) {
	rows, err := q.db.QueryContext(ctx, listPostsByUser,
		arg.UserID,
		arg.Column2,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListPostsByUserRow{}
	for rows.Next() {
		var i ListPostsByUserRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Title,
			&i.Content,
			&i.Type,
			&i.Status,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Likes,
			&i.Username,
			&i.CommentCount,
			&i.Tags,
			&i.Likes_2,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listPostsOrderByLikes = `-- name: ListPostsOrderByLikes :many
SELECT 
  p.id, p.user_id, p.title, p.content, p.type, p.status, p.created_at, p.updated_at, p.likes,
  u.username,
  COUNT(DISTINCT c.id) as comment_count,
  COALESCE(array_agg(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL), ARRAY[]::text[]) as tags
FROM posts p
LEFT JOIN users u ON p.user_id = u.id
LEFT JOIN comments c ON p.id = c.post_id
LEFT JOIN post_tags pt ON p.id = pt.post_id
LEFT JOIN tags t ON pt.tag_id = t.id
WHERE ($1::text IS NULL OR p.status = $1)
GROUP BY p.id, u.id
ORDER BY p.likes DESC, p.created_at DESC
LIMIT $2 OFFSET $3
`

type ListPostsOrderByLikesParams struct {
	Column1 string `json:"column_1"`
	Limit   int32  `json:"limit"`
	Offset  int32  `json:"offset"`
}

type ListPostsOrderByLikesRow struct {
	ID           int32          `json:"id"`
	UserID       sql.NullInt32  `json:"user_id"`
	Title        string         `json:"title"`
	Content      string         `json:"content"`
	Type         string         `json:"type"`
	Status       sql.NullString `json:"status"`
	CreatedAt    sql.NullTime   `json:"created_at"`
	UpdatedAt    sql.NullTime   `json:"updated_at"`
	Likes        int32          `json:"likes"`
	Username     sql.NullString `json:"username"`
	CommentCount int64          `json:"comment_count"`
	Tags         interface{}    `json:"tags"`
}

func (q *Queries) ListPostsOrderByLikes(ctx context.Context, arg ListPostsOrderByLikesParams) ([]ListPostsOrderByLikesRow, error) {
	rows, err := q.db.QueryContext(ctx, listPostsOrderByLikes, arg.Column1, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListPostsOrderByLikesRow{}
	for rows.Next() {
		var i ListPostsOrderByLikesRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Title,
			&i.Content,
			&i.Type,
			&i.Status,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Likes,
			&i.Username,
			&i.CommentCount,
			&i.Tags,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listTags = `-- name: ListTags :many
SELECT id, name FROM tags
ORDER BY name
`

func (q *Queries) ListTags(ctx context.Context) ([]Tag, error) {
	rows, err := q.db.QueryContext(ctx, listTags)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Tag{}
	for rows.Next() {
		var i Tag
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listUserImages = `-- name: ListUserImages :many
SELECT id, user_id, file_path, alt_text, uploaded_at FROM images
WHERE user_id = $1
ORDER BY uploaded_at DESC
LIMIT $2 OFFSET $3
`

type ListUserImagesParams struct {
	UserID sql.NullInt32 `json:"user_id"`
	Limit  int32         `json:"limit"`
	Offset int32         `json:"offset"`
}

func (q *Queries) ListUserImages(ctx context.Context, arg ListUserImagesParams) ([]Image, error) {
	rows, err := q.db.QueryContext(ctx, listUserImages, arg.UserID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Image{}
	for rows.Next() {
		var i Image
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.FilePath,
			&i.AltText,
			&i.UploadedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listUsers = `-- name: ListUsers :many
SELECT id, username, email, first_name, last_name, bio, created_at, updated_at
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
`

type ListUsersParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListUsersRow struct {
	ID        int32          `json:"id"`
	Username  string         `json:"username"`
	Email     string         `json:"email"`
	FirstName sql.NullString `json:"first_name"`
	LastName  sql.NullString `json:"last_name"`
	Bio       sql.NullString `json:"bio"`
	CreatedAt sql.NullTime   `json:"created_at"`
	UpdatedAt sql.NullTime   `json:"updated_at"`
}

func (q *Queries) ListUsers(ctx context.Context, arg ListUsersParams) ([]ListUsersRow, error) {
	rows, err := q.db.QueryContext(ctx, listUsers, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListUsersRow{}
	for rows.Next() {
		var i ListUsersRow
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.Email,
			&i.FirstName,
			&i.LastName,
			&i.Bio,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listUsersOrderByPostLikes = `-- name: ListUsersOrderByPostLikes :many
SELECT 
  u.id, u.username, u.email, u.first_name, u.last_name, u.bio, u.created_at, u.updated_at,
  COALESCE(SUM(p.likes), 0) as total_likes,
  COUNT(DISTINCT p.id) as post_count
FROM users u
LEFT JOIN posts p ON u.id = p.user_id
GROUP BY u.id
ORDER BY total_likes DESC
LIMIT $1 OFFSET $2
`

type ListUsersOrderByPostLikesParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListUsersOrderByPostLikesRow struct {
	ID         int32          `json:"id"`
	Username   string         `json:"username"`
	Email      string         `json:"email"`
	FirstName  sql.NullString `json:"first_name"`
	LastName   sql.NullString `json:"last_name"`
	Bio        sql.NullString `json:"bio"`
	CreatedAt  sql.NullTime   `json:"created_at"`
	UpdatedAt  sql.NullTime   `json:"updated_at"`
	TotalLikes interface{}    `json:"total_likes"`
	PostCount  int64          `json:"post_count"`
}

func (q *Queries) ListUsersOrderByPostLikes(ctx context.Context, arg ListUsersOrderByPostLikesParams) ([]ListUsersOrderByPostLikesRow, error) {
	rows, err := q.db.QueryContext(ctx, listUsersOrderByPostLikes, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListUsersOrderByPostLikesRow{}
	for rows.Next() {
		var i ListUsersOrderByPostLikesRow
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.Email,
			&i.FirstName,
			&i.LastName,
			&i.Bio,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.TotalLikes,
			&i.PostCount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updatePost = `-- name: UpdatePost :one
UPDATE posts
SET 
  title = COALESCE($2, title),
  content = COALESCE($3, content),
  type = COALESCE($4, type),
  status = COALESCE($5, status),
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, user_id, title, content, type, status, created_at, updated_at, likes
`

type UpdatePostParams struct {
	ID      int32          `json:"id"`
	Title   string         `json:"title"`
	Content string         `json:"content"`
	Type    string         `json:"type"`
	Status  sql.NullString `json:"status"`
}

func (q *Queries) UpdatePost(ctx context.Context, arg UpdatePostParams) (Post, error) {
	row := q.db.QueryRowContext(ctx, updatePost,
		arg.ID,
		arg.Title,
		arg.Content,
		arg.Type,
		arg.Status,
	)
	var i Post
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Title,
		&i.Content,
		&i.Type,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Likes,
	)
	return i, err
}

const updateUser = `-- name: UpdateUser :one
UPDATE users
SET 
  username = COALESCE($2, username),
  email = COALESCE($3, email),
  first_name = COALESCE($4, first_name),
  last_name = COALESCE($5, last_name),
  bio = COALESCE($6, bio),
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, username, email, password_hash, first_name, last_name, bio, created_at, updated_at
`

type UpdateUserParams struct {
	ID        int32          `json:"id"`
	Username  string         `json:"username"`
	Email     string         `json:"email"`
	FirstName sql.NullString `json:"first_name"`
	LastName  sql.NullString `json:"last_name"`
	Bio       sql.NullString `json:"bio"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, updateUser,
		arg.ID,
		arg.Username,
		arg.Email,
		arg.FirstName,
		arg.LastName,
		arg.Bio,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.PasswordHash,
		&i.FirstName,
		&i.LastName,
		&i.Bio,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateUserPassword = `-- name: UpdateUserPassword :one
UPDATE users
SET 
  password_hash = $2,
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, email, username, updated_at
`

type UpdateUserPasswordParams struct {
	ID           int32  `json:"id"`
	PasswordHash string `json:"password_hash"`
}

type UpdateUserPasswordRow struct {
	ID        int32        `json:"id"`
	Email     string       `json:"email"`
	Username  string       `json:"username"`
	UpdatedAt sql.NullTime `json:"updated_at"`
}

func (q *Queries) UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) (UpdateUserPasswordRow, error) {
	row := q.db.QueryRowContext(ctx, updateUserPassword, arg.ID, arg.PasswordHash)
	var i UpdateUserPasswordRow
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Username,
		&i.UpdatedAt,
	)
	return i, err
}

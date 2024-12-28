-- name: CreateUser :one
INSERT INTO users (
  username,
  email,
  password_hash,
  first_name,
  last_name,
  bio
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: ListUsers :many
SELECT id, username, email, first_name, last_name, bio, created_at, updated_at
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateUser :one
UPDATE users
SET 
  username = COALESCE($2, username),
  email = COALESCE($3, email),
  first_name = COALESCE($4, first_name),
  last_name = COALESCE($5, last_name),
  bio = COALESCE($6, bio),
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users
SET 
  password_hash = $2,
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, email, username, updated_at;

-- name: CreatePost :one
INSERT INTO posts (
  user_id,
  title,
  content,
  type,
  status
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetPost :one
SELECT 
  p.*,
  u.username,
  u.first_name,
  u.last_name,
  array_agg(DISTINCT t.name) as tags,
  array_agg(DISTINCT i.file_path) as images
FROM posts p
LEFT JOIN users u ON p.user_id = u.id
LEFT JOIN post_tags pt ON p.id = pt.post_id
LEFT JOIN tags t ON pt.tag_id = t.id
LEFT JOIN post_images pi ON p.id = pi.post_id
LEFT JOIN images i ON pi.image_id = i.id
WHERE p.id = $1
GROUP BY p.id, u.id;

-- name: ListPosts :many
SELECT 
  p.*,
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
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdatePost :one
UPDATE posts
SET 
  title = COALESCE($2, title),
  content = COALESCE($3, content),
  type = COALESCE($4, type),
  status = COALESCE($5, status),
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = $1;

-- name: CreateComment :one
INSERT INTO comments (
  post_id,
  user_id,
  content
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: ListPostComments :many
SELECT 
  c.*,
  u.username,
  u.first_name,
  u.last_name
FROM comments c
JOIN users u ON c.user_id = u.id
WHERE c.post_id = $1
ORDER BY c.created_at DESC;

-- name: AddPostImage :exec
INSERT INTO post_images (
  post_id,
  image_id,
  display_order
) VALUES (
  $1, $2, $3
);

-- name: CreateTag :one
INSERT INTO tags (name)
VALUES ($1)
ON CONFLICT (name) DO UPDATE
SET name = EXCLUDED.name
RETURNING *;

-- name: AddPostTag :exec
INSERT INTO post_tags (post_id, tag_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: CreateImage :one
INSERT INTO images (
  user_id,
  file_path,
  alt_text,
  uploaded_at
) VALUES (
  $1, $2, $3, CURRENT_TIMESTAMP
) RETURNING *;

-- name: GetImage :one
SELECT * FROM images
WHERE id = $1 LIMIT 1;

-- name: ListUserImages :many
SELECT * FROM images
WHERE user_id = $1
ORDER BY uploaded_at DESC
LIMIT $2 OFFSET $3;

-- name: DeleteImage :exec
DELETE FROM images WHERE id = $1;

-- name: DeleteTag :exec
DELETE FROM tags WHERE id = $1;

-- name: DeletePostTags :exec
DELETE FROM post_tags WHERE post_id = $1;

-- name: DeletePostTag :exec
DELETE FROM post_tags 
WHERE post_id = $1 AND tag_id = $2;

-- name: GetTag :one
SELECT * FROM tags WHERE id = $1 LIMIT 1;

-- name: ListTags :many
SELECT * FROM tags
ORDER BY name;

-- name: ListPostTags :many
SELECT t.* 
FROM tags t
JOIN post_tags pt ON t.id = pt.tag_id
WHERE pt.post_id = $1;

-- name: CreatePostTag :one
INSERT INTO post_tags (
    post_id,
    tag_id
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetPostTag :one
SELECT * FROM post_tags
WHERE post_id = $1 AND tag_id = $2 LIMIT 1;


-- name: DeleteTagFromPosts :exec
DELETE FROM post_tags 
WHERE tag_id = $1;

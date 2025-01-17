// db/sqlc/store.go

package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Store interface {
	Querier
	CreatePostTx(ctx context.Context, arg CreatePostTxParams) (Post, error)
	UploadPostImageTx(ctx context.Context, arg UploadPostImageTxParams) error
	UpdatePostTx(ctx context.Context, arg UpdatePostTxParams) (UpdatePostTxResult, error)
	AddPostTagTx(ctx context.Context, arg PostTagTxParams) (PostTag, error)
	BatchAddPostTagsTx(ctx context.Context, arg BatchAddPostTagsParams) ([]PostTag, error)
	FilterPosts(ctx context.Context, filter FilterParams) ([]FilteredPost, error)
}

type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *SQLStore {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil) // Use default isolation level
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// db/sqlc/store.go

type UploadPostImageTxParams struct {
	PostID   int32  `json:"post_id"`
	UserID   int32  `json:"user_id"`
	FilePath string `json:"file_path"`
	AltText  string `json:"alt_text"`
	Order    int32  `json:"order"`
}

func (store *SQLStore) UploadPostImageTx(ctx context.Context, arg UploadPostImageTxParams) error {
	return store.execTx(ctx, func(q *Queries) error {
		// 1. Create image record first
		image, err := q.CreateImage(ctx, CreateImageParams{
			UserID: sql.NullInt32{
				Int32: arg.UserID,
				Valid: true,
			},
			FilePath: arg.FilePath,
			AltText: sql.NullString{
				String: arg.AltText,
				Valid:  true,
			},
		})
		if err != nil {
			return err
		}

		// 2. Then link it to the post
		err = q.AddPostImage(ctx, AddPostImageParams{
			PostID:  arg.PostID,
			ImageID: image.ID,
			DisplayOrder: sql.NullInt32{
				Int32: arg.Order,
				Valid: true,
			},
		})
		if err != nil {
			return err
		}

		return nil
	})
}

type CreatePostTxParams struct {
	UserID  int32             `json:"user_id"`
	Title   string            `json:"title"`
	Content string            `json:"content"`
	Images  []CreatePostImage `json:"images"` // Changed to struct
	Tags    []int32           `json:"tags"`
}

type CreatePostImage struct {
	FilePath string `json:"file_path"`
	AltText  string `json:"alt_text"`
}

func (store *SQLStore) CreatePostTx(ctx context.Context, arg CreatePostTxParams) (Post, error) {
	var post Post

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// 1. Create post first
		post, err = q.CreatePost(ctx, CreatePostParams{
			UserID: sql.NullInt32{
				Int32: arg.UserID,
				Valid: true,
			},
			Title:   arg.Title,
			Content: arg.Content,
			Type:    "blog",
			Status: sql.NullString{
				String: "published",
				Valid:  true,
			},
		})
		if err != nil {
			return err
		}

		// 2. Then add images
		for i, img := range arg.Images {
			// Create image record
			image, err := q.CreateImage(ctx, CreateImageParams{
				UserID: sql.NullInt32{
					Int32: arg.UserID,
					Valid: true,
				},
				FilePath: img.FilePath,
				AltText: sql.NullString{
					String: img.AltText,
					Valid:  true,
				},
			})
			if err != nil {
				return err
			}

			// Link image to post
			err = q.AddPostImage(ctx, AddPostImageParams{
				PostID:  post.ID,
				ImageID: image.ID,
				DisplayOrder: sql.NullInt32{
					Int32: int32(i),
					Valid: true,
				},
			})
			if err != nil {
				return err
			}
		}

		// 3. Finally add tags
		for _, tagID := range arg.Tags {
			err = q.AddPostTag(ctx, AddPostTagParams{
				PostID: post.ID,
				TagID:  tagID,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return post, err
}

// db/sqlc/store.go

type UpdatePostTxParams struct {
	ID      int32   `json:"id"`
	Title   string  `json:"title"`
	Content string  `json:"content"`
	Tags    []int32 `json:"tags"`
}

type UpdatePostTxResult struct {
	Post Post      `json:"post"`
	Tags []PostTag `json:"tags"`
}

func (store *SQLStore) UpdatePostTx(ctx context.Context, arg UpdatePostTxParams) (UpdatePostTxResult, error) {
	var result UpdatePostTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// 1. Update post
		post, err := q.UpdatePost(ctx, UpdatePostParams{
			ID:      arg.ID,
			Title:   arg.Title,
			Content: arg.Content,
		})
		if err != nil {
			return err
		}
		result.Post = post

		// 2. Delete existing tags - always access tables in the same order to prevent deadlocks
		err = q.DeletePostTags(ctx, arg.ID)
		if err != nil {
			return err
		}

		// 3. Add new tags
		for _, tagID := range arg.Tags {
			err := q.AddPostTag(ctx, AddPostTagParams{
				PostID: arg.ID,
				TagID:  tagID,
			})
			if err != nil {
				return err
			}
			tag, err := q.GetPostTag(ctx, GetPostTagParams{PostID: arg.ID,
				TagID: tagID})
			if err != nil {
				return err
			}
			result.Tags = append(result.Tags, tag)
		}

		return nil
	})

	return result, err
}

type PostTagTxParams struct {
	PostID int32 `json:"post_id"`
	TagID  int32 `json:"tag_id"`
}

func (store *SQLStore) AddPostTagTx(ctx context.Context, arg PostTagTxParams) (PostTag, error) {
	var result PostTag

	err := store.execTx(ctx, func(q *Queries) error {
		// Verify post exists
		_, err := q.GetPost(ctx, arg.PostID)
		if err != nil {
			return err
		}

		// Verify tag exists
		_, err = q.GetTag(ctx, arg.TagID)
		if err != nil {
			return err
		}

		// Create post_tag relation
		postTag, err := q.CreatePostTag(ctx, CreatePostTagParams{
			PostID: arg.PostID,
			TagID:  arg.TagID,
		})
		if err != nil {
			return err
		}

		result = postTag
		return nil
	})

	return result, err
}

type BatchAddPostTagsParams struct {
	PostID int32   `json:"post_id"`
	TagIDs []int32 `json:"tag_ids"`
}

func (store *SQLStore) BatchAddPostTagsTx(ctx context.Context, arg BatchAddPostTagsParams) ([]PostTag, error) {
	var result []PostTag

	err := store.execTx(ctx, func(q *Queries) error {
		// Verify post exists
		_, err := q.GetPost(ctx, arg.PostID)
		if err != nil {
			return err
		}

		// Add each tag
		for _, tagID := range arg.TagIDs {
			postTag, err := q.CreatePostTag(ctx, CreatePostTagParams{
				PostID: arg.PostID,
				TagID:  tagID,
			})
			if err != nil {
				return err
			}
			result = append(result, postTag)
		}

		return nil
	})

	return result, err
}

type UpdatePostTagsParams struct {
	PostID int32   `json:"post_id"`
	TagIDs []int32 `json:"tag_ids"`
}

func (store *SQLStore) UpdatePostTagsTx(ctx context.Context, arg UpdatePostTagsParams) error {
	return store.execTx(ctx, func(q *Queries) error {
		// 1. Delete existing tags
		err := q.DeletePostTags(ctx, arg.PostID)
		if err != nil {
			return err
		}

		// 2. Add new tags
		for _, tagID := range arg.TagIDs {
			_, err = q.CreatePostTag(ctx, CreatePostTagParams{
				PostID: arg.PostID,
				TagID:  tagID,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
}

type FilterParams struct {
	UserID    *int32
	Status    *string
	SortBy    string
	SortOrder string
	Limit     int32
	Offset    int32
}
type FilteredPost struct {
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

// Implementation of FilterPosts for SQLStore
func (store *SQLStore) FilterPosts(ctx context.Context, filter FilterParams) ([]FilteredPost, error) {
	query := buildFilterQuery(filter)
	rows, err := store.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []FilteredPost
	for rows.Next() {
		var post FilteredPost
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.Type,
			&post.Status,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Likes,
			&post.Username,
			&post.CommentCount,
			&post.Tags,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, rows.Err()
}

func buildFilterQuery(filter FilterParams) string {
	baseQuery := `
        SELECT 
            p.id, p.user_id, p.title, p.content, p.type, p.status, 
            p.created_at, p.updated_at, p.likes,
            u.username,
            COUNT(DISTINCT c.id) as comment_count,
            COALESCE(array_agg(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL), ARRAY[]::text[]) as tags
        FROM posts p
        LEFT JOIN users u ON p.user_id = u.id
        LEFT JOIN comments c ON p.id = c.post_id
        LEFT JOIN post_tags pt ON p.id = pt.post_id
        LEFT JOIN tags t ON pt.tag_id = t.id
        WHERE 1=1
    `

	// Add filters
	if filter.UserID != nil {
		baseQuery += fmt.Sprintf(" AND p.user_id = %d", *filter.UserID)
	}
	if filter.Status != nil && *filter.Status != "" {
		baseQuery += fmt.Sprintf(" AND p.status = '%s'", *filter.Status)
	}

	// Add group by
	baseQuery += " GROUP BY p.id, u.id"

	// Add sorting
	sortField := "p.created_at"
	switch filter.SortBy {
	case "likes":
		sortField = "p.likes"
	case "title":
		sortField = "p.title"
	}
	baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortField, strings.ToUpper(filter.SortOrder))

	// Add pagination
	baseQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.Limit, filter.Offset)

	return baseQuery
}

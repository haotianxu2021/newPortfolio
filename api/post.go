package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	db "github.com/haotianxu2021/newPortfolio/db/sqlc"
	"github.com/haotianxu2021/newPortfolio/util"
)

type createPostRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	UserID  int32  `json:"user_id" binding:"required"`
	Type    string `json:"type" binding:"required"`
	Status  string `json:"status"`
}

type updatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Type    string `json:"type"`
	Status  string `json:"status"`
}

type addImageRequest struct {
	UserID   int32  `json:"user_id" binding:"required"`
	FilePath string `json:"file_path" binding:"required"`
	AltText  string `json:"alt_text"`
}

type addTagRequest struct {
	Name string `json:"name" binding:"required"`
}

type postResponse struct {
	ID        int32          `json:"id"`
	UserID    sql.NullInt32  `json:"user_id"`
	Title     string         `json:"title"`
	Content   string         `json:"content"`
	Type      string         `json:"type"`
	Status    sql.NullString `json:"status"`
	Likes     int32          `json:"likes"`
	CreatedAt sql.NullTime   `json:"created_at"`
	UpdatedAt sql.NullTime   `json:"updated_at"`
}

func (server *Server) createPost(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*util.Payload)

	// Get authenticated user
	user, err := server.store.GetUserByUsername(ctx, authPayload.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Verify request userID matches authenticated user
	var req createPostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.UserID != int32(user.ID) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "you can only create posts for yourself"})
		return
	}

	arg := db.CreatePostParams{
		Title:   req.Title,
		Content: req.Content,
		UserID: sql.NullInt32{
			Int32: req.UserID,
			Valid: true,
		},
		Type: req.Type,
		Status: sql.NullString{
			String: req.Status,
			Valid:  true,
		},
	}

	post, err := server.store.CreatePost(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, postResponse{
		ID:        post.ID,
		UserID:    post.UserID,
		Title:     post.Title,
		Content:   post.Content,
		Type:      post.Type,
		Status:    post.Status,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	})
}

func (server *Server) updatePost(ctx *gin.Context) {
	// Get auth payload
	authPayload := ctx.MustGet(authorizationPayloadKey).(*util.Payload)

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Verify post exists and belongs to user
	post, err := server.store.GetPost(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Verify ownership
	if post.Username.String != authPayload.Username {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "you can only update your own posts"})
		return
	}

	var req updatePostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	arg := db.UpdatePostParams{
		ID:      int32(id),
		Title:   req.Title,
		Content: req.Content,
		Type:    req.Type,
		Status: sql.NullString{
			String: req.Status,
			Valid:  req.Status != "",
		},
	}

	updatedPost, err := server.store.UpdatePost(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedPost)
}

func (server *Server) getPost(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	post, err := server.store.GetPost(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":         post.ID,
		"user_id":    post.UserID.Int32,
		"title":      post.Title,
		"content":    post.Content,
		"type":       post.Type,
		"status":     post.Status.String,
		"created_at": post.CreatedAt,
		"updated_at": post.UpdatedAt,
		"username":   post.Username.String,
		"first_name": post.FirstName,
		"last_name":  post.LastName,
		"tags":       post.Tags,
		"images":     post.Images,
		"likes":      post.Likes,
	})
}

func (server *Server) listPosts(ctx *gin.Context) {
	var limit int32 = 10 // default limit
	var offset int32 = 0 // default offset
	var status string = "published"
	// Parse limit from query parameter
	if limitStr := ctx.Query("limit"); limitStr != "" {
		limitInt, err := strconv.ParseInt(limitStr, 10, 32)
		if err == nil {
			limit = int32(limitInt)
		}
	}

	// Parse offset from query parameter
	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		offsetInt, err := strconv.ParseInt(offsetStr, 10, 32)
		if err == nil {
			offset = int32(offsetInt)
		}
	}

	// Get status from query parameter
	if statusStr := ctx.Query("status"); statusStr != "" {
		status = statusStr
	}

	arg := db.ListPostsParams{
		Column1: status, // Pass the raw status string, it can be empty
		Limit:   limit,
		Offset:  offset,
	}
	posts, err := server.store.ListPosts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]gin.H, len(posts))
	for i, post := range posts {
		response[i] = gin.H{
			"id":            post.ID,
			"user_id":       post.UserID,
			"title":         post.Title,
			"content":       post.Content,
			"type":          post.Type,
			"status":        post.Status.String,
			"created_at":    post.CreatedAt,
			"updated_at":    post.UpdatedAt,
			"username":      post.Username,
			"comment_count": post.CommentCount,
			"tags":          post.Tags,
			"likes":         post.Likes,
		}
	}

	ctx.JSON(http.StatusOK, response)
}

func (server *Server) incrementPostLikes(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	post, err := server.store.IncrementPostLikes(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":    post.ID,
		"likes": post.Likes,
	})
}

func (server *Server) decrementPostLikes(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	post, err := server.store.DecrementPostLikes(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":    post.ID,
		"likes": post.Likes,
	})
}

func (server *Server) deletePost(ctx *gin.Context) {
	// Get auth payload
	authPayload := ctx.MustGet(authorizationPayloadKey).(*util.Payload)

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Verify post exists and belongs to user
	post, err := server.store.GetPost(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Verify ownership
	if post.Username.String != authPayload.Username {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "you can only delete your own posts"})
		return
	}

	err = server.store.DeletePost(ctx, int32(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "post deleted successfully"})
}

func (server *Server) addImage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	postID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	var req addImageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// First create the image
	imageArg := db.CreateImageParams{
		UserID: sql.NullInt32{
			Int32: req.UserID,
			Valid: true,
		},
		FilePath: req.FilePath,
		AltText: sql.NullString{
			String: req.AltText,
			Valid:  req.AltText != "",
		},
	}

	image, err := server.store.CreateImage(ctx, imageArg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Then associate it with the post
	err = server.store.AddPostImage(ctx, db.AddPostImageParams{
		PostID:  int32(postID),
		ImageID: image.ID,
		DisplayOrder: sql.NullInt32{
			Int32: 1,
			Valid: true,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, image)
}

func (server *Server) deleteImage(ctx *gin.Context) {
	// Get auth payload
	authPayload := ctx.MustGet(authorizationPayloadKey).(*util.Payload)

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid image id"})
		return
	}

	// Get the image to verify ownership
	image, err := server.store.GetImage(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get user by username
	user, err := server.store.GetUserByUsername(ctx, authPayload.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Verify ownership
	if image.UserID.Int32 != user.ID {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "you can only delete your own images"})
		return
	}

	err = server.store.DeleteImage(ctx, int32(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "image deleted successfully"})
}

func (server *Server) addTag(ctx *gin.Context) {
	idStr := ctx.Param("id")
	postID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	var req addTagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// First create or get the tag
	tag, err := server.store.CreateTag(ctx, req.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Then associate it with the post
	err = server.store.AddPostTag(ctx, db.AddPostTagParams{
		PostID: int32(postID),
		TagID:  tag.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, tag)
}

func (server *Server) deleteTag(ctx *gin.Context) {
	// Get auth payload
	authPayload := ctx.MustGet(authorizationPayloadKey).(*util.Payload)

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	// Get user by username
	user, err := server.store.GetUserByUsername(ctx, authPayload.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get all posts that use this tag
	posts, err := server.store.GetPostsByTagID(ctx, int32(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if user owns any posts with this tag
	hasAccess := false
	for _, post := range posts {
		if post.UserID.Int32 == user.ID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "you can only delete tags used in your own posts"})
		return
	}

	// First remove the tag from all posts
	err = server.store.DeleteTagFromPosts(ctx, int32(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Then delete the tag itself
	err = server.store.DeleteTag(ctx, int32(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "tag deleted successfully"})
}

func (server *Server) removeTagFromPost(ctx *gin.Context) {
	// Get auth payload
	authPayload := ctx.MustGet(authorizationPayloadKey).(*util.Payload)

	postIDStr := ctx.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	// Verify post exists and belongs to user
	post, err := server.store.GetPost(ctx, int32(postID))
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Verify ownership
	if post.Username.String != authPayload.Username {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "you can only modify your own posts"})
		return
	}

	tagIDStr := ctx.Param("tagId")
	tagID, err := strconv.ParseInt(tagIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	err = server.store.DeletePostTag(ctx, db.DeletePostTagParams{
		PostID: int32(postID),
		TagID:  int32(tagID),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "tag removed from post successfully"})
}

func (server *Server) getPostByTagID(ctx *gin.Context) {
	// Parse tag ID from URL parameter
	tagIDStr := ctx.Param("tagId")
	tagID, err := strconv.ParseInt(tagIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	// Get posts by tag ID
	posts, err := server.store.GetPostsByTagID(ctx, int32(tagID))
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "no posts found with this tag"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert posts to response format
	response := make([]gin.H, len(posts))
	for i, post := range posts {
		response[i] = gin.H{
			"id":         post.ID,
			"user_id":    post.UserID,
			"title":      post.Title,
			"content":    post.Content,
			"type":       post.Type,
			"status":     post.Status.String,
			"created_at": post.CreatedAt,
			"updated_at": post.UpdatedAt,
			"username":   post.Username,
		}
	}

	ctx.JSON(http.StatusOK, response)
}

// type listPostsByUserParams struct {
// 	UserID int32  `form:"user_id" binding:"required,min=1"`
// 	Status string `form:"status"`
// 	Limit  int32  `form:"limit,default=10"`
// 	Offset int32  `form:"offset,default=0"`
// }

// listPostsByUser handles retrieving all posts for a specific user
func (server *Server) listPostsByUser(ctx *gin.Context) {
	// Get user ID from URL
	userIDStr := ctx.Param("userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil || userID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// Set up default values
	var limit int32 = 10            // default limit
	var offset int32 = 0            // default offset
	var status string = "published" // default status

	// Parse limit from query parameter
	if limitStr := ctx.Query("limit"); limitStr != "" {
		limitInt, err := strconv.ParseInt(limitStr, 10, 32)
		if err == nil && limitInt > 0 {
			limit = int32(limitInt)
		}
	}

	// Parse offset from query parameter
	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		offsetInt, err := strconv.ParseInt(offsetStr, 10, 32)
		if err == nil && offsetInt >= 0 {
			offset = int32(offsetInt)
		}
	}

	// Get status from query parameter
	if statusStr := ctx.Query("status"); statusStr != "" {
		status = statusStr
	}

	// Check if the user exists first
	user, err := server.store.GetUser(ctx, int32(userID))
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch posts by user ID with pagination and status filter
	arg := db.ListPostsByUserParams{
		UserID: sql.NullInt32{
			Int32: user.ID,
			Valid: true,
		},
		Column2: status,
		Limit:   limit,
		Offset:  offset,
	}

	posts, err := server.store.ListPostsByUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format the response
	response := make([]gin.H, len(posts))
	for i, post := range posts {
		response[i] = gin.H{
			"id":            post.ID,
			"title":         post.Title,
			"content":       post.Content,
			"type":          post.Type,
			"status":        post.Status.String,
			"created_at":    post.CreatedAt,
			"updated_at":    post.UpdatedAt,
			"username":      post.Username,
			"comment_count": post.CommentCount,
			"tags":          post.Tags,
			"likes":         post.Likes,
		}
	}

	ctx.JSON(http.StatusOK, response)
}

func (server *Server) listPostsByLikes(ctx *gin.Context) {
	var limit int32 = 10            // default limit
	var offset int32 = 0            // default offset
	var status string = "published" // default status

	// Parse limit from query parameter
	if limitStr := ctx.Query("limit"); limitStr != "" {
		limitInt, err := strconv.ParseInt(limitStr, 10, 32)
		if err == nil && limitInt > 0 {
			limit = int32(limitInt)
		}
	}

	// Parse offset from query parameter
	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		offsetInt, err := strconv.ParseInt(offsetStr, 10, 32)
		if err == nil && offsetInt >= 0 {
			offset = int32(offsetInt)
		}
	}

	// Get status from query parameter
	if statusStr := ctx.Query("status"); statusStr != "" {
		status = statusStr
	}

	arg := db.ListPostsOrderByLikesParams{
		Column1: status,
		Limit:   limit,
		Offset:  offset,
	}

	posts, err := server.store.ListPostsOrderByLikes(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]gin.H, len(posts))
	for i, post := range posts {
		response[i] = gin.H{
			"id":            post.ID,
			"user_id":       post.UserID,
			"title":         post.Title,
			"content":       post.Content,
			"type":          post.Type,
			"status":        post.Status.String,
			"created_at":    post.CreatedAt,
			"updated_at":    post.UpdatedAt,
			"username":      post.Username,
			"comment_count": post.CommentCount,
			"tags":          post.Tags,
			"likes":         post.Likes,
		}
	}

	ctx.JSON(http.StatusOK, response)
}

// listUsersByPostLikes handles retrieving users ordered by their total post likes
func (server *Server) listUsersByPostLikes(ctx *gin.Context) {
	var limit int32 = 10 // default limit
	var offset int32 = 0 // default offset

	// Parse limit from query parameter
	if limitStr := ctx.Query("limit"); limitStr != "" {
		limitInt, err := strconv.ParseInt(limitStr, 10, 32)
		if err == nil && limitInt > 0 {
			limit = int32(limitInt)
		}
	}

	// Parse offset from query parameter
	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		offsetInt, err := strconv.ParseInt(offsetStr, 10, 32)
		if err == nil && offsetInt >= 0 {
			offset = int32(offsetInt)
		}
	}

	arg := db.ListUsersOrderByPostLikesParams{
		Limit:  limit,
		Offset: offset,
	}

	users, err := server.store.ListUsersOrderByPostLikes(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]gin.H, len(users))
	for i, user := range users {
		response[i] = gin.H{
			"id":          user.ID,
			"username":    user.Username,
			"email":       user.Email,
			"first_name":  user.FirstName,
			"last_name":   user.LastName,
			"bio":         user.Bio,
			"created_at":  user.CreatedAt,
			"updated_at":  user.UpdatedAt,
			"total_likes": user.TotalLikes,
			"post_count":  user.PostCount,
		}
	}

	ctx.JSON(http.StatusOK, response)
}

type PostFilter struct {
	UserID    *int32  `form:"user_id"`
	Status    *string `form:"status"`
	SortBy    string  `form:"sort_by"`    // Possible values: "created_at", "likes", "title"
	SortOrder string  `form:"sort_order"` // Possible values: "asc", "desc"
	Limit     int32   `form:"limit,default=10"`
	Offset    int32   `form:"offset,default=0"`
}

func (f *PostFilter) ValidateSortParams() error {
	// Default values
	if f.SortBy == "" {
		f.SortBy = "created_at"
	}
	if f.SortOrder == "" {
		f.SortOrder = "desc"
	}

	// Validate sort_by
	validSortBy := map[string]bool{
		"created_at": true,
		"likes":      true,
		"title":      true,
	}
	if !validSortBy[f.SortBy] {
		return fmt.Errorf("invalid sort_by parameter: %s", f.SortBy)
	}

	// Validate sort_order
	f.SortOrder = strings.ToLower(f.SortOrder)
	if f.SortOrder != "asc" && f.SortOrder != "desc" {
		return fmt.Errorf("invalid sort_order parameter: %s", f.SortOrder)
	}

	// Validate limit and offset
	if f.Limit <= 0 {
		f.Limit = 10
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	return nil
}

// FilterPosts handles retrieving posts with filtering and sorting
func (server *Server) FilterPosts(ctx *gin.Context) {
	var filter PostFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := filter.ValidateSortParams(); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert PostFilter to FilterParams
	params := db.FilterParams{
		UserID:    filter.UserID,
		Status:    filter.Status,
		SortBy:    filter.SortBy,
		SortOrder: filter.SortOrder,
		Limit:     filter.Limit,
		Offset:    filter.Offset,
	}

	// Use the store interface to filter posts
	posts, err := server.store.FilterPosts(ctx, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	response := make([]gin.H, len(posts))
	for i, post := range posts {
		response[i] = gin.H{
			"id":            post.ID,
			"user_id":       post.UserID.Int32,
			"title":         post.Title,
			"content":       post.Content,
			"type":          post.Type,
			"status":        post.Status.String,
			"created_at":    post.CreatedAt,
			"updated_at":    post.UpdatedAt,
			"likes":         post.Likes,
			"username":      post.Username.String,
			"comment_count": post.CommentCount,
			"tags":          post.Tags,
		}
	}

	ctx.JSON(http.StatusOK, response)
}

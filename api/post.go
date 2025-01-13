package api

import (
	"database/sql"
	"net/http"
	"strconv"

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
		}
	}

	ctx.JSON(http.StatusOK, response)
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

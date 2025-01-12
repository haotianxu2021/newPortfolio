package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mockdb "github.com/haotianxu2021/newPortfolio/db/mock"
	db "github.com/haotianxu2021/newPortfolio/db/sqlc"
	"github.com/haotianxu2021/newPortfolio/util"
	"github.com/stretchr/testify/require"
)

func createTestToken(t *testing.T, maker token.TokenMaker, username string) string {
	token, err := maker.CreateToken(username, 24*time.Hour)
	require.NoError(t, err)
	return token
}

func addAuthHeader(request *http.Request, token string) {
	request.Header.Set("Authorization", "Bearer "+token)
}

func TestCreatePost(t *testing.T) {
	post := db.Post{
		ID:      1,
		Title:   "Test Post",
		Content: "Test Content",
		UserID: sql.NullInt32{
			Int32: 1,
			Valid: true,
		},
		Type: "blog",
		Status: sql.NullString{
			String: "draft",
			Valid:  true,
		},
	}

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"title":   post.Title,
				"content": post.Content,
				"user_id": post.UserID.Int32,
				"type":    post.Type,
				"status":  post.Status.String,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreatePostParams{
					Title:   post.Title,
					Content: post.Content,
					UserID:  post.UserID,
					Type:    post.Type,
					Status:  post.Status,
				}
				store.EXPECT().
					GetUserByUsername(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{ID: 1}, nil)
				store.EXPECT().
					CreatePost(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(post, nil)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
				// Use tokenMaker directly instead of server
				token := createTestToken(t, tokenMaker, "testuser1")
				addAuthHeader(request, token)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPost(t, recorder.Body, post)
			},
		},
		{
			name: "InvalidRequest",
			body: gin.H{
				"content": post.Content,
				"user_id": post.UserID.Int32,
				"type":    post.Type,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByUsername(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{ID: 1}, nil)
				store.EXPECT().
					CreatePost(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(post, nil)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
				// Use tokenMaker directly instead of server
				token := createTestToken(t, tokenMaker, "testuser1")
				addAuthHeader(request, token)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "UnauthorizedError",
			body: gin.H{
				"title":   post.Title,
				"content": post.Content,
				"user_id": post.UserID.Int32,
				"type":    post.Type,
				"status":  post.Status.String,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByUsername(gomock.Any(), gomock.Any()).
					CreatePost(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"title":   post.Title,
				"content": post.Content,
				"user_id": post.UserID.Int32,
				"type":    post.Type,
				"status":  post.Status.String,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByUsername(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{ID: 1}, nil)
				store.EXPECT().
					CreatePost(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(post, nil)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
				// Use tokenMaker directly instead of server
				token := createTestToken(t, tokenMaker, "testuser1")
				addAuthHeader(request, token)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// Create test config
			config := util.Config{
				TokenSymmetricKey: "12345678901234567890123456789012",
			}

			server, err := NewServer(store, config)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/posts"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			if tc.name != "UnauthorizedError" {
				// Create token for the post owner (user_id 1)
				token := createTestToken(t, server, "testuser1")
				addAuthHeader(request, token)
			}

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchPost(t *testing.T, body *bytes.Buffer, post db.Post) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotPost postResponse
	err = json.Unmarshal(data, &gotPost)
	require.NoError(t, err)

	require.Equal(t, post.ID, gotPost.ID)
	require.Equal(t, post.Title, gotPost.Title)
	require.Equal(t, post.Content, gotPost.Content)
	require.Equal(t, post.UserID, gotPost.UserID)
	require.Equal(t, post.Type, gotPost.Type)
	require.Equal(t, post.Status, gotPost.Status)
}

func TestGetPost(t *testing.T) {
	post := db.GetPostRow{
		ID:      1,
		Title:   "Test Post",
		Content: "Test Content",
		UserID: sql.NullInt32{
			Int32: 1,
			Valid: true,
		},
		Type: "blog",
		Status: sql.NullString{
			String: "published",
			Valid:  true,
		},
		Username: sql.NullString{
			String: "testuser",
			Valid:  true,
		},
		FirstName: sql.NullString{
			String: "Test",
			Valid:  true,
		},
		LastName: sql.NullString{
			String: "User",
			Valid:  true,
		},
	}

	testCases := []struct {
		name          string
		postID        int32
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			postID: post.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPost(gomock.Any(), gomock.Eq(post.ID)).
					Times(1).
					Return(post, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchGetPost(t, recorder.Body, post)
			},
		},
		{
			name:   "NotFound",
			postID: post.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPost(gomock.Any(), gomock.Eq(post.ID)).
					Times(1).
					Return(db.GetPostRow{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			postID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPost(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			postID: post.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPost(gomock.Any(), gomock.Eq(post.ID)).
					Times(1).
					Return(db.GetPostRow{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// Create test config
			config := util.Config{
				TokenSymmetricKey: "12345678901234567890123456789012",
			}

			server, err := NewServer(store, config)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v1/posts/%d", tc.postID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			if tc.name != "UnauthorizedError" {
				token := createTestToken(t, server, "test_user")
				addAuthHeader(request, token)
			}

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchGetPost(t *testing.T, body *bytes.Buffer, post db.GetPostRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotPost gin.H
	err = json.Unmarshal(data, &gotPost)
	require.NoError(t, err)

	require.Equal(t, post.ID, int32(gotPost["id"].(float64)))
	require.Equal(t, post.Title, gotPost["title"])
	require.Equal(t, post.Content, gotPost["content"])
	require.Equal(t, post.Type, gotPost["type"])
	require.Equal(t, post.Status.String, gotPost["status"])
	require.Equal(t, post.Username.String, gotPost["username"])
}

func TestListPosts(t *testing.T) {
	n := 5
	posts := make([]db.ListPostsRow, n)
	for i := 0; i < n; i++ {
		posts[i] = db.ListPostsRow{
			ID:      int32(i + 1),
			Title:   fmt.Sprintf("Title %d", i+1),
			Content: fmt.Sprintf("Content %d", i+1),
			Type:    "blog",
			Status: sql.NullString{
				String: "published",
				Valid:  true,
			},
		}
	}

	testCases := []struct {
		name          string
		query         string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: "?limit=5&offset=0",
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListPostsParams{
					Column1: "published",
					Limit:   5,
					Offset:  0,
				}
				store.EXPECT().
					ListPosts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(posts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPosts(t, recorder.Body, posts)
			},
		},
		{
			name:  "InternalError",
			query: "?limit=5&offset=0",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListPosts(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListPostsRow{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// Create test config
			config := util.Config{
				TokenSymmetricKey: "12345678901234567890123456789012",
			}

			server, err := NewServer(store, config)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()

			url := "/api/v1/posts" + tc.query
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			if tc.name != "UnauthorizedError" {
				token := createTestToken(t, server, "test_user")
				addAuthHeader(request, token)
			}

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchPosts(t *testing.T, body *bytes.Buffer, posts []db.ListPostsRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotPosts []gin.H
	err = json.Unmarshal(data, &gotPosts)
	require.NoError(t, err)
	require.Equal(t, len(posts), len(gotPosts))

	for i := range posts {
		require.Equal(t, posts[i].ID, int32(gotPosts[i]["id"].(float64)))
		require.Equal(t, posts[i].Title, gotPosts[i]["title"])
		require.Equal(t, posts[i].Content, gotPosts[i]["content"])
		require.Equal(t, posts[i].Type, gotPosts[i]["type"])
		require.Equal(t, posts[i].Status.String, gotPosts[i]["status"])
	}
}

func TestUpdatePost(t *testing.T) {
	post := db.Post{
		ID:      1,
		Title:   "Updated Title",
		Content: "Updated Content",
		Type:    "blog",
		Status: sql.NullString{
			String: "published",
			Valid:  true,
		},
	}

	testCases := []struct {
		name          string
		postID        int32
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			postID: post.ID,
			body: gin.H{
				"title":   post.Title,
				"content": post.Content,
				"type":    post.Type,
				"status":  post.Status.String,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdatePostParams{
					ID:      post.ID,
					Title:   post.Title,
					Content: post.Content,
					Type:    post.Type,
					Status:  post.Status,
				}
				store.EXPECT().
					GetPost(gomock.Any(), gomock.Eq(post.ID)).
					Times(1).
					Return(post, nil)
				store.EXPECT().
					UpdatePost(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(post, nil)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
				// Use tokenMaker directly instead of server
				token := createTestToken(t, tokenMaker, "testuser1")
				addAuthHeader(request, token)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPost(t, recorder.Body, post)
			},
		},
		{
			name:   "UnauthorizedError",
			postID: post.ID,
			body: gin.H{
				"title":   post.Title,
				"content": post.Content,
				"type":    post.Type,
				"status":  post.Status.String,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPost(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdatePost(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:   "NotFound",
			postID: post.ID,
			body: gin.H{
				"title":   post.Title,
				"content": post.Content,
				"type":    post.Type,
				"status":  post.Status.String,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPost(gomock.Any(), gomock.Eq(post.ID)).
					Times(1).
					Return(db.Post{}, sql.ErrNoRows)
				store.EXPECT().
					UpdatePost(gomock.Any(), gomock.Any()).
					Times(0)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
				// Use tokenMaker directly instead of server
				token := createTestToken(t, tokenMaker, "testuser1")
				addAuthHeader(request, token)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// Create test config
			config := util.Config{
				TokenSymmetricKey: "12345678901234567890123456789012",
			}

			server, err := NewServer(store, config)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/api/v1/posts/%d", tc.postID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			if tc.name != "UnauthorizedError" {
				token := createTestToken(t, server, "test_user")
				addAuthHeader(request, token)
			}

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeletePost(t *testing.T) {
	testCases := []struct {
		name          string
		postID        int32
		buildStubs    func(store *mockdb.MockStore)
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			postID: 1,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeletePost(gomock.Any(), gomock.Eq(int32(1))).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			postID: 1,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeletePost(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "UnauthorizedError",
			postID: 1,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeletePost(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			postID: 1,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeletePost(gomock.Any(), gomock.Eq(int32(1))).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// Create test config
			config := util.Config{
				TokenSymmetricKey: "12345678901234567890123456789012",
			}

			server, err := NewServer(store, config)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()

			var url string
			if tc.name == "InvalidID" {
				url = "/api/v1/posts/abc"
			} else {
				url = fmt.Sprintf("/api/v1/posts/%d", tc.postID)
			}
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			if tc.name != "UnauthorizedError" {
				token := createTestToken(t, server, "test_user")
				addAuthHeader(request, token)
			}

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

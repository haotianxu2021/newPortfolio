package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/haotianxu2021/newPortfolio/db/sqlc"
	"github.com/haotianxu2021/newPortfolio/util"
)

const authorizationPayloadKey = "authorization_payload"

func (server *Server) getAuthPayload(ctx *gin.Context) (*util.Payload, error) {
	payload, exists := ctx.Get(authorizationPayloadKey)
	if !exists {
		return nil, fmt.Errorf("authorization payload not found in context")
	}

	authPayload, ok := payload.(*util.Payload)
	if !ok {
		return nil, fmt.Errorf("invalid authorization payload type")
	}

	return authPayload, nil
}

// Server serves HTTP requests for our service
type Server struct {
	store      db.Store
	router     *gin.Engine
	httpServer *http.Server
	tokenMaker util.TokenMaker
	config     util.Config
}

// NewServer creates a new HTTP server and sets up routing
func NewServer(store db.Store, config util.Config) (*Server, error) {
	tokenMaker, err := util.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		store:      store,
		router:     gin.Default(),
		tokenMaker: tokenMaker,
		config:     config,
	}

	// Add CORS middleware
	server.router.Use(corsMiddleware())

	// setup routes
	server.setupRouter()

	return server, nil
}

// ServeHTTP implements http.Handler interface
func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server.router.ServeHTTP(w, r)
}

// Start runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
	server.httpServer = &http.Server{
		Addr:    address,
		Handler: server.router,
	}
	return server.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (server *Server) Shutdown(ctx context.Context) error {
	if server.httpServer != nil {
		return server.httpServer.Shutdown(ctx)
	}
	return nil
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// authMiddleware verifies the Authorization header and sets the user in context
func (server *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			return
		}

		tokenString := authHeader[len("Bearer "):]
		payload, err := server.tokenMaker.VerifyToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Store the full payload in context for handlers to access
		c.Set(authorizationPayloadKey, payload)
		// Also store username for backward compatibility
		c.Set("username", payload.Username)
		c.Next()
	}
}

// setupRouter sets up all the routes for our API
func (server *Server) setupRouter() {
	router := server.router

	// Add routes to the router
	v1 := router.Group("/api/v1")
	{
		// Public routes
		v1.POST("/users", server.createUser)
		v1.POST("/login", server.loginUser)
		v1.GET("/users/:id", server.getUser)
		v1.GET("/users", server.listUsers)
		v1.GET("/posts/:id", server.getPost)
		v1.GET("/posts", server.listPosts)

		// Protected routes
		protected := v1.Group("")
		protected.Use(server.authMiddleware())
		{
			// User routes
			protected.PUT("/users/:id", server.updateUser)
			protected.PUT("/users/:id/password", server.updateUserPassword)

			// Post routes
			protected.POST("/posts", server.createPost)
			protected.PUT("/posts/:id", server.updatePost)
			protected.DELETE("/posts/:id", server.deletePost)

			// Post images routes
			protected.POST("/posts/:id/images", server.addImage)
			protected.DELETE("/images/:id", server.deleteImage)

			// Post tags routes
			protected.POST("/posts/:id/tags", server.addTag)
			protected.DELETE("/tags/:id", server.deleteTag)
			protected.DELETE("/posts/:id/tags/:tagId", server.removeTagFromPost)
		}
	}
}

package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/haotianxu2021/newPortfolio/db/sqlc"
)

// Server serves HTTP requests for our service
type Server struct {
	store      db.Store
	router     *gin.Engine
	httpServer *http.Server
}

// NewServer creates a new HTTP server and sets up routing
func NewServer(store db.Store) *Server {
	server := &Server{
		store:  store,
		router: gin.Default(),
	}

	// Add CORS middleware
	server.router.Use(corsMiddleware())

	// setup routes
	server.setupRouter()

	return server
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

// setupRouter sets up all the routes for our API
func (server *Server) setupRouter() {
	router := server.router

	// Add routes to the router
	v1 := router.Group("/api/v1")
	{
		// User routes
		v1.POST("/users", server.createUser)
		v1.GET("/users/:id", server.getUser)
		v1.GET("/users", server.listUsers)
		v1.PUT("/users/:id", server.updateUser)
		v1.PUT("/users/:id/password", server.updateUserPassword)

		// Post routes
		v1.POST("/posts", server.createPost)
		v1.GET("/posts/:id", server.getPost)
		v1.GET("/posts", server.listPosts)
		v1.PUT("/posts/:id", server.updatePost)
		v1.DELETE("/posts/:id", server.deletePost)

		// Post images routes
		v1.POST("/posts/:id/images", server.addImage)
		v1.DELETE("/images/:id", server.deleteImage)

		// Post tags routes
		v1.POST("/posts/:id/tags", server.addTag)
		v1.DELETE("/tags/:id", server.deleteTag)
		v1.DELETE("/posts/:id/tags/:tagId", server.removeTagFromPost)
	}
}

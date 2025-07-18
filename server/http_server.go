package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"twt/config"
	"twt/models"
	"twt/services"

	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	db            *models.DB
	githubService *services.GitHubService
	router        *gin.Engine
}

func NewHTTPServer(db *models.DB, githubService *services.GitHubService) *HTTPServer {
	router := gin.Default()
	server := &HTTPServer{
		db:            db,
		githubService: githubService,
		router:        router,
	}

	server.setupRoutes()
	return server
}

func (s *HTTPServer) setupRoutes() {
	api := s.router.Group("/api/v1")
	{
		api.GET("/repositories", s.getRepositories)
		api.GET("/repositories/:owner/:name", s.getRepository)
		api.POST("/sync", s.syncRepositories)
		api.GET("/health", s.healthCheck)
	}

	// Serve static files for a simple web UI
	s.router.Static("/static", "./web/static")
	s.router.LoadHTMLGlob("web/templates/*")
	s.router.GET("/", s.indexPage)
}

func (s *HTTPServer) getRepositories(c *gin.Context) {
	repos, err := s.db.GetRepositories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get repositories",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"repositories": repos,
		"total":        len(repos),
	})
}

func (s *HTTPServer) getRepository(c *gin.Context) {
	owner := c.Param("owner")
	name := c.Param("name")
	fullName := fmt.Sprintf("%s/%s", owner, name)

	repo, err := s.db.GetRepositoryByName(fullName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Repository not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"repository": repo,
	})
}

type SyncRequest struct {
	RepositoryURLs []string `json:"repository_urls"`
}

func (s *HTTPServer) syncRepositories(c *gin.Context) {
	var req SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body provided, use config repositories
		cfg := config.GetConfig()
		req.RepositoryURLs = cfg.Github.Repositories
	}

	if len(req.RepositoryURLs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No repository URLs provided",
		})
		return
	}

	syncedCount, err := s.githubService.SyncRepositories(req.RepositoryURLs, s.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to sync repositories",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      fmt.Sprintf("Successfully synced %d repositories", syncedCount),
		"synced_count": syncedCount,
	})
}

func (s *HTTPServer) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "twt-repository-service",
	})
}

func (s *HTTPServer) indexPage(c *gin.Context) {
	repos, err := s.db.GetRepositories()
	if err != nil {
		log.Printf("Failed to get repositories for index page: %v", err)
		repos = []*models.Repository{}
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":        "GitHub Repository Dashboard",
		"repositories": repos,
	})
}

func StartHTTPServer(db *models.DB, githubService *services.GitHubService) error {
	cfg := config.GetConfig()
	port := cfg.Server.HTTPPort

	server := NewHTTPServer(db, githubService)

	log.Printf("HTTP server starting on port %d", port)
	return server.router.Run(":" + strconv.Itoa(port))
}

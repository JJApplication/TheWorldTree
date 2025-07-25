package server

import (
	"fmt"
	"log"
	"net/http"

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
	gin.SetMode(gin.ReleaseMode)
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
		api.GET("/commits/:owner/:name", s.getCommits)
		api.POST("/repositories/sync", s.syncRepositories)
		api.POST("/commits/sync/:owner/:name", s.syncCommits)
		api.POST("/commits/sync", s.syncCommitsAll)
		api.GET("/health", s.healthCheck)
	}
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

func (s *HTTPServer) getCommits(c *gin.Context) {
	owner := c.Param("owner")
	name := c.Param("name")
	fullName := fmt.Sprintf("%s/%s", owner, name)
	commits, err := s.db.GetCommits(fullName, 50, 0)
	count, err := s.db.GetCommitCount(fullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get repositories",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"commits": commits,
		"total":   count,
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

func (s *HTTPServer) syncCommits(c *gin.Context) {
	owner := c.Param("owner")
	name := c.Param("name")
	fullName := fmt.Sprintf("%s/%s", owner, name)

	if fullName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No repository URLs provided",
		})
		return
	}

	syncedCount, err := s.githubService.SyncCommits(fullName, 50, s.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to sync commits",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      fmt.Sprintf("Successfully synced %d commits", syncedCount),
		"synced_count": syncedCount,
	})
}

func (s *HTTPServer) syncCommitsAll(c *gin.Context) {
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

	syncedCount, err := s.githubService.SyncCommitsAll(req.RepositoryURLs, 50, s.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to sync commits",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      fmt.Sprintf("Successfully synced %d commits", syncedCount),
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
	host := cfg.Server.HTTP.Host
	port := cfg.Server.HTTP.Port

	server := NewHTTPServer(db, githubService)

	log.Printf("HTTP server starting on %s:%d", host, port)
	return server.router.Run(fmt.Sprintf("%s:%d", host, port))
}

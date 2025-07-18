package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"twt/models"
)

type GitHubService struct {
	token  string
	client *http.Client
}

type GitHubRepo struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	FullName    string    `json:"full_name"`
	Description *string   `json:"description"`
	HTMLURL     string    `json:"html_url"`
	Language    *string   `json:"language"`
	Stars       int       `json:"stargazers_count"`
	Forks       int       `json:"forks_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GitHubCommit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
}

func NewGitHubService(token string) *GitHubService {
	return &GitHubService{
		token:  token,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (g *GitHubService) GetRepositoryInfo(repoURL string) (*models.Repository, error) {
	// Extract owner and repo name from URL
	// e.g., https://github.com/gin-gonic/gin -> gin-gonic/gin
	parts := strings.Split(strings.TrimPrefix(repoURL, "https://github.com/"), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid repository URL: %s", repoURL)
	}

	owner := parts[0]
	repoName := parts[1]
	fullName := fmt.Sprintf("%s/%s", owner, repoName)

	// GitHub API URL
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s", fullName)

	// Create request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authorization header if token is provided
	if g.token != "" && g.token != "your_github_token_here" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", g.token))
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Make request
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var githubRepo GitHubRepo
	if err := json.NewDecoder(resp.Body).Decode(&githubRepo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to our model
	repo := &models.Repository{
		Name:      githubRepo.Name,
		FullName:  githubRepo.FullName,
		URL:       githubRepo.HTMLURL,
		Stars:     githubRepo.Stars,
		Forks:     githubRepo.Forks,
		CreatedAt: githubRepo.CreatedAt,
		UpdatedAt: githubRepo.UpdatedAt,
	}

	if githubRepo.Description != nil {
		repo.Description = *githubRepo.Description
	}

	if githubRepo.Language != nil {
		repo.Language = *githubRepo.Language
	}

	return repo, nil
}

func (g *GitHubService) GetCommits(repositoryFullName string, limit int) ([]*models.Commit, error) {
	if limit <= 0 {
		limit = 50 // default limit
	}

	// GitHub API URL for commits
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/commits?per_page=%d", repositoryFullName, limit)

	// Create request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authorization header if token is provided
	if g.token != "" && g.token != "your_github_token_here" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", g.token))
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Make request
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var githubCommits []GitHubCommit
	if err := json.NewDecoder(resp.Body).Decode(&githubCommits); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to our model
	var commits []*models.Commit
	for _, gc := range githubCommits {
		commit := &models.Commit{
			SHA:                gc.SHA,
			Message:            gc.Commit.Message,
			AuthorName:         gc.Commit.Author.Name,
			AuthorEmail:        gc.Commit.Author.Email,
			CommitDate:         gc.Commit.Author.Date,
			RepositoryFullName: repositoryFullName,
		}
		commits = append(commits, commit)
	}

	return commits, nil
}

func (g *GitHubService) SyncCommits(repositoryFullName string, limit int, db *models.DB) (int, error) {
	commits, err := g.GetCommits(repositoryFullName, limit)
	if err != nil {
		return 0, fmt.Errorf("failed to get commits: %w", err)
	}

	syncedCount := 0
	for _, commit := range commits {
		if err := db.SaveCommit(commit); err != nil {
			fmt.Printf("Failed to save commit %s: %v\n", commit.SHA, err)
			continue
		}
		syncedCount++
	}

	fmt.Printf("Successfully synced %d commits for repository: %s\n", syncedCount, repositoryFullName)
	return syncedCount, nil
}

func (g *GitHubService) SyncRepositories(repoURLs []string, db *models.DB) (int, error) {
	syncedCount := 0

	for _, repoURL := range repoURLs {
		repo, err := g.GetRepositoryInfo(repoURL)
		if err != nil {
			fmt.Printf("Failed to sync repository %s: %v\n", repoURL, err)
			continue
		}

		if err := db.SaveRepository(repo); err != nil {
			fmt.Printf("Failed to save repository %s: %v\n", repo.FullName, err)
			continue
		}

		fmt.Printf("Successfully synced repository: %s\n", repo.FullName)
		syncedCount++
	}

	return syncedCount, nil
}

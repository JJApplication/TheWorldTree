package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
	"twt/config"
	"twt/models"
	"twt/proto"
	"twt/services"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GRPCServer struct {
	proto.UnimplementedRepositoryServiceServer
	db            *models.DB
	githubService *services.GitHubService
}

func NewGRPCServer(db *models.DB, githubService *services.GitHubService) *GRPCServer {
	return &GRPCServer{
		db:            db,
		githubService: githubService,
	}
}

func (s *GRPCServer) GetRepositories(ctx context.Context, req *proto.GetRepositoriesRequest) (*proto.GetRepositoriesResponse, error) {
	repos, err := s.db.GetRepositories()
	if err != nil {
		return nil, fmt.Errorf("failed to get repositories: %w", err)
	}

	var protoRepos []*proto.Repository
	for _, repo := range repos {
		protoRepo := &proto.Repository{
			Id:          int32(repo.ID),
			Name:        repo.Name,
			FullName:    repo.FullName,
			Description: repo.Description,
			Url:         repo.URL,
			Language:    repo.Language,
			Stars:       int32(repo.Stars),
			Forks:       int32(repo.Forks),
			CreatedAt:   timestamppb.New(repo.CreatedAt),
			UpdatedAt:   timestamppb.New(repo.UpdatedAt),
			SyncedAt:    timestamppb.New(repo.SyncedAt),
		}
		protoRepos = append(protoRepos, protoRepo)
	}

	return &proto.GetRepositoriesResponse{
		Repositories: protoRepos,
		Total:        int32(len(protoRepos)),
	}, nil
}

func (s *GRPCServer) GetRepository(ctx context.Context, req *proto.GetRepositoryRequest) (*proto.GetRepositoryResponse, error) {
	repo, err := s.db.GetRepositoryByName(req.FullName)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	protoRepo := &proto.Repository{
		Id:          int32(repo.ID),
		Name:        repo.Name,
		FullName:    repo.FullName,
		Description: repo.Description,
		Url:         repo.URL,
		Language:    repo.Language,
		Stars:       int32(repo.Stars),
		Forks:       int32(repo.Forks),
		CreatedAt:   timestamppb.New(repo.CreatedAt),
		UpdatedAt:   timestamppb.New(repo.UpdatedAt),
		SyncedAt:    timestamppb.New(repo.SyncedAt),
	}

	return &proto.GetRepositoryResponse{
		Repository: protoRepo,
	}, nil
}

func (s *GRPCServer) SyncRepositories(ctx context.Context, req *proto.SyncRepositoriesRequest) (*proto.SyncRepositoriesResponse, error) {
	repoURLs := req.RepositoryUrls
	if len(repoURLs) == 0 {
		// Use URLs from config if none provided
		cfg := config.GetConfig()
		repoURLs = cfg.Github.Repositories
	}

	syncedCount, err := s.githubService.SyncRepositories(repoURLs, s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to sync repositories: %w", err)
	}

	return &proto.SyncRepositoriesResponse{
		Message:     fmt.Sprintf("Successfully synced %d repositories", syncedCount),
		SyncedCount: int32(syncedCount),
	}, nil
}

func StartGRPCServer(db *models.DB, githubService *services.GitHubService) error {
	cfg := config.GetConfig()
	port := cfg.Server.GRPCPort

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	s := grpc.NewServer()
	grpcServer := NewGRPCServer(db, githubService)
	proto.RegisterRepositoryServiceServer(s, grpcServer)

	log.Printf("gRPC server starting on port %d", port)
	return s.Serve(lis)
}

func initializeDatabase() (*models.DB, error) {
	cfg := config.GetConfig()
	dbPath := cfg.Database.Path

	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := models.NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return db, nil
}

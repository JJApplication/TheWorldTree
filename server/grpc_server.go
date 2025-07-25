package server

import (
	"context"
	"fmt"
	"log"
	"net"

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

func (s *GRPCServer) GetCommits(ctx context.Context, req *proto.GetCommitsRequest) (*proto.GetCommitsResponse, error) {
	commits, err := s.db.GetCommits(req.RepositoryFullName, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	var protoCommits []*proto.Commit
	for _, commit := range commits {
		protoCommit := &proto.Commit{
			Id:          int32(commit.ID),
			Message:     commit.Message,
			Sha:         commit.SHA,
			AuthorName:  commit.AuthorName,
			AuthorEmail: commit.AuthorEmail,
			CommitDate:  timestamppb.New(commit.CommitDate),
			SyncedAt:    timestamppb.New(commit.SyncedAt),
		}
		protoCommits = append(protoCommits, protoCommit)
	}
	return &proto.GetCommitsResponse{
		Commits: protoCommits,
		Total:   int32(len(protoCommits)),
	}, nil
}

func (s *GRPCServer) SyncCommits(ctx context.Context, req *proto.SyncCommitsRequest) (*proto.SyncCommitsResponse, error) {
	syncedCount, err := s.githubService.SyncCommits(req.RepositoryFullName, int(req.Limit), s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to sync commits: %w", err)
	}
	return &proto.SyncCommitsResponse{
		Message:     fmt.Sprintf("Successfully synced %d commits", syncedCount),
		SyncedCount: int32(syncedCount),
	}, nil
}

func (s *GRPCServer) SyncCommitsAll(ctx context.Context, req *proto.SyncCommitsAllRequest) (*proto.SyncCommitsResponse, error) {
	repoURLs := req.RepositoryUrls
	if len(repoURLs) == 0 {
		// Use URLs from config if none provided
		cfg := config.GetConfig()
		repoURLs = cfg.Github.Repositories
	}
	syncedCount, err := s.githubService.SyncCommitsAll(repoURLs, int(req.Limit), s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to sync commits: %w", err)
	}
	return &proto.SyncCommitsResponse{
		Message:     fmt.Sprintf("Successfully synced %d repositories", syncedCount),
		SyncedCount: int32(syncedCount),
	}, nil
}

func StartGRPCServer(db *models.DB, githubService *services.GitHubService) error {
	cfg := config.GetConfig()
	address := cfg.Server.GRPC.Address

	addr, err := net.ResolveUnixAddr("unix", address)
	if err != nil {
		return fmt.Errorf("resolve unix addr: %v", err)
	}
	udsLis, err := net.ListenUnix("unix", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on address %s: %v", address, err)
	}

	s := grpc.NewServer()
	grpcServer := NewGRPCServer(db, githubService)
	proto.RegisterRepositoryServiceServer(s, grpcServer)

	log.Printf("gRPC server starting on: %s", address)
	return s.Serve(udsLis)
}

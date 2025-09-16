package services

import (
	"context"
	"fmt"
	"go-starter/internal/modules/auth/dto"
	"go-starter/internal/modules/auth/models"
	"go-starter/internal/modules/auth/repositories"
	"go-starter/pkg/config"

	"github.com/google/uuid"
)

type AuthService struct {
	repo       *repositories.UserRepository
	jwtService *JWTService
	config     *config.Config
}

func NewAuthService(repo *repositories.UserRepository, jwtService *JWTService, config *config.Config) *AuthService {
	return &AuthService{
		repo:       repo,
		jwtService: jwtService,
		config:     config,
	}
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	exists, err := s.repo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Check user limit if MAX_AVAILABLE_USER is set (> 0)
	if s.config.MaxAvailableUser > 0 {
		userCount, err := s.repo.GetUserCount(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get user count: %w", err)
		}
		if userCount >= int64(s.config.MaxAvailableUser) {
			return nil, fmt.Errorf("maximum number of users reached (%d)", s.config.MaxAvailableUser)
		}
	}

	user := &models.User{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if err := user.HashPassword(); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := s.jwtService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AuthResponse{
		Success: true,
		Token:   token,
		User:    *user,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.CheckPassword(req.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	token, err := s.jwtService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AuthResponse{
		Success: true,
		Token:   token,
		User:    *user,
	}, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

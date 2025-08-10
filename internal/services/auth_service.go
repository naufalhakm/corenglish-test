package services

import (
	"go-corenglish/internal/commons/response"
	"go-corenglish/internal/config"
	"go-corenglish/internal/models"
	"go-corenglish/internal/params"
	"go-corenglish/internal/repositories"
	"go-corenglish/pkg/token"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(req *params.RegisterRequest) (*params.AuthResponse, *response.CustomError)
	Login(req *params.LoginRequest) (*params.AuthResponse, *response.CustomError)
}

type authService struct {
	userRepo   repositories.UserRepository
	config     *config.Config
	logger     *logrus.Logger
	jwtManager *token.TokenManager
}

func NewAuthService(userRepo repositories.UserRepository, config *config.Config, logger *logrus.Logger, jwtManager *token.TokenManager) AuthService {
	return &authService{
		userRepo:   userRepo,
		config:     config,
		logger:     logger,
		jwtManager: jwtManager,
	}
}

func (s *authService) Register(req *params.RegisterRequest) (*params.AuthResponse, *response.CustomError) {
	// Check if user already exists by email
	if _, err := s.userRepo.GetByEmail(req.Email); err == nil {
		s.logger.WithField("email", req.Email).Warn("Registration attempt with existing email")
		return nil, response.BadRequestError("user with this email already exists")
	}

	// Check if username is taken
	if _, err := s.userRepo.GetByUsername(req.Username); err == nil {
		s.logger.WithField("username", req.Username).Warn("Registration attempt with existing username")
		return nil, response.BadRequestError("username is already taken")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.config.BcryptCost)
	if err != nil {
		s.logger.WithError(err).Error("Failed to hash password")
		return nil, response.GeneralError("failed to hash password")
	}

	// Create user
	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(user); err != nil {
		s.logger.WithError(err).WithField("email", req.Email).Error("Failed to create user")
		return nil, response.RepositoryError("failed to create user")
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", user.ID).Error("Failed to generate token")
		return nil, response.GeneralError("failed to generate token")
	}

	response := &params.AuthResponse{
		Token: token,
	}
	response.User.ID = user.ID
	response.User.Username = user.Username
	response.User.Email = user.Email

	s.logger.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	}).Info("User registered successfully")

	return response, nil
}

func (s *authService) Login(req *params.LoginRequest) (*params.AuthResponse, *response.CustomError) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		s.logger.WithField("email", req.Email).Warn("Login attempt with non-existing email")
		return nil, response.BadRequestError("invalid email or password")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"email":   req.Email,
		}).Warn("Login attempt with invalid password")
		return nil, response.BadRequestError("invalid email or password")
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", user.ID).Error("Failed to generate token")
		return nil, response.GeneralError("failed to generate token")
	}

	response := &params.AuthResponse{
		Token: token,
	}
	response.User.ID = user.ID
	response.User.Username = user.Username
	response.User.Email = user.Email

	s.logger.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	}).Info("User logged in successfully")

	return response, nil
}

package repositories

import (
	"fmt"
	"go-corenglish/internal/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByEmail(email string) (*models.User, error)
	GetByID(id uuid.UUID) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
}

type userRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewUserRepository(db *gorm.DB, logger *logrus.Logger) UserRepository {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

func (r *userRepository) Create(user *models.User) error {
	if err := r.db.Create(user).Error; err != nil {
		r.logger.WithError(err).WithField("email", user.Email).Error("Failed to create user")
		return fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	}).Info("User created successfully")

	return nil
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.WithError(err).WithField("email", email).Error("Failed to get user by email")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.WithError(err).WithField("user_id", id).Error("Failed to get user by ID")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.WithError(err).WithField("username", username).Error("Failed to get user by username")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

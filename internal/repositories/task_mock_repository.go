package repositories

import (
	"go-corenglish/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockBookRepository struct {
	mock.Mock
}

func (m *MockBookRepository) Create(task *models.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockBookRepository) GetByID(id uuid.UUID, userID uuid.UUID) (*models.Task, error) {
	args := m.Called(id, userID)
	if args.Get(0) != nil || args.Get(1) != nil {
		return args.Get(0).(*models.Task), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockBookRepository) GetAll(userID uuid.UUID, status string, page, limit int) ([]models.Task, int64, error) {
	args := m.Called(userID, status, page, limit)
	if args.Get(0) != nil {
		return args.Get(0).([]models.Task), args.Get(1).(int64), args.Error(2)
	}
	return nil, 0, args.Error(2)
}

func (m *MockBookRepository) Update(task *models.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockBookRepository) Delete(id uuid.UUID, userID uuid.UUID) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

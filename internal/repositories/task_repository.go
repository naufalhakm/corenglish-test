package repositories

import (
	"fmt"
	"go-corenglish/internal/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(task *models.Task) error
	GetByID(id uuid.UUID, userID uuid.UUID) (*models.Task, error)
	GetAll(userID uuid.UUID, status string, page, limit int) ([]models.Task, int64, error)
	Update(task *models.Task) error
	Delete(id uuid.UUID, userID uuid.UUID) error
}

type taskRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewTaskRepository(db *gorm.DB, logger *logrus.Logger) TaskRepository {
	return &taskRepository{
		db:     db,
		logger: logger,
	}
}

func (r *taskRepository) Create(task *models.Task) error {
	if err := r.db.Create(task).Error; err != nil {
		r.logger.WithError(err).Error("Failed to create task")
		return fmt.Errorf("failed to create task: %w", err)
	}

	r.logger.WithField("task_id", task.ID).Info("Task created successfully")
	return nil
}

func (r *taskRepository) GetByID(id uuid.UUID, userID uuid.UUID) (*models.Task, error) {
	var task models.Task
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&task).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WithField("task_id", id).Warn("Task not found")
			return nil, fmt.Errorf("task not found")
		}
		r.logger.WithError(err).WithField("task_id", id).Error("Failed to get task")
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &task, nil
}

func (r *taskRepository) GetAll(userID uuid.UUID, status string, page, limit int) ([]models.Task, int64, error) {
	var tasks []models.Task
	var total int64

	offset := (page - 1) * limit

	query := r.db.Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Model(&models.Task{}).Count(&total).Error; err != nil {
		r.logger.WithError(err).Error("Failed to count tasks")
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&tasks).Error; err != nil {
		r.logger.WithError(err).Error("Failed to get tasks")
		return nil, 0, fmt.Errorf("failed to get tasks: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"status":  status,
		"page":    page,
		"limit":   limit,
		"total":   total,
		"count":   len(tasks),
	}).Info("Tasks retrieved successfully")

	return tasks, total, nil
}

func (r *taskRepository) Update(task *models.Task) error {
	result := r.db.Model(task).Where("id = ? AND user_id = ?", task.ID, task.UserID).Updates(task)
	if result.Error != nil {
		r.logger.WithError(result.Error).WithField("task_id", task.ID).Error("Failed to update task")
		return fmt.Errorf("failed to update task: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.WithField("task_id", task.ID).Warn("Task not found for update")
		return fmt.Errorf("task not found")
	}

	r.logger.WithField("task_id", task.ID).Info("Task updated successfully")
	return nil
}

func (r *taskRepository) Delete(id uuid.UUID, userID uuid.UUID) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Task{})
	if result.Error != nil {
		r.logger.WithError(result.Error).WithField("task_id", id).Error("Failed to delete task")
		return fmt.Errorf("failed to delete task: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.WithField("task_id", id).Warn("Task not found for deletion")
		return fmt.Errorf("task not found")
	}

	r.logger.WithField("task_id", id).Info("Task deleted successfully")
	return nil
}
